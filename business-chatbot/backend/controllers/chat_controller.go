package controllers

import (
	"context"
	"net/http"
	"time"

	"business-ai-agent/config"
	"business-ai-agent/models"
	"business-ai-agent/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// Ask is the endpoint any frontend (your test app today, another frontend
// later) calls when a customer asks the business chatbot a question.
// It never touches Mongo or the vector store directly for the RAG lookup -
// it just validates the request, looks up which model this business has
// selected, and delegates to the RAG service. It does, however, log the
// exchange to Mongo afterward so admins can see usage per model.
func Ask(c *gin.Context) {
	var req models.AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_REQUEST"})
		return
	}

	// The frontend never picks the model - it's whatever the business's
	// admin configured. Look it up here so switching models in the admin
	// panel takes effect on the very next question, no rebuild needed.
	req.ModelKey = lookupActiveModel(req.BusinessID)

	answer, err := services.QueryRAG(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to generate answer: " + err.Error(), "code": "RAG_FAILURE"})
		return
	}

	go logChatExchange(req, answer)

	c.JSON(http.StatusOK, answer)
}

// lookupActiveModel checks Mongo for this business's saved model choice.
// Returns "" if none set, which the RAG service treats as "use the default".
func lookupActiveModel(businessID string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var setting models.ModelSetting
	err := config.Collection("model_settings").FindOne(ctx, bson.M{"business_id": businessID}).Decode(&setting)
	if err != nil {
		return ""
	}
	return setting.ModelKey
}

// logChatExchange saves the Q&A pair asynchronously so a slow/failed Mongo
// write never delays or breaks the response the customer is waiting on.
func logChatExchange(req models.AskRequest, answer *models.AskResponse) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logEntry := models.ChatLog{
		BusinessID: req.BusinessID,
		SessionID:  req.SessionID,
		Question:   req.Question,
		Answer:     answer.Answer,
		ModelKey:   answer.ModelKey,
		TokensUsed: answer.TokensUsed,
		CreatedAt:  time.Now().Unix(),
	}

	config.Collection("chat_logs").InsertOne(ctx, logEntry)
}
