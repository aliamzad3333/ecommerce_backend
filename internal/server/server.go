package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce-backend/internal/config"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/handlers"
	"ecommerce-backend/internal/logger"
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Server represents the HTTP server
type Server struct {
	config *config.Config
	db     *database.Client
	logger *slog.Logger
	router *gin.Engine
}

// New creates a new server instance
func New() (*Server, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using system environment variables")
	}

	cfg := config.Load()
	log := logger.New()

	// Initialize database
	db, err := database.NewClient(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize JWT manager
	jwtManager := utils.NewJWTManager(&cfg.JWT)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, jwtManager)

	// Setup router
	router := setupRouter(cfg, log, authHandler, jwtManager)

	return &Server{
		config: cfg,
		db:     db,
		logger: log,
		router: router,
	}, nil
}

// setupRouter configures the HTTP router
func setupRouter(cfg *config.Config, log *slog.Logger, authHandler *handlers.AuthHandler, jwtManager *utils.JWTManager) *gin.Engine {
	// Set Gin mode
	if cfg.Server.Port == "8080" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware(log))
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now()})
	})

	// API routes
	api := router.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			protected.GET("/profile", authHandler.GetProfile)
			protected.GET("/admin/dashboard", middleware.AdminMiddleware(), authHandler.AdminDashboard)
		}
	}

	return router
}

// Start starts the HTTP server
func (s *Server) Start() error {
	srv := &http.Server{
		Addr:         ":" + s.config.Server.Port,
		Handler:      s.router,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		s.logger.Info("Starting server", "port", s.config.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	// Close database connection
	if err := s.db.Close(ctx); err != nil {
		s.logger.Error("Failed to close database connection", "error", err)
		return err
	}

	s.logger.Info("Server exited")
	return nil
}
