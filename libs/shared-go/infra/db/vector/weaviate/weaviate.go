package weaviate

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/weaviate/weaviate-go-client/v5/weaviate"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/auth"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/grpc"
)

const (
	defaultScheme         = "http"
	defaultStartupTimeout = 60 * time.Second
)

type WeaviateConfig struct {
	Host           string
	Scheme         string
	APIKey         string
	Headers        map[string]string
	StartupTimeout time.Duration
	Timeout        time.Duration
	GrpcConfig     *grpc.Config
}

func NewWeaviateClient(cfg WeaviateConfig) (*weaviate.Client, error) {
	cfg = cfg.withDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	client, err := weaviate.NewClient(weaviate.Config{
		Host:           cfg.Host,
		Scheme:         cfg.Scheme,
		AuthConfig:     cfg.authConfig(),
		Headers:        cfg.Headers,
		StartupTimeout: cfg.StartupTimeout,
		Timeout:        cfg.Timeout,
		GrpcConfig:     cfg.GrpcConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to weaviate: %w", err)
	}

	return client, nil
}

func (cfg WeaviateConfig) withDefaults() WeaviateConfig {
	if cfg.Scheme == "" {
		cfg.Scheme = defaultScheme
	}
	if cfg.StartupTimeout == 0 {
		cfg.StartupTimeout = defaultStartupTimeout
	}
	return cfg
}

func (cfg WeaviateConfig) validate() error {
	if strings.TrimSpace(cfg.Host) == "" {
		return errors.New("weaviate host is required")
	}
	switch cfg.Scheme {
	case "http", "https":
	default:
		return fmt.Errorf("unsupported weaviate scheme: %s", cfg.Scheme)
	}
	return nil
}

func (cfg WeaviateConfig) authConfig() auth.Config {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil
	}
	return auth.ApiKey{Value: cfg.APIKey}
}

