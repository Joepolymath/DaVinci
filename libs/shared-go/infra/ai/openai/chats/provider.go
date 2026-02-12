package chats

import "context"

// Provider wraps the OpenAI chat completion client and exposes
// a high-level interface for chat completions.
type Provider struct {
	client *Client
}

// NewProvider creates a new chat completion provider backed by OpenAI.
func NewProvider(client *Client) *Provider {
	return &Provider{client: client}
}

// Completion sends a non-streaming chat completion request.
func (p *Provider) Completion(ctx context.Context, messages []Message, opts *Options) (*CompletionResponse, error) {
	return p.client.Completion(ctx, messages, opts)
}

// CompletionStream sends a streaming chat completion request.
func (p *Provider) CompletionStream(ctx context.Context, messages []Message, opts *Options, onChunk func(chunk StreamChunk) error) error {
	return p.client.CompletionStream(ctx, messages, opts, onChunk)
}

// IsEnabled returns whether the underlying client is enabled.
func (p *Provider) IsEnabled() bool {
	return p.client != nil && p.client.IsEnabled()
}

// GetModel returns the configured model name.
func (p *Provider) GetModel() string {
	return p.client.GetModel()
}
