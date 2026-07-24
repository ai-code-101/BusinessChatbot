package controllers

import (
	"context"
	"io"
	"net/http"
	"time"

	"business-ai-agent/config"
	"business-ai-agent/models"
	"business-ai-agent/services"

	"github.com/gin-gonic/gin"
	
)

func preview(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// UploadFile handles admin uploads of a .txt file. It reads the file,
// forwards the content to the RAG service for chunking + embedding, and
// stores a metadata record in Postgres.
func UploadFile(c *gin.Context) {
	businessID := c.GetString("business_id")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required", "code": "MISSING_FILE"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not open uploaded file", "code": "SERVER_ERROR"})
		return
	}
	defer file.Close()

	contentBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read uploaded file", "code": "SERVER_ERROR"})
		return
	}
	content := string(contentBytes)

	saveDocument(c, businessID, fileHeader.Filename, "file", content)
}

// UploadText handles admin-pasted text (no file involved).
func UploadText(c *gin.Context) {
	businessID := c.GetString("business_id")

	var req models.UploadTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_REQUEST"})
		return
	}

	saveDocument(c, businessID, req.Title, "text", req.Content)
}

// saveDocument is shared logic: create the Postgres record first (to get an
// id via RETURNING), ingest the content into the vector store, then update
// the chunk count once the RAG service confirms how many chunks were created.
func saveDocument(c *gin.Context, businessID, title, sourceType, content string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := models.Document{
		BusinessID: businessID,
		Title:      title,
		SourceType: sourceType,
		Preview:    preview(content, 200),
		ChunkCount: 0,
		CreatedAt:  time.Now().Unix(),
	}

	err := config.Pool.QueryRow(ctx,
		`INSERT INTO documents (business_id, title, source_type, preview, chunk_count, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		doc.BusinessID, doc.Title, doc.SourceType, doc.Preview, doc.ChunkCount, doc.CreatedAt,
	).Scan(&doc.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save document record", "code": "DB_ERROR"})
		return
	}

	result, err := services.IngestDocument(businessID, doc.ID, title, content)
	if err != nil {
		// Roll back the Postgres record if ingestion into the vector store failed,
		// so the admin doesn't see a "ghost" document with no searchable content.
		config.Pool.Exec(ctx, `DELETE FROM documents WHERE id=$1`, doc.ID)
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to process document: " + err.Error(), "code": "RAG_FAILURE"})
		return
	}

	config.Pool.Exec(ctx, `UPDATE documents SET chunk_count=$1 WHERE id=$2`, result.ChunkCount, doc.ID)
	doc.ChunkCount = result.ChunkCount

	c.JSON(http.StatusCreated, doc)
}

// ListDocuments returns every document ingested for the admin's business.
func ListDocuments(c *gin.Context) {
	businessID := c.GetString("business_id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := config.Pool.Query(ctx,
		`SELECT id, business_id, title, source_type, preview, chunk_count, created_at
		 FROM documents WHERE business_id=$1 ORDER BY created_at DESC`,
		businessID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch documents", "code": "DB_ERROR"})
		return
	}
	defer rows.Close()

	docs := []models.Document{}
	for rows.Next() {
		var d models.Document
		if err := rows.Scan(&d.ID, &d.BusinessID, &d.Title, &d.SourceType, &d.Preview, &d.ChunkCount, &d.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode documents", "code": "DB_ERROR"})
			return
		}
		docs = append(docs, d)
	}

	c.JSON(http.StatusOK, gin.H{"documents": docs})
}

// DeleteDocument removes both the Postgres record and the associated vectors
// in the RAG service, scoped to the admin's own business so one business
// can never delete another's data.
func DeleteDocument(c *gin.Context) {
	businessID := c.GetString("business_id")
	docID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tag, err := config.Pool.Exec(ctx,
		`DELETE FROM documents WHERE id=$1 AND business_id=$2`,
		docID, businessID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id", "code": "INVALID_REQUEST"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found", "code": "NOT_FOUND"})
		return
	}

	if err := services.DeleteDocumentVectors(businessID, docID); err != nil {
		// Postgres record is already gone; log-worthy but not fatal to the request.
		c.JSON(http.StatusOK, gin.H{"warning": "document metadata deleted, but vector cleanup failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document deleted"})
}