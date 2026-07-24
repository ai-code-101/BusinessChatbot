package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims embedded in the JWT issued at login.
type Claims struct {
	AdminID    string `json:"admin_id"`
	BusinessID string `json:"business_id"`
	Email      string `json:"email"`
	jwt.RegisteredClaims
}

// AuthRequired protects admin-only routes. It expects:
//
//	Authorization: Bearer <token>
//
// On success it sets "business_id" and "admin_id" in the Gin context so
// downstream handlers know which business the admin manages.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing or malformed Authorization header",
				"code":  "AUTH_FAILED",
			})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
				"code":  "AUTH_FAILED",
			})
			return
		}

		c.Set("business_id", claims.BusinessID)
		c.Set("admin_id", claims.AdminID)
		c.Next()
	}
}
