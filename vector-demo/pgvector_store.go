package main

import (
	"context"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

// Connect to PGVector and return vectorstore
func connectToPGVector(ctx context.Context, embedder embeddings.Embedder) (vectorstores.VectorStore, error) {
	connStr := "postgres://veeam:Passw0rd999!@192.168.169.105:5432/vector_db?sslmode=disable"

	// ✅ Now create PGVector store with connection URL + embedder
	store, err := pgvector.New(ctx,
		pgvector.WithConnectionURL(connStr),
		pgvector.WithEmbedder(embedder),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create PGVector store: %w", err)
	}

	fmt.Println("✅ PGVector store successfully connected.")
	return store, nil
}

// Store documents (chunks) into PGVector
func storeDocuments(ctx context.Context, store vectorstores.VectorStore, chunks []string) error {
	// Convert to schema.Documents
	var documents []schema.Document
	for _, chunk := range chunks {
		documents = append(documents, schema.Document{PageContent: chunk})
	}

	// Add documents
	_, err := store.AddDocuments(ctx, documents)
	if err != nil {
		return fmt.Errorf("failed to store documents: %w", err)
	}

	fmt.Println("✅ Successfully stored documents in PGVector.")
	return nil
}
