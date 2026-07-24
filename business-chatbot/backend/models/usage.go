package models

type ChatLog struct {
	ID         string `json:"id"`
	BusinessID string `json:"business_id"`
	SessionID  string `json:"session_id"`
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	ModelKey   string `json:"model_key"`
	TokensUsed int    `json:"tokens_used"`
	CreatedAt  int64  `json:"created_at"`
}

type UsageSummary struct {
	TotalMessages int          `json:"total_messages"`
	TotalTokens   int          `json:"total_tokens"`
	TodayTokens   int          `json:"today_tokens"`
	TodayMessages int          `json:"today_messages"`
	ByModel       []ModelUsage `json:"by_model"`
}

type ModelUsage struct {
	ModelKey string `json:"model_key"`
	Tokens   int    `json:"tokens"`
	Messages int    `json:"messages"`
}
