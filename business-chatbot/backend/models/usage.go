package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// ChatLog records one question/answer exchange, so admins can see usage
// and cost trends over time instead of only getting a number back once
// and losing it.
type ChatLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BusinessID string             `bson:"business_id" json:"business_id"`
	SessionID  string             `bson:"session_id" json:"session_id"`
	Question   string             `bson:"question" json:"question"`
	Answer     string             `bson:"answer" json:"answer"`
	ModelKey   string             `bson:"model_key" json:"model_key"`
	TokensUsed int                `bson:"tokens_used" json:"tokens_used"`
	CreatedAt  int64              `bson:"created_at" json:"created_at"`
}

// UsageSummary is the aggregated view shown at the top of the admin
// usage dashboard.
type UsageSummary struct {
	TotalMessages int             `json:"total_messages"`
	TotalTokens   int             `json:"total_tokens"`
	TodayTokens   int             `json:"today_tokens"`
	TodayMessages int             `json:"today_messages"`
	ByModel       []ModelUsage    `json:"by_model"`
}

// ModelUsage is one row in the per-model breakdown - lets an admin see,
// e.g., "claude-haiku-4-5: 8,200 tokens across 40 messages" vs
// "gpt-4.1-mini: 4,100 tokens across 15 messages" side by side.
type ModelUsage struct {
	ModelKey string `json:"model_key" bson:"_id"`
	Tokens   int    `json:"tokens" bson:"tokens"`
	Messages int    `json:"messages" bson:"messages"`
}
