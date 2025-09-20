package store

import (
	"context"
	"testing"

	"go-rag/internal/embedding"
	"go-rag/internal/types"

	"github.com/qdrant/go-client/qdrant"
)

// MockEmbeddingService for testing
type MockEmbeddingService struct {
	dimensions int
}

func (m *MockEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	// Return a simple mock embedding
	embedding := make([]float64, m.dimensions)
	for i := range embedding {
		embedding[i] = 0.1 * float64(i+1)
	}
	return embedding, nil
}

func (m *MockEmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))
	for i := range texts {
		embedding, err := m.GenerateEmbedding(ctx, texts[i])
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}
	return embeddings, nil
}

func (m *MockEmbeddingService) GetDimensions() int {
	return m.dimensions
}

func (m *MockEmbeddingService) GetConfig() types.EmbeddingConfig {
	return types.EmbeddingConfig{
		Provider:   "mock",
		Model:      "mock-model",
		Dimensions: m.dimensions,
	}
}

func TestNewQdrantStore(t *testing.T) {
	config := types.VectorStoreConfig{
		Provider:       "qdrant",
		Host:           "localhost",
		Port:           6333,
		CollectionName: "test_collection",
	}

	mockEmbedding := &MockEmbeddingService{dimensions: 384}

	store, err := NewQdrantStore(config, mockEmbedding)
	if err != nil {
		t.Fatalf("Failed to create QdrantStore: %v", err)
	}

	if store == nil {
		t.Fatal("QdrantStore is nil")
	}

	if store.config.Provider != "qdrant" {
		t.Errorf("Expected provider 'qdrant', got '%s'", store.config.Provider)
	}

	if store.embeddingService == nil {
		t.Error("Embedding service is nil")
	}
}

func TestNewQdrantStore_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config types.VectorStoreConfig
		embSvc embedding.Service
		errMsg string
	}{
		{
			name: "unsupported provider",
			config: types.VectorStoreConfig{
				Provider: "invalid",
			},
			embSvc: &MockEmbeddingService{dimensions: 384},
			errMsg: "unsupported vector store provider",
		},
		{
			name: "missing host",
			config: types.VectorStoreConfig{
				Provider: "qdrant",
				Host:     "",
			},
			embSvc: &MockEmbeddingService{dimensions: 384},
			errMsg: "qdrant host is required",
		},
		{
			name: "missing collection name",
			config: types.VectorStoreConfig{
				Provider: "qdrant",
				Host:     "localhost",
			},
			embSvc: &MockEmbeddingService{dimensions: 384},
			errMsg: "collection name is required",
		},
		{
			name: "missing embedding service",
			config: types.VectorStoreConfig{
				Provider:       "qdrant",
				Host:           "localhost",
				CollectionName: "test",
			},
			embSvc: nil,
			errMsg: "embedding service is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewQdrantStore(tt.config, tt.embSvc)
			if err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
				return
			}
			if err.Error() == "" || len(err.Error()) == 0 {
				t.Errorf("Expected error containing '%s', got empty error", tt.errMsg)
			}
		})
	}
}

func TestDocumentChunkConversion(t *testing.T) {
	config := types.VectorStoreConfig{
		Provider:       "qdrant",
		Host:           "localhost",
		Port:           6333,
		CollectionName: "test_collection",
	}

	mockEmbedding := &MockEmbeddingService{dimensions: 384}
	store, err := NewQdrantStore(config, mockEmbedding)
	if err != nil {
		t.Fatalf("Failed to create QdrantStore: %v", err)
	}

	// Test helper functions
	payload := map[string]*qdrant.Value{
		"test_string": qdrant.NewValueString("test_value"),
		"test_int":    qdrant.NewValueInt(42),
	}

	stringVal := store.getStringFromPayload(payload, "test_string")
	if stringVal != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", stringVal)
	}

	intVal := store.getIntFromPayload(payload, "test_int")
	if intVal != 42 {
		t.Errorf("Expected 42, got %d", intVal)
	}

	// Test missing keys
	missingString := store.getStringFromPayload(payload, "missing")
	if missingString != "" {
		t.Errorf("Expected empty string for missing key, got '%s'", missingString)
	}

	missingInt := store.getIntFromPayload(payload, "missing")
	if missingInt != 0 {
		t.Errorf("Expected 0 for missing key, got %d", missingInt)
	}
}

func TestEmbeddingGeneration(t *testing.T) {
	mockEmbedding := &MockEmbeddingService{dimensions: 3}

	ctx := context.Background()

	// Test single embedding
	embedding, err := mockEmbedding.GenerateEmbedding(ctx, "test text")
	if err != nil {
		t.Fatalf("Failed to generate embedding: %v", err)
	}

	if len(embedding) != 3 {
		t.Errorf("Expected embedding length 3, got %d", len(embedding))
	}

	// Test batch embeddings
	texts := []string{"text1", "text2", "text3"}
	embeddings, err := mockEmbedding.GenerateEmbeddings(ctx, texts)
	if err != nil {
		t.Fatalf("Failed to generate embeddings: %v", err)
	}

	if len(embeddings) != 3 {
		t.Errorf("Expected 3 embeddings, got %d", len(embeddings))
	}

	for i, emb := range embeddings {
		if len(emb) != 3 {
			t.Errorf("Expected embedding %d length 3, got %d", i, len(emb))
		}
	}
}
