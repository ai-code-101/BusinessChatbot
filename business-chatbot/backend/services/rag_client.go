package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"business-ai-agent/models"
)

func ragBaseURL() string {
	url := os.Getenv("RAG_SERVICE_URL")
	if url == "" {
		url = "http://localhost:8000"
	}
	return url
}

var httpClient = &http.Client{Timeout: 120 * time.Second}

// IngestResult mirrors the RAG service's response after chunking + embedding
// a piece of content.
type IngestResult struct {
	ChunkCount int `json:"chunk_count"`
}

// IngestDocument sends raw text to the Python RAG service to be chunked,
// embedded, and stored in that business's vector collection.
func IngestDocument(businessID, docID, title, content string) (*IngestResult, error) {
	payload, _ := json.Marshal(map[string]string{
		"business_id": businessID,
		"doc_id":      docID,
		"title":       title,
		"content":     content,
	})

	resp, err := httpClient.Post(ragBaseURL()+"/ingest", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rag service ingest failed (%d): %s", resp.StatusCode, string(body))
	}

	var result IngestResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryRAG asks the Python RAG service to answer a customer question using
// only that business's ingested knowledge.
func QueryRAG(req models.AskRequest) (*models.AskResponse, error) {
	payload, _ := json.Marshal(req)

	resp, err := httpClient.Post(ragBaseURL()+"/query", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rag service query failed (%d): %s", resp.StatusCode, string(body))
	}

	var result models.AskResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteDocumentVectors removes all chunks belonging to a document from the
// business's vector collection when an admin deletes it.
func DeleteDocumentVectors(businessID, docID string) error {
	url := fmt.Sprintf("%s/documents/%s?business_id=%s", ragBaseURL(), docID, businessID)
	httpReq, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.New(string(body))
	}
	return nil
}

// FetchAvailableModels asks the RAG service which model keys it currently
// supports, so the admin dashboard's dropdown always matches what's
// actually implemented in llm.py - no need to keep two lists in sync by hand.
func FetchAvailableModels() ([]string, error) {
	resp, err := httpClient.Get(ragBaseURL() + "/models")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rag service models failed (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Models []string `json:"models"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Models, nil
}
