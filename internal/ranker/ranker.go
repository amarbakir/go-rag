package ranker

import (
	"context"
	"sort"
	"strings"

	"go-rag/internal/types"
)

// Service handles ranking and reranking of retrieved chunks
type Service struct {
	// Add any dependencies like reranking models here
}

// NewService creates a new ranking service
func NewService() *Service {
	return &Service{}
}

// RankChunks reranks chunks based on relevance to the query
func (s *Service) RankChunks(ctx context.Context, query string, chunks []types.DocumentChunk) ([]types.RankedChunk, error) {
	var rankedChunks []types.RankedChunk
	
	for _, chunk := range chunks {
		score := s.calculateRelevanceScore(query, chunk.Content)
		rankedChunks = append(rankedChunks, types.RankedChunk{
			DocumentChunk: chunk,
			Score:         score,
		})
	}
	
	// Sort by score in descending order
	sort.Slice(rankedChunks, func(i, j int) bool {
		return rankedChunks[i].Score > rankedChunks[j].Score
	})
	
	return rankedChunks, nil
}

// calculateRelevanceScore calculates a simple relevance score
// In a real implementation, this would use a more sophisticated reranking model
func (s *Service) calculateRelevanceScore(query, content string) float64 {
	queryLower := strings.ToLower(query)
	contentLower := strings.ToLower(content)
	
	// Simple keyword matching score
	queryWords := strings.Fields(queryLower)
	score := 0.0
	
	for _, word := range queryWords {
		if strings.Contains(contentLower, word) {
			score += 1.0
		}
	}
	
	// Normalize by query length
	if len(queryWords) > 0 {
		score = score / float64(len(queryWords))
	}
	
	return score
}

// FilterByThreshold filters chunks by minimum score threshold
func (s *Service) FilterByThreshold(rankedChunks []types.RankedChunk, threshold float64) []types.RankedChunk {
	var filtered []types.RankedChunk
	
	for _, chunk := range rankedChunks {
		if chunk.Score >= threshold {
			filtered = append(filtered, chunk)
		}
	}
	
	return filtered
}

// GetTopK returns the top K ranked chunks
func (s *Service) GetTopK(rankedChunks []types.RankedChunk, k int) []types.RankedChunk {
	if k <= 0 || k >= len(rankedChunks) {
		return rankedChunks
	}
	
	return rankedChunks[:k]
}
