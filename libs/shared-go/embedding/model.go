package embedding

type Embedding []float32

type CreateEmbeddingRequest struct {
	Text string `json:"text" validate:"required"`
}

type CreateEmbeddingsRequest struct {
	Texts []string `json:"texts" validate:"required,min=1"`
}

type CreateEmbeddingResponse struct {
	Embedding Embedding `json:"embedding"`
	Dimension int       `json:"dimension"`
}

type CreateEmbeddingsResponse struct {
	Embeddings []Embedding `json:"embeddings"`
	Dimension  int         `json:"dimension"`
}

