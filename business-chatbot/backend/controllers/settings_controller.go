package controllers

import (
	"context"
	"net/http"
	"time"

	"business-ai-agent/config"
	"business-ai-agent/models"
	"business-ai-agent/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// GetModelSetting returns the currently active model for this business,
// plus the full list of models the RAG service knows how to use.
func GetModelSetting(c *gin.Context) {
	businessID := c.GetString("business_id")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var activeModel string
	err := config.Pool.QueryRow(ctx,
		`SELECT model_key FROM model_settings WHERE business_id=$1`,
		businessID,
	).Scan(&activeModel)
	if err != nil && err != pgx.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch model setting", "code": "DB_ERROR"})
		return
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

	_, err := config.Pool.Exec(ctx,
		`INSERT INTO model_settings (business_id, model_key) VALUES ($1, $2)
		 ON CONFLICT (business_id) DO UPDATE SET model_key = EXCLUDED.model_key`,
		businessID, req.ModelKey,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save model setting", "code": "DB_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"active_model": req.ModelKey})
}