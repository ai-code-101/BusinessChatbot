package models

type AdminUser struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	BusinessID   string `json:"business_id"`
	CreatedAt    int64  `json:"created_at"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token      string `json:"token"`
	BusinessID string `json:"business_id"`
	Email      string `json:"email"`
}
