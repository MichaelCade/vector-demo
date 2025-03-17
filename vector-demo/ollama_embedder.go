package main

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
)

type Embedder interface {
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
}

type OllamaEmbedder struct {
	embedder embeddings.Embedder
}

func createEmbedder() (Embedder, error) {
	llm, err := ollama.New(ollama.WithModel("mxbai-embed-large"))
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}

	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	return &OllamaEmbedder{
		embedder: embedder,
	}, nil
}

func (e *OllamaEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	embeddingsData, err := e.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("failed to embed documents: %w", err)
	}
	return embeddingsData, nil
}

func (e *OllamaEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	embeddingData, err := e.embedder.EmbedQuery(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}
	return embeddingData, nil
}
