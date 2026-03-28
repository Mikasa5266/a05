package rag

import "context"

// Embedder converts text into dense vectors through an external model provider.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorStore stores vectors and performs top-k cosine similarity search.
type VectorStore interface {
	Upsert(ctx context.Context, points []VectorPoint) error
	Search(ctx context.Context, queryVector []float32, topK int) ([]SearchResult, error)
}

// DocumentSplitter splits a source document into retrievable chunks.
type DocumentSplitter interface {
	Split(ctx context.Context, doc Document) ([]Chunk, error)
}

type VectorPoint struct {
	ID       string
	Vector   []float32
	Content  string
	Metadata map[string]string
}

type SearchResult struct {
	ID       string
	Score    float32
	Content  string
	Metadata map[string]string
}

type Document struct {
	ID       string
	Content  string
	Metadata map[string]string
}

type Chunk struct {
	ID       string
	Content  string
	Metadata map[string]string
}
