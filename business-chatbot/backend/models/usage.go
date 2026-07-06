package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type ChatLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BusinessID string             `bson:"business_id" json:"business_id"`
	SessionID  string             `bson:"session_id" json:"session_id"`
	Question   string             `bson:"question" json:"question"`
	Answer     string             `bson:"answer" json:"answer"`
	TokensUsed int                `bson:"tokens_used" json:"tokens_used"`
	CreatedAt  int64              `bson:"created_at" json:"created_at"`
}

type UsageSummary struct {
	TotalMessages int `json:"total_messages"`
	TotalTokens   int `json:"total_tokens"`
	TodayTokens   int `json:"today_tokens"`
	TodayMessages int `json:"today_messages"`
}