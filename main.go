package main

import (
	"context"
	"kaizen-hq/config"
	"kaizen-hq/internal/auth"
	"kaizen-hq/internal/database"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize the database
	db, err := database.NewDB(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize auth components
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, cfg)
	authHandler := *auth.NewHandler(authService)

	// Create Gin router
	r := gin.Default()

	// Public routes
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	// Protected routes
	protected := r.Group("/")
	protected.Use(auth.AuthMiddleware(cfg))
	{
		protected.GET("/protected", func(c *gin.Context) {
			tornID := c.MustGet("torn_id")
			c.JSON(http.StatusOK, gin.H{"message": "Hello Torn User", "torn_id": tornID})
		})
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on: %s", port)
	r.Run(":" + port)

}
