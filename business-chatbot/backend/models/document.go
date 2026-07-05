package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Document is the metadata record stored in Mongo for every piece of
// knowledge ingested into a business's chatbot (uploaded file or pasted text).
// The actual text is chunked + embedded and stored in the RAG service's
// vector store; Mongo just tracks what exists and lets the admin manage it.
type Document struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BusinessID string             `bson:"business_id" json:"business_id"`
	Title      string             `bson:"title" json:"title"`
	SourceType string             `bson:"source_type" json:"source_type"` // "file" or "text"
	Preview    string             `bson:"preview" json:"preview"`         // first ~200 chars, for admin UI
	ChunkCount int                `bson:"chunk_count" json:"chunk_count"`
	CreatedAt  int64              `bson:"created_at" json:"created_at"`
}

type UploadTextRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}
