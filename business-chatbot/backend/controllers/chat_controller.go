package controllers

import (
	"net/http"

	"business-ai-agent/models"
	"business-ai-agent/services"

	"github.com/gin-gonic/gin"
)

// Ask is the endpoint any frontend (your test app today, another frontend
// later) calls when a customer asks the business chatbot a question.
// It never touches Mongo or the vector store directly - it just validates
// the request and delegates to the RAG service.
func Ask(c *gin.Context) {
	var req models.AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_REQUEST"})
		return
	}

	answer, err := services.QueryRAG(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to generate answer: " + err.Error(), "code": "RAG_FAILURE"})
		return
	}

	c.JSON(http.StatusOK, answer)
}
