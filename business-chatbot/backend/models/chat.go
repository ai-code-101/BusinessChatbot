package models

// AskRequest is what the (any) frontend sends when a customer asks the
// business chatbot a question.
type AskRequest struct {
	BusinessID string `json:"business_id" binding:"required"`
	Question   string `json:"question" binding:"required"`
	SessionID  string `json:"session_id"`
	ModelKey   string `json:"model_key,omitempty"` // set internally from the business's saved setting, not by the frontend
}

// AskResponse mirrors exactly what the RAG service returns, so the Go
// layer can pass it straight through to the frontend.
type AskResponse struct {
	Answer     string   `json:"answer"`
	Sources    []string `json:"sources"`
	SessionID  string   `json:"session_id"`
	TokensUsed int      `json:"tokens_used"`
	ModelKey   string   `json:"model_key"`
}
