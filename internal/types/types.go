package types

import "time"

// DocumentChunk represents a chunk of a document with metadata
type DocumentChunk struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	Content    string    `json:"content"`
	ChunkIndex int       `json:"chunk_index"`
	Metadata   Metadata  `json:"metadata,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Metadata contains additional information about a document chunk
type Metadata struct {
	Title       string            `json:"title,omitempty"`
	Author      string            `json:"author,omitempty"`
	Source      string            `json:"source,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Language    string            `json:"language,omitempty"`
	ContentType string            `json:"content_type,omitempty"`
	Custom      map[string]string `json:"custom,omitempty"`
}

// RankedChunk represents a document chunk with a relevance score
type RankedChunk struct {
	DocumentChunk
	Score float64 `json:"score"`
}

// SearchRequest represents a search query request
type SearchRequest struct {
	Query     string            `json:"query" binding:"required"`
	Limit     int               `json:"limit,omitempty"`
	Threshold float64           `json:"threshold,omitempty"`
	Filters   map[string]string `json:"filters,omitempty"`
}

// SearchResponse represents the response to a search query
type SearchResponse struct {
	Query   string        `json:"query"`
	Results []RankedChunk `json:"results"`
	Total   int           `json:"total"`
}

// GeneratedResponse represents an AI-generated response
type GeneratedResponse struct {
	Response string   `json:"response"`
	Sources  []string `json:"sources"`
}

// RAGRequest represents a complete RAG (Retrieve-Augment-Generate) request
type RAGRequest struct {
	Query     string            `json:"query" binding:"required"`
	Limit     int               `json:"limit,omitempty"`
	Threshold float64           `json:"threshold,omitempty"`
	Filters   map[string]string `json:"filters,omitempty"`
}

// RAGResponse represents the response to a RAG request
type RAGResponse struct {
	Query            string        `json:"query"`
	GeneratedResponse GeneratedResponse `json:"generated_response"`
	RetrievedChunks  []RankedChunk `json:"retrieved_chunks"`
	ProcessingTime   string        `json:"processing_time"`
}

// IngestRequest represents a document ingestion request
type IngestRequest struct {
	DocumentID string   `json:"document_id" binding:"required"`
	Content    string   `json:"content" binding:"required"`
	Metadata   Metadata `json:"metadata,omitempty"`
}

// IngestResponse represents the response to an ingestion request
type IngestResponse struct {
	DocumentID   string `json:"document_id"`
	ChunksCount  int    `json:"chunks_count"`
	Status       string `json:"status"`
	ProcessingTime string `json:"processing_time"`
}

// HealthCheckResponse represents a health check response
type HealthCheckResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Document represents a complete document
type Document struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Metadata  Metadata  `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChunkingConfig represents configuration for text chunking
type ChunkingConfig struct {
	ChunkSize    int    `json:"chunk_size"`
	ChunkOverlap int    `json:"chunk_overlap"`
	Strategy     string `json:"strategy"` // "fixed", "sentence", "paragraph"
}

// EmbeddingConfig represents configuration for embeddings
type EmbeddingConfig struct {
	Model      string `json:"model"`
	Dimensions int    `json:"dimensions"`
	Provider   string `json:"provider"` // "openai", "huggingface", etc.
}

// VectorStoreConfig represents configuration for vector storage
type VectorStoreConfig struct {
	Provider       string `json:"provider"` // "qdrant", "pinecone", "weaviate"
	Host           string `json:"host"`
	Port           int    `json:"port"`
	CollectionName string `json:"collection_name"`
	APIKey         string `json:"api_key,omitempty"`
}

// GenerationConfig represents configuration for response generation
type GenerationConfig struct {
	Provider    string  `json:"provider"` // "openai", "anthropic", "huggingface"
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	APIKey      string  `json:"api_key,omitempty"`
}
