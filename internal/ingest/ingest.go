package ingest

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

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
func (s *Service) IngestDocument(ctx context.Context, docID string, content io.Reader) (int, error) {
	// Read content
	contentBytes, err := io.ReadAll(content)
	if err != nil {
		return 0, fmt.Errorf("failed to read content: %w", err)
	}
	
	text := string(contentBytes)

	// Chunk the document using sentence-based chunking
	chunks, err := s.chunker.ChunkBySentences(text)
	if err != nil {
		return 0, fmt.Errorf("failed to chunk document: %w", err)
	}
	
	// Convert to document chunks
	var docChunks []types.DocumentChunk
	for i, chunk := range chunks {
		docChunks = append(docChunks, types.DocumentChunk{
			ID:         types.GenerateChunkID(docID, i),
			DocumentID: docID,
			Content:    chunk,
			ChunkIndex: i,
		})
	}
	
	// Store chunks in vector database
	err = s.store.StoreChunks(ctx, docChunks)
	if err != nil {
		return 0, err
	}

	return len(chunks), nil
}

// IngestText processes and stores raw text
func (s *Service) IngestText(ctx context.Context, docID, text string) (int, error) {
	return s.IngestDocument(ctx, docID, strings.NewReader(text))
}

// DeleteDocument removes a document and all its chunks
func (s *Service) DeleteDocument(ctx context.Context, docID string) error {
	return s.store.DeleteDocument(ctx, docID)
}

// IngestDirectory processes and stores all files from a directory
func (s *Service) IngestDirectory(ctx context.Context, req types.DirectoryIngestRequest) (*types.DirectoryIngestResponse, error) {
	start := time.Now()

	// Scan directory for files
	files, err := s.scanDirectory(req.DirectoryPath, req.Recursive, req.FilePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	var successfulIngestions []types.IngestResponse
	var errors []string

	// Process each file
	for _, filePath := range files {
		result := s.processFile(ctx, filePath, req.Metadata)
		if result.Error != "" {
			errors = append(errors, fmt.Sprintf("%s: %s", result.FilePath, result.Error))
		} else {
			successfulIngestions = append(successfulIngestions, types.IngestResponse{
				DocumentID:     result.DocumentID,
				Status:         result.Status,
				ProcessingTime: "", // Individual file processing time could be added if needed
			})
		}
	}

	return &types.DirectoryIngestResponse{
		DirectoryPath:        req.DirectoryPath,
		ProcessedFiles:       len(files),
		SuccessfulIngestions: successfulIngestions,
		Errors:               errors,
		ProcessingTime:       time.Since(start).String(),
	}, nil
}

// scanDirectory scans a directory for files matching the pattern
func (s *Service) scanDirectory(dirPath string, recursive bool, pattern string) ([]string, error) {
	var files []string

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", dirPath)
	}

	// Walk through directory
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// If not recursive, skip subdirectories
			if !recursive && path != dirPath {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file pattern if specified
		if pattern != "" {
			matched, err := s.matchesPattern(filepath.Base(path), pattern)
			if err != nil {
				return fmt.Errorf("pattern matching error: %w", err)
			}
			if !matched {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return files, nil
}

// processFile processes a single file and returns the result
func (s *Service) processFile(ctx context.Context, filePath string, metadata types.Metadata) types.FileIngestResult {
	// Generate document ID from file path
	docID := s.generateDocumentID(filePath)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return types.FileIngestResult{
			FilePath:   filePath,
			DocumentID: docID,
			Status:     "failed",
			Error:      fmt.Sprintf("failed to read file: %v", err),
		}
	}

	// Check if file is empty
	if len(content) == 0 {
		return types.FileIngestResult{
			FilePath:   filePath,
			DocumentID: docID,
			Status:     "skipped",
			Error:      "file is empty",
		}
	}

	// Ingest the text content
	_, err = s.IngestText(ctx, docID, string(content))
	if err != nil {
		return types.FileIngestResult{
			FilePath:   filePath,
			DocumentID: docID,
			Status:     "failed",
			Error:      fmt.Sprintf("failed to ingest: %v", err),
		}
	}

	return types.FileIngestResult{
		FilePath:   filePath,
		DocumentID: docID,
		Status:     "success",
	}
}

// generateDocumentID creates a document ID from file path
func (s *Service) generateDocumentID(filePath string) string {
	// Use the relative path as document ID, replacing path separators with underscores
	docID := strings.ReplaceAll(filePath, string(filepath.Separator), "_")
	// Remove any leading dots or slashes
	docID = strings.TrimLeft(docID, "./")
	return docID
}

// matchesPattern checks if a filename matches the given pattern
func (s *Service) matchesPattern(filename, pattern string) (bool, error) {
	if pattern == "" {
		return true, nil
	}

	// Split pattern by comma for multiple patterns
	patterns := strings.Split(pattern, ",")
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		matched, err := filepath.Match(p, filename)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}
