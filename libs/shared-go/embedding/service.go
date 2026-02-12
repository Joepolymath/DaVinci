package embedding

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type embeddingService struct {
	provider Provider
	logger   *zap.Logger
}

func NewService(provider Provider, logger *zap.Logger) Service {
	return &embeddingService{
		provider: provider,
		logger:   logger,
	}
}

func (s *embeddingService) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	if !s.provider.IsEnabled() {
		return nil, fmt.Errorf("embedding provider is not enabled")
	}

	s.logger.Debug("Creating embedding",
		zap.String("text_length", fmt.Sprintf("%d", len(text))))

	embedding, err := s.provider.CreateEmbedding(ctx, text)
	if err != nil {
		s.logger.Error("Failed to create embedding",
			zap.Error(err))
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	s.logger.Debug("Embedding created successfully",
		zap.Int("dimension", len(embedding)))

	return embedding, nil
}

func (s *embeddingService) CreateEmbeddings(ctx context.Context, texts []string) ([]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	if !s.provider.IsEnabled() {
		return nil, fmt.Errorf("embedding provider is not enabled")
	}

	s.logger.Debug("Creating embeddings",
		zap.Int("text_count", len(texts)))

	// For now, we'll create embeddings for all texts and return the first one
	// This can be extended to return all embeddings if needed
	embedding, err := s.provider.CreateEmbeddings(ctx, texts)
	if err != nil {
		s.logger.Error("Failed to create embeddings",
			zap.Error(err))
		return nil, fmt.Errorf("failed to create embeddings: %w", err)
	}

	s.logger.Debug("Embeddings created successfully",
		zap.Int("dimension", len(embedding)),
		zap.Int("text_count", len(texts)))

	return embedding, nil
}

