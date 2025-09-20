package embedding

import (
	"context"
	"testing"

	"go-rag/internal/types"
)

func TestNewOpenAIService(t *testing.T) {
	config := types.EmbeddingConfig{
		Provider:   "openai",
		Model:      "text-embedding-ada-002",
		Dimensions: 1536,
		APIKey:     "test-api-key",
	}

	service, err := NewOpenAIService(config)
	if err != nil {
		t.Fatalf("Failed to create OpenAI service: %v", err)
	}

	if service == nil {
		t.Fatal("OpenAI service is nil")
	}

	if service.GetDimensions() != 1536 {
		t.Errorf("Expected dimensions 1536, got %d", service.GetDimensions())
	}

	serviceConfig := service.GetConfig()
	if serviceConfig.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", serviceConfig.Provider)
	}

	if serviceConfig.Model != "text-embedding-ada-002" {
		t.Errorf("Expected model 'text-embedding-ada-002', got '%s'", serviceConfig.Model)
	}
}

func TestNewOpenAIService_MissingAPIKey(t *testing.T) {
	config := types.EmbeddingConfig{
		Provider:   "openai",
		Model:      "text-embedding-ada-002",
		Dimensions: 1536,
		APIKey:     "",
	}

	_, err := NewOpenAIService(config)
	if err == nil {
		t.Error("Expected error for missing API key, got nil")
	}

	expectedMsg := "OpenAI API key is required"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGenerateEmbedding_EmptyText(t *testing.T) {
	config := types.EmbeddingConfig{
		Provider:   "openai",
		Model:      "text-embedding-ada-002",
		Dimensions: 1536,
		APIKey:     "test-api-key",
	}

	service, err := NewOpenAIService(config)
	if err != nil {
		t.Fatalf("Failed to create OpenAI service: %v", err)
	}

	ctx := context.Background()
	_, err = service.GenerateEmbedding(ctx, "")
	if err == nil {
		t.Error("Expected error for empty text, got nil")
	}

	expectedMsg := "text cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGenerateEmbeddings_EmptyTexts(t *testing.T) {
	config := types.EmbeddingConfig{
		Provider:   "openai",
		Model:      "text-embedding-ada-002",
		Dimensions: 1536,
		APIKey:     "test-api-key",
	}

	service, err := NewOpenAIService(config)
	if err != nil {
		t.Fatalf("Failed to create OpenAI service: %v", err)
	}

	ctx := context.Background()

	// Test empty slice
	_, err = service.GenerateEmbeddings(ctx, []string{})
	if err == nil {
		t.Error("Expected error for empty texts slice, got nil")
	}

	// Test slice with only empty strings
	_, err = service.GenerateEmbeddings(ctx, []string{"", "", ""})
	if err == nil {
		t.Error("Expected error for slice with only empty strings, got nil")
	}
}

func TestNewService_Factory(t *testing.T) {
	config := types.EmbeddingConfig{
		Provider:   "openai",
		Model:      "text-embedding-ada-002",
		Dimensions: 1536,
		APIKey:     "test-api-key",
	}

	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service via factory: %v", err)
	}

	if service == nil {
		t.Fatal("Service is nil")
	}

	// Verify it's an OpenAI service
	openaiService, ok := service.(*OpenAIService)
	if !ok {
		t.Error("Expected OpenAIService, got different type")
	}

	if openaiService.GetDimensions() != 1536 {
		t.Errorf("Expected dimensions 1536, got %d", openaiService.GetDimensions())
	}
}

func TestNewService_UnsupportedProvider(t *testing.T) {
	config := types.EmbeddingConfig{
		Provider:   "unsupported",
		Model:      "some-model",
		Dimensions: 512,
		APIKey:     "test-api-key",
	}

	_, err := NewService(config)
	if err == nil {
		t.Error("Expected error for unsupported provider, got nil")
	}

	expectedMsg := "unsupported embedding provider: unsupported"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}
