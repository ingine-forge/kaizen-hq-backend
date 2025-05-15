package main

import (
	"context"
	"fmt"
	"kaizen-hq/bootstrap"
	"kaizen-hq/config"
	"kaizen-hq/internal/auth"
	"kaizen-hq/internal/bot"
	"kaizen-hq/internal/client/torn"
	"kaizen-hq/internal/database"
	"kaizen-hq/internal/faction"
	"kaizen-hq/internal/permission"
	"kaizen-hq/internal/role"
	"kaizen-hq/internal/user"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"slices"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
	_ "github.com/joho/godotenv/autoload"
)

var BotID string

func RunMidnightTask(factionService *faction.Service) {
	fmt.Println("Running task at:", time.Now().UTC())
	factionService.UpdateGymEnergy()
}

func main() {
	token := config.Load().DiscordBotToken

	if token == "" {
		fmt.Println("Missing DISCORD_TOKEN env variable")
		return
	}

	bot, err := bot.NewBot(token)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	go func() {
		if err := bot.Start(); err != nil {
			log.Fatalf("Error starting bot: %v", err)
		}
	}()

	ctx := context.Background()
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
	factionRepo := faction.NewRepository(db)
	roleRepo := role.NewRepository(db)
	permissionRepo := permission.NewRepository(db)

	// Services
	userService := user.NewService(userRepo, cfg)
	authService := auth.NewService(userService, cfg)
	factionService := faction.NewService(factionRepo, cfg, tornClient)
	roleService := role.NewService(roleRepo, cfg)
	permissionService := permission.NewService(permissionRepo, cfg)

	// Handlers
	authHandler := *auth.NewHandler(authService)
	userHandler := *user.NewHandler(userService)

	// Handle admin creation
	bootstrap.SeedSystem(ctx, userService, roleService, permissionService)

	location, err := time.LoadLocation(time.UTC.String())
	fmt.Println(err)

	// Create a new scheduler
	s, err := gocron.NewScheduler(gocron.WithLocation(location))
	if err != nil {
		// Handle error
		fmt.Println("Error creating scheduler:", err)
		return
	}

	// Define the cron expression for midnight UTC
	cronExpr := "00 00 * * *" // At 00:00 UTC every day

	// Create a new job with the cron expression
	_, err = s.NewJob(
		gocron.CronJob(cronExpr, false),
		gocron.NewTask(func() {
			currentTime := time.Now().UTC()
			if currentTime.Second() != 0 {
				waitForSecond := time.Second * time.Duration(60-currentTime.Second())
				time.Sleep(waitForSecond)
			}
			RunMidnightTask(factionService)
		}),
	)
	if err != nil {
		// Handle error
		fmt.Println("Error creating job:", err)
		return
	}

	// Start the scheduler
	s.Start()

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
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create the HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("HTTP server running on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %s\n", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shut down bot
	if err := bot.Stop(); err != nil {
		log.Printf("Error stopping bot: %v", err)
	}

	// Graceful HTTP shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server forced to shutdown: %v", err)
	}

	log.Println("Server exited")

}

// CORS middleware function definition
func corsMiddleware() gin.HandlerFunc {
	// Define allowed origins as a comma-separated string
	originsString := "http://localhost:5173"
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
