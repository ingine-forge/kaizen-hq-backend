package main

import (
	"context"
	"kaizen-hq/config"
	"kaizen-hq/internal/auth"
	"kaizen-hq/internal/client/torn"
	"kaizen-hq/internal/database"
	"kaizen-hq/internal/energy"
	"kaizen-hq/internal/faction"
	"kaizen-hq/internal/user"
	"log"
	"os"
	"strings"
	"time"

	"slices"

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

	// Clients
	tornClient := torn.NewTornClient(os.Getenv("API_KEY"))

	// Repositories
	userRepo := user.NewRepository(db)
	energyRepo := energy.NewRepository(db)
	factionRepo := faction.NewRepository(db)

	// Services
	authService := auth.NewService(userRepo, cfg)
	userService := user.NewService(userRepo, cfg)
	energyService := energy.NewService(energyRepo, cfg.TornAPI.BaseURL)
	factionService := faction.NewService(factionRepo, cfg, tornClient)

	// Handlers
	authHandler := *auth.NewHandler(authService)
	energyHandler := *energy.NewHandler(energyService)
	userHandler := *user.NewHandler(userService)

	// Run immediately on startup (for testing)
	// go func() {
	// 	if err := energyService.ProcessAllUsers(); err != nil {
	// 		log.Printf("Initial energy tracking failed: %v", err)
	// 	}
	// }()

	go func() {
		if err := factionService.UpdateGymEnergy(); err != nil {
			log.Printf("Error updating faction data: %v", err)
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

	gin.SetMode(gin.ReleaseMode)
	// Create Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(corsMiddleware())

	// Public routes
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	// Protected routes
	protected := r.Group("/")
	protected.Use(auth.AuthMiddleware(cfg))
	{
		protected.GET("/user/:tornID", userHandler.GetUserByTornID)
		protected.GET("/user/me", userHandler.GetCurrentUser)
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

// CORS middleware function definition
func corsMiddleware() gin.HandlerFunc {
	// Define allowed origins as a comma-separated string
	originsString := "http://" + config.Load().CORS.ClientDomain + ":" + config.Load().CORS.ClientPort
	var allowedOrigins []string
	if originsString != "" {
		// Split the originsString into individual origins and store them in allowedOrigins slice
		allowedOrigins = strings.Split(originsString, ",")
	}

	// Return the actual middleware handler function
	return func(c *gin.Context) {
		// Function to check if a given origin is allowed
		isOriginAllowed := func(origin string, allowedOrigins []string) bool {
			return slices.Contains(allowedOrigins, origin)
		}

		// Get the Origin header from the request
		origin := c.Request.Header.Get("Origin")

		// Check if the origin is allowed
		if isOriginAllowed(origin, allowedOrigins) {
			// If the origin is allowed, set CORS headers in the response
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE")
		}

		// Handle preflight OPTIONS requests by aborting with status 204
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// Call the next handler
		c.Next()
	}
}
