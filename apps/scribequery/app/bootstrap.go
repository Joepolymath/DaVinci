package app

import (
	"github.com/Joepolymath/DaVinci/apps/scribequery/internal/domain/chat"
	"github.com/Joepolymath/DaVinci/libs/shared-go/config"
	"github.com/Joepolymath/DaVinci/libs/shared-go/infra/ai"
	"go.uber.org/zap"
)

type Services struct {
	ChatService chat.Service
}

func InitServices(cfg *config.Config, logger *zap.Logger) *Services {
	chatProviderConfig := &ai.ChatProviderConfig{
		Provider:     ai.ProviderLocal,
		OpenAIAPIKey: cfg.OpenAIAPIKey,
		OpenAIModel:  cfg.OpenAIModel,
		LocalHost:    cfg.LocalHost,
		LocalModel:   cfg.LocalModel,
	}

	chatProvider, err := ai.NewChatProvider(chatProviderConfig, logger)
	if err != nil {
		logger.Error("Failed to create chat provider", zap.Error(err))
		return nil
	}

	return &Services{
		ChatService: chat.NewService(chatProvider),
	}
}
