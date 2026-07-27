package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"business-ai-agent/config"
	"business-ai-agent/controllers"
	"business-ai-agent/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// seedAdmin creates a default admin account on first run, using
// ADMIN_EMAIL / ADMIN_PASSWORD / ADMIN_BUSINESS_ID from the environment,
// so you have something to log in with immediately without a separate
// signup flow. Safe to call every startup - it's a no-op if the admin
// already exists.
func seedAdmin() {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	businessID := os.Getenv("ADMIN_BUSINESS_ID")
	if email == "" || password == "" {
		log.Println("ADMIN_EMAIL/ADMIN_PASSWORD not set - skipping admin seed")
		return
	}
	if businessID == "" {
		businessID = "default_business"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existingID string
	err := config.Pool.QueryRow(ctx,
		`SELECT id FROM admin_users WHERE email=$1`, email,
	).Scan(&existingID)
	if err == nil {
		// admin already exists, nothing to do
		return
	}
	if err != pgx.ErrNoRows {
		log.Printf("could not check for existing admin: %v", err)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("could not hash admin password: %v", err)
		return
	}

	_, err = config.Pool.Exec(ctx,
		`INSERT INTO admin_users (email, password_hash, business_id, created_at)
		 VALUES ($1, $2, $3, $4)`,
		email, string(hash), businessID, time.Now().Unix(),
	)
	if err != nil {
		log.Printf("could not seed admin: %v", err)
		return
	}
	log.Printf("seeded admin account: %s (business_id: %s)", email, businessID)
}

func allowedOrigins() []string {
	raw := os.Getenv("ALLOWED_ORIGINS")
	if raw == "" {
		return []string{"http://localhost:5173"}
	}
	return strings.Split(raw, ",")
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on environment variables")
	}

	config.ConnectPostgres()
	seedAdmin()

	router := gin.Default()

	origins := allowedOrigins()
	router.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			for _, o := range origins {
				if o == origin {
					return true
				}
			}
			return false
		},
		AllowHeaders: []string{"Authorization", "Content-Type"},
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := router.Group("/v1")
	{
		// Public: admin login
		v1.POST("/admin/login", controllers.Login)

		// Admin-only: manage the knowledge base
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthRequired())
		{
			admin.POST("/documents/upload", controllers.UploadFile)
			admin.POST("/documents/text", controllers.UploadText)
			admin.GET("/documents", controllers.ListDocuments)
			admin.DELETE("/documents/:id", controllers.DeleteDocument)
			admin.GET("/usage/summary", controllers.GetUsageSummary)
			admin.GET("/usage/logs", controllers.GetUsageLogs)
			admin.GET("/settings/model", controllers.GetModelSetting)
			admin.PUT("/settings/model", controllers.SetModelSetting)
		}

		// Public: the actual chatbot endpoint any frontend calls
		v1.POST("/chat/ask", controllers.Ask)

		// Public: submits captured onboarding details (name/phone), which
		// gets forwarded to the rag-service to validate and email sales
		v1.POST("/onboarding/submit", controllers.OnboardingSubmit)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("backend listening on :%s", port)
	router.Run(":" + port)
}