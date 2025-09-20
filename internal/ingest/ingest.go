package ingest

import (
	"context"
	"fmt"
	"io"
	"strings"

	"go-rag/internal/chunk"
	"go-rag/internal/store"
	"go-rag/internal/types"
)

// Service handles document ingestion
type Service struct {
	chunker chunk.Service
	store   store.VectorStore
}

// NewService creates a new ingestion service
func NewService(chunker chunk.Service, store store.VectorStore) *Service {
	return &Service{
		chunker: chunker,
		store:   store,
	}
}

// IngestDocument processes and stores a document
func (s *Service) IngestDocument(ctx context.Context, docID string, content io.Reader) error {
	// Read content
	contentBytes, err := io.ReadAll(content)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}
	
	text := string(contentBytes)
	
	// Chunk the document
	chunks, err := s.chunker.ChunkText(text)
	if err != nil {
		return fmt.Errorf("failed to chunk document: %w", err)
	}
	
	// Convert to document chunks
	var docChunks []types.DocumentChunk
	for i, chunk := range chunks {
		docChunks = append(docChunks, types.DocumentChunk{
			ID:         fmt.Sprintf("%s_chunk_%d", docID, i),
			DocumentID: docID,
			Content:    chunk,
			ChunkIndex: i,
		})
	}
	
	// Store chunks in vector database
	return s.store.StoreChunks(ctx, docChunks)
}

// IngestText processes and stores raw text
func (s *Service) IngestText(ctx context.Context, docID, text string) error {
	return s.IngestDocument(ctx, docID, strings.NewReader(text))
}

// DeleteDocument removes a document and all its chunks
func (s *Service) DeleteDocument(ctx context.Context, docID string) error {
	return s.store.DeleteDocument(ctx, docID)
}
