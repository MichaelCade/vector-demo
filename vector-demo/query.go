package main

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/vectorstores"
)

func queryVectorStore(ctx context.Context, store vectorstores.VectorStore, query string) ([]string, error) {
	// Perform similarity search
	docs, err := store.SimilaritySearch(ctx, query, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to perform similarity search: %w", err)
	}

	// Check if any results returned
	if len(docs) == 0 {
		return nil, nil // or return []string{"No results found."}, nil if you want to handle empty gracefully
	}

	// Collect all page contents as chunks
	var chunks []string
	for _, doc := range docs {
		chunks = append(chunks, doc.PageContent)
	}

	return chunks, nil
}
