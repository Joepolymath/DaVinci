package ai

import (
	"context"
	"fmt"

	localchats "github.com/Joepolymath/DaVinci/libs/shared-go/infra/ai/local/chats"
	openaichats "github.com/Joepolymath/DaVinci/libs/shared-go/infra/ai/openai/chats"
	"go.uber.org/zap"
)

type ChatProviderConfig struct {
	Provider ProviderType

	// OpenAI-specific
	OpenAIAPIKey string
	OpenAIModel  string

	// Local (Ollama)-specific
	LocalHost  string
	LocalModel string
}

func NewChatProvider(cfg *ChatProviderConfig, logger *zap.Logger) (ChatProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("chat provider config is required")
	}

	switch cfg.Provider {
	case ProviderOpenAI:
		return newOpenAIAdapter(cfg, logger)
	case ProviderLocal:
		return newLocalAdapter(cfg, logger)
	default:
		return nil, fmt.Errorf("unsupported chat provider: %q (supported: %q, %q)", cfg.Provider, ProviderOpenAI, ProviderLocal)
	}
}

// ---------------------------------------------------------------------------
// OpenAI adapter
// ---------------------------------------------------------------------------

type openAIAdapter struct {
	client *openaichats.Client
}

func newOpenAIAdapter(cfg *ChatProviderConfig, logger *zap.Logger) (*openAIAdapter, error) {
	client, err := openaichats.NewClient(&openaichats.Config{
		APIKey: cfg.OpenAIAPIKey,
		Model:  cfg.OpenAIModel,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI chat client: %w", err)
	}
	return &openAIAdapter{client: client}, nil
}

func (a *openAIAdapter) Completion(ctx context.Context, messages []Message, opts *ChatOptions) (*ChatResponse, error) {
	oaiMsgs := toOpenAIMessages(messages)
	oaiOpts := toOpenAIOptions(opts)

	resp, err := a.client.Completion(ctx, oaiMsgs, oaiOpts)
	if err != nil {
		return nil, err
	}

	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}

	return &ChatResponse{
		Model:   resp.Model,
		Content: content,
		Usage: ChatUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func (a *openAIAdapter) CompletionStream(ctx context.Context, messages []Message, opts *ChatOptions, onDelta func(delta ChatStreamDelta) error) error {
	oaiMsgs := toOpenAIMessages(messages)
	oaiOpts := toOpenAIOptions(opts)

	return a.client.CompletionStream(ctx, oaiMsgs, oaiOpts, func(chunk openaichats.StreamChunk) error {
		content := ""
		finishReason := ""
		done := false

		if len(chunk.Choices) > 0 {
			content = chunk.Choices[0].Delta.Content
			finishReason = chunk.Choices[0].FinishReason
			done = finishReason == "stop"
		}

		return onDelta(ChatStreamDelta{
			Content:      content,
			Done:         done,
			FinishReason: finishReason,
		})
	})
}

func (a *openAIAdapter) Health(ctx context.Context) error {
	return a.client.Health(ctx)
}

func (a *openAIAdapter) IsEnabled() bool {
	return a.client.IsEnabled()
}

func (a *openAIAdapter) GetModel() string {
	return a.client.GetModel()
}

// ---------------------------------------------------------------------------
// Local (Ollama) adapter
// ---------------------------------------------------------------------------

type localAdapter struct {
	client *localchats.Client
}

func newLocalAdapter(cfg *ChatProviderConfig, logger *zap.Logger) (*localAdapter, error) {
	client, err := localchats.NewClient(&localchats.Config{
		Host:  cfg.LocalHost,
		Model: cfg.LocalModel,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create local LLM chat client: %w", err)
	}
	return &localAdapter{client: client}, nil
}

func (a *localAdapter) Completion(ctx context.Context, messages []Message, opts *ChatOptions) (*ChatResponse, error) {
	localMsgs := toLocalMessages(messages)
	localOpts := toLocalOptions(opts)

	resp, err := a.client.Completion(ctx, localMsgs, localOpts)
	if err != nil {
		return nil, err
	}

	// Ollama doesn't report standard token counts; approximate from eval counts.
	return &ChatResponse{
		Model:   resp.Model,
		Content: resp.Message.Content,
		Usage: ChatUsage{
			PromptTokens:     resp.PromptEvalCount,
			CompletionTokens: resp.EvalCount,
			TotalTokens:      resp.PromptEvalCount + resp.EvalCount,
		},
	}, nil
}

func (a *localAdapter) CompletionStream(ctx context.Context, messages []Message, opts *ChatOptions, onDelta func(delta ChatStreamDelta) error) error {
	localMsgs := toLocalMessages(messages)
	localOpts := toLocalOptions(opts)

	return a.client.CompletionStream(ctx, localMsgs, localOpts, func(chunk localchats.StreamChunk) error {
		return onDelta(ChatStreamDelta{
			Content: chunk.Message.Content,
			Done:    chunk.Done,
		})
	})
}

func (a *localAdapter) Health(ctx context.Context) error {
	return a.client.Health(ctx)
}

func (a *localAdapter) IsEnabled() bool {
	return a.client.IsEnabled()
}

func (a *localAdapter) GetModel() string {
	return a.client.GetModel()
}

// ---------------------------------------------------------------------------
// Type conversion helpers
// ---------------------------------------------------------------------------

func toOpenAIMessages(msgs []Message) []openaichats.Message {
	out := make([]openaichats.Message, len(msgs))
	for i, m := range msgs {
		out[i] = openaichats.Message{Role: m.Role, Content: m.Content}
	}
	return out
}

func toOpenAIOptions(opts *ChatOptions) *openaichats.Options {
	if opts == nil {
		return nil
	}
	return &openaichats.Options{
		Temperature: opts.Temperature,
		TopP:        opts.TopP,
		MaxTokens:   opts.MaxTokens,
		Stop:        opts.Stop,
	}
}

func toLocalMessages(msgs []Message) []localchats.Message {
	out := make([]localchats.Message, len(msgs))
	for i, m := range msgs {
		out[i] = localchats.Message{Role: m.Role, Content: m.Content}
	}
	return out
}

func toLocalOptions(opts *ChatOptions) *localchats.Options {
	if opts == nil {
		return nil
	}
	return &localchats.Options{
		Temperature: opts.Temperature,
		TopP:        opts.TopP,
		MaxTokens:   opts.MaxTokens,
		Stop:        opts.Stop,
	}
}
