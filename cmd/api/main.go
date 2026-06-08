package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "fixapp/docs"

	"fixapp/internal/auth"
	"fixapp/internal/auth/provider"
	"fixapp/internal/auth/token"
	"fixapp/internal/catalog"
	"fixapp/internal/dispatch"
	"fixapp/internal/handyman"
	"fixapp/internal/health"
	"fixapp/internal/job"
	"fixapp/internal/lead"
	"fixapp/internal/review"
	"fixapp/internal/scoring"
	"fixapp/internal/user"
	"fixapp/internal/wallet"
	"fixapp/pkg/database"
	"fixapp/pkg/logger"
	"fixapp/pkg/middleware"

	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// @title           FixApp API
// @version         1.0
// @description     FixApp backend API server
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@fixapp.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Initialize logger
	if err := logger.Initialize(os.Getenv("LOG_LEVEL")); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Log.Info("Starting server...")

	// Connect to database
	dbConfig := database.DefaultConfig()
	db, err := database.Connect(dbConfig)
	if err != nil {
		logger.Log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close(db)
	logger.Log.Info("Connected to database")

	// Initialize repositories
	userRepo := user.NewPostgresRepository(db)

	// Initialize JWT token service
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production" // TODO: require in production
		logger.Log.Warn("JWT_SECRET not set, using insecure default")
	}
	tokenService := token.NewService(token.DefaultConfig(jwtSecret))

	// Initialize auth providers
	providerRegistry := provider.NewRegistry()

	// Register Google OAuth (if configured)
	if clientID := os.Getenv("GOOGLE_CLIENT_ID"); clientID != "" {
		googleProvider := provider.NewGoogleProvider(provider.GoogleConfig{
			ClientID:     clientID,
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		})
		providerRegistry.Register(googleProvider)
		logger.Log.Info("Registered Google OAuth provider")
	}

	// Register Facebook OAuth (if configured)
	if appID := os.Getenv("FACEBOOK_APP_ID"); appID != "" {
		facebookProvider := provider.NewFacebookProvider(provider.FacebookConfig{
			ClientID:     appID,
			ClientSecret: os.Getenv("FACEBOOK_APP_SECRET"),
			RedirectURL:  os.Getenv("FACEBOOK_REDIRECT_URL"),
		})
		providerRegistry.Register(facebookProvider)
		logger.Log.Info("Registered Facebook OAuth provider")
	}

	// Initialize repositories
	jobRepo := job.NewPostgresRepository(db)
	leadRepo := lead.NewPostgresRepository(db)
	walletRepo := wallet.NewPostgresRepository(db)
	handymanRepo := handyman.NewPostgresRepository(db)
	scoringRepo := scoring.NewPostgresRepository(db)
	catalogRepo := catalog.NewPostgresRepository(db)
	reviewRepo := review.NewPostgresRepository(db)

	// Initialize services
	userService := user.NewService(userRepo, logger.Log)
	authService := auth.NewService(
		providerRegistry,
		tokenService,
		userRepo,
		logger.Log,
		auth.DefaultConfig(),
	)
	jobService := job.NewService(jobRepo, logger.Log)
	walletService := wallet.NewService(walletRepo, logger.Log)
	leadService := lead.NewService(leadRepo, jobRepo, walletRepo, logger.Log)
	handymanService := handyman.NewService(handymanRepo, logger.Log)
	scoringService := scoring.NewService(scoringRepo, logger.Log)

	// Dispatch service: matches jobs to handymen and creates leads
	dispatchService := dispatch.NewService(handymanRepo, catalogRepo, leadRepo, scoringRepo, logger.Log)
	jobService.SetDispatcher(dispatch.NewJobDispatcher(dispatchService))

	// Review service: reviews feed into ProScore
	reviewService := review.NewService(reviewRepo, jobRepo, scoringService, logger.Log)
	scoringService.SetReviewCounter(reviewRepo)

	// Initialize JWT middleware
	jwtMiddleware := middleware.NewJWTAuth(tokenService)

	// Initialize handlers
	userHandler := user.NewHandler(userService, logger.Log)
	authHandler := auth.NewHandler(authService, logger.Log)
	jobHandler := job.NewHandler(jobService, logger.Log)
	leadHandler := lead.NewHandler(leadService, logger.Log)
	walletHandler := wallet.NewHandler(walletService, logger.Log)
	handymanHandler := handyman.NewHandler(handymanService, logger.Log)
	scoringHandler := scoring.NewHandler(scoringService, logger.Log)
	catalogHandler := catalog.NewHandler(catalogRepo, logger.Log)
	reviewHandler := review.NewHandler(reviewService, logger.Log)

	// Setup router
	mux := http.NewServeMux()

	// Register handlers
	healthHandler := health.New("1.0.0")
	healthHandler.Register(mux)
	authHandler.Register(mux)
	userHandler.Register(mux)
	jobHandler.Register(mux)
	leadHandler.Register(mux)
	walletHandler.Register(mux)
	handymanHandler.Register(mux)
	scoringHandler.Register(mux)
	catalogHandler.Register(mux)
	reviewHandler.Register(mux)

	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)
	mux.HandleFunc("GET /swagger/*", httpSwagger.WrapHandler)

	// Apply middleware (order matters!)
	// JWT middleware runs after logging, extracts user from token
	handler := middleware.RequestID(
		middleware.AttachLogger(logger.Log)(
			middleware.AccessLogger(logger.Log)(
				jwtMiddleware.Middleware(mux),
			),
		),
	)

	// Server configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Log.Info("Server listening",
			zap.String("port", port),
			zap.String("version", "1.0.0"),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed to start",
				zap.Error(err),
			)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("Server stopped")
}
