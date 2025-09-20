package generate

import (
	"context"
	"fmt"
	"strings"

	"go-rag/internal/types"
)

// MockService implements a mock generation service for testing
type MockService struct {
	config types.GenerationConfig
}

// NewMockService creates a new mock generation service
func NewMockService(config types.GenerationConfig) (*MockService, error) {
	return &MockService{
		config: config,
	}, nil
}

// GenerateResponse generates a mock response based on the query and relevant chunks
func (s *MockService) GenerateResponse(ctx context.Context, query string, chunks []types.RankedChunk) (*types.GeneratedResponse, error) {
	if len(chunks) == 0 {
		return &types.GeneratedResponse{
			Response: "I don't have enough information to answer your question.",
			Sources:  []string{},
		}, nil
	}

	// Build a simple mock response based on the chunks
	var contextParts []string
	var sources []string
	
	for i, chunk := range chunks {
		if i < 3 { // Use first 3 chunks for context
			contextParts = append(contextParts, chunk.Content)
		}
		sources = append(sources, chunk.DocumentID)
	}

	// Create a mock response that incorporates the query and context
	response := fmt.Sprintf("Based on the provided information about %s, here's what I found: %s", 
		query, 
		strings.Join(contextParts, " "))

	// Deduplicate sources
	uniqueSources := make(map[string]bool)
	var finalSources []string
	for _, source := range sources {
		if !uniqueSources[source] {
			uniqueSources[source] = true
			finalSources = append(finalSources, source)
		}
	}

	return &types.GeneratedResponse{
		Response: response,
		Sources:  finalSources,
	}, nil
}
