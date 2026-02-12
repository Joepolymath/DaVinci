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
	defaultModel   = "gpt-4o-mini"
	defaultTimeout = 2 * time.Minute
	chatAPIURL     = "https://api.openai.com/v1/chat/completions"
)

type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
	logger     *zap.Logger
	enabled    bool
}

func NewClient(cfg *Config, logger *zap.Logger) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if !cfg.IsValid() {
		return nil, errors.New("invalid OpenAI configuration: API key is required")
	}

	model := cfg.Model
	if model == "" {
		model = defaultModel
	}

	client := &Client{
		apiKey: cfg.APIKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		logger:  logger,
		enabled: true,
	}

	logger.Info("OpenAI chat client initialized",
		zap.String("model", model))

	return client, nil
}

func (c *Client) Completion(ctx context.Context, messages []Message, opts *Options) (*CompletionResponse, error) {
	if !c.enabled {
		return nil, errors.New("OpenAI chat client is not enabled")
	}
	if len(messages) == 0 {
		return nil, errors.New("at least one message is required")
	}

	reqBody := c.buildRequest(messages, false, opts)

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
		zap.Int("prompt_tokens", resp.Usage.PromptTokens),
		zap.Int("completion_tokens", resp.Usage.CompletionTokens),
		zap.Int("total_tokens", resp.Usage.TotalTokens))

	return &resp, nil
}

// CompletionStream sends a streaming chat completion request.
// Each chunk is delivered to the provided callback function.
// The callback receives the chunk and can return an error to stop streaming early.
func (c *Client) CompletionStream(ctx context.Context, messages []Message, opts *Options, onChunk func(chunk StreamChunk) error) error {
	if !c.enabled {
		return errors.New("OpenAI chat client is not enabled")
	}
	if len(messages) == 0 {
		return errors.New("at least one message is required")
	}
	if onChunk == nil {
		return errors.New("onChunk callback is required")
	}

	reqBody := c.buildRequest(messages, true, opts)

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

		// OpenAI streaming uses SSE format: "data: {...}" or "data: [DONE]"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			c.logger.Debug("Stream completed")
			break
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			c.logger.Error("Failed to unmarshal stream chunk",
				zap.Error(err),
				zap.String("raw", data))
			return fmt.Errorf("failed to unmarshal stream chunk: %w", err)
		}

		if err := onChunk(chunk); err != nil {
			c.logger.Debug("Streaming stopped by callback", zap.Error(err))
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		c.logger.Error("Error reading stream", zap.Error(err))
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

// Health checks if the OpenAI API is reachable by listing models.
func (c *Client) Health(ctx context.Context) error {
	if !c.enabled {
		return errors.New("OpenAI chat client is not enabled")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.openai.com/v1/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("OpenAI health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OpenAI health check returned status %d", resp.StatusCode)
	}

	c.logger.Info("OpenAI health check passed")
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

// buildRequest constructs the CompletionRequest, flattening Options into the top-level fields.
func (c *Client) buildRequest(messages []Message, stream bool, opts *Options) CompletionRequest {
	req := CompletionRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   stream,
	}

	if opts != nil {
		if opts.Temperature != 0 {
			req.Temperature = &opts.Temperature
		}
		if opts.TopP != 0 {
			req.TopP = &opts.TopP
		}
		if opts.MaxTokens != 0 {
			req.MaxTokens = &opts.MaxTokens
		}
		if len(opts.Stop) > 0 {
			req.Stop = opts.Stop
		}
	}

	return req
}

// doRequest marshals the request body and sends the HTTP POST to the OpenAI API.
// Returns the response body (caller must close it).
func (c *Client) doRequest(ctx context.Context, reqBody CompletionRequest) (io.ReadCloser, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Error("Failed to marshal request", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, chatAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.Error("Failed to create HTTP request", zap.Error(err))
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

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

		var apiErr APIError
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr == nil && apiErr.Error.Message != "" {
			c.logger.Error("OpenAI API error",
				zap.Int("status", resp.StatusCode),
				zap.String("type", apiErr.Error.Type),
				zap.String("message", apiErr.Error.Message))
			return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, apiErr.Error.Message)
		}

		c.logger.Error("OpenAI API error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)))
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}
