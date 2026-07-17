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
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetModelSetting returns the currently active model for this business,
// plus the full list of models the RAG service knows how to use.
func GetModelSetting(c *gin.Context) {
	businessID := c.GetString("business_id")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var setting models.ModelSetting
	err := config.Collection("model_settings").FindOne(ctx, bson.M{"business_id": businessID}).Decode(&setting)
	activeModel := ""
	if err == nil {
		activeModel = setting.ModelKey
	}

	availableModels, err := services.FetchAvailableModels()
	if err != nil {
		// Not fatal - the admin can still see/set their current choice even
		// if the RAG service is briefly unreachable.
		availableModels = []string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"active_model":     activeModel,
		"available_models": availableModels,
	})
}

// SetModelSetting lets an admin change which model their chatbot uses.
// Takes effect immediately on the next question - no rebuild required.
func SetModelSetting(c *gin.Context) {
	businessID := c.GetString("business_id")

	var req models.SetModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_REQUEST"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := config.Collection("model_settings").UpdateOne(
		ctx,
		bson.M{"business_id": businessID},
		bson.M{"$set": bson.M{"business_id": businessID, "model_key": req.ModelKey}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save model setting", "code": "DB_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"active_model": req.ModelKey})
}
