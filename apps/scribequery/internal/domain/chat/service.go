package chat

import (
	"context"

	"github.com/Joepolymath/DaVinci/libs/shared-go/infra/ai"
)

type service struct {
	aiProvider ai.ChatProvider
}

func NewService(aiProvider ai.ChatProvider) Service {
	return &service{
		aiProvider: aiProvider,
	}
}

func (s *service) Chat(ctx context.Context, messages []ai.Message) (ai.ChatResponse, error) {
	resp, err := s.aiProvider.Completion(ctx, messages, nil)
	if err != nil {
		return ai.ChatResponse{}, err
	}
	return *resp, nil
}
