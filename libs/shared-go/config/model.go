package config

type Config struct {
	ScribeQueryPort  string `mapstructure:"SCRIBE_QUERY_PORT"`
	WeaviateScheme   string `mapstructure:"WEAVIATE_SCHEME"`
	WeaviateHost     string `mapstructure:"WEAVIATE_HOST"`
	WeaviateAPIKey   string `mapstructure:"WEAVIATE_API_KEY"`
	WeaviateGrpcHost string `mapstructure:"WEAVIATE_GRPC_HOST"`
	ORIGINS          string `mapstructure:"ORIGINS"`
	OpenAIAPIKey     string `mapstructure:"OPENAI_API_KEY"`
	OpenAIModel      string `mapstructure:"OPENAI_MODEL"`
	LocalHost        string `mapstructure:"LOCAL_HOST"`
	LocalModel       string `mapstructure:"LOCAL_MODEL"`
	Provider         string `mapstructure:"PROVIDER"`
}
