package weaviate

import "context"

// Service defines the high-level vector store operations for RAG:
// create collection, upsert points, similarity search, delete, and get by IDs.
type Service interface {
	CreateCollection(ctx context.Context, req *CreateCollectionRequest) error
	UpsertPoints(ctx context.Context, req *UpsertPointsRequest) error
	Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
	DeletePoints(ctx context.Context, req *DeletePointsRequest) error
	GetPointsByIDs(ctx context.Context, req *GetPointsByIDsRequest) (*GetPointsByIDsResponse, error)
}

// Client is the low-level Weaviate connection handle (e.g. *weaviate.Client).
// Health can be used for liveness/readiness checks.
type Client interface {
	Health(ctx context.Context) error
}

