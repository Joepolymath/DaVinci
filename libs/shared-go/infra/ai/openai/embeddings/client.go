package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const embeddingsAPIURL = "https://api.openai.com/v1/embeddings"

type Config struct {
	APIKey string
	Model  string
}

func (c *Config) IsValid() bool {
	return c.APIKey != ""
}

type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
	logger     *zap.Logger
	enabled    bool
}

type EmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

type APIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func NewClient(config *Config, logger *zap.Logger) (*Client, error) {
	if !config.IsValid() {
		logger.Error("Invalid OpenAI configuration")
		return nil, errors.New("invalid OpenAI configuration")
	}

	model := config.Model
	if model == "" {
		model = "text-embedding-3-small"
	}

	client := &Client{
		apiKey: config.APIKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger:  logger,
		enabled: true,
	}

	logger.Info("OpenAI embeddings client initialized successfully",
		zap.String("model", model))

	return client, nil
}

func (c *Client) CreateEmbeddings(ctx context.Context, input []string) ([]float32, error) {
	if len(input) == 0 {
		c.logger.Error("Input cannot be empty")
		return nil, errors.New("input cannot be empty")
	}

	if !c.enabled {
		c.logger.Error("Embedding provider is not enabled")
		return nil, errors.New("embedding provider is not enabled")
	}

	c.logger.Debug("Creating embedding",
		zap.Int("input_length", len(input)))

	request := EmbeddingRequest{
		Input: input,
		Model: c.model,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		c.logger.Error("Failed to marshal embedding request",
			zap.Error(err))
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, embeddingsAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.Error("Failed to create HTTP request",
			zap.Error(err))
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		c.logger.Error("Failed to send HTTP request",
			zap.Error(err))
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		apiError := APIError{}
		if err := json.NewDecoder(response.Body).Decode(&apiError); err != nil {
			c.logger.Error("Failed to decode API error response",
				zap.Error(err))
			return nil, fmt.Errorf("failed to decode API error response: %w", err)
		}
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.logger.Error("Failed to read response body",
			zap.Error(err))
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var embeddingResponse EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResponse); err != nil {
		c.logger.Error("Failed to unmarshal embedding response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal embedding response: %w", err)
	}

	if len(embeddingResponse.Data) == 0 {
		c.logger.Error("No embedding data returned")
		return nil, fmt.Errorf("no embedding data returned from open ai api")
	}

	embedding := embeddingResponse.Data[0].Embedding
	if len(embedding) == 0 {
		c.logger.Error("No embedding returned")
		return nil, fmt.Errorf("no embedding returned from open ai api")
	}

	return embedding, nil
}

func (c *Client) CreateEmbedding(ctx context.Context, input string) ([]float32, error) {
	return c.CreateEmbeddings(ctx, []string{input})
}

func (c *Client) IsEnabled() bool {
	return c.enabled
}

