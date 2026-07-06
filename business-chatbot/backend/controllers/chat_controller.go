package controllers

import (
	"context"
	"net/http"
	"time"

	"business-ai-agent/config"
	"business-ai-agent/models"
	"business-ai-agent/services"

	"github.com/gin-gonic/gin"
)

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

	go logChatExchange(req, answer)

	c.JSON(http.StatusOK, answer)
}

func logChatExchange(req models.AskRequest, answer *models.AskResponse) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logEntry := models.ChatLog{
		BusinessID: req.BusinessID,
		SessionID:  req.SessionID,
		Question:   req.Question,
		Answer:     answer.Answer,
		TokensUsed: answer.TokensUsed,
		CreatedAt:  time.Now().Unix(),
	}

	config.Collection("chat_logs").InsertOne(ctx, logEntry)
}