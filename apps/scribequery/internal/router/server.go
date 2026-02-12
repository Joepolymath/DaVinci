package router

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Joepolymath/DaVinci/apps/scribequery/internal/handlers"
	"github.com/Joepolymath/DaVinci/libs/shared-go/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func InitRouterWithConfig(cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		IdleTimeout:  5 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	origins := cfg.ORIGINS
	if origins == "" {
		origins = "*"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:  origins,
		AllowMethods:  "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:  "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders: "Content-Length",
		MaxAge:        300,
	}))

	return app
}

func RunWithGracefulShutdown(app *fiber.App, cfg *config.Config) error {
	go func() {
		if err := app.Listen("0.0.0.0:" + cfg.ScribeQueryPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Println("🚀 🚀 Server is running on port", cfg.ScribeQueryPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server shutdown complete.")

	return nil
}

func InitHandlers(env *handlers.Environment, handlers []handlers.IHandler) error {

	for _, handler := range handlers {
		if err := handler.Init("/api", env); err != nil {
			return fmt.Errorf("failed to initialize handler: %v", err)
		}
	}

	return nil
}
