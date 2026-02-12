package chat

import (
	"github.com/Joepolymath/DaVinci/apps/scribequery/internal/domain/chat"
	"github.com/Joepolymath/DaVinci/apps/scribequery/internal/handlers"
	"github.com/Joepolymath/DaVinci/libs/shared-go/infra/ai"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service chat.Service
	env     *handlers.Environment
}

func (h *Handler) Init(basePath string, env *handlers.Environment) error {
	h.env = env
	h.service = env.Services.ChatService

	group := env.Fiber.Group(basePath + "/chats")

	group.Post("/", h.chat)

	return nil
}

func (h *Handler) chat(c *fiber.Ctx) error {
	var request ai.Message
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	response, err := h.service.Chat(c.Context(), []ai.Message{request})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to chat",
		})
	}

	return c.JSON(response)
}
