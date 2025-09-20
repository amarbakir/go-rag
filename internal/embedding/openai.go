package embedding

import (
	"context"
	"fmt"

	"go-rag/internal/types"

	"github.com/sashabaranov/go-openai"
)

// OpenAIService implements the embedding Service interface using OpenAI
type OpenAIService struct {
	client *openai.Client
	config types.EmbeddingConfig
}

// NewOpenAIService creates a new OpenAI embedding service
func NewOpenAIService(config types.EmbeddingConfig) (*OpenAIService, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	client := openai.NewClient(config.APIKey)

	return &OpenAIService{
		client: client,
		config: config,
	}, nil
}

// GenerateEmbedding generates an embedding vector for a single text
func (s *OpenAIService) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.EmbeddingModel(s.config.Model),
	}

	resp, err := s.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	// Convert float32 to float64
	embedding := make([]float64, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		embedding[i] = float64(v)
	}

	return embedding, nil
}

// GenerateEmbeddings generates embedding vectors for multiple texts
func (s *OpenAIService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	// Filter out empty texts
	validTexts := make([]string, 0, len(texts))
	for _, text := range texts {
		if text != "" {
			validTexts = append(validTexts, text)
		}
	}

	if len(validTexts) == 0 {
		return nil, fmt.Errorf("no valid texts provided")
	}

	req := openai.EmbeddingRequest{
		Input: validTexts,
		Model: openai.EmbeddingModel(s.config.Model),
	}

	resp, err := s.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create embeddings: %w", err)
	}

	if len(resp.Data) != len(validTexts) {
		return nil, fmt.Errorf("embedding count mismatch: expected %d, got %d", len(validTexts), len(resp.Data))
	}

	embeddings := make([][]float64, len(resp.Data))
	for i, data := range resp.Data {
		embedding := make([]float64, len(data.Embedding))
		for j, v := range data.Embedding {
			embedding[j] = float64(v)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// GetDimensions returns the dimension size of the embeddings
func (s *OpenAIService) GetDimensions() int {
	return s.config.Dimensions
}

// GetConfig returns the embedding configuration
func (s *OpenAIService) GetConfig() types.EmbeddingConfig {
	return s.config
}
