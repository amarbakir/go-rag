package httpapi

import (
	"net/http"
	"time"

	"go-rag/internal/chunk"
	"go-rag/internal/generate"
	"go-rag/internal/ingest"
	"go-rag/internal/ranker"
	"go-rag/internal/retriever"
	"go-rag/internal/store"
	"go-rag/internal/types"

	"github.com/gin-gonic/gin"
)

// Handler contains all the service dependencies
type Handler struct {
	ingestService    *ingest.Service
	retrieverService *retriever.Service
	rankerService    *ranker.Service
	generateService  *generate.Service
	vectorStore      store.VectorStore
}

// NewHandler creates a new HTTP handler with all dependencies
func NewHandler() *Handler {
	// Initialize services
	chunker := chunk.NewService(1000, 200)
	vectorStore, _ := store.NewQdrantStore("localhost", 6333, "documents")

	return &Handler{
		ingestService:    ingest.NewService(*chunker, vectorStore),
		retrieverService: retriever.NewService(vectorStore),
		rankerService:    ranker.NewService(),
		generateService:  generate.NewService(),
		vectorStore:      vectorStore,
	}
}

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine) {
	handler := NewHandler()

	// Health check
	router.GET("/health", handler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Document ingestion
		v1.POST("/ingest", handler.IngestDocument)
		v1.DELETE("/documents/:id", handler.DeleteDocument)

		// Search and retrieval
		v1.POST("/search", handler.SearchDocuments)
		v1.GET("/documents/:id/chunks", handler.GetDocumentChunks)
		v1.GET("/chunks/:id", handler.GetChunk)

		// RAG endpoint
		v1.POST("/rag", handler.RAGQuery)
	}
}

// HealthCheck checks the health of all services
func (h *Handler) HealthCheck(c *gin.Context) {
	response := types.HealthCheckResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services: map[string]string{
			"api":          "healthy",
			"vector_store": "healthy", // TODO: implement actual health check
		},
	}

	c.JSON(http.StatusOK, response)
}

// IngestDocument handles document ingestion requests
func (h *Handler) IngestDocument(c *gin.Context) {
	var req types.IngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_request",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	start := time.Now()

	err := h.ingestService.IngestText(c.Request.Context(), req.DocumentID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "ingestion_failed",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	response := types.IngestResponse{
		DocumentID:     req.DocumentID,
		Status:         "success",
		ProcessingTime: time.Since(start).String(),
	}

	c.JSON(http.StatusOK, response)
}

// DeleteDocument handles document deletion requests
func (h *Handler) DeleteDocument(c *gin.Context) {
	documentID := c.Param("id")

	err := h.ingestService.DeleteDocument(c.Request.Context(), documentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "deletion_failed",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted", "document_id": documentID})
}

// SearchDocuments handles search requests
func (h *Handler) SearchDocuments(c *gin.Context) {
	var req types.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_request",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Retrieve relevant chunks
	chunks, err := h.retrieverService.RetrieveRelevantChunks(c.Request.Context(), req.Query, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "search_failed",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Rank chunks
	rankedChunks, err := h.rankerService.RankChunks(c.Request.Context(), req.Query, chunks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "ranking_failed",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Apply threshold filter if specified
	if req.Threshold > 0 {
		rankedChunks = h.rankerService.FilterByThreshold(rankedChunks, req.Threshold)
	}

	response := types.SearchResponse{
		Query:   req.Query,
		Results: rankedChunks,
		Total:   len(rankedChunks),
	}

	c.JSON(http.StatusOK, response)
}

// GetDocumentChunks retrieves all chunks for a specific document
func (h *Handler) GetDocumentChunks(c *gin.Context) {
	documentID := c.Param("id")

	chunks, err := h.retrieverService.RetrieveByDocumentID(c.Request.Context(), documentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "retrieval_failed",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"document_id": documentID,
		"chunks":      chunks,
		"total":       len(chunks),
	})
}

// GetChunk retrieves a specific chunk by ID
func (h *Handler) GetChunk(c *gin.Context) {
	chunkID := c.Param("id")

	chunk, err := h.retrieverService.RetrieveChunkByID(c.Request.Context(), chunkID)
	if err != nil {
		c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error:   "chunk_not_found",
			Code:    http.StatusNotFound,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, chunk)
}

// RAGQuery handles complete RAG (Retrieve-Augment-Generate) requests
func (h *Handler) RAGQuery(c *gin.Context) {
	var req types.RAGRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_request",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	start := time.Now()

	if req.Limit <= 0 {
		req.Limit = 5 // Default for RAG
	}

	// Retrieve relevant chunks
	chunks, err := h.retrieverService.RetrieveRelevantChunks(c.Request.Context(), req.Query, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "retrieval_failed",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Rank chunks
	rankedChunks, err := h.rankerService.RankChunks(c.Request.Context(), req.Query, chunks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "ranking_failed",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Apply threshold filter if specified
	if req.Threshold > 0 {
		rankedChunks = h.rankerService.FilterByThreshold(rankedChunks, req.Threshold)
	}

	// Generate response
	generatedResponse, err := h.generateService.GenerateResponse(c.Request.Context(), req.Query, rankedChunks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "generation_failed",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	response := types.RAGResponse{
		Query:             req.Query,
		GeneratedResponse: *generatedResponse,
		RetrievedChunks:   rankedChunks,
		ProcessingTime:    time.Since(start).String(),
	}

	c.JSON(http.StatusOK, response)
}
