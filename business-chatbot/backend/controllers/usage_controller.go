package controllers

import (
	"context"
	"net/http"
	"time"

	"business-ai-agent/config"
	"business-ai-agent/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// startOfTodayUnix returns midnight today (UTC) as a unix timestamp, used
// to filter "today's" usage out of the full history.
func startOfTodayUnix() int64 {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return start.Unix()
}

// GetUsageSummary returns aggregate token/message counts for the admin's
// business - both all-time and for today, so they can see cost trends
// without needing to check GitHub's billing dashboard.
func GetUsageSummary(c *gin.Context) {
	businessID := c.GetString("business_id")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := config.Collection("chat_logs")

	totalMessages, _ := collection.CountDocuments(ctx, bson.M{"business_id": businessID})
	todayMessages, _ := collection.CountDocuments(ctx, bson.M{
		"business_id": businessID,
		"created_at":  bson.M{"$gte": startOfTodayUnix()},
	})

	totalTokens := sumTokens(ctx, businessID, bson.M{"business_id": businessID})
	todayTokens := sumTokens(ctx, businessID, bson.M{
		"business_id": businessID,
		"created_at":  bson.M{"$gte": startOfTodayUnix()},
	})

	byModel := usageByModel(ctx, businessID)

	c.JSON(http.StatusOK, models.UsageSummary{
		TotalMessages: int(totalMessages),
		TotalTokens:   totalTokens,
		TodayTokens:   todayTokens,
		TodayMessages: int(todayMessages),
		ByModel:       byModel,
	})
}

// usageByModel groups all-time token/message counts by which model
// generated each answer, so an admin comparing Haiku vs GPT-4.1-mini vs
// Sonnet can see cost and volume side by side.
func usageByModel(ctx context.Context, businessID string) []models.ModelUsage {
	pipeline := []bson.M{
		{"$match": bson.M{"business_id": businessID}},
		{"$group": bson.M{
			"_id":      "$model_key",
			"tokens":   bson.M{"$sum": "$tokens_used"},
			"messages": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"tokens": -1}},
	}

	cursor, err := config.Collection("chat_logs").Aggregate(ctx, pipeline)
	if err != nil {
		return []models.ModelUsage{}
	}
	defer cursor.Close(ctx)

	results := []models.ModelUsage{}
	if err := cursor.All(ctx, &results); err != nil {
		return []models.ModelUsage{}
	}

	for i := range results {
		if results[i].ModelKey == "" {
			results[i].ModelKey = "unknown"
		}
	}
	return results
}

// sumTokens runs a Mongo aggregation pipeline to sum tokens_used across
// all matching chat_logs documents.
func sumTokens(ctx context.Context, businessID string, filter bson.M) int {
	pipeline := []bson.M{
		{"$match": filter},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$tokens_used"}}},
	}

	cursor, err := config.Collection("chat_logs").Aggregate(ctx, pipeline)
	if err != nil {
		return 0
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil || len(result) == 0 {
		return 0
	}

	total, ok := result[0]["total"].(int32)
	if ok {
		return int(total)
	}
	if totalInt64, ok := result[0]["total"].(int64); ok {
		return int(totalInt64)
	}
	return 0
}

// GetUsageLogs returns the most recent chat exchanges for this business,
// newest first, so admins can see what customers are actually asking and
// how many tokens each answer cost.
func GetUsageLogs(c *gin.Context) {
	businessID := c.GetString("business_id")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	findOptions := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(50)

	cursor, err := config.Collection("chat_logs").Find(ctx, bson.M{"business_id": businessID}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch usage logs", "code": "DB_ERROR"})
		return
	}
	defer cursor.Close(ctx)

	logs := []models.ChatLog{}
	if err := cursor.All(ctx, &logs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode usage logs", "code": "DB_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}
