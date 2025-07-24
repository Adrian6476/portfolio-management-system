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
		WHERE u.username = $1
		ORDER BY ph.created_at DESC
	`

	rows, err := h.services.DB.Query(query, "default_user")
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
	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolio summary"})
		return
	}

	// Get user ID
	var userID string
	err := h.services.DB.QueryRow("SELECT id FROM users WHERE username = $1", "default_user").Scan(&userID)
	if err != nil {
		h.logger.Error("Failed to get user ID", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Get portfolio summary data
	query := `
		SELECT
			COUNT(*) as total_holdings,
			COALESCE(SUM(ph.quantity * ph.average_cost), 0) as total_cost,
			COALESCE(SUM(ph.quantity), 0) as total_shares
		FROM portfolio_holdings ph
		WHERE ph.user_id = $1
	`

	var totalHoldings int
	var totalCost, totalShares float64
	err = h.services.DB.QueryRow(query, userID).Scan(&totalHoldings, &totalCost, &totalShares)
	if err != nil {
		h.logger.Error("Failed to query portfolio summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolio summary"})
		return
	}

	// Get asset allocation by type
	allocationQuery := `
		SELECT
			a.asset_type,
			COUNT(*) as count,
			COALESCE(SUM(ph.quantity * ph.average_cost), 0) as total_value
		FROM portfolio_holdings ph
		JOIN assets a ON ph.asset_id = a.id
		WHERE ph.user_id = $1
		GROUP BY a.asset_type
		ORDER BY total_value DESC
	`

	rows, err := h.services.DB.Query(allocationQuery, userID)
	if err != nil {
		h.logger.Error("Failed to query asset allocation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolio summary"})
		return
	}
	defer rows.Close()

	var allocations []map[string]interface{}
	for rows.Next() {
		var assetType string
		var count int
		var totalValue float64

		err := rows.Scan(&assetType, &count, &totalValue)
		if err != nil {
			h.logger.Error("Failed to scan allocation row", zap.Error(err))
			continue
		}

		percentage := 0.0
		if totalCost > 0 {
			percentage = (totalValue / totalCost) * 100
		}

		allocations = append(allocations, map[string]interface{}{
			"asset_type":  assetType,
			"count":       count,
			"total_value": totalValue,
			"percentage":  percentage,
		})
	}

	// Get top holdings
	topHoldingsQuery := `
		SELECT
			a.symbol,
			a.name,
			ph.quantity,
			ph.average_cost,
			(ph.quantity * ph.average_cost) as total_value
		FROM portfolio_holdings ph
		JOIN assets a ON ph.asset_id = a.id
		WHERE ph.user_id = $1
		ORDER BY (ph.quantity * ph.average_cost) DESC
		LIMIT 5
	`

	topRows, err := h.services.DB.Query(topHoldingsQuery, userID)
	if err != nil {
		h.logger.Error("Failed to query top holdings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolio summary"})
		return
	}
	defer topRows.Close()

	var topHoldings []map[string]interface{}
	for topRows.Next() {
		var symbol, name string
		var quantity, averageCost, totalValue float64

		err := topRows.Scan(&symbol, &name, &quantity, &averageCost, &totalValue)
		if err != nil {
			h.logger.Error("Failed to scan top holding row", zap.Error(err))
			continue
		}

		topHoldings = append(topHoldings, map[string]interface{}{
			"symbol":       symbol,
			"name":         name,
			"quantity":     quantity,
			"average_cost": averageCost,
			"total_value":  totalValue,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": map[string]interface{}{
			"total_holdings": totalHoldings,
			"total_cost":     totalCost,
			"total_shares":   totalShares,
		},
		"asset_allocation": allocations,
		"top_holdings":     topHoldings,
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
	err := h.services.DB.QueryRow("SELECT id FROM users WHERE username = $1", "default_user").Scan(&userID)
	if err != nil {
		h.logger.Error("Failed to get user ID", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Get or create asset
	var assetID string
	err = h.services.DB.QueryRow("SELECT id FROM assets WHERE symbol = $1", request.Symbol).Scan(&assetID)
	if err != nil {
		// Asset doesn't exist, create it with real company data from Finnhub
		assetName := request.Symbol // fallback to symbol
		if h.services.Finnhub != nil {
			if profile, err := h.services.Finnhub.GetCompanyProfile(request.Symbol); err == nil && profile.Name != "" {
				assetName = profile.Name
				h.logger.Info("Fetched company name from Finnhub", zap.String("symbol", request.Symbol), zap.String("name", assetName))
			} else {
				h.logger.Warn("Failed to fetch company profile from Finnhub, using symbol as name", zap.String("symbol", request.Symbol), zap.Error(err))
			}
		}

		err = h.services.DB.QueryRow(`
			INSERT INTO assets (symbol, name, asset_type, currency)
			VALUES ($1, $2, 'STOCK', 'USD')
			RETURNING id
		`, request.Symbol, assetName).Scan(&assetID)
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
	holdingID := c.Param("id")
	if holdingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Holding ID is required"})
		return
	}

	var request struct {
		Quantity    *float64 `json:"quantity" binding:"omitempty,gt=0"`
		AverageCost *float64 `json:"average_cost" binding:"omitempty,gt=0"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if at least one field is provided for update
	if request.Quantity == nil && request.AverageCost == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field (quantity or average_cost) must be provided"})
		return
	}

	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update holding"})
		return
	}

	// Get user ID
	var userID string
	err := h.services.DB.QueryRow("SELECT id FROM users WHERE username = $1", "default_user").Scan(&userID)
	if err != nil {
		h.logger.Error("Failed to get user ID", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Check if holding exists and belongs to the user
	var existingQuantity, existingCost float64
	var assetSymbol string
	err = h.services.DB.QueryRow(`
		SELECT ph.quantity, ph.average_cost, a.symbol
		FROM portfolio_holdings ph
		JOIN assets a ON ph.asset_id = a.id
		WHERE ph.id = $1 AND ph.user_id = $2
	`, holdingID, userID).Scan(&existingQuantity, &existingCost, &assetSymbol)
	if err != nil {
		h.logger.Error("Failed to find holding", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Holding not found"})
		return
	}

	// Prepare update values
	newQuantity := existingQuantity
	newCost := existingCost
	if request.Quantity != nil {
		newQuantity = *request.Quantity
	}
	if request.AverageCost != nil {
		newCost = *request.AverageCost
	}

	// Update the holding
	_, err = h.services.DB.Exec(`
		UPDATE portfolio_holdings
		SET quantity = $1, average_cost = $2, updated_at = NOW()
		WHERE id = $3 AND user_id = $4
	`, newQuantity, newCost, holdingID, userID)

	if err != nil {
		h.logger.Error("Failed to update holding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update holding"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Holding updated successfully",
		"id":           holdingID,
		"symbol":       assetSymbol,
		"quantity":     newQuantity,
		"average_cost": newCost,
	})
}

func (h *Handler) RemoveHolding(c *gin.Context) {
	holdingID := c.Param("id")
	if holdingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Holding ID is required"})
		return
	}

	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove holding"})
		return
	}

	// Get user ID
	var userID string
	err := h.services.DB.QueryRow("SELECT id FROM users WHERE username = $1", "default_user").Scan(&userID)
	if err != nil {
		h.logger.Error("Failed to get user ID", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Check if holding exists and belongs to the user, and get asset symbol for response
	var assetSymbol string
	var quantity float64
	err = h.services.DB.QueryRow(`
		SELECT a.symbol, ph.quantity
		FROM portfolio_holdings ph
		JOIN assets a ON ph.asset_id = a.id
		WHERE ph.id = $1 AND ph.user_id = $2
	`, holdingID, userID).Scan(&assetSymbol, &quantity)
	if err != nil {
		h.logger.Error("Failed to find holding", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Holding not found"})
		return
	}

	// Delete the holding
	result, err := h.services.DB.Exec(`
		DELETE FROM portfolio_holdings
		WHERE id = $1 AND user_id = $2
	`, holdingID, userID)

	if err != nil {
		h.logger.Error("Failed to delete holding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove holding"})
		return
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.Error("Failed to get rows affected", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove holding"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Holding not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Holding removed successfully",
		"id":       holdingID,
		"symbol":   assetSymbol,
		"quantity": quantity,
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
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
		return
	}

	if h.services.Finnhub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Market data service not available"})
		return
	}

	quote, err := h.services.Finnhub.GetQuote(symbol)
	if err != nil {
		h.logger.Error("Failed to fetch quote from Finnhub", zap.String("symbol", symbol), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch current price"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":         symbol,
		"current_price":  quote.CurrentPrice,
		"change":         quote.Change,
		"change_percent": quote.PercentChange,
		"high":           quote.HighPriceOfDay,
		"low":            quote.LowPriceOfDay,
		"open":           quote.OpenPriceOfDay,
		"previous_close": quote.PreviousClosePrice,
		"timestamp":      quote.Timestamp,
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
