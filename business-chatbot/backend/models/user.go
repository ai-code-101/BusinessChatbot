package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// AdminUser represents a business admin who can log in and manage
// the knowledge base (upload documents / paste text) for their business.
type AdminUser struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	BusinessID   string             `bson:"business_id" json:"business_id"`
	CreatedAt    int64              `bson:"created_at" json:"created_at"`
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
