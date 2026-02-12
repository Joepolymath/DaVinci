package handlers

import (
	"github.com/Joepolymath/DaVinci/apps/scribequery/app"
	"github.com/Joepolymath/DaVinci/libs/shared-go/config"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type IHandler interface {
	Init(string, *Environment) error
}

type Environment struct {
	Config *config.Config
	Fiber  *fiber.App
	Logger *zap.Logger

	Services *app.Services
}

func NewEnvironment(cfg *config.Config, fiber *fiber.App, logger *zap.Logger, services *app.Services) *Environment {
	return &Environment{
		Config:   cfg,
		Fiber:    fiber,
		Logger:   logger,
		Services: services,
	}
}
