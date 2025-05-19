package main

import (
	"context"
	"fmt"
	"kaizen-hq/bootstrap"
	"kaizen-hq/config"
	"kaizen-hq/internal/account"
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
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

var BotID string

func RunMidnightTask(factionService *faction.Service) {
	fmt.Println("Running task at:", time.Now().UTC())
	factionService.UpdateGymEnergy()
}

func main() {
	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load application configuration
	cfg := config.Load()
	if err := validateConfig(cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize components
	app, err := initializeApp(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.Cleanup()

	// Start all services in separate goroutines
	errChan := make(chan error, 3) // Buffer for potential errors from services
	app.StartServices(ctx, errChan)

	// Wait for shutdown signal or error
	shutdownApp(ctx, app, errChan, cancel)
}

// App holds all application components and services
type App struct {
	Bot        *bot.Bot
	DB         *pgxpool.Pool
	HTTPServer *http.Server
	Scheduler  gocron.Scheduler
}

// Cleanup handles graceful shutdown of all components
func (a *App) Cleanup() {
	// Only attempt cleanup for initialized components
	if a.DB != nil {
		a.DB.Close()
	}
}

// StartServices launches all application services in separate goroutines
func (a *App) StartServices(ctx context.Context, errChan chan error) {
	// Start Discord bot
	go func() {
		log.Println("Starting Discord bot...")
		if err := a.Bot.Start(); err != nil {
			errChan <- fmt.Errorf("discord bot error: %w", err)
		}
	}()

	// Start HTTP server
	go func() {
		log.Printf("Starting HTTP server on %s...", a.HTTPServer.Addr)
		if err := a.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Start scheduler
	go func() {
		log.Println("Starting task scheduler...")
		a.Scheduler.Start()
	}()
}

// Shutdown performs graceful shutdown of all components
func (a *App) Shutdown(ctx context.Context) {
	// Create a timeout context for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	log.Println("Shutting down services...")

	// Stop scheduler first
	if a.Scheduler != nil {
		a.Scheduler.Shutdown()
		log.Println("Scheduler stopped")
	}

	// Stop HTTP server
	if a.HTTPServer != nil {
		if err := a.HTTPServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server forced to shutdown: %v", err)
		} else {
			log.Println("HTTP server stopped gracefully")
		}
	}

	// Stop Discord bot
	if a.Bot != nil {
		if err := a.Bot.Stop(); err != nil {
			log.Printf("Error stopping bot: %v", err)
		} else {
			log.Println("Discord bot stopped gracefully")
		}
	}

	log.Println("All services shut down successfully")
}

// validateConfig ensures configuration is valid before starting
func validateConfig(cfg *config.Config) error {
	if cfg.DiscordBotToken == "" {
		return fmt.Errorf("missing Discord bot token")
	}

	// Add more validation as needed

	return nil
}

// initializeApp sets up all application components
func initializeApp(ctx context.Context, cfg *config.Config) (*App, error) {
	app := &App{}

	// Initialize database
	db, err := initializeDB(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	app.DB = db

	// Initialize repositories and services
	repos := initializeRepositories(db)
	services := initializeServices(repos, cfg)

	// Seed system data if needed
	if err := bootstrap.SeedSystem(ctx, services.Account, services.User, services.Role, services.Permission); err != nil {
		return nil, fmt.Errorf("failed to seed system data: %w", err)
	}

	// Initialize HTTP server
	server, err := initializeHTTPServer(cfg, services)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP server: %w", err)
	}
	app.HTTPServer = server

	// Initialize scheduler
	scheduler, err := initializeScheduler(services)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize scheduler: %w", err)
	}
	app.Scheduler = scheduler

	// Initialize Discord bot
	bot, err := initializeBot(cfg.DiscordBotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize bot: %w", err)
	}
	app.Bot = bot

	return app, nil
}

// initializeBot creates and configures the Discord bot
func initializeBot(token string) (*bot.Bot, error) {
	return bot.NewBot(token)
}

// initializeDB sets up the database connection
func initializeDB(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	return database.NewDB(ctx, cfg)
}

// Repositories holds all data access repositories
type Repositories struct {
	Account    *account.Repository
	User       *user.Repository
	Faction    *faction.Repository
	Role       *role.Repository
	Permission *permission.Repository
}

// initializeRepositories creates all data repositories
func initializeRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		Account:    account.NewRepository(db),
		User:       user.NewRepository(db),
		Faction:    faction.NewRepository(db),
		Role:       role.NewRepository(db),
		Permission: permission.NewRepository(db),
	}
}

// Services holds all business logic services
type Services struct {
	Account    *account.Service
	User       *user.Service
	Auth       *auth.Service
	Faction    *faction.Service
	Role       *role.Service
	Permission *permission.Service
	TornClient torn.Client
}

// initializeServices creates all business logic services
func initializeServices(repos *Repositories, cfg *config.Config) *Services {
	tornClient := torn.NewTornClient(os.Getenv("API_KEY"))

	accountService := account.NewService(repos.Account, cfg)
	userService := user.NewService(repos.User, cfg, tornClient)
	authService := auth.NewService(accountService, cfg)
	factionService := faction.NewService(repos.Faction, cfg, tornClient)
	roleService := role.NewService(repos.Role, cfg)
	permissionService := permission.NewService(repos.Permission, cfg)

	return &Services{
		Account:    accountService,
		User:       userService,
		Auth:       authService,
		Faction:    factionService,
		Role:       roleService,
		Permission: permissionService,
		TornClient: tornClient,
	}
}

// initializeHTTPServer sets up the HTTP server and routes
func initializeHTTPServer(cfg *config.Config, services *Services) (*http.Server, error) {
	// Initialize router
	router := gin.Default()

	// Apply middleware
	router.Use(corsMiddleware())

	// Create handlers
	authHandler := auth.NewHandler(services.Auth)
	accountHandler := account.NewHandler(services.Account)

	// Register routes
	registerRoutes(router, authHandler, accountHandler, cfg)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	return server, nil
}

// registerRoutes configures all API endpoints
func registerRoutes(r *gin.Engine, authHandler *auth.Handler, userHandler *account.Handler, cfg *config.Config) {
	// Public routes
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	// Protected routes
	protected := r.Group("/")
	protected.Use(auth.AuthMiddleware(cfg))
	{
		protected.GET("/user/:tornID", userHandler.GetAccountByTornID)
		// Add more protected routes here
	}
}

// initializeScheduler sets up scheduled tasks
func initializeScheduler(services *Services) (gocron.Scheduler, error) {
	// Create scheduler with UTC timezone
	location, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, fmt.Errorf("error loading timezone: %w", err)
	}

	scheduler, err := gocron.NewScheduler(gocron.WithLocation(location))
	if err != nil {
		return nil, fmt.Errorf("error creating scheduler: %w", err)
	}

	// Register midnight task
	midnightCron := "0 0 * * *" // Midnight every day
	_, err = scheduler.NewJob(
		gocron.CronJob(midnightCron, false),
		gocron.NewTask(func() {
			// Original midnight task logic
			currentTime := time.Now().UTC()
			if currentTime.Second() != 0 {
				waitForSecond := time.Second * time.Duration(60-currentTime.Second())
				time.Sleep(waitForSecond)
			}
			// Run the midnight task with faction service
			RunMidnightTask(services.Faction)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("error scheduling midnight task: %w", err)
	}

	return scheduler, nil
}

// shutdownApp handles application shutdown on signal or error
func shutdownApp(ctx context.Context, app *App, errChan chan error, cancel context.CancelFunc) {
	// Create channel for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either error, context cancellation or termination signal
	select {
	case err := <-errChan:
		log.Printf("Error occurred: %v", err)
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	case <-ctx.Done():
		log.Println("Context cancelled")
	}

	// Cancel context to notify all components
	cancel()

	// Perform graceful shutdown
	app.Shutdown(ctx)
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
