package models

type OnboardingState string

const (
	OnboardingNone            OnboardingState = "none"
	OnboardingAwaitingConsent OnboardingState = "awaiting_consent"
	OnboardingAwaitingDetails OnboardingState = "awaiting_details"
	OnboardingCompleted       OnboardingState = "completed"
)

type ChatSession struct {
	SessionID       string          `bson:"session_id" json:"session_id"`
	BusinessID      string          `bson:"business_id" json:"business_id"`
	OnboardingState OnboardingState `bson:"onboarding_state" json:"onboarding_state"`
}

type OnboardingSubmitRequest struct {
	SessionID  string `json:"session_id" binding:"required"`
	BusinessID string `json:"business_id"`
	Name       string `json:"name" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
}