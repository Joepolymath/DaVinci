package ai

import "context"

type ChatProvider interface {
	Completion(ctx context.Context, messages []Message, opts *ChatOptions) (*ChatResponse, error)

	CompletionStream(ctx context.Context, messages []Message, opts *ChatOptions, onDelta func(delta ChatStreamDelta) error) error

	Health(ctx context.Context) error

	IsEnabled() bool

	GetModel() string
}
