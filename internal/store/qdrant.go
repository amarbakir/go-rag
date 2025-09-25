package store

import (
	"context"
	"fmt"
	"time"

	"go-rag/internal/embedding"
	"go-rag/internal/types"

	"github.com/qdrant/go-client/qdrant"
)

// VectorStore interface defines the contract for vector storage operations
type VectorStore interface {
	StoreChunks(ctx context.Context, chunks []types.DocumentChunk) error
	SearchSimilar(ctx context.Context, query string, limit int) ([]types.DocumentChunk, error)
	GetChunksByDocumentID(ctx context.Context, documentID string) ([]types.DocumentChunk, error)
	GetChunkByID(ctx context.Context, chunkID uint64) (*types.DocumentChunk, error)
	DeleteDocument(ctx context.Context, documentID string) error
	DeleteChunk(ctx context.Context, chunkID uint64) error
}

// QdrantStore implements VectorStore using Qdrant
type QdrantStore struct {
	config          types.VectorStoreConfig
	client          *qdrant.Client
	embeddingService embedding.Service
}

// NewQdrantStore creates a new Qdrant vector store using configuration
func NewQdrantStore(config types.VectorStoreConfig, embeddingService embedding.Service) (*QdrantStore, error) {
	// Validate config
	if config.Provider != "qdrant" {
		return nil, fmt.Errorf("unsupported vector store provider: %s", config.Provider)
	}

	if config.Host == "" {
		return nil, fmt.Errorf("qdrant host is required")
	}

	if config.CollectionName == "" {
		return nil, fmt.Errorf("collection name is required")
	}

	if embeddingService == nil {
		return nil, fmt.Errorf("embedding service is required")
	}

	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   config.Host,
		Port:   config.Port,
		APIKey: config.APIKey,
		UseTLS: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	return &QdrantStore{
		config:          config,
		client:          client,
		embeddingService: embeddingService,
	}, nil
}

// GetConfig returns the vector store configuration
func (q *QdrantStore) GetConfig() types.VectorStoreConfig {
	return q.config
}

// StoreChunks stores document chunks in Qdrant
func (q *QdrantStore) StoreChunks(ctx context.Context, chunks []types.DocumentChunk) error {
	if len(chunks) == 0 {
		return nil
	}

	// Extract texts for batch embedding generation
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	// Generate embeddings for all chunks
	embeddings, err := q.embeddingService.GenerateEmbeddings(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Prepare points for Qdrant
	points := make([]*qdrant.PointStruct, len(chunks))
	for i, chunk := range chunks {
		// Convert embedding to float32
		vector := make([]float32, len(embeddings[i]))
		for j, v := range embeddings[i] {
			vector[j] = float32(v)
		}

		// Prepare payload (metadata)
		payload := map[string]*qdrant.Value{
			"document_id":  qdrant.NewValueString(chunk.DocumentID),
			"content":      qdrant.NewValueString(chunk.Content),
			"chunk_index":  qdrant.NewValueInt(int64(chunk.ChunkIndex)),
			"created_at":   qdrant.NewValueString(chunk.CreatedAt.Format(time.RFC3339)),
			"updated_at":   qdrant.NewValueString(chunk.UpdatedAt.Format(time.RFC3339)),
		}

		// Add metadata fields
		if chunk.Metadata.Title != "" {
			payload["title"] = qdrant.NewValueString(chunk.Metadata.Title)
		}
		if chunk.Metadata.Author != "" {
			payload["author"] = qdrant.NewValueString(chunk.Metadata.Author)
		}
		if chunk.Metadata.Source != "" {
			payload["source"] = qdrant.NewValueString(chunk.Metadata.Source)
		}
		if chunk.Metadata.Language != "" {
			payload["language"] = qdrant.NewValueString(chunk.Metadata.Language)
		}
		if chunk.Metadata.ContentType != "" {
			payload["content_type"] = qdrant.NewValueString(chunk.Metadata.ContentType)
		}

		// Add tags as a list
		if len(chunk.Metadata.Tags) > 0 {
			tagInterfaces := make([]interface{}, len(chunk.Metadata.Tags))
			for j, tag := range chunk.Metadata.Tags {
				tagInterfaces[j] = tag
			}
			listValue, _ := qdrant.NewListValue(tagInterfaces)
			payload["tags"] = qdrant.NewValueList(listValue)
		}

		// Add custom metadata
		for key, value := range chunk.Metadata.Custom {
			payload["custom_"+key] = qdrant.NewValueString(value)
		}

		points[i] = &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(chunk.ID),
			Vectors: qdrant.NewVectors(vector...),
			Payload: payload,
		}
	}

	// Upsert points to Qdrant
	_, err = q.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: q.config.CollectionName,
		Points:         points,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert points to Qdrant: %w", err)
	}

	return nil
}

// SearchSimilar searches for similar chunks using vector similarity
func (q *QdrantStore) SearchSimilar(ctx context.Context, query string, limit int) ([]types.DocumentChunk, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if limit <= 0 {
		limit = 10
	}

	// Generate embedding for the query
	queryEmbedding, err := q.embeddingService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert to float32
	queryVector := make([]float32, len(queryEmbedding))
	for i, v := range queryEmbedding {
		queryVector[i] = float32(v)
	}

	// Search in Qdrant using Query
	searchResult, err := q.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: q.config.CollectionName,
		Query:          qdrant.NewQuery(queryVector...),
		Limit:          qdrant.PtrOf(uint64(limit)),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search in Qdrant: %w", err)
	}

	// Convert results to DocumentChunk
	chunks := make([]types.DocumentChunk, len(searchResult))
	for i, point := range searchResult {
		chunk, err := q.pointToDocumentChunk(point)
		if err != nil {
			return nil, fmt.Errorf("failed to convert point to document chunk: %w", err)
		}
		chunks[i] = *chunk
	}

	return chunks, nil
}

// pointToDocumentChunk converts a Qdrant point to a DocumentChunk
func (q *QdrantStore) pointToDocumentChunk(point *qdrant.ScoredPoint) (*types.DocumentChunk, error) {
	// Extract ID
	var id uint64
	if point.Id != nil {
		if numID := point.Id.GetNum(); numID != 0 {
			id = numID
		} else {
			return nil, fmt.Errorf("point ID must be numeric")
		}
	}

	if id == 0 {
		return nil, fmt.Errorf("point ID is missing")
	}

	// Extract payload
	payload := point.Payload
	if payload == nil {
		return nil, fmt.Errorf("point payload is missing")
	}

	// Extract required fields
	documentID := q.getStringFromPayload(payload, "document_id")
	content := q.getStringFromPayload(payload, "content")
	chunkIndex := int(q.getIntFromPayload(payload, "chunk_index"))

	// Parse timestamps
	createdAt, _ := time.Parse(time.RFC3339, q.getStringFromPayload(payload, "created_at"))
	updatedAt, _ := time.Parse(time.RFC3339, q.getStringFromPayload(payload, "updated_at"))

	// Extract metadata
	metadata := types.Metadata{
		Title:       q.getStringFromPayload(payload, "title"),
		Author:      q.getStringFromPayload(payload, "author"),
		Source:      q.getStringFromPayload(payload, "source"),
		Language:    q.getStringFromPayload(payload, "language"),
		ContentType: q.getStringFromPayload(payload, "content_type"),
		Custom:      make(map[string]string),
	}

	// Extract tags
	if tagsValue, exists := payload["tags"]; exists && tagsValue.GetListValue() != nil {
		tags := make([]string, 0)
		for _, tagValue := range tagsValue.GetListValue().Values {
			if tag := tagValue.GetStringValue(); tag != "" {
				tags = append(tags, tag)
			}
		}
		metadata.Tags = tags
	}

	// Extract custom metadata
	for key, value := range payload {
		if len(key) > 7 && key[:7] == "custom_" {
			customKey := key[7:]
			metadata.Custom[customKey] = value.GetStringValue()
		}
	}

	return &types.DocumentChunk{
		ID:         id,
		DocumentID: documentID,
		Content:    content,
		ChunkIndex: chunkIndex,
		Metadata:   metadata,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

// Helper functions for payload extraction
func (q *QdrantStore) getStringFromPayload(payload map[string]*qdrant.Value, key string) string {
	if value, exists := payload[key]; exists {
		return value.GetStringValue()
	}
	return ""
}

func (q *QdrantStore) getIntFromPayload(payload map[string]*qdrant.Value, key string) int64 {
	if value, exists := payload[key]; exists {
		return value.GetIntegerValue()
	}
	return 0
}

// GetChunksByDocumentID retrieves all chunks for a specific document
func (q *QdrantStore) GetChunksByDocumentID(ctx context.Context, documentID string) ([]types.DocumentChunk, error) {
	if documentID == "" {
		return nil, fmt.Errorf("document ID cannot be empty")
	}

	// Create filter for document ID
	filter := &qdrant.Filter{
		Must: []*qdrant.Condition{
			{
				ConditionOneOf: &qdrant.Condition_Field{
					Field: &qdrant.FieldCondition{
						Key: "document_id",
						Match: &qdrant.Match{
							MatchValue: &qdrant.Match_Text{
								Text: documentID,
							},
						},
					},
				},
			},
		},
	}

	// Scroll through all points with the filter
	scrollResult, err := q.client.Scroll(ctx, &qdrant.ScrollPoints{
		CollectionName: q.config.CollectionName,
		Filter:         filter,
		WithPayload:    qdrant.NewWithPayload(true),
		Limit:          qdrant.PtrOf(uint32(1000)), // Adjust as needed
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scroll points in Qdrant: %w", err)
	}

	// Convert results to DocumentChunk
	chunks := make([]types.DocumentChunk, len(scrollResult))
	for i, point := range scrollResult {
		chunk, err := q.pointToDocumentChunk(&qdrant.ScoredPoint{
			Id:      point.Id,
			Payload: point.Payload,
			Vectors: point.Vectors,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to convert point to document chunk: %w", err)
		}
		chunks[i] = *chunk
	}

	return chunks, nil
}

// GetChunkByID retrieves a specific chunk by its ID
func (q *QdrantStore) GetChunkByID(ctx context.Context, chunkID uint64) (*types.DocumentChunk, error) {
	if chunkID == 0 {
		return nil, fmt.Errorf("chunk ID cannot be zero")
	}

	// Retrieve point by ID
	getResult, err := q.client.Get(ctx, &qdrant.GetPoints{
		CollectionName: q.config.CollectionName,
		Ids:            []*qdrant.PointId{qdrant.NewIDNum(chunkID)},
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get point from Qdrant: %w", err)
	}

	if len(getResult) == 0 {
		return nil, fmt.Errorf("chunk not found: %d", chunkID)
	}

	// Convert result to DocumentChunk
	point := getResult[0]
	chunk, err := q.pointToDocumentChunk(&qdrant.ScoredPoint{
		Id:      point.Id,
		Payload: point.Payload,
		Vectors: point.Vectors,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert point to document chunk: %w", err)
	}

	return chunk, nil
}

// DeleteDocument removes all chunks for a specific document
func (q *QdrantStore) DeleteDocument(ctx context.Context, documentID string) error {
	if documentID == "" {
		return fmt.Errorf("document ID cannot be empty")
	}

	// Create filter for document ID
	filter := &qdrant.Filter{
		Must: []*qdrant.Condition{
			{
				ConditionOneOf: &qdrant.Condition_Field{
					Field: &qdrant.FieldCondition{
						Key: "document_id",
						Match: &qdrant.Match{
							MatchValue: &qdrant.Match_Text{
								Text: documentID,
							},
						},
					},
				},
			},
		},
	}

	// Delete points with the filter
	_, err := q.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: q.config.CollectionName,
		Points:         qdrant.NewPointsSelectorFilter(filter),
	})
	if err != nil {
		return fmt.Errorf("failed to delete document from Qdrant: %w", err)
	}

	return nil
}

// DeleteChunk removes a specific chunk
func (q *QdrantStore) DeleteChunk(ctx context.Context, chunkID uint64) error {
	if chunkID == 0 {
		return fmt.Errorf("chunk ID cannot be zero")
	}

	// Delete point by ID
	_, err := q.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: q.config.CollectionName,
		Points: qdrant.NewPointsSelector(qdrant.NewIDNum(chunkID)),
	})
	if err != nil {
		return fmt.Errorf("failed to delete chunk from Qdrant: %w", err)
	}

	return nil
}

// CreateCollection creates a new collection in Qdrant
func (q *QdrantStore) CreateCollection(ctx context.Context, vectorSize int) error {
	if vectorSize <= 0 {
		vectorSize = q.embeddingService.GetDimensions()
	}

	// Check if collection already exists
	collections, err := q.client.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	for _, collectionName := range collections {
		if collectionName == q.config.CollectionName {
			// Collection already exists
			return nil
		}
	}

	// Create collection
	err = q.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: q.config.CollectionName,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     uint64(vectorSize),
					Distance: qdrant.Distance_Cosine,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	return nil
}

// HealthCheck checks if Qdrant is accessible
func (q *QdrantStore) HealthCheck(ctx context.Context) error {
	// Try to list collections as a health check
	_, err := q.client.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("Qdrant health check failed: %w", err)
	}

	return nil
}


