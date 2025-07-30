package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/portfolio-management/api-gateway/internal/config"
	"github.com/portfolio-management/api-gateway/internal/handlers"
	"github.com/portfolio-management/api-gateway/internal/middleware"
	"github.com/portfolio-management/api-gateway/internal/services"
)

// @title Portfolio Management API Gateway
// @version 1.0
// @description Microservices-based Portfolio Management System API Gateway
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg := config.Load()

	// Initialize services
	svc, err := services.NewServices(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize services", zap.Error(err))
	}
	defer svc.Close()

	// Initialize handlers
	handler := handlers.NewHandler(svc, logger)

	// Setup router
	router := setupRouter(handler, logger)

	// Setup server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func setupRouter(handler *handlers.Handler, logger *zap.Logger) *gin.Engine {
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(logger))

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Request-ID"}
	router.Use(cors.New(config))

	// Health check
	router.GET("/health", handler.HealthCheck)

	// Development endpoint to create sample data
	router.POST("/dev/sample-data", func(c *gin.Context) {
		if err := handler.CreateSampleData(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sample data"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Sample data created successfully"})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Portfolio routes
		portfolio := v1.Group("/portfolio")
		{
			portfolio.GET("/", handler.GetPortfolio)
			portfolio.GET("/summary", handler.GetPortfolioSummary)
			portfolio.GET("/performance", handler.GetPortfolioPerformance)
			portfolio.POST("/holdings", handler.AddHolding)
			portfolio.PUT("/holdings/:id", handler.UpdateHolding)
			portfolio.DELETE("/holdings/:id", handler.RemoveHolding)
		}

		// Transactions routes
		transactions := v1.Group("/transactions")
		{
			transactions.GET("/", handler.GetTransactions)
			transactions.POST("/", handler.CreateTransaction)
			transactions.GET("/:id", handler.GetTransaction)
			transactions.PUT("/:id", handler.UpdateTransaction)
			transactions.DELETE("/:id", handler.DeleteTransaction)
		}

		// Market data routes
		market := v1.Group("/market")
		{
			market.GET("/assets", handler.GetAssets)
			market.GET("/assets/:symbol", handler.GetAsset)
			market.GET("/prices/:symbol", handler.GetCurrentPrice)
			market.GET("/prices/:symbol/history", handler.GetPriceHistory)
		}

		// Analytics routes
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/performance", handler.GetPerformanceAnalytics)
			analytics.GET("/risk", handler.GetRiskMetrics)
			analytics.GET("/allocation", handler.GetAssetAllocation)
			analytics.POST("/whatif", handler.WhatIfAnalysis)
		}

		// Notifications routes
		notifications := v1.Group("/notifications")
		{
			notifications.GET("/", handler.GetNotifications)
			notifications.PUT("/:id/read", handler.MarkNotificationRead)
			notifications.POST("/settings", handler.UpdateNotificationSettings)
		}

		// WebSocket for real-time updates
		v1.GET("/ws", handler.WebSocketHandler)
	}

	// Swagger documentation
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
