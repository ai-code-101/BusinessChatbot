package models

// ModelSetting stores which model a business's chatbot currently uses.
// Changing it takes effect immediately on the next question - no rebuild,
// no redeploy, since the RAG service reads this per-request rather than
// from a fixed env var.
type ModelSetting struct {
	BusinessID string `bson:"business_id" json:"business_id"`
	ModelKey   string `bson:"model_key" json:"model_key"`
}

type SetModelRequest struct {
	ModelKey string `json:"model_key" binding:"required"`
}
