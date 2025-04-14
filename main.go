package main

import (
	"context"
	"fmt"
	"kaizen-hq/auth"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbConnString := os.Getenv("DATABASE_URL")

	if dbConnString == "" {
		log.Fatal("Missing DATABASE_URL environment variable")
	}

	dbpool, err := pgxpool.New(context.Background(), dbConnString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	// Verify connection
	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}

	log.Println("Connected to Supabase database!")

	// Create auth components
	store := &auth.Store{Pool: dbpool}
	service := auth.NewAuthService(store)

	// Create Gin router
	router := gin.Default()

	// Set up routes
	router.POST("/api/register", auth.RegisterHandler(service))
	router.POST("/api/login", auth.LoginHandler(service))

	// Protected routes (authentication required)
	protected := router.Group("/api")
	protected.Use(auth.AuthMiddleware())
	{
		// Example of a protected route
		protected.GET("/user", func(c *gin.Context) {
			// Get user ID from context (set by middleware)
			tornID, _ := c.Get("tornID")

			// Get user data
			user, err := service.GetUserByTornID(tornID.(int))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}

			c.JSON(http.StatusOK, user)
		})
	}

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)

	router.Run(":" + port)
}
