package pinecone

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	pineconeSDK "github.com/pinecone-io/go-pinecone/pinecone"
	"go.uber.org/zap"
)

const (
	defaultTimeout   = 60 * time.Second
	defaultDimension = 1536 // Default for OpenAI text-embedding-3-small
)

type PineconeConfig struct {
	APIKey    string
	Host      string // Local Docker or explicit control plane host
	Region    string // Used for cloud serverless index creation (default: us-east-1)
	Cloud     string // Used for cloud serverless index creation (default: aws)
	Namespace string // Optional namespace for data-plane operations
	SourceTag string
	Timeout   time.Duration
	Dimension int // Vector dimension (default: 1536 for OpenAI text-embedding-3-small)
}

type pineconeClient struct {
	client       *pineconeSDK.Client
	indexClients map[string]*pineconeSDK.IndexConnection
	host         string
	namespace    string
	logger       *zap.Logger
}

func NewPineconeClient(cfg PineconeConfig, logger *zap.Logger) (*pineconeClient, error) {
	cfg = cfg.withDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	var (
		client *pineconeSDK.Client
		err    error
	)

	if strings.TrimSpace(cfg.APIKey) == "" {
		// Local Docker or custom auth-less setups.
		client, err = pineconeSDK.NewClientBase(pineconeSDK.NewClientBaseParams{
			Host:      cfg.Host,
			SourceTag: cfg.SourceTag,
		})
	} else {
		client, err = pineconeSDK.NewClient(pineconeSDK.NewClientParams{
			ApiKey: cfg.APIKey,
			// Host:      cfg.Host,
			// SourceTag: cfg.SourceTag,
		})
		logger.Info("created pinecone client")
	}
	if err != nil {
		return nil, fmt.Errorf("connect to pinecone failed: %w", err)
	}

	return &pineconeClient{
		client:       client,
		indexClients: make(map[string]*pineconeSDK.IndexConnection),
		host:         cfg.Host,
		namespace:    cfg.Namespace,
		logger:       logger,
	}, nil
}

func (pc *pineconeClient) GetIndexClient(ctx context.Context, indexName string) (*pineconeSDK.IndexConnection, error) {
	if indexConn, ok := pc.indexClients[indexName]; ok {
		return indexConn, nil
	}

	host := strings.TrimSpace(pc.host)
	if host == "" {
		idx, err := pc.client.DescribeIndex(ctx, indexName)
		if err != nil {
			pc.logger.Error("describe index failed", zap.Error(err))
			return nil, fmt.Errorf("describe index failed for %q: %w", indexName, err)
		}
		host = idx.Host
	}

	indexConn, err := pc.client.Index(pineconeSDK.NewIndexConnParams{
		Host:      host,
		Namespace: pc.namespace,
	})
	if err != nil {
		pc.logger.Error("create index connection failed", zap.Error(err))
		return nil, fmt.Errorf("create index connection failed for %q: %w", indexName, err)
	}

	pc.indexClients[indexName] = indexConn
	return indexConn, nil
}

func (pc *pineconeClient) Health(ctx context.Context) error {
	if pc.client == nil {
		return errors.New("pinecone client is not initialized")
	}
	indexes, err := pc.client.ListIndexes(ctx)
	if err != nil && strings.TrimSpace(pc.host) != "" {
		// Local docker deployments may not expose full control-plane APIs.
		pc.logger.Error("list indexes failed", zap.Error(err))
		return nil
	}
	if err != nil {
		return fmt.Errorf("pinecone health check failed: %w", err)
	}
	pc.logger.Info("list indexes", zap.Any("indexes", indexes))
	return nil
}

func (pc *pineconeClient) CreateIndex(ctx context.Context, cfg *PineconeConfig, indexName string) error {
	// vectorType := "sparse"
	metric := pineconeSDK.Dotproduct
	deletionProtection := pineconeSDK.DeletionProtectionDisabled

	dimension := cfg.Dimension
	if dimension == 0 {
		dimension = defaultDimension
	}

	idx, err := pc.client.CreateServerlessIndex(ctx, &pineconeSDK.CreateServerlessIndexRequest{
		Name:               indexName,
		Metric:             metric,
		Cloud:              pineconeSDK.Aws,
		Region:             cfg.Region,
		DeletionProtection: deletionProtection,
		Dimension:          int32(dimension),
	})
	if err != nil {
		pc.logger.Error("Failed to create serverless index", zap.Error(err))
		return fmt.Errorf("failed to create serverless index: %w", err)
	} else {
		pc.logger.Info("Successfully created serverless index", zap.String("index", idx.Name))
	}
	return nil
}

func (cfg PineconeConfig) withDefaults() PineconeConfig {
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	if strings.TrimSpace(cfg.Region) == "" {
		cfg.Region = "us-east-1"
	}
	if strings.TrimSpace(cfg.Cloud) == "" {
		cfg.Cloud = "aws"
	}
	if cfg.Dimension == 0 {
		cfg.Dimension = defaultDimension
	}
	return cfg
}

func (cfg PineconeConfig) validate() error {
	isLocal := strings.HasPrefix(strings.TrimSpace(cfg.Host), "http://localhost") ||
		strings.HasPrefix(strings.TrimSpace(cfg.Host), "http://127.0.0.1")

	if !isLocal && strings.TrimSpace(cfg.APIKey) == "" {
		return errors.New("pinecone API key is required for non-local deployments")
	}
	return nil
}

