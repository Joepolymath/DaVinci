package chats

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	defaultHost    = "http://localhost:11434"
	defaultModel   = "llama3:8b"
	defaultTimeout = 5 * time.Minute
	chatEndpoint   = "/api/chat"
)

type Client struct {
	host       string
	model      string
	httpClient *http.Client
	logger     *zap.Logger
	enabled    bool
}

func NewClient(cfg *Config, logger *zap.Logger) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	host := strings.TrimRight(cfg.Host, "/")
	if host == "" {
		host = defaultHost
	}

	model := cfg.Model
	if model == "" {
		model = defaultModel
	}

	client := &Client{
		host:  host,
		model: model,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		logger:  logger,
		enabled: true,
	}

	logger.Info("Local LLM chat client initialized",
		zap.String("host", host),
		zap.String("model", model))

	return client, nil
}

func (c *Client) Completion(ctx context.Context, messages []Message, opts *Options) (*CompletionResponse, error) {
	if !c.enabled {
		return nil, errors.New("local LLM client is not enabled")
	}
	if len(messages) == 0 {
		return nil, errors.New("at least one message is required")
	}

	reqBody := CompletionRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
		Options:  opts,
	}

	c.logger.Debug("Sending completion request",
		zap.String("model", c.model),
		zap.Int("message_count", len(messages)))

	body, err := c.doRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	raw, err := io.ReadAll(body)
	if err != nil {
		c.logger.Error("Failed to read response body", zap.Error(err))
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var resp CompletionResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		c.logger.Error("Failed to unmarshal completion response", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal completion response: %w", err)
	}

	c.logger.Debug("Completion response received",
		zap.String("model", resp.Model),
		zap.Int("eval_count", resp.EvalCount))

	return &resp, nil
}

// CompletionStream sends a chat completion request and streams the response.
// Each chunk is delivered to the provided callback function.
// The callback receives the chunk and returns an error to stop streaming early.
// The final chunk (Done=true) includes usage statistics.
func (c *Client) CompletionStream(ctx context.Context, messages []Message, opts *Options, onChunk func(chunk StreamChunk) error) error {
	if !c.enabled {
		return errors.New("local LLM client is not enabled")
	}
	if len(messages) == 0 {
		return errors.New("at least one message is required")
	}
	if onChunk == nil {
		return errors.New("onChunk callback is required")
	}

	reqBody := CompletionRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
		Options:  opts,
	}

	c.logger.Debug("Sending streaming completion request",
		zap.String("model", c.model),
		zap.Int("message_count", len(messages)))

	body, err := c.doRequest(ctx, reqBody)
	if err != nil {
		return err
	}
	defer body.Close()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			c.logger.Error("Failed to unmarshal stream chunk",
				zap.Error(err),
				zap.String("raw", line))
			return fmt.Errorf("failed to unmarshal stream chunk: %w", err)
		}

		if err := onChunk(chunk); err != nil {
			c.logger.Debug("Streaming stopped by callback", zap.Error(err))
			return err
		}

		if chunk.Done {
			c.logger.Debug("Stream completed",
				zap.String("model", chunk.Model),
				zap.Int("eval_count", chunk.EvalCount))
			break
		}
	}

	if err := scanner.Err(); err != nil {
		c.logger.Error("Error reading stream", zap.Error(err))
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

// Health checks if the local LLM is reachable.
func (c *Client) Health(ctx context.Context) error {
	if !c.enabled {
		return errors.New("local LLM client is not enabled")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.host, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("local LLM health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("local LLM health check returned status %d", resp.StatusCode)
	}

	c.logger.Info("Local LLM health check passed", zap.String("host", c.host))
	return nil
}

// IsEnabled returns whether the client is enabled.
func (c *Client) IsEnabled() bool {
	return c.enabled
}

// GetModel returns the configured model name.
func (c *Client) GetModel() string {
	return c.model
}

// doRequest marshals the request body and sends the HTTP POST to the chat endpoint.
// Returns the response body (caller must close it).
func (c *Client) doRequest(ctx context.Context, reqBody CompletionRequest) (io.ReadCloser, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Error("Failed to marshal request", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.host + chatEndpoint
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.Error("Failed to create HTTP request", zap.Error(err))
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// For streaming requests, use a client without a timeout
	// so the connection stays open for the duration of generation.
	httpClient := c.httpClient
	if reqBody.Stream {
		httpClient = &http.Client{} // no timeout for streaming
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.logger.Error("Failed to send HTTP request", zap.Error(err))
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("LLM API error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)))
		return nil, fmt.Errorf("LLM API error (status %d): %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}
