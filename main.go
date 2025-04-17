package main

import (
	"context"
	"kaizen-hq/config"
	"kaizen-hq/internal/auth"
	"kaizen-hq/internal/database"
	"kaizen-hq/internal/energy"
	"log"
	"net/http"
	"os"
	"time"

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

	// Repositories
	authRepo := auth.NewRepository(db)
	energyRepo := energy.NewRepository(db)

	// Services
	authService := auth.NewService(authRepo, cfg)
	energyService := energy.NewService(energyRepo, cfg.TornAPI.BaseURL)

	// Handlers
	authHandler := *auth.NewHandler(authService)
	energyHandler := *energy.NewHandler(energyService)

	// Run immediately on startup (for testing)
	go func() {
		if err := energyService.ProcessAllUsers(); err != nil {
			log.Printf("Initial energy tracking failed: %v", err)
		}
	}()

	// Then schedule daily runs
	go func() {
		for {
			now := time.Now().UTC()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 5, 0, time.UTC) // 00:05 UTC
			time.Sleep(time.Until(next))

			log.Println("Starting daily energy tracking...")
			if err := energyService.ProcessAllUsers(); err != nil {
				log.Printf("Daily energy tracking failed: %v", err)
			}
		}
	}()

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

		protected.GET("/energyUsage", energyHandler.GetUserEnergyByID)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on: %s", port)
	r.Run(":" + port)

}
