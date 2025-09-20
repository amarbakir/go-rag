package chunk

import (
	"strings"
	"unicode"
)

// Service handles text chunking operations
type Service struct {
	chunkSize    int
	chunkOverlap int
}

// NewService creates a new chunking service
func NewService(chunkSize, chunkOverlap int) *Service {
	if chunkSize <= 0 {
		chunkSize = 1000 // default chunk size
	}
	if chunkOverlap < 0 {
		chunkOverlap = 200 // default overlap
	}
	if chunkOverlap >= chunkSize {
		chunkOverlap = chunkSize / 4 // ensure overlap is less than chunk size
	}
	
	return &Service{
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
	}
}

// ChunkText splits text into overlapping chunks
func (s *Service) ChunkText(text string) ([]string, error) {
	if text == "" {
		return []string{}, nil
	}
	
	// Clean and normalize text
	text = s.cleanText(text)
	
	// If text is shorter than chunk size, return as single chunk
	if len(text) <= s.chunkSize {
		return []string{text}, nil
	}
	
	var chunks []string
	start := 0
	
	for start < len(text) {
		end := start + s.chunkSize
		
		// Don't exceed text length
		if end > len(text) {
			end = len(text)
		}
		
		// Try to break at sentence or word boundary
		if end < len(text) {
			end = s.findBestBreakPoint(text, start, end)
		}
		
		chunk := text[start:end]
		chunks = append(chunks, strings.TrimSpace(chunk))
		
		// Move start position with overlap
		start = end - s.chunkOverlap
		
		// Ensure we make progress
		if start <= 0 {
			start = end
		}
	}
	
	return chunks, nil
}

// cleanText removes excessive whitespace and normalizes text
func (s *Service) cleanText(text string) string {
	// Replace multiple whitespace with single space
	text = strings.Join(strings.Fields(text), " ")
	
	// Remove excessive newlines
	text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	
	return strings.TrimSpace(text)
}

// findBestBreakPoint finds the best place to break text (sentence > paragraph > word boundary)
func (s *Service) findBestBreakPoint(text string, start, maxEnd int) int {
	// Look for sentence endings first
	for i := maxEnd - 1; i > start+s.chunkSize/2; i-- {
		if text[i] == '.' || text[i] == '!' || text[i] == '?' {
			// Check if it's followed by whitespace (likely sentence end)
			if i+1 < len(text) && unicode.IsSpace(rune(text[i+1])) {
				return i + 1
			}
		}
	}
	
	// Look for paragraph breaks
	for i := maxEnd - 1; i > start+s.chunkSize/2; i-- {
		if text[i] == '\n' {
			return i + 1
		}
	}
	
	// Look for word boundaries
	for i := maxEnd - 1; i > start+s.chunkSize/2; i-- {
		if unicode.IsSpace(rune(text[i])) {
			return i + 1
		}
	}
	
	// If no good break point found, use max end
	return maxEnd
}

// ChunkByParagraphs splits text by paragraphs and then chunks large paragraphs
func (s *Service) ChunkByParagraphs(text string) ([]string, error) {
	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		
		if len(paragraph) <= s.chunkSize {
			chunks = append(chunks, paragraph)
		} else {
			// Chunk large paragraphs
			subChunks, err := s.ChunkText(paragraph)
			if err != nil {
				return nil, err
			}
			chunks = append(chunks, subChunks...)
		}
	}
	
	return chunks, nil
}

// ChunkBySentences splits text by sentences
func (s *Service) ChunkBySentences(text string) ([]string, error) {
	sentences := s.splitIntoSentences(text)
	var chunks []string
	var currentChunk strings.Builder
	
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}
		
		// Check if adding this sentence would exceed chunk size
		if currentChunk.Len()+len(sentence)+1 > s.chunkSize && currentChunk.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
			currentChunk.Reset()
		}
		
		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(sentence)
	}
	
	// Add the last chunk if it has content
	if currentChunk.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}
	
	return chunks, nil
}

// splitIntoSentences splits text into sentences (simple implementation)
func (s *Service) splitIntoSentences(text string) []string {
	// Simple sentence splitting - can be improved with more sophisticated NLP
	text = strings.ReplaceAll(text, ".", ".|")
	text = strings.ReplaceAll(text, "!", "!|")
	text = strings.ReplaceAll(text, "?", "?|")
	
	sentences := strings.Split(text, "|")
	var result []string
	
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" {
			result = append(result, sentence)
		}
	}
	
	return result
}
