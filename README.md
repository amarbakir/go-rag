# RAG AI Project

A Retrieval-Augmented Generation (RAG) system built in Go, providing intelligent document search and AI-powered question answering.

## Features

- **Document Ingestion**: Process and store documents with intelligent chunking
- **Vector Search**: Fast similarity search using Qdrant vector database
- **Smart Ranking**: Advanced reranking of search results for better relevance
- **AI Generation**: Generate contextual responses using retrieved information
- **RESTful API**: Clean HTTP API for all operations
- **Configurable**: Flexible configuration for different use cases

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP API      │    │   Ingestion     │    │   Chunking      │
│   (Gin Router)  │────│   Service       │────│   Service       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       ▼                       ▼
         │              ┌─────────────────┐    ┌─────────────────┐
         │              │  Vector Store   │    │   Document      │
         │              │   (Qdrant)      │    │   Chunks        │
         │              └─────────────────┘    └─────────────────┘
         │
         ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Retrieval     │    │    Ranking      │    │   Generation    │
│   Service       │────│    Service      │────│    Service      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Project Structure

```
rag-go/
├── cmd/server/main.go              # Application entry point
├── internal/
│   ├── ingest/ingest.go           # Document ingestion logic
│   ├── retriever/retriever.go     # Document retrieval logic
│   ├── ranker/ranker.go           # Result ranking logic
│   ├── generate/generate.go       # Response generation logic
│   ├── store/qdrant.go           # Vector store implementation
│   ├── chunk/chunk.go            # Text chunking logic
│   └── types/types.go            # Shared data types
├── pkg/httpapi/router.go          # HTTP API routes and handlers
├── docker-compose.yaml           # Docker services configuration
├── .env.example                  # Environment variables template
└── README.md                     # This file
```

## Quick Start

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- Qdrant vector database

### 1. Clone and Setup

```bash
git clone <repository-url>
cd rag-go
cp .env.example .env
# Edit .env with your configuration
```

### 2. Start Dependencies

```bash
docker-compose up -d
```

This will start:
- Qdrant vector database on port 6333

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Run the Application

```bash
go run cmd/server/main.go
```

The API will be available at `http://localhost:8080`

## API Endpoints

### Health Check
```bash
GET /health
```

### Document Ingestion
```bash
POST /api/v1/ingest
Content-Type: application/json

{
  "document_id": "doc1",
  "content": "Your document content here...",
  "metadata": {
    "title": "Document Title",
    "author": "Author Name"
  }
}
```

### Search Documents
```bash
POST /api/v1/search
Content-Type: application/json

{
  "query": "What is machine learning?",
  "limit": 10,
  "threshold": 0.7
}
```

### RAG Query (Retrieve + Generate)
```bash
POST /api/v1/rag
Content-Type: application/json

{
  "query": "Explain the concept of neural networks",
  "limit": 5
}
```

### Get Document Chunks
```bash
GET /api/v1/documents/{document_id}/chunks
```

### Delete Document
```bash
DELETE /api/v1/documents/{document_id}
```

## Configuration

The application uses environment variables for configuration. Copy `.env.example` to `.env` and modify as needed:

### Key Configuration Options

- **Vector Database**: Configure Qdrant connection
- **Embedding Service**: Choose embedding provider (OpenAI, HuggingFace)
- **LLM Provider**: Configure generation service (OpenAI, Anthropic)
- **Chunking**: Adjust chunk size and overlap
- **Search**: Set default limits and thresholds

## Development

### Adding New Features

1. **New Services**: Add to `internal/` directory
2. **API Endpoints**: Add to `pkg/httpapi/router.go`
3. **Data Types**: Define in `internal/types/types.go`

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/chunk
```

### Building

```bash
# Build for current platform
go build -o bin/rag-server cmd/server/main.go

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/rag-server-linux cmd/server/main.go
```

## Docker Deployment

```bash
# Build Docker image
docker build -t rag-go .

# Run with Docker Compose
docker-compose up
```

## TODO

- [x] Implement actual Qdrant integration
- [x] Add embedding service integration (OpenAI, HuggingFace)
- [x] Implement LLM integration for generation
- [ ] Add authentication and authorization
- [ ] Implement caching layer
- [ ] Add comprehensive tests
- [ ] Add metrics and monitoring
- [ ] Implement streaming responses
- [ ] Add document format support (PDF, DOCX, etc.)
- [ ] Implement advanced chunking strategies

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
