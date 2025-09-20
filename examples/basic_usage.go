package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-rag/internal/embedding"
	"go-rag/internal/store"
	"go-rag/internal/types"
)

func main() {
	// Example of how to use the completed Qdrant store and OpenAI embedding service
	fmt.Println("RAG System Basic Usage Example")
	fmt.Println("==============================")

	// 1. Configure embedding service
	embeddingConfig := types.EmbeddingConfig{
		Provider:   "openai",
		Model:      "text-embedding-ada-002",
		Dimensions: 1536,
		APIKey:     "your-openai-api-key-here", // Replace with actual API key
	}

	// 2. Configure vector store
	vectorStoreConfig := types.VectorStoreConfig{
		Provider:       "qdrant",
		Host:           "localhost",
		Port:           6333,
		CollectionName: "documents",
		APIKey:         "", // Optional for local Qdrant
	}

	// 3. Create embedding service
	embeddingService, err := embedding.NewService(embeddingConfig)
	if err != nil {
		log.Fatalf("Failed to create embedding service: %v", err)
	}

	fmt.Printf("✓ Created embedding service: %s\n", embeddingConfig.Provider)

	// 4. Create vector store
	vectorStore, err := store.NewQdrantStore(vectorStoreConfig, embeddingService)
	if err != nil {
		log.Fatalf("Failed to create vector store: %v", err)
	}

	fmt.Printf("✓ Created vector store: %s\n", vectorStoreConfig.Provider)

	ctx := context.Background()

	// 5. Create collection (optional - will be created automatically if it doesn't exist)
	err = vectorStore.CreateCollection(ctx, embeddingService.GetDimensions())
	if err != nil {
		log.Printf("Warning: Failed to create collection (may already exist): %v", err)
	} else {
		fmt.Printf("✓ Created/verified collection: %s\n", vectorStoreConfig.CollectionName)
	}

	// 6. Health check
	err = vectorStore.HealthCheck(ctx)
	if err != nil {
		log.Printf("Warning: Health check failed: %v", err)
	} else {
		fmt.Println("✓ Vector store health check passed")
	}

	// 7. Example document chunks
	chunks := []types.DocumentChunk{
		{
			ID:         types.GenerateChunkID("doc-1", 0),
			DocumentID: "doc-1",
			Content:    "Artificial intelligence is a branch of computer science that aims to create intelligent machines.",
			ChunkIndex: 0,
			Metadata: types.Metadata{
				Title:       "Introduction to AI",
				Author:      "AI Researcher",
				Source:      "AI Textbook",
				Language:    "en",
				ContentType: "text",
				Tags:        []string{"ai", "computer-science", "technology"},
				Custom: map[string]string{
					"chapter": "1",
					"section": "introduction",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:         types.GenerateChunkID("doc-1", 1),
			DocumentID: "doc-1",
			Content:    "Machine learning is a subset of AI that enables computers to learn and improve from experience.",
			ChunkIndex: 1,
			Metadata: types.Metadata{
				Title:       "Introduction to AI",
				Author:      "AI Researcher",
				Source:      "AI Textbook",
				Language:    "en",
				ContentType: "text",
				Tags:        []string{"ai", "machine-learning", "technology"},
				Custom: map[string]string{
					"chapter": "1",
					"section": "machine-learning",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// 8. Store chunks (this will generate embeddings and store them)
	fmt.Println("\n📝 Storing document chunks...")
	err = vectorStore.StoreChunks(ctx, chunks)
	if err != nil {
		log.Printf("Warning: Failed to store chunks: %v", err)
	} else {
		fmt.Printf("✓ Stored %d chunks successfully\n", len(chunks))
	}

	// 9. Search for similar content
	fmt.Println("\n🔍 Searching for similar content...")
	query := "What is machine learning?"
	results, err := vectorStore.SearchSimilar(ctx, query, 5)
	if err != nil {
		log.Printf("Warning: Search failed: %v", err)
	} else {
		fmt.Printf("✓ Found %d similar chunks for query: '%s'\n", len(results), query)
		for i, chunk := range results {
			fmt.Printf("  %d. %s (Document: %s)\n", i+1, chunk.Content[:50]+"...", chunk.DocumentID)
		}
	}

	// 10. Retrieve chunks by document ID
	fmt.Println("\n📄 Retrieving chunks by document ID...")
	docChunks, err := vectorStore.GetChunksByDocumentID(ctx, "doc-1")
	if err != nil {
		log.Printf("Warning: Failed to retrieve chunks: %v", err)
	} else {
		fmt.Printf("✓ Retrieved %d chunks for document 'doc-1'\n", len(docChunks))
	}

	// 11. Get specific chunk by ID
	fmt.Println("\n📋 Retrieving specific chunk...")
	chunkID := types.GenerateChunkID("doc-1", 0)
	chunk, err := vectorStore.GetChunkByID(ctx, chunkID)
	if err != nil {
		log.Printf("Warning: Failed to retrieve chunk: %v", err)
	} else {
		fmt.Printf("✓ Retrieved chunk: %s\n", chunk.Content[:50]+"...")
	}

	fmt.Println("\n🎉 Basic usage example completed!")
	fmt.Println("\nNext steps:")
	fmt.Println("- Start Qdrant server: docker-compose up -d")
	fmt.Println("- Set your OpenAI API key in the configuration")
	fmt.Println("- Run the server: go run cmd/server/main.go")
	fmt.Println("- Test the API endpoints with curl or Postman")
}
