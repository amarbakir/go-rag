package embedding

import (
	"context"
	"crypto/md5"
	"fmt"
	"math"

	"go-rag/internal/types"
)

// MockService implements the embedding Service interface for testing
type MockService struct {
	config types.EmbeddingConfig
}

// NewMockService creates a new mock embedding service
func NewMockService(config types.EmbeddingConfig) (*MockService, error) {
	return &MockService{
		config: config,
	}, nil
}

// GenerateEmbedding generates a deterministic mock embedding vector for a single text
func (s *MockService) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// Create a deterministic embedding based on text hash
	hash := md5.Sum([]byte(text))
	
	// Generate embedding vector of specified dimensions
	embedding := make([]float64, s.config.Dimensions)
	
	// Use hash bytes to seed the embedding values
	for i := 0; i < s.config.Dimensions; i++ {
		// Use different parts of the hash to create variation
		byteIndex := i % len(hash)
		value := float64(hash[byteIndex]) / 255.0 // Normalize to 0-1
		
		// Add some mathematical transformation to create more realistic embeddings
		angle := 2 * math.Pi * value
		embedding[i] = math.Sin(angle + float64(i)*0.1)
	}
	
	// Normalize the vector
	return normalizeVector(embedding), nil
}

// GenerateEmbeddings generates embedding vectors for multiple texts
func (s *MockService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	embeddings := make([][]float64, len(texts))
	for i, text := range texts {
		if text == "" {
			continue // Skip empty texts
		}
		
		embedding, err := s.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// GetDimensions returns the dimension size of the embeddings
func (s *MockService) GetDimensions() int {
	return s.config.Dimensions
}

// GetConfig returns the embedding configuration
func (s *MockService) GetConfig() types.EmbeddingConfig {
	return s.config
}

// normalizeVector normalizes a vector to unit length
func normalizeVector(vector []float64) []float64 {
	var magnitude float64
	for _, val := range vector {
		magnitude += val * val
	}
	magnitude = math.Sqrt(magnitude)
	
	if magnitude == 0 {
		return vector
	}
	
	normalized := make([]float64, len(vector))
	for i, val := range vector {
		normalized[i] = val / magnitude
	}
	
	return normalized
}
