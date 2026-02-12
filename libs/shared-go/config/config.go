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
	paths := []string{".env", "../.env", "../../.env", "apps/scribequery/.env"}
	var lastErr error
	for _, path := range paths {
		if err := godotenv.Overload(path); err == nil {
			log.Printf("Loaded environment variables from %s", path)
			return nil
		} else {
			lastErr = err
		}
	}

	// If no .env file found, that's okay - we'll use environment variables directly
	if lastErr != nil {
		log.Printf("No .env file found, using environment variables directly")
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
