package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"business-ai-agent/config"
	"business-ai-agent/models"
	"business-ai-agent/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

var onboardingTriggers = []string{
	"sign up", "get started", "onboard", "talk to sales", "contact sales",
}

func Ask(c *gin.Context) {
	var req models.AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_REQUEST"})
		return
	}

	session, err := GetOrCreateSession(req.SessionID, req.BusinessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session lookup failed", "code": "SESSION_ERROR"})
		return
	}

	switch session.OnboardingState {
	case models.OnboardingAwaitingConsent:
		if isAffirmative(strings.TrimSpace(strings.ToLower(req.Question))) {
			SetOnboardingState(req.SessionID, models.OnboardingAwaitingDetails)
			c.JSON(http.StatusOK, gin.H{
				"answer":     "Great! Please enter your name and phone number below.",
				"action":     "show_form",
				"session_id": req.SessionID,
			})
			return
		}
		SetOnboardingState(req.SessionID, models.OnboardingNone)
		c.JSON(http.StatusOK, gin.H{
			"answer":     "No problem, how can I help you today?",
			"session_id": req.SessionID,
		})
		return

	case models.OnboardingAwaitingDetails:
		c.JSON(http.StatusOK, gin.H{
			"answer":     "Whenever you're ready, just fill in the form above with your name and phone number.",
			"action":     "show_form",
			"session_id": req.SessionID,
		})
		return
	}

	if containsTrigger(strings.ToLower(req.Question)) {
		SetOnboardingState(req.SessionID, models.OnboardingAwaitingConsent)
		c.JSON(http.StatusOK, gin.H{
			"answer":     "Would you like to be onboarded? I can pass your details to our sales team.",
			"action":     "offer_onboarding",
			"session_id": req.SessionID,
		})
		return
	}

	req.ModelKey = lookupActiveModel(req.BusinessID)

	answer, err := services.QueryRAG(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "reach out to our team +254 759 422 480: " + err.Error(), "code": "RAG_FAILURE"})
		return
	}

	go logChatExchange(req, answer)

	c.JSON(http.StatusOK, answer)
}

func containsTrigger(question string) bool {
	for _, trigger := range onboardingTriggers {
		if strings.Contains(question, trigger) {
			return true
		}
	}
	return false
}

func lookupActiveModel(businessID string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var modelKey string
	err := config.Pool.QueryRow(ctx,
		`SELECT model_key FROM model_settings WHERE business_id=$1`,
		businessID,
	).Scan(&modelKey)
	if err != nil {
		if err != pgx.ErrNoRows {
			// non-fatal: fall back to empty so the RAG service uses its own default
		}
		return ""
	}
	return modelKey
}

func logChatExchange(req models.AskRequest, answer *models.AskResponse) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config.Pool.Exec(ctx,
		`INSERT INTO chat_logs (business_id, session_id, question, answer, model_key, tokens_used, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		req.BusinessID, req.SessionID, req.Question, answer.Answer, answer.ModelKey, answer.TokensUsed, time.Now().Unix(),
	)
}