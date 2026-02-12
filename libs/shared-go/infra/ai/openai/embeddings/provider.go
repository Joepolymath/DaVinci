package embeddings

import (
	"context"

	"github.com/Joepolymath/DaVinci/libs/shared-go/embedding"
)

type EmbeddingProvider struct {
	client *Client
}

func NewEmbeddingProvider(client *Client) embedding.Provider {
	return &EmbeddingProvider{client: client}
}

func (p *EmbeddingProvider) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return p.client.CreateEmbedding(ctx, text)
}

func (p *EmbeddingProvider) CreateEmbeddings(ctx context.Context, texts []string) ([]float32, error) {
	return p.client.CreateEmbeddings(ctx, texts)
}

func (p *EmbeddingProvider) IsEnabled() bool {
	return p.client != nil && p.client.IsEnabled()
}

