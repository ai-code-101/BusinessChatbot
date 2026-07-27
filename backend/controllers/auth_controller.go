package controllers

import (
	"context"
	"net/http"
	"os"
	"time"

	"business-ai-agent/config"
	"business-ai-agent/middleware"
	"business-ai-agent/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Login authenticates an admin using email + password and returns a JWT
// that scopes every subsequent admin action to that admin's business_id.
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_REQUEST"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var admin models.AdminUser
	err := config.Pool.QueryRow(ctx,
		`SELECT id, email, password_hash, business_id, created_at FROM admin_users WHERE email=$1`,
		req.Email,
	).Scan(&admin.ID, &admin.Email, &admin.PasswordHash, &admin.BusinessID, &admin.CreatedAt)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password", "code": "AUTH_FAILED"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password", "code": "AUTH_FAILED"})
		return
	}

	claims := middleware.Claims{
		AdminID:    admin.ID,
		BusinessID: admin.BusinessID,
		Email:      admin.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token", "code": "SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:      signed,
		BusinessID: admin.BusinessID,
		Email:      admin.Email,
	})
}