package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Joepolymath/DaVinci/apps/scribequery/app"
	"github.com/Joepolymath/DaVinci/apps/scribequery/internal/handlers"
	"github.com/Joepolymath/DaVinci/apps/scribequery/internal/handlers/chat"
	"github.com/Joepolymath/DaVinci/apps/scribequery/internal/router"
	sharedgo "github.com/Joepolymath/DaVinci/libs/shared-go"
	"github.com/Joepolymath/DaVinci/libs/shared-go/config"
	vector "github.com/Joepolymath/DaVinci/libs/shared-go/infra/db/vector/pinecone"
	"go.uber.org/zap"
)

func main() {
	log.Println("Starting ScribeQuery backend")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("Config loaded successfully", cfg)

	dimension := 0 // Will use default (1536) if not set
	if dimStr := os.Getenv("PINECONE_DIMENSION"); dimStr != "" {
		if d, err := strconv.Atoi(dimStr); err == nil {
			dimension = d
		}
	}

	pineconeConfig := vector.PineconeConfig{
		APIKey:    os.Getenv("PINECONE_API_KEY"),
		Host:      os.Getenv("PINECONE_HOST"),
		Namespace: os.Getenv("PINECONE_NAMESPACE"),
		Region:    os.Getenv("PINECONE_REGION"),
		Cloud:     os.Getenv("PINECONE_CLOUD"),
		Timeout:   10 * time.Second,
		Dimension: dimension,
	}

	logger, _ := zap.NewProduction()

	pineconeClient, err := vector.NewPineconeClient(pineconeConfig, logger)
	if err != nil {
		log.Fatalf("Failed to create pinecone client: %v", err)
	}

	if err := pineconeClient.Health(context.Background()); err != nil {
		log.Printf("Pinecone health check warning: %v", err)
	}
	log.Println("Pinecone client connected successfully")

	if err := pineconeClient.CreateIndex(context.Background(), &pineconeConfig, sharedgo.ScribeQueryIndex); err != nil {
		logger.Info("Index already exists", zap.Error(err))
	} else {
		logger.Info("Index created successfully")
	}

	services := app.InitServices(cfg, logger)
	if services == nil {
		logger.Error("Failed to initialize services")
		return
	}

	appEnv := router.InitRouterWithConfig(cfg)

	env := handlers.NewEnvironment(cfg, appEnv, logger, services)

	if err := router.InitHandlers(env, []handlers.IHandler{
		&chat.Handler{},
	}); err != nil {
		logger.Error("Failed to initialize handlers", zap.Error(err))
		return
	}

	go func() {
		router.RunWithGracefulShutdown(appEnv, cfg)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
