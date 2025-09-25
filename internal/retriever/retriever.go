package retriever

import (
	"context"
	"fmt"

	"go-rag/internal/store"
	"go-rag/internal/types"
)

// Service handles document retrieval
type Service struct {
	store store.VectorStore
}

// NewService creates a new retrieval service
func NewService(store store.VectorStore) *Service {
	return &Service{
		store: store,
	}
}

// RetrieveRelevantChunks finds the most relevant document chunks for a query
func (s *Service) RetrieveRelevantChunks(ctx context.Context, query string, limit int) ([]types.DocumentChunk, error) {
	if limit <= 0 {
		limit = 10 // default limit
	}
	
	chunks, err := s.store.SearchSimilar(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar chunks: %w", err)
	}
	
	return chunks, nil
}

// RetrieveByDocumentID gets all chunks for a specific document
func (s *Service) RetrieveByDocumentID(ctx context.Context, documentID string) ([]types.DocumentChunk, error) {
	chunks, err := s.store.GetChunksByDocumentID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chunks by document ID: %w", err)
	}
	
	return chunks, nil
}

// RetrieveChunkByID gets a specific chunk by its ID
func (s *Service) RetrieveChunkByID(ctx context.Context, chunkID uint64) (*types.DocumentChunk, error) {
	chunk, err := s.store.GetChunkByID(ctx, chunkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chunk by ID: %w", err)
	}

	return chunk, nil
}
