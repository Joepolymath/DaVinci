package weaviate

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-openapi/strfmt"
	weavLib "github.com/weaviate/weaviate-go-client/v5/weaviate"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
	"go.uber.org/zap"
)

type weaviateService struct {
	client *weavLib.Client
	logger *zap.Logger
}

func NewService(client *weavLib.Client, logger *zap.Logger) Service {
	return &weaviateService{
		client: client,
		logger: logger,
	}
}

func (s *weaviateService) CreateCollection(ctx context.Context, req *CreateCollectionRequest) error {
	if req == nil {
		return errors.New("CreateCollectionRequest is required")
	}
	if strings.TrimSpace(req.CollectionName) == "" {
		return errors.New("collection name is required")
	}
	if req.VectorSize == 0 {
		return errors.New("vector size is required")
	}

	distance := normalizeDistance(req.Distance)
	class := &models.Class{
		Class:             req.CollectionName,
		Vectorizer:        "none",
		VectorIndexType:   "hnsw",
		VectorIndexConfig: map[string]interface{}{"distance": distance},
		Properties:        []*models.Property{},
	}

	err := s.client.Schema().ClassCreator().WithClass(class).Do(ctx)
	if err != nil {
		return fmt.Errorf("create collection %q: %w", req.CollectionName, err)
	}
	s.logger.Info("created collection", zap.String("collection", req.CollectionName))
	return nil
}

func (s *weaviateService) UpsertPoints(ctx context.Context, req *UpsertPointsRequest) error {
	if req == nil {
		return errors.New("UpsertPointsRequest is required")
	}
	if strings.TrimSpace(req.CollectionName) == "" {
		return errors.New("collection name is required")
	}
	if len(req.Points) == 0 {
		return errors.New("at least one point is required")
	}

	objects := make([]*models.Object, 0, len(req.Points))
	for _, p := range req.Points {
		if strings.TrimSpace(p.ID) == "" {
			return errors.New("point ID is required")
		}
		obj := &models.Object{
			Class:      req.CollectionName,
			ID:         strfmt.UUID(p.ID),
			Properties: models.PropertySchema(p.Payload),
			Vector:     []float32(p.Vector),
		}
		objects = append(objects, obj)
	}

	batcher := s.client.Batch().ObjectsBatcher()
	for _, obj := range objects {
		batcher = batcher.WithObject(obj)
	}
	_, err := batcher.Do(ctx)
	if err != nil {
		return fmt.Errorf("upsert points to %q: %w", req.CollectionName, err)
	}
	s.logger.Debug("upserted points", zap.String("collection", req.CollectionName), zap.Int("count", len(req.Points)))
	return nil
}

func (s *weaviateService) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	if req == nil {
		return nil, errors.New("SearchRequest is required")
	}
	if strings.TrimSpace(req.CollectionName) == "" {
		return nil, errors.New("collection name is required")
	}
	if len(req.Vector) == 0 {
		return nil, errors.New("search vector is required")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	nearVector := s.client.GraphQL().NearVectorArgBuilder().
		WithVector(req.Vector)
	if req.ScoreThreshold > 0 {
		nearVector = nearVector.WithCertainty(req.ScoreThreshold)
	}

	additionalFields := "_additional { id certainty"
	if req.WithVector {
		additionalFields += " vector"
	}
	additionalFields += " }"

	builder := s.client.GraphQL().Get().
		WithClassName(req.CollectionName).
		WithLimit(limit).
		WithNearVector(nearVector).
		WithFields(graphql.Field{Name: additionalFields})

	resp, err := builder.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("search %q: %w", req.CollectionName, err)
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("search GraphQL errors: %v", resp.Errors)
	}

	results := parseSearchResults(resp, req.CollectionName, req.WithPayload, req.WithVector)
	return &SearchResponse{Results: results}, nil
}

func (s *weaviateService) DeletePoints(ctx context.Context, req *DeletePointsRequest) error {
	if req == nil {
		return errors.New("DeletePointsRequest is required")
	}
	if strings.TrimSpace(req.CollectionName) == "" {
		return errors.New("collection name is required")
	}
	if len(req.PointIDs) == 0 {
		return errors.New("at least one point id is required")
	}

	if len(req.PointIDs) == 1 {
		err := s.client.Data().Deleter().
			WithClassName(req.CollectionName).
			WithID(req.PointIDs[0]).
			Do(ctx)
		if err != nil {
			return fmt.Errorf("delete point from %q: %w", req.CollectionName, err)
		}
		return nil
	}

	operands := make([]*filters.WhereBuilder, 0, len(req.PointIDs))
	for _, id := range req.PointIDs {
		operands = append(operands, filters.Where().
			WithPath([]string{"id"}).
			WithOperator(filters.Equal).
			WithValueString(id))
	}
	whereFilter := filters.Where().
		WithOperator(filters.Or).
		WithOperands(operands)

	_, err := s.client.Batch().ObjectsBatchDeleter().
		WithClassName(req.CollectionName).
		WithWhere(whereFilter).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("delete points from %q: %w", req.CollectionName, err)
	}
	s.logger.Debug("deleted points", zap.String("collection", req.CollectionName), zap.Int("count", len(req.PointIDs)))
	return nil
}

func (s *weaviateService) GetPointsByIDs(ctx context.Context, req *GetPointsByIDsRequest) (*GetPointsByIDsResponse, error) {
	if req == nil {
		return nil, errors.New("GetPointsByIDsRequest is required")
	}
	if strings.TrimSpace(req.CollectionName) == "" {
		return nil, errors.New("collection name is required")
	}
	if len(req.PointIDs) == 0 {
		return nil, errors.New("at least one point id is required")
	}

	points := make([]Point, 0, len(req.PointIDs))
	for _, id := range req.PointIDs {
		getter := s.client.Data().ObjectsGetter().
			WithClassName(req.CollectionName).
			WithID(id)
		if req.WithVector {
			getter = getter.WithVector()
		}
		objs, err := getter.Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("get point %q from %q: %w", id, req.CollectionName, err)
		}
		if len(objs) == 0 {
			continue
		}
		obj := objs[0]
		p := pointFromObject(obj, req.WithPayload, req.WithVector)
		p.ID = id
		points = append(points, p)
	}
	return &GetPointsByIDsResponse{Points: points}, nil
}

func normalizeDistance(d string) string {
	switch strings.ToLower(strings.TrimSpace(d)) {
	case "cosine", "":
		return "cosine"
	case "euclid", "l2":
		return "l2"
	case "dot":
		return "dot"
	default:
		return "cosine"
	}
}

func parseSearchResults(resp *models.GraphQLResponse, className string, withPayload, withVector bool) []SearchResult {
	if resp.Data == nil {
		return nil
	}
	getData, ok := resp.Data["Get"].(map[string]interface{})
	if !ok {
		return nil
	}
	classResults, ok := getData[className].([]interface{})
	if !ok {
		return nil
	}
	results := make([]SearchResult, 0, len(classResults))
	for _, r := range classResults {
		item, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		sr := SearchResult{}
		if id, ok := item["_additional"].(map[string]interface{}); ok {
			if v, ok := id["id"].(string); ok {
				sr.ID = v
			}
			if c, ok := id["certainty"].(float64); ok {
				sr.Score = float32(c)
			}
			if withVector && id["vector"] != nil {
				if vec, ok := id["vector"].([]interface{}); ok {
					v := make(Vector, 0, len(vec))
					for _, f := range vec {
						if fl, ok := f.(float64); ok {
							v = append(v, float32(fl))
						}
					}
					sr.Vector = &v
				}
			}
		}
		if withPayload {
			sr.Payload = make(Payload)
			for k, val := range item {
				if k != "_additional" {
					sr.Payload[k] = val
				}
			}
		}
		results = append(results, sr)
	}
	return results
}

func pointFromObject(obj *models.Object, withPayload, withVector bool) Point {
	p := Point{ID: obj.ID.String()}
	if withPayload && obj.Properties != nil {
		if m, ok := obj.Properties.(map[string]interface{}); ok {
			p.Payload = Payload(m)
		}
	}
	if withVector && len(obj.Vector) > 0 {
		p.Vector = Vector(obj.Vector)
	}
	return p
}

