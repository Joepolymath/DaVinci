package pinecone

import (
	"context"
	"errors"
	"fmt"
	"strings"

	pineconeSDK "github.com/pinecone-io/go-pinecone/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
	"go.uber.org/zap"
)

type pineconeService struct {
	client       *pineconeClient
	currentIndex string
	indexConn    *pineconeSDK.IndexConnection
	logger       *zap.Logger
}

func NewService(client *pineconeClient, logger *zap.Logger) Service {
	return &pineconeService{
		client: client,
		logger: logger,
	}
}

func (s *pineconeService) CreateCollection(ctx context.Context, req *CreateCollectionRequest) error {
	if req == nil {
		return errors.New("CreateCollectionRequest is required")
	}
	if strings.TrimSpace(req.CollectionName) == "" {
		return errors.New("collection name is required")
	}
	if req.VectorSize == 0 {
		return errors.New("vector size is required")
	}

	// Local docker mode: use configured host as data-plane and skip control-plane creation.
	if strings.TrimSpace(s.client.host) != "" {
		if err := s.ensureIndexClient(ctx, req.CollectionName); err != nil {
			return err
		}
		s.logger.Info("connected to local pinecone index", zap.String("collection", req.CollectionName))
		return nil
	}

	createReq := &pineconeSDK.CreateServerlessIndexRequest{
		Name:      req.CollectionName,
		Dimension: int32(req.VectorSize),
		Metric:    toIndexMetric(req.Distance),
		Cloud:     toCloud(s.client),
		Region:    "us-east-1",
	}

	_, err := s.client.client.CreateServerlessIndex(ctx, createReq)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "already exist") {
		return fmt.Errorf("create collection %q: %w", req.CollectionName, err)
	}

	if err := s.ensureIndexClient(ctx, req.CollectionName); err != nil {
		return err
	}

	s.logger.Info("created collection", zap.String("collection", req.CollectionName))
	return nil
}

func (s *pineconeService) UpsertPoints(ctx context.Context, req *UpsertPointsRequest) error {
	if req == nil {
		return errors.New("UpsertPointsRequest is required")
	}
	if strings.TrimSpace(req.CollectionName) == "" {
		return errors.New("collection name is required")
	}
	if len(req.Points) == 0 {
		return errors.New("at least one point is required")
	}
	if err := s.ensureIndexClient(ctx, req.CollectionName); err != nil {
		return err
	}

	vectors := make([]*pineconeSDK.Vector, 0, len(req.Points))
	for _, p := range req.Points {
		if strings.TrimSpace(p.ID) == "" {
			return errors.New("point ID is required")
		}

		var metadata *pineconeSDK.Metadata
		if len(p.Payload) > 0 {
			m, err := structpb.NewStruct(map[string]interface{}(p.Payload))
			if err != nil {
				return fmt.Errorf("invalid point payload for %q: %w", p.ID, err)
			}
			metadata = m
		}

		vectors = append(vectors, &pineconeSDK.Vector{
			Id:       p.ID,
			Values:   []float32(p.Vector),
			Metadata: metadata,
		})
	}

	_, err := s.indexConn.UpsertVectors(ctx, vectors)
	if err != nil {
		return fmt.Errorf("upsert points to %q: %w", req.CollectionName, err)
	}

	s.logger.Debug("upserted points", zap.String("collection", req.CollectionName), zap.Int("count", len(req.Points)))
	return nil
}

func (s *pineconeService) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	if req == nil {
		return nil, errors.New("SearchRequest is required")
	}
	if strings.TrimSpace(req.CollectionName) == "" {
		return nil, errors.New("collection name is required")
	}
	if len(req.Vector) == 0 {
		return nil, errors.New("search vector is required")
	}
	if err := s.ensureIndexClient(ctx, req.CollectionName); err != nil {
		return nil, err
	}

	limit := uint32(req.Limit)
	if limit == 0 {
		limit = 10
	}

	var filter *pineconeSDK.MetadataFilter
	if req.Filter != nil {
		f, err := structpb.NewStruct(map[string]interface{}(*req.Filter))
		if err != nil {
			return nil, fmt.Errorf("invalid search filter: %w", err)
		}
		filter = f
	}

	resp, err := s.indexConn.QueryByVectorValues(ctx, &pineconeSDK.QueryByVectorValuesRequest{
		Vector:          []float32(req.Vector),
		TopK:            limit,
		MetadataFilter:  filter,
		IncludeMetadata: req.WithPayload,
		IncludeValues:   req.WithVector,
	})
	if err != nil {
		return nil, fmt.Errorf("search %q: %w", req.CollectionName, err)
	}

	return &SearchResponse{Results: parseSearchResults(resp, req.WithPayload, req.WithVector, req.ScoreThreshold)}, nil
}

func (s *pineconeService) DeletePoints(ctx context.Context, req *DeletePointsRequest) error {
	if req == nil {
		return errors.New("DeletePointsRequest is required")
	}
	if strings.TrimSpace(req.CollectionName) == "" {
		return errors.New("collection name is required")
	}
	if len(req.PointIDs) == 0 {
		return errors.New("at least one point id is required")
	}
	if err := s.ensureIndexClient(ctx, req.CollectionName); err != nil {
		return err
	}

	if err := s.indexConn.DeleteVectorsById(ctx, req.PointIDs); err != nil {
		return fmt.Errorf("delete points from %q: %w", req.CollectionName, err)
	}
	s.logger.Debug("deleted points", zap.String("collection", req.CollectionName), zap.Int("count", len(req.PointIDs)))
	return nil
}

func (s *pineconeService) GetPointsByIDs(ctx context.Context, req *GetPointsByIDsRequest) (*GetPointsByIDsResponse, error) {
	if req == nil {
		return nil, errors.New("GetPointsByIDsRequest is required")
	}
	if strings.TrimSpace(req.CollectionName) == "" {
		return nil, errors.New("collection name is required")
	}
	if len(req.PointIDs) == 0 {
		return nil, errors.New("at least one point id is required")
	}
	if err := s.ensureIndexClient(ctx, req.CollectionName); err != nil {
		return nil, err
	}

	resp, err := s.indexConn.FetchVectors(ctx, req.PointIDs)
	if err != nil {
		return nil, fmt.Errorf("get points from %q: %w", req.CollectionName, err)
	}
	return &GetPointsByIDsResponse{Points: parseFetchResults(resp, req.WithPayload, req.WithVector)}, nil
}

func (s *pineconeService) ensureIndexClient(ctx context.Context, indexName string) error {
	if s.currentIndex == indexName && s.indexConn != nil {
		return nil
	}
	indexConn, err := s.client.GetIndexClient(ctx, indexName)
	if err != nil {
		return fmt.Errorf("get index connection for %q: %w", indexName, err)
	}
	s.indexConn = indexConn
	s.currentIndex = indexName
	return nil
}

func normalizeDistance(d string) string {
	switch strings.ToLower(strings.TrimSpace(d)) {
	case "cosine", "":
		return "cosine"
	case "euclid", "l2", "euclidean":
		return "euclidean"
	case "dot", "dotproduct":
		return "dotproduct"
	default:
		return "cosine"
	}
}

func toIndexMetric(distance string) pineconeSDK.IndexMetric {
	switch normalizeDistance(distance) {
	case "euclidean":
		return pineconeSDK.Euclidean
	case "dotproduct":
		return pineconeSDK.Dotproduct
	default:
		return pineconeSDK.Cosine
	}
}

func toCloud(_ *pineconeClient) pineconeSDK.Cloud {
	return pineconeSDK.Aws
}

func parseSearchResults(resp *pineconeSDK.QueryVectorsResponse, withPayload, withVector bool, scoreThreshold float32) []SearchResult {
	if resp == nil || resp.Matches == nil {
		return nil
	}

	results := make([]SearchResult, 0, len(resp.Matches))
	for _, match := range resp.Matches {
		if match == nil || match.Vector == nil {
			continue
		}
		if scoreThreshold > 0 && match.Score < scoreThreshold {
			continue
		}

		sr := SearchResult{
			ID:    match.Vector.Id,
			Score: match.Score,
		}

		if withPayload && match.Vector.Metadata != nil {
			sr.Payload = Payload(match.Vector.Metadata.AsMap())
		}
		if withVector && len(match.Vector.Values) > 0 {
			v := Vector(match.Vector.Values)
			sr.Vector = &v
		}
		results = append(results, sr)
	}
	return results
}

func parseFetchResults(resp *pineconeSDK.FetchVectorsResponse, withPayload, withVector bool) []Point {
	if resp == nil || resp.Vectors == nil {
		return nil
	}

	points := make([]Point, 0, len(resp.Vectors))
	for id, vectorData := range resp.Vectors {
		if vectorData == nil {
			continue
		}
		p := Point{ID: id}
		if withPayload && vectorData.Metadata != nil {
			p.Payload = Payload(vectorData.Metadata.AsMap())
		}
		if withVector && len(vectorData.Values) > 0 {
			p.Vector = Vector(vectorData.Values)
		}
		points = append(points, p)
	}
	return points
}

