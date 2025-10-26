package main

import (
	"fixapp/internal/health"
	"fixapp/pkg/logger"
	"fixapp/pkg/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "fixapp/docs"

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

	// Setup router
	mux := http.NewServeMux()

	// Register handlers
	healthHandler := health.New("1.0.0")
	healthHandler.Register(mux)

	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	// Future handlers go here:
	// userHandler := users.New(logger.Log, db)
	// userHandler.Register(mux)

	// Apply middleware (order matters!)
	handler := middleware.RequestID(
		middleware.AttachLogger(logger.Log)(
			middleware.AccessLogger(logger.Log)(mux),
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
}
