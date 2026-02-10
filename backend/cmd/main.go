package main

import (
	"log"
	"time"

	"github.com/Joepolymath/ScribeQuery/internal/config"
	"github.com/Joepolymath/ScribeQuery/internal/infra/db/vector"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/grpc"
)

func main() {
	log.Println("Starting ScribeQuery backend")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("Config loaded successfully", cfg)

	weaviateConfig := vector.WeaviateConfig{
		Host:    cfg.WeaviateHost,
		Scheme:  cfg.WeaviateScheme,
		APIKey:  cfg.WeaviateAPIKey,
		Headers: map[string]string{},
		Timeout: 10 * time.Second,
		GrpcConfig: &grpc.Config{
			Host: cfg.WeaviateGrpcHost,
		},
	}

	weaviateClient, err := vector.NewWeaviateClient(weaviateConfig)
	if err != nil {
		log.Fatalf("Failed to create weaviate client: %v", err)
	}

	log.Println("Weaviate client connected successfully", weaviateClient)
}
