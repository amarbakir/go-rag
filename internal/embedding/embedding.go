package embedding

import (
	"context"
	"fmt"
	"go-rag/internal/types"
)

// Service interface defines the contract for embedding operations
type Service interface {
	// GenerateEmbedding generates an embedding vector for a single text
	GenerateEmbedding(ctx context.Context, text string) ([]float64, error)

	// GenerateEmbeddings generates embedding vectors for multiple texts
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float64, error)

	// GetDimensions returns the dimension size of the embeddings
	GetDimensions() int

	// GetConfig returns the embedding configuration
	GetConfig() types.EmbeddingConfig
}

// NewService creates a new embedding service based on the provider configuration
func NewService(config types.EmbeddingConfig) (Service, error) {
	switch config.Provider {
	case "openai":
		return NewOpenAIService(config)
	case "mock":
		return NewMockService(config)
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", config.Provider)
	}
}
