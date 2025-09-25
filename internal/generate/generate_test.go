package generate

import (
	"context"
	"testing"

	"go-rag/internal/types"
)

func TestNewService_Success(t *testing.T) {
	config := types.GenerationConfig{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		MaxTokens:   1000,
		APIKey:      "test-api-key",
	}

	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if service == nil {
		t.Fatal("Service is nil")
	}

	if service.config.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", service.config.Provider)
	}

	if service.config.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected model 'gpt-3.5-turbo', got '%s'", service.config.Model)
	}
}

func TestNewService_MissingAPIKey(t *testing.T) {
	config := types.GenerationConfig{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		MaxTokens:   1000,
		APIKey:      "",
	}

	_, err := NewService(config)
	if err == nil {
		t.Error("Expected error for missing API key, got nil")
	}

	expectedMsg := "API key is required for generation service"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestNewService_UnsupportedProvider(t *testing.T) {
	config := types.GenerationConfig{
		Provider:    "unsupported",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		MaxTokens:   1000,
		APIKey:      "test-api-key",
	}

	_, err := NewService(config)
	if err == nil {
		t.Error("Expected error for unsupported provider, got nil")
	}

	expectedMsg := "unsupported generation provider: unsupported"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGenerateResponse_EmptyChunks(t *testing.T) {
	config := types.GenerationConfig{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		MaxTokens:   1000,
		APIKey:      "test-api-key",
	}

	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()
	response, err := service.GenerateResponse(ctx, "test query", []types.RankedChunk{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedResponse := "I don't have enough information to answer your question."
	if response.Response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, response.Response)
	}

	if len(response.Sources) != 0 {
		t.Errorf("Expected 0 sources, got %d", len(response.Sources))
	}
}

func TestBuildContext(t *testing.T) {
	config := types.GenerationConfig{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		MaxTokens:   1000,
		APIKey:      "test-api-key",
	}

	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	chunks := []types.RankedChunk{
		{
			DocumentChunk: types.DocumentChunk{
				Content: "First chunk content",
			},
		},
		{
			DocumentChunk: types.DocumentChunk{
				Content: "Second chunk content",
			},
		},
	}

	context := service.buildContext(chunks)
	expected := "Context 1: First chunk content\n\nContext 2: Second chunk content"
	if context != expected {
		t.Errorf("Expected context '%s', got '%s'", expected, context)
	}
}

func TestBuildPrompt(t *testing.T) {
	config := types.GenerationConfig{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		MaxTokens:   1000,
		APIKey:      "test-api-key",
	}

	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	query := "What is AI?"
	context := "AI is artificial intelligence"
	prompt := service.buildPrompt(query, context)

	if !contains(prompt, query) {
		t.Errorf("Prompt should contain query '%s'", query)
	}

	if !contains(prompt, context) {
		t.Errorf("Prompt should contain context '%s'", context)
	}

	if !contains(prompt, "Based on the following context") {
		t.Error("Prompt should contain instruction text")
	}
}

func TestExtractSources(t *testing.T) {
	config := types.GenerationConfig{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		MaxTokens:   1000,
		APIKey:      "test-api-key",
	}

	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	chunks := []types.RankedChunk{
		{
			DocumentChunk: types.DocumentChunk{
				DocumentID: "doc-1",
			},
		},
		{
			DocumentChunk: types.DocumentChunk{
				DocumentID: "doc-2",
			},
		},
		{
			DocumentChunk: types.DocumentChunk{
				DocumentID: "doc-1", // Duplicate
			},
		},
	}

	sources := service.extractSources(chunks)
	if len(sources) != 2 {
		t.Errorf("Expected 2 unique sources, got %d", len(sources))
	}

	if sources[0] != "doc-1" || sources[1] != "doc-2" {
		t.Errorf("Expected sources ['doc-1', 'doc-2'], got %v", sources)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
