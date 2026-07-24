package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"business-ai-agent/config"
	"business-ai-agent/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func ragServiceURL() string {
	url := os.Getenv("RAG_SERVICE_URL")
	if url == "" {
		url = "http://localhost:8000"
	}
	return url
}

// GetOrCreateSession fetches session state, defaulting to "none".
func GetOrCreateSession(sessionID, businessID string) (*models.ChatSession, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var session models.ChatSession
	err := config.Pool.QueryRow(ctx,
		`SELECT session_id, business_id, onboarding_state FROM chat_sessions WHERE session_id=$1`,
		sessionID,
	).Scan(&session.SessionID, &session.BusinessID, &session.OnboardingState)
	if err == nil {
		return &session, nil
	}
	if err != pgx.ErrNoRows {
		return nil, err
	}

	session = models.ChatSession{
		SessionID:       sessionID,
		BusinessID:      businessID,
		OnboardingState: models.OnboardingNone,
	}
	_, err = config.Pool.Exec(ctx,
		`INSERT INTO chat_sessions (session_id, business_id, onboarding_state) VALUES ($1, $2, $3)`,
		session.SessionID, session.BusinessID, session.OnboardingState,
	)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func SetOnboardingState(sessionID string, state models.OnboardingState) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := config.Pool.Exec(ctx,
		`UPDATE chat_sessions SET onboarding_state=$1 WHERE session_id=$2`,
		state, sessionID,
	)
	return err
}

func isAffirmative(msg string) bool {
	affirmatives := map[string]bool{
		"yes": true, "yeah": true, "sure": true, "ok": true, "okay": true, "y": true, "yep": true,
	}
	return affirmatives[msg]
}

// OnboardingSubmit handles the form submission from the widget, forwards
// it to the Python rag-service to validate + send the email.
func OnboardingSubmit(c *gin.Context) {
	var req models.OnboardingSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload, _ := json.Marshal(map[string]string{
		"business_id": req.BusinessID,
		"session_id":  req.SessionID,
		"name":        req.Name,
		"phone":       req.Phone,
	})

	resp, err := http.Post(ragServiceURL()+"/onboarding/submit", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not reach rag-service"})
		return
	}
	defer resp.Body.Close()

	var ragResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&ragResult)

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, ragResult)
		return
	}

	SetOnboardingState(req.SessionID, models.OnboardingCompleted)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"reply":   "Thanks! We'll be in touch.",
	})
}