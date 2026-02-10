package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var (
	configInstance *Config
	configOnce     sync.Once
)

func parseEnv() error {
	err := godotenv.Overload()
	if err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}
	return nil
}

func loadConfig() *Config {
	parseEnv()

	return &Config{
		Port:             os.Getenv("PORT"),
		WeaviateScheme:   os.Getenv("WEAVIATE_SCHEME"),
		WeaviateHost:     os.Getenv("WEAVIATE_HOST"),
		WeaviateAPIKey:   os.Getenv("WEAVIATE_API_KEY"),
		WeaviateGrpcHost: os.Getenv("WEAVIATE_GRPC_HOST"),
	}
}

func LoadConfig() (*Config, error) {
	configOnce.Do(func() {
		configInstance = loadConfig()
	})
	return configInstance, nil
}
