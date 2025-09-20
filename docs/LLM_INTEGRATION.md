# LLM Integration Implementation

This document describes the implementation of the LLM (Large Language Model) integration for the RAG system's response generation component.

## Overview

The LLM integration replaces the placeholder implementation in `internal/generate/generate.go` with actual OpenAI API calls to generate contextual responses based on retrieved document chunks.

## Implementation Details

### Service Structure

The `generate.Service` struct now includes:
- **OpenAI Client**: Uses the `github.com/sashabaranov/go-openai` library
- **Configuration**: Stores generation settings (model, temperature, max tokens, API key)

```go
type Service struct {
    client *openai.Client
    config types.GenerationConfig
}
```

### Key Features

1. **OpenAI Integration**: Direct integration with OpenAI's Chat Completion API
2. **Configuration Support**: Supports all key parameters (model, temperature, max tokens)
3. **Error Handling**: Comprehensive error handling for API failures
4. **Validation**: Input validation for prompts and configuration
5. **Source Tracking**: Extracts and deduplicates source document IDs

### Configuration

The service is configured via environment variables:

```bash
# LLM Configuration
LLM_PROVIDER=openai
LLM_MODEL=gpt-3.5-turbo
LLM_TEMPERATURE=0.7
LLM_MAX_TOKENS=1000
OPENAI_API_KEY=your_openai_api_key_here
```

### API Integration

The implementation uses OpenAI's Chat Completion API with the following request structure:

```go
req := openai.ChatCompletionRequest{
    Model: s.config.Model,
    Messages: []openai.ChatCompletionMessage{
        {
            Role:    openai.ChatMessageRoleUser,
            Content: prompt,
        },
    },
    Temperature: float32(s.config.Temperature),
    MaxTokens:   s.config.MaxTokens,
}
```

### Prompt Engineering

The system constructs prompts with:
1. **Context**: Retrieved document chunks formatted as numbered contexts
2. **Instructions**: Clear instructions for the LLM to answer based on context
3. **Query**: The user's original question

Example prompt structure:
```
Based on the following context, please answer the question. If the context doesn't contain enough information to answer the question, please say so.

Context:
Context 1: [chunk content]

Context 2: [chunk content]

Question: [user query]

Answer:
```

## Usage Examples

### Basic Usage

```go
// Configure generation service
config := types.GenerationConfig{
    Provider:    "openai",
    Model:       "gpt-3.5-turbo",
    Temperature: 0.7,
    MaxTokens:   1000,
    APIKey:      "your-api-key",
}

// Create service
service, err := generate.NewService(config)
if err != nil {
    log.Fatal(err)
}

// Generate response
response, err := service.GenerateResponse(ctx, query, chunks)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Response: %s\n", response.Response)
fmt.Printf("Sources: %v\n", response.Sources)
```

### Integration with RAG Pipeline

The service integrates seamlessly with the existing RAG pipeline:

1. **Document Retrieval**: Vector store returns relevant chunks
2. **Ranking**: Ranker scores and filters chunks
3. **Generation**: LLM service generates response from top chunks
4. **Response**: Returns generated text with source attribution

## Error Handling

The implementation includes comprehensive error handling:

- **Configuration Validation**: Checks for required API keys and supported providers
- **Input Validation**: Validates prompts and parameters
- **API Error Handling**: Wraps OpenAI API errors with context
- **Response Validation**: Ensures valid responses are returned

## Testing

Comprehensive test suite covers:
- Service creation and configuration
- Input validation
- Context building
- Prompt construction
- Source extraction
- Error scenarios

Run tests with:
```bash
go test ./internal/generate -v
```

## Configuration Validation

The system validates configuration at startup:
- Requires `OPENAI_API_KEY` when using OpenAI provider
- Validates provider support (currently OpenAI only)
- Ensures required configuration fields are present

## Future Enhancements

The implementation is designed for extensibility:

1. **Multiple Providers**: Easy to add Anthropic, Hugging Face, etc.
2. **Streaming Responses**: Framework for streaming implementation
3. **Advanced Prompting**: Support for system messages, few-shot examples
4. **Response Caching**: Integration with caching layer
5. **Rate Limiting**: Built-in rate limiting for API calls

## Files Modified

- `internal/generate/generate.go`: Main implementation
- `pkg/httpapi/router.go`: Service initialization
- `internal/config/config.go`: Configuration validation
- `internal/generate/generate_test.go`: Test suite (new)
- `examples/llm_integration_example.go`: Usage example (new)

## Dependencies

- `github.com/sashabaranov/go-openai`: OpenAI API client (already present)

## Security Considerations

- API keys are validated but not logged
- Prompts are constructed safely to prevent injection
- Error messages don't expose sensitive information
- Configuration supports environment variable injection

## Performance

- Single API call per generation request
- Configurable token limits to control costs
- Efficient context building from chunks
- Minimal memory allocation for large responses
