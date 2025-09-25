package generate

import (
	"context"
	"fmt"
	"strings"

	"go-rag/internal/types"

	"github.com/sashabaranov/go-openai"
)

// Service handles response generation
type Service struct {
	client *openai.Client
	config types.GenerationConfig
}

// GenerationService interface defines the contract for generation operations
type GenerationService interface {
	GenerateResponse(ctx context.Context, query string, chunks []types.RankedChunk) (*types.GeneratedResponse, error)
}

// NewService creates a new generation service
func NewService(config types.GenerationConfig) (GenerationService, error) {
	switch config.Provider {
	case "openai":
		if config.APIKey == "" {
			return nil, fmt.Errorf("API key is required for OpenAI generation service")
		}
		client := openai.NewClient(config.APIKey)
		return &Service{
			client: client,
			config: config,
		}, nil
	case "mock":
		return NewMockService(config)
	default:
		return nil, fmt.Errorf("unsupported generation provider: %s", config.Provider)
	}
}

// GenerateResponse generates a response based on the query and relevant chunks
func (s *Service) GenerateResponse(ctx context.Context, query string, chunks []types.RankedChunk) (*types.GeneratedResponse, error) {
	if len(chunks) == 0 {
		return &types.GeneratedResponse{
			Response: "I don't have enough information to answer your question.",
			Sources:  []string{},
		}, nil
	}

	// Build responseContext from chunks
	responseContext := s.buildContext(chunks)

	// Create prompt
	prompt := s.buildPrompt(query, responseContext)

	// Generate response
	response, err := s.generateWithLLM(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

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

// generateWithLLM generates a response using an LLM
func (s *Service) generateWithLLM(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	req := openai.ChatCompletionRequest{
		Model: s.config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: float32(s.config.Temperature),
		MaxTokens:   s.config.MaxTokens,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return resp.Choices[0].Message.Content, nil
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
		response, err := s.GenerateResponse(ctx, query, chunks)
		if err != nil {
			responseChan <- fmt.Sprintf("Error generating response: %v", err)
		} else {
			responseChan <- response.Response
		}
	}()

	return responseChan, nil
}
