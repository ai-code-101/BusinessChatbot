package models

// Document is the metadata record stored for every piece of
// knowledge ingested into a business's chatbot (uploaded file or pasted text).
// The actual text is chunked + embedded and stored in the RAG service's
// vector store; Postgres just tracks what exists and lets the admin manage it.
type Document struct {
	ID         string `json:"id"`
	BusinessID string `json:"business_id"`
	Title      string `json:"title"`
	SourceType string `json:"source_type"` // "file" or "text"
	Preview    string `json:"preview"`
	ChunkCount int    `json:"chunk_count"`
	CreatedAt  int64  `json:"created_at"`
}

type UploadTextRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}