package weaviate

type Vector []float32

type Payload map[string]interface{}

type Point struct {
	ID      string  `json:"id"`
	Vector  Vector  `json:"vector"`
	Payload Payload `json:"payload,omitempty"`
}

type CreateCollectionRequest struct {
	CollectionName string `json:"collection_name" validate:"required"`
	VectorSize     uint64 `json:"vector_size" validate:"required,min=1"`
	Distance       string `json:"distance,omitempty"` // Cosine, Euclid, Dot
}

type UpsertPointsRequest struct {
	CollectionName string  `json:"collection_name" validate:"required"`
	Points         []Point `json:"points" validate:"required,min=1"`
	Wait           bool    `json:"wait,omitempty"` // Wait for indexing to complete
}

type SearchRequest struct {
	CollectionName string   `json:"collection_name" validate:"required"`
	Vector         Vector   `json:"vector" validate:"required"`
	Limit          uint64   `json:"limit,omitempty"`           // Number of results
	ScoreThreshold float32  `json:"score_threshold,omitempty"` // Minimum similarity score
	Filter         *Payload `json:"filter,omitempty"`          // Optional metadata filter
	WithPayload    bool     `json:"with_payload,omitempty"`    // Include payload in results
	WithVector     bool     `json:"with_vector,omitempty"`     // Include vector in results
}

type SearchResult struct {
	ID      string  `json:"id"`
	Score   float32 `json:"score"`
	Payload Payload `json:"payload,omitempty"`
	Vector  *Vector `json:"vector,omitempty"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

type DeletePointsRequest struct {
	CollectionName string   `json:"collection_name" validate:"required"`
	PointIDs       []string `json:"point_ids" validate:"required,min=1"`
	Wait           bool     `json:"wait,omitempty"`
}

type GetPointsByIDsRequest struct {
	CollectionName string   `json:"collection_name" validate:"required"`
	PointIDs       []string `json:"point_ids" validate:"required,min=1"`
	WithPayload    bool     `json:"with_payload,omitempty"` // Include payload in results
	WithVector     bool     `json:"with_vector,omitempty"`  // Include vector in results
}

type GetPointsByIDsResponse struct {
	Points []Point `json:"points"`
}

type CollectionInfo struct {
	Name                string `json:"name"`
	VectorsCount        uint64 `json:"vectors_count"`
	IndexedVectorsCount uint64 `json:"indexed_vectors_count"`
	PointsCount         uint64 `json:"points_count"`
}

