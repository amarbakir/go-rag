package config

import (
	"fmt"
	"os"
	"strconv"

	"go-rag/internal/types"
	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server      ServerConfig              `json:"server"`
	VectorStore types.VectorStoreConfig   `json:"vector_store"`
	Embedding   types.EmbeddingConfig     `json:"embedding"`
	Generation  types.GenerationConfig    `json:"generation"`
	Chunking    types.ChunkingConfig      `json:"chunking"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port    int    `json:"port"`
	Host    string `json:"host"`
	GinMode string `json:"gin_mode"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()
	config := &Config{
		Server: ServerConfig{
			Port:    getEnvAsInt("PORT", 8080),
			Host:    getEnv("HOST", "localhost"),
			GinMode: getEnv("GIN_MODE", "release"),
		},
		VectorStore: types.VectorStoreConfig{
			Provider:       getEnv("QDRANT_PROVIDER", "qdrant"),
			Host:           getEnv("QDRANT_HOST", "localhost"),
			Port:           getEnvAsInt("QDRANT_PORT", 6333),
			CollectionName: getEnv("QDRANT_COLLECTION_NAME", "documents"),
			APIKey:         getEnv("QDRANT_API_KEY", ""),
		},
		Embedding: types.EmbeddingConfig{
			Provider:   getEnv("EMBEDDING_PROVIDER", "openai"),
			Model:      getEnv("EMBEDDING_MODEL", "text-embedding-ada-002"),
			Dimensions: getEnvAsInt("EMBEDDING_DIMENSIONS", 1536),
			APIKey:     getEnv("OPENAI_API_KEY", ""),
		},
		Generation: types.GenerationConfig{
			Provider:    getEnv("LLM_PROVIDER", "openai"),
			Model:       getEnv("LLM_MODEL", "gpt-3.5-turbo"),
			Temperature: getEnvAsFloat("LLM_TEMPERATURE", 0.7),
			MaxTokens:   getEnvAsInt("LLM_MAX_TOKENS", 1000),
			APIKey:      getEnv("OPENAI_API_KEY", ""),
		},
		Chunking: types.ChunkingConfig{
			ChunkSize:    getEnvAsInt("CHUNK_SIZE", 1000),
			ChunkOverlap: getEnvAsInt("CHUNK_OVERLAP", 200),
			Strategy:     getEnv("CHUNKING_STRATEGY", "fixed"),
		},
	}

	// Validate required fields
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// validateConfig ensures required configuration is present
func validateConfig(config *Config) error {
	if config.VectorStore.Host == "" {
		return fmt.Errorf("QDRANT_HOST is required")
	}
	if config.VectorStore.CollectionName == "" {
		return fmt.Errorf("QDRANT_COLLECTION_NAME is required")
	}
	if config.Embedding.Provider == "openai" && config.Embedding.APIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required when using OpenAI for embeddings")
	}
	if config.Generation.Provider == "openai" && config.Generation.APIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required when using OpenAI for generation")
	}
	return nil
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
