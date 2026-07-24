package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"business-ai-agent/config"
	"business-ai-agent/models"

	"github.com/gin-gonic/gin"
)

// startOfTodayUnix returns midnight today (UTC) as a unix timestamp, used
// to filter "today's" usage out of the full history.
func startOfTodayUnix() int64 {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return start.Unix()
}

// GetUsageSummary returns aggregate token/message counts for the admin's
// business - both all-time and for today.
func GetUsageSummary(c *gin.Context) {
	businessID := c.GetString("business_id")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	todayStart := startOfTodayUnix()

	var totalMessages, todayMessages int
	var totalTokens, todayTokens int

	config.Pool.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(SUM(tokens_used), 0) FROM chat_logs WHERE business_id=$1`,
		businessID,
	).Scan(&totalMessages, &totalTokens)

	config.Pool.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(SUM(tokens_used), 0) FROM chat_logs WHERE business_id=$1 AND created_at >= $2`,
		businessID, todayStart,
	).Scan(&todayMessages, &todayTokens)

	byModel := usageByModel(ctx, businessID)

	c.JSON(http.StatusOK, models.UsageSummary{
		TotalMessages: totalMessages,
		TotalTokens:   totalTokens,
		TodayTokens:   todayTokens,
		TodayMessages: todayMessages,
		ByModel:       byModel,
	})
}

// usageByModel groups all-time token/message counts by which model
// generated each answer.
func usageByModel(ctx context.Context, businessID string) []models.ModelUsage {
	rows, err := config.Pool.Query(ctx,
		`SELECT COALESCE(NULLIF(model_key, ''), 'unknown') AS model_key,
		        COALESCE(SUM(tokens_used), 0) AS tokens,
		        COUNT(*) AS messages
		 FROM chat_logs
		 WHERE business_id=$1
		 GROUP BY COALESCE(NULLIF(model_key, ''), 'unknown')
		 ORDER BY tokens DESC`,
		businessID,
	)
	if err != nil {
		return []models.ModelUsage{}
	}
	defer rows.Close()

	results := []models.ModelUsage{}
	for rows.Next() {
		var m models.ModelUsage
		if err := rows.Scan(&m.ModelKey, &m.Tokens, &m.Messages); err != nil {
			return []models.ModelUsage{}
		}
		results = append(results, m)
	}
	return results
}

// GetUsageLogs returns the most recent chat exchanges for this business,
// newest first.
func GetUsageLogs(c *gin.Context) {
	businessID := c.GetString("business_id")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := config.Pool.Query(ctx,
		`SELECT id, business_id, session_id, question, answer, model_key, tokens_used, created_at
		 FROM chat_logs WHERE business_id=$1 ORDER BY created_at DESC LIMIT 50`,
		businessID,
	)
	if err != nil {
		log.Printf("usage logs query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch usage logs", "code": "DB_ERROR"})
		return
	}
	defer rows.Close()

	logs := []models.ChatLog{}
	for rows.Next() {
		var l models.ChatLog
		if err := rows.Scan(&l.ID, &l.BusinessID, &l.SessionID, &l.Question, &l.Answer, &l.ModelKey, &l.TokensUsed, &l.CreatedAt); err != nil {
			log.Printf("usage logs scan error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode usage logs", "code": "DB_ERROR"})
			return
		}
		logs = append(logs, l)
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}