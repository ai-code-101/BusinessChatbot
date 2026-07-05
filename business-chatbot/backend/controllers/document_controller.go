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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func preview(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// UploadFile handles admin uploads of a .txt file. It reads the file,
// forwards the content to the RAG service for chunking + embedding, and
// stores a metadata record in Mongo.
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

// saveDocument is shared logic: create the Mongo record first (to get an
// ID), ingest the content into the vector store, then update the chunk
// count once the RAG service confirms how many chunks were created.
func saveDocument(c *gin.Context, businessID, title, sourceType, content string) {
	doc := models.Document{
		ID:         primitive.NewObjectID(),
		BusinessID: businessID,
		Title:      title,
		SourceType: sourceType,
		Preview:    preview(content, 200),
		ChunkCount: 0,
		CreatedAt:  time.Now().Unix(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := config.Collection("documents").InsertOne(ctx, doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save document record", "code": "DB_ERROR"})
		return
	}

	result, err := services.IngestDocument(businessID, doc.ID.Hex(), title, content)
	if err != nil {
		// Roll back the Mongo record if ingestion into the vector store failed,
		// so the admin doesn't see a "ghost" document with no searchable content.
		config.Collection("documents").DeleteOne(ctx, bson.M{"_id": doc.ID})
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to process document: " + err.Error(), "code": "RAG_FAILURE"})
		return
	}

	config.Collection("documents").UpdateOne(ctx, bson.M{"_id": doc.ID}, bson.M{
		"$set": bson.M{"chunk_count": result.ChunkCount},
	})
	doc.ChunkCount = result.ChunkCount

	c.JSON(http.StatusCreated, doc)
}

// ListDocuments returns every document ingested for the admin's business.
func ListDocuments(c *gin.Context) {
	businessID := c.GetString("business_id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := config.Collection("documents").Find(ctx, bson.M{"business_id": businessID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch documents", "code": "DB_ERROR"})
		return
	}
	defer cursor.Close(ctx)

	docs := []models.Document{}
	if err := cursor.All(ctx, &docs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode documents", "code": "DB_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"documents": docs})
}

// DeleteDocument removes both the Mongo record and the associated vectors
// in the RAG service, scoped to the admin's own business so one business
// can never delete another's data.
func DeleteDocument(c *gin.Context) {
	businessID := c.GetString("business_id")
	docID := c.Param("id")

	objID, err := primitive.ObjectIDFromHex(docID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id", "code": "INVALID_REQUEST"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := config.Collection("documents").DeleteOne(ctx, bson.M{"_id": objID, "business_id": businessID})
	if err != nil || res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found", "code": "NOT_FOUND"})
		return
	}

	if err := services.DeleteDocumentVectors(businessID, docID); err != nil {
		// Mongo record is already gone; log-worthy but not fatal to the request.
		c.JSON(http.StatusOK, gin.H{"warning": "document metadata deleted, but vector cleanup failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document deleted"})
}
