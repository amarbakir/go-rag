package generate

import (
	"context"
	"fmt"
	"strings"

	"go-rag/internal/types"
)

// Service handles response generation
type Service struct {
	// Add LLM client dependencies here (e.g., OpenAI, Anthropic, etc.)
}

// NewService creates a new generation service
func NewService() *Service {
	return &Service{}
}

// GenerateResponse generates a response based on the query and relevant chunks
func (s *Service) GenerateResponse(ctx context.Context, query string, chunks []types.RankedChunk) (*types.GeneratedResponse, error) {
	if len(chunks) == 0 {
		return &types.GeneratedResponse{
			Response: "I don't have enough information to answer your question.",
			Sources:  []string{},
		}, nil
	}
	
	// Build context from chunks
	context := s.buildContext(chunks)
	
	// Create prompt
	prompt := s.buildPrompt(query, context)
	
	// Generate response (placeholder implementation)
	response := s.generateWithLLM(ctx, prompt)
	
	// Extract sources
	sources := s.extractSources(chunks)
	
	return &types.GeneratedResponse{
		Response: response,
		Sources:  sources,
	}, nil
}

// buildContext combines relevant chunks into a context string
func (s *Service) buildContext(chunks []types.RankedChunk) string {
	var contextParts []string
	
	for i, chunk := range chunks {
		contextParts = append(contextParts, fmt.Sprintf("Context %d: %s", i+1, chunk.Content))
	}
	
	return strings.Join(contextParts, "\n\n")
}

// buildPrompt creates a prompt for the LLM
func (s *Service) buildPrompt(query, context string) string {
	return fmt.Sprintf(`Based on the following context, please answer the question. If the context doesn't contain enough information to answer the question, please say so.

Context:
%s

Question: %s

Answer:`, context, query)
}

// generateWithLLM generates a response using an LLM (placeholder implementation)
func (s *Service) generateWithLLM(ctx context.Context, prompt string) string {
	// TODO: Implement actual LLM integration (OpenAI, Anthropic, etc.)
	// This is a placeholder implementation
	return "This is a placeholder response. Please integrate with your preferred LLM service."
}

// extractSources extracts source information from chunks
func (s *Service) extractSources(chunks []types.RankedChunk) []string {
	var sources []string
	seenDocs := make(map[string]bool)
	
	for _, chunk := range chunks {
		if !seenDocs[chunk.DocumentID] {
			sources = append(sources, chunk.DocumentID)
			seenDocs[chunk.DocumentID] = true
		}
	}
	
	return sources
}

// StreamResponse generates a streaming response (for future implementation)
func (s *Service) StreamResponse(ctx context.Context, query string, chunks []types.RankedChunk) (<-chan string, error) {
	// TODO: Implement streaming response
	responseChan := make(chan string, 1)
	
	go func() {
		defer close(responseChan)
		response, _ := s.GenerateResponse(ctx, query, chunks)
		responseChan <- response.Response
	}()
	
	return responseChan, nil
}
