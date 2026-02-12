package config

type Config struct {
	Port             string `mapstructure:"PORT"`
	WeaviateScheme   string `mapstructure:"WEAVIATE_SCHEME"`
	WeaviateHost     string `mapstructure:"WEAVIATE_HOST"`
	WeaviateAPIKey   string `mapstructure:"WEAVIATE_API_KEY"`
	WeaviateGrpcHost string `mapstructure:"WEAVIATE_GRPC_HOST"`
	ORIGINS          string `mapstructure:"ORIGINS"`
}

