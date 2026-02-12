package embedding

import (
	"context"
)

type Provider interface {
	CreateEmbedding(ctx context.Context, text string) ([]float32, error)

	CreateEmbeddings(ctx context.Context, texts []string) ([]float32, error)

	IsEnabled() bool
}

type Service interface {
	CreateEmbedding(ctx context.Context, text string) ([]float32, error)

	CreateEmbeddings(ctx context.Context, texts []string) ([]float32, error)
}

