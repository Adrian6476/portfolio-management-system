package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/portfolio-management/api-gateway/internal/services"
)

type Handler struct {
	services *services.Services
	logger   *zap.Logger
}

func NewHandler(services *services.Services, logger *zap.Logger) *Handler {
	return &Handler{
		services: services,
		logger:   logger,
	}
}

// HealthCheck checks the health of the service
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "api-gateway",
		"version": "1.0.0",
	})
}

// Portfolio handlers
func (h *Handler) GetPortfolio(c *gin.Context) {
	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolio"})
		return
	}

	// Get default user's portfolio holdings
	query := `
		SELECT
			ph.id,
			a.symbol,
			a.name,
			a.asset_type,
			ph.quantity,
			ph.average_cost,
			ph.purchase_date
		FROM portfolio_holdings ph
		JOIN assets a ON ph.asset_id = a.id
		JOIN users u ON ph.user_id = u.id
		WHERE u.username = 'default_user'
		ORDER BY ph.created_at DESC
	`

	rows, err := h.services.DB.Query(query)
	if err != nil {
		h.logger.Error("Failed to query portfolio", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolio"})
		return
	}
	defer rows.Close()

	var holdings []map[string]interface{}
	for rows.Next() {
		var id, symbol, name, assetType, purchaseDate string
		var quantity, averageCost float64

		err := rows.Scan(&id, &symbol, &name, &assetType, &quantity, &averageCost, &purchaseDate)
		if err != nil {
			h.logger.Error("Failed to scan portfolio row", zap.Error(err))
			continue
		}

		holdings = append(holdings, map[string]interface{}{
			"id":            id,
			"symbol":        symbol,
			"name":          name,
			"asset_type":    assetType,
			"quantity":      quantity,
			"average_cost":  averageCost,
			"purchase_date": purchaseDate,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"holdings":       holdings,
		"total_holdings": len(holdings),
	})
}

func (h *Handler) GetPortfolioSummary(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get portfolio summary - Coming Soon",
	})
}

func (h *Handler) GetPortfolioPerformance(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get portfolio performance - Coming Soon",
	})
}

func (h *Handler) AddHolding(c *gin.Context) {
	var request struct {
		Symbol      string  `json:"symbol" binding:"required"`
		Quantity    float64 `json:"quantity" binding:"required,gt=0"`
		AverageCost float64 `json:"average_cost" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Get user ID
	var userID string
	err := h.services.DB.QueryRow("SELECT id FROM users WHERE username = 'default_user'").Scan(&userID)
	if err != nil {
		h.logger.Error("Failed to get user ID", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Get or create asset
	var assetID string
	err = h.services.DB.QueryRow("SELECT id FROM assets WHERE symbol = $1", request.Symbol).Scan(&assetID)
	if err != nil {
		// Asset doesn't exist, create it with minimal info
		err = h.services.DB.QueryRow(`
			INSERT INTO assets (symbol, name, asset_type, currency)
			VALUES ($1, $1, 'STOCK', 'USD')
			RETURNING id
		`, request.Symbol).Scan(&assetID)
		if err != nil {
			h.logger.Error("Failed to create asset", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create asset"})
			return
		}
	}

	// Insert or update holding
	_, err = h.services.DB.Exec(`
		INSERT INTO portfolio_holdings (user_id, asset_id, quantity, average_cost)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, asset_id)
		DO UPDATE SET
			quantity = portfolio_holdings.quantity + EXCLUDED.quantity,
			average_cost = ((portfolio_holdings.quantity * portfolio_holdings.average_cost) +
							(EXCLUDED.quantity * EXCLUDED.average_cost)) /
							(portfolio_holdings.quantity + EXCLUDED.quantity),
			updated_at = NOW()
	`, userID, assetID, request.Quantity, request.AverageCost)

	if err != nil {
		h.logger.Error("Failed to add holding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add holding"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Holding added successfully",
		"symbol":       request.Symbol,
		"quantity":     request.Quantity,
		"average_cost": request.AverageCost,
	})
}

func (h *Handler) UpdateHolding(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Update holding - Coming Soon",
	})
}

func (h *Handler) RemoveHolding(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Remove holding - Coming Soon",
	})
}

// Market data handlers
func (h *Handler) GetAssets(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get assets - Coming Soon",
	})
}

func (h *Handler) GetAsset(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get asset - Coming Soon",
	})
}

func (h *Handler) GetCurrentPrice(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get current price - Coming Soon",
	})
}

func (h *Handler) GetPriceHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get price history - Coming Soon",
	})
}

// Analytics handlers
func (h *Handler) GetPerformanceAnalytics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get performance analytics - Coming Soon",
	})
}

func (h *Handler) GetRiskMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get risk metrics - Coming Soon",
	})
}

func (h *Handler) GetAssetAllocation(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get asset allocation - Coming Soon",
	})
}

func (h *Handler) WhatIfAnalysis(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "What-if analysis - Coming Soon",
	})
}

// Notification handlers
func (h *Handler) GetNotifications(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get notifications - Coming Soon",
	})
}

func (h *Handler) MarkNotificationRead(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Mark notification read - Coming Soon",
	})
}

func (h *Handler) UpdateNotificationSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Update notification settings - Coming Soon",
	})
}

// WebSocket handler
func (h *Handler) WebSocketHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "WebSocket handler - Coming Soon",
	})
}
