package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go-rag/internal/generate"
	"go-rag/internal/types"
)

func main() {
	fmt.Println("LLM Integration Example")
	fmt.Println("======================")

	// Check if API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️  OPENAI_API_KEY environment variable not set")
		fmt.Println("   This example will show the configuration but won't make actual API calls")
		apiKey = "demo-key-for-testing"
	}

	// Configure generation service
	config := types.GenerationConfig{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		MaxTokens:   1000,
		APIKey:      apiKey,
	}

	fmt.Printf("✓ Configuration:\n")
	fmt.Printf("  Provider: %s\n", config.Provider)
	fmt.Printf("  Model: %s\n", config.Model)
	fmt.Printf("  Temperature: %.1f\n", config.Temperature)
	fmt.Printf("  Max Tokens: %d\n", config.MaxTokens)
	fmt.Printf("  API Key: %s\n", maskAPIKey(config.APIKey))

	// Create generation service
	service, err := generate.NewService(config)
	if err != nil {
		log.Fatalf("Failed to create generation service: %v", err)
	}

	fmt.Println("\n✓ Generation service created successfully")

	// Create sample chunks (simulating retrieved documents)
	chunks := []types.RankedChunk{
		{
			DocumentChunk: types.DocumentChunk{
				ID:         types.GenerateChunkID("doc-1", 0),
				DocumentID: "doc-1",
				Content:    "Machine learning is a subset of artificial intelligence that enables computers to learn and improve from experience without being explicitly programmed.",
				ChunkIndex: 0,
			},
			Score: 0.95,
		},
		{
			DocumentChunk: types.DocumentChunk{
				ID:         types.GenerateChunkID("doc-2", 0),
				DocumentID: "doc-2",
				Content:    "Deep learning is a machine learning technique that uses neural networks with multiple layers to model and understand complex patterns in data.",
				ChunkIndex: 0,
			},
			Score: 0.87,
		},
		{
			DocumentChunk: types.DocumentChunk{
				ID:         types.GenerateChunkID("doc-1", 1),
				DocumentID: "doc-1",
				Content:    "Natural language processing (NLP) is a branch of AI that helps computers understand, interpret and manipulate human language.",
				ChunkIndex: 1,
			},
			Score: 0.82,
		},
	}

	fmt.Printf("\n📄 Sample chunks prepared (%d chunks)\n", len(chunks))
	for i, chunk := range chunks {
		fmt.Printf("  %d. %s (Score: %.2f, Doc: %s)\n", 
			i+1, 
			truncateString(chunk.Content, 60), 
			chunk.Score, 
			chunk.DocumentID)
	}

	// Test query
	query := "What is machine learning and how does it relate to AI?"
	fmt.Printf("\n❓ Query: %s\n", query)

	ctx := context.Background()

	// Test with empty chunks first
	fmt.Println("\n🧪 Testing with empty chunks...")
	emptyResponse, err := service.GenerateResponse(ctx, query, []types.RankedChunk{})
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("✓ Response: %s\n", emptyResponse.Response)
		fmt.Printf("✓ Sources: %v\n", emptyResponse.Sources)
	}

	// Test with actual chunks (only if real API key is provided)
	if os.Getenv("OPENAI_API_KEY") != "" {
		fmt.Println("\n🤖 Generating response with LLM...")
		response, err := service.GenerateResponse(ctx, query, chunks)
		if err != nil {
			log.Printf("Error generating response: %v", err)
		} else {
			fmt.Printf("✓ Generated Response:\n%s\n", response.Response)
			fmt.Printf("✓ Sources: %v\n", response.Sources)
		}

		// Test streaming response
		fmt.Println("\n📡 Testing streaming response...")
		streamChan, err := service.StreamResponse(ctx, query, chunks)
		if err != nil {
			log.Printf("Error creating stream: %v", err)
		} else {
			for response := range streamChan {
				fmt.Printf("✓ Streamed Response:\n%s\n", response)
			}
		}
	} else {
		fmt.Println("\n⏭️  Skipping actual LLM calls (no API key provided)")
		fmt.Println("   To test with real API calls, set OPENAI_API_KEY environment variable")
	}

	fmt.Println("\n🎉 LLM integration example completed!")
	fmt.Println("\nNext steps:")
	fmt.Println("- Set OPENAI_API_KEY environment variable to test with real API calls")
	fmt.Println("- Run the full server: go run cmd/server/main.go")
	fmt.Println("- Test the /api/v1/rag endpoint with curl or Postman")
	fmt.Println("- Try different models like 'gpt-4' or 'gpt-3.5-turbo-16k'")
}

// maskAPIKey masks the API key for display purposes
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
