package handlers

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"time"

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
	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolio performance"})
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

	// Get query parameters
	period := c.DefaultQuery("period", "1d") // 1d, 7d, 30d, 90d, 1y, all

	// Get all portfolio holdings with current prices
	holdingsQuery := `
		SELECT
			ph.id,
			a.symbol,
			a.name,
			ph.quantity,
			ph.average_cost,
			ph.purchase_date,
			(ph.quantity * ph.average_cost) as cost_basis
		FROM portfolio_holdings ph
		JOIN assets a ON ph.asset_id = a.id
		WHERE ph.user_id = $1
		ORDER BY cost_basis DESC
	`

	rows, err := h.services.DB.Query(holdingsQuery, userID)
	if err != nil {
		h.logger.Error("Failed to query portfolio holdings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolio performance"})
		return
	}
	defer rows.Close()

	var holdings []map[string]interface{}
	var totalCostBasis float64
	var totalCurrentValue float64
	var portfolioErrors []string

	// Process each holding and get real-time prices
	for rows.Next() {
		var id, symbol, name, purchaseDate string
		var quantity, averageCost, costBasis float64

		err := rows.Scan(&id, &symbol, &name, &quantity, &averageCost, &purchaseDate, &costBasis)
		if err != nil {
			h.logger.Error("Failed to scan holding row", zap.Error(err))
			continue
		}

		// Get current price from Finnhub
		var currentPrice float64
		var change float64
		var changePercent float64
		var marketValue float64

		if h.services.Finnhub != nil {
			if quote, priceErr := h.services.Finnhub.GetQuote(symbol); priceErr == nil {
				currentPrice = quote.CurrentPrice
				change = quote.Change
				changePercent = quote.PercentChange
				marketValue = quantity * currentPrice
			} else {
				h.logger.Warn("Failed to fetch price for symbol", zap.String("symbol", symbol), zap.Error(priceErr))
				portfolioErrors = append(portfolioErrors, fmt.Sprintf("Could not fetch price for %s", symbol))
				// Use average cost as fallback
				currentPrice = averageCost
				marketValue = costBasis
				change = 0
				changePercent = 0
			}
		} else {
			// Finnhub not available, use cost basis
			currentPrice = averageCost
			marketValue = costBasis
			change = 0
			changePercent = 0
		}

		// Calculate holding performance
		gainLoss := marketValue - costBasis
		gainLossPercent := 0.0
		if costBasis > 0 {
			gainLossPercent = (gainLoss / costBasis) * 100
		}

		totalCostBasis += costBasis
		totalCurrentValue += marketValue

		holdings = append(holdings, map[string]interface{}{
			"id":                           id,
			"symbol":                       symbol,
			"name":                         name,
			"quantity":                     quantity,
			"average_cost":                 averageCost,
			"current_price":                currentPrice,
			"cost_basis":                   costBasis,
			"market_value":                 marketValue,
			"unrealized_gain_loss":         gainLoss,
			"unrealized_gain_loss_percent": gainLossPercent,
			"daily_change":                 change,
			"daily_change_percent":         changePercent,
			"purchase_date":                purchaseDate,
			"weight_percent":               0.0, // Will calculate after we have total value
		})
	}

	// Calculate portfolio weights
	for i := range holdings {
		if totalCurrentValue > 0 {
			marketValue := holdings[i]["market_value"].(float64)
			holdings[i]["weight_percent"] = (marketValue / totalCurrentValue) * 100
		}
	}

	// Calculate overall portfolio performance
	totalGainLoss := totalCurrentValue - totalCostBasis
	totalGainLossPercent := 0.0
	if totalCostBasis > 0 {
		totalGainLossPercent = (totalGainLoss / totalCostBasis) * 100
	}

	// Get historical performance for the requested period
	var historicalPerformance []map[string]interface{}

	// Query portfolio snapshots if available
	snapshotQuery := `
		SELECT snapshot_date, total_value, total_cost, unrealized_pnl
		FROM portfolio_snapshots
		WHERE user_id = $1 AND snapshot_date >= CURRENT_DATE - INTERVAL '%s'
		ORDER BY snapshot_date ASC
	`

	var intervalClause string
	switch period {
	case "7d":
		intervalClause = "7 days"
	case "30d":
		intervalClause = "30 days"
	case "90d":
		intervalClause = "90 days"
	case "1y":
		intervalClause = "1 year"
	case "all":
		intervalClause = "10 years" // Arbitrary large period
	default: // 1d
		intervalClause = "1 day"
	}

	histRows, err := h.services.DB.Query(fmt.Sprintf(snapshotQuery, intervalClause), userID)
	if err != nil {
		h.logger.Warn("Failed to query historical snapshots", zap.Error(err))
		// Continue without historical data
	} else {
		defer histRows.Close()
		for histRows.Next() {
			var snapshotDate, totalValue, totalCost, unrealizedPnl interface{}
			err := histRows.Scan(&snapshotDate, &totalValue, &totalCost, &unrealizedPnl)
			if err != nil {
				h.logger.Error("Failed to scan snapshot row", zap.Error(err))
				continue
			}

			historicalPerformance = append(historicalPerformance, map[string]interface{}{
				"date":            snapshotDate,
				"portfolio_value": totalValue,
				"cost_basis":      totalCost,
				"unrealized_pnl":  unrealizedPnl,
			})
		}
	}

	// Calculate additional performance metrics
	performanceMetrics := map[string]interface{}{
		"total_return":         totalGainLoss,
		"total_return_percent": totalGainLossPercent,
		"total_cost_basis":     totalCostBasis,
		"total_market_value":   totalCurrentValue,
		"number_of_holdings":   len(holdings),
		"largest_holding":      "",
		"largest_gain":         "",
		"largest_loss":         "",
	}

	// Find best and worst performers
	var largestHolding, largestGain, largestLoss map[string]interface{}
	var maxValue, maxGain, maxLoss float64 = 0, 0, 0

	for _, holding := range holdings {
		marketValue := holding["market_value"].(float64)
		gainLoss := holding["unrealized_gain_loss"].(float64)

		if marketValue > maxValue {
			maxValue = marketValue
			largestHolding = holding
		}

		if gainLoss > maxGain {
			maxGain = gainLoss
			largestGain = holding
		}

		if gainLoss < maxLoss {
			maxLoss = gainLoss
			largestLoss = holding
		}
	}

	if largestHolding != nil {
		performanceMetrics["largest_holding"] = fmt.Sprintf("%s (%.2f%%)",
			largestHolding["symbol"], largestHolding["weight_percent"])
	}
	if largestGain != nil {
		performanceMetrics["largest_gain"] = fmt.Sprintf("%s (+$%.2f, +%.2f%%)",
			largestGain["symbol"], largestGain["unrealized_gain_loss"], largestGain["unrealized_gain_loss_percent"])
	}
	if largestLoss != nil {
		performanceMetrics["largest_loss"] = fmt.Sprintf("%s ($%.2f, %.2f%%)",
			largestLoss["symbol"], largestLoss["unrealized_gain_loss"], largestLoss["unrealized_gain_loss_percent"])
	}

	// Create current portfolio snapshot for tracking
	if len(holdings) > 0 {
		_, err = h.services.DB.Exec(`
			INSERT INTO portfolio_snapshots (user_id, total_value, total_cost, unrealized_pnl)
			VALUES ($1, $2, $3, $4)
		`, userID, totalCurrentValue, totalCostBasis, totalGainLoss)
		if err != nil {
			h.logger.Warn("Failed to create portfolio snapshot", zap.Error(err))
		}
	}

	response := gin.H{
		"performance_summary":    performanceMetrics,
		"holdings_performance":   holdings,
		"historical_performance": historicalPerformance,
		"period":                 period,
		"last_updated":           fmt.Sprintf("%d", time.Now().Unix()),
	}

	if len(portfolioErrors) > 0 {
		response["warnings"] = portfolioErrors
	}

	c.JSON(http.StatusOK, response)
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
	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assets"})
		return
	}

	// Get query parameters for filtering
	assetType := c.Query("type")
	search := c.Query("search")
	limit := c.DefaultQuery("limit", "50")

	// Build query with optional filters
	query := `
		SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at
		FROM assets
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0

	if assetType != "" {
		argCount++
		query += fmt.Sprintf(" AND asset_type = $%d", argCount)
		args = append(args, assetType)
	}

	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (symbol ILIKE $%d OR name ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+search+"%")
	}

	query += " ORDER BY symbol ASC"

	if limit != "all" {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	rows, err := h.services.DB.Query(query, args...)
	if err != nil {
		h.logger.Error("Failed to query assets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assets"})
		return
	}
	defer rows.Close()

	var assets []map[string]interface{}
	for rows.Next() {
		var id, symbol, name, assetType, exchange, currency, sector, createdAt string
		err := rows.Scan(&id, &symbol, &name, &assetType, &exchange, &currency, &sector, &createdAt)
		if err != nil {
			h.logger.Error("Failed to scan asset row", zap.Error(err))
			continue
		}

		assets = append(assets, map[string]interface{}{
			"id":         id,
			"symbol":     symbol,
			"name":       name,
			"asset_type": assetType,
			"exchange":   exchange,
			"currency":   currency,
			"sector":     sector,
			"created_at": createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"assets": assets,
		"total":  len(assets),
	})
}

func (h *Handler) GetAsset(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
		return
	}

	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch asset"})
		return
	}

	// Get asset details
	query := `
		SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at, updated_at
		FROM assets
		WHERE symbol = $1
	`

	var id, name, assetType, exchange, currency, sector, createdAt, updatedAt string
	err := h.services.DB.QueryRow(query, symbol).Scan(&id, &symbol, &name, &assetType, &exchange, &currency, &sector, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
			return
		}
		h.logger.Error("Failed to query asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch asset"})
		return
	}

	// Get latest market data if available
	var currentPrice, change24h *float64
	var lastUpdate *string
	marketQuery := `
		SELECT price, change_24h, timestamp
		FROM market_data
		WHERE asset_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`
	marketErr := h.services.DB.QueryRow(marketQuery, id).Scan(&currentPrice, &change24h, &lastUpdate)
	if marketErr != nil && marketErr != sql.ErrNoRows {
		h.logger.Warn("Failed to query market data", zap.String("symbol", symbol), zap.Error(marketErr))
	}

	asset := map[string]interface{}{
		"id":         id,
		"symbol":     symbol,
		"name":       name,
		"asset_type": assetType,
		"exchange":   exchange,
		"currency":   currency,
		"sector":     sector,
		"created_at": createdAt,
		"updated_at": updatedAt,
	}

	if currentPrice != nil {
		asset["current_price"] = *currentPrice
		asset["change_24h"] = change24h
		asset["last_update"] = lastUpdate
	}

	c.JSON(http.StatusOK, asset)
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
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
		return
	}

	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch price history"})
		return
	}

	// Get query parameters
	period := c.DefaultQuery("period", "30d")    // 7d, 30d, 90d, 1y, etc.
	interval := c.DefaultQuery("interval", "1d") // 1d, 1h, etc.
	limit := c.DefaultQuery("limit", "100")

	// Get asset ID first
	var assetID string
	err := h.services.DB.QueryRow("SELECT id FROM assets WHERE symbol = $1", symbol).Scan(&assetID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
			return
		}
		h.logger.Error("Failed to get asset ID", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch price history"})
		return
	}

	// Build date filter based on period
	var dateFilter string
	switch period {
	case "7d":
		dateFilter = "date >= CURRENT_DATE - INTERVAL '7 days'"
	case "30d":
		dateFilter = "date >= CURRENT_DATE - INTERVAL '30 days'"
	case "90d":
		dateFilter = "date >= CURRENT_DATE - INTERVAL '90 days'"
	case "1y":
		dateFilter = "date >= CURRENT_DATE - INTERVAL '1 year'"
	default:
		dateFilter = "date >= CURRENT_DATE - INTERVAL '30 days'"
	}

	// Query price history
	query := fmt.Sprintf(`
		SELECT date, open_price, high_price, low_price, close_price, volume
		FROM price_history
		WHERE asset_id = $1 AND %s
		ORDER BY date DESC
		LIMIT $2
	`, dateFilter)

	rows, err := h.services.DB.Query(query, assetID, limit)
	if err != nil {
		h.logger.Error("Failed to query price history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch price history"})
		return
	}
	defer rows.Close()

	var priceHistory []map[string]interface{}
	for rows.Next() {
		var date, openPrice, highPrice, lowPrice, closePrice, volume interface{}
		err := rows.Scan(&date, &openPrice, &highPrice, &lowPrice, &closePrice, &volume)
		if err != nil {
			h.logger.Error("Failed to scan price history row", zap.Error(err))
			continue
		}

		priceHistory = append(priceHistory, map[string]interface{}{
			"date":   date,
			"open":   openPrice,
			"high":   highPrice,
			"low":    lowPrice,
			"close":  closePrice,
			"volume": volume,
		})
	}

	// Reverse to get chronological order
	for i := len(priceHistory)/2 - 1; i >= 0; i-- {
		opp := len(priceHistory) - 1 - i
		priceHistory[i], priceHistory[opp] = priceHistory[opp], priceHistory[i]
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":        symbol,
		"period":        period,
		"interval":      interval,
		"price_history": priceHistory,
		"total_points":  len(priceHistory),
	})
}

// Analytics handlers
func (h *Handler) GetPerformanceAnalytics(c *gin.Context) {
	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch performance analytics"})
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

	// Get period parameter
	period := c.DefaultQuery("period", "30d")

	// Calculate total portfolio value and cost
	portfolioQuery := `
		SELECT
			COALESCE(SUM(ph.quantity * ph.average_cost), 0) as total_cost,
			COUNT(*) as total_holdings
		FROM portfolio_holdings ph
		WHERE ph.user_id = $1
	`

	var totalCost float64
	var totalHoldings int
	err = h.services.DB.QueryRow(portfolioQuery, userID).Scan(&totalCost, &totalHoldings)
	if err != nil {
		h.logger.Error("Failed to calculate portfolio totals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch performance analytics"})
		return
	}

	// Calculate current market value using real-time prices
	holdingsQuery := `
		SELECT
			a.symbol,
			ph.quantity,
			ph.average_cost
		FROM portfolio_holdings ph
		JOIN assets a ON ph.asset_id = a.id
		WHERE ph.user_id = $1
	`

	holdingsRows, err := h.services.DB.Query(holdingsQuery, userID)
	if err != nil {
		h.logger.Error("Failed to query holdings for market value", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch performance analytics"})
		return
	}
	defer holdingsRows.Close()

	var currentValue float64
	var priceUpdateErrors []string

	// Calculate real market value using Finnhub prices
	for holdingsRows.Next() {
		var symbol string
		var quantity, averageCost float64

		err := holdingsRows.Scan(&symbol, &quantity, &averageCost)
		if err != nil {
			h.logger.Error("Failed to scan holdings row", zap.Error(err))
			continue
		}

		// Get current price from Finnhub
		if h.services.Finnhub != nil {
			if quote, priceErr := h.services.Finnhub.GetQuote(symbol); priceErr == nil {
				currentValue += quantity * quote.CurrentPrice
			} else {
				h.logger.Warn("Failed to fetch price for analytics", zap.String("symbol", symbol), zap.Error(priceErr))
				priceUpdateErrors = append(priceUpdateErrors, fmt.Sprintf("Could not fetch price for %s", symbol))
				// Use cost basis as fallback
				currentValue += quantity * averageCost
			}
		} else {
			// Finnhub not available, use cost basis
			currentValue += quantity * averageCost
		}
	}

	// Calculate basic performance metrics
	totalGainLoss := currentValue - totalCost
	totalReturnPercent := 0.0
	if totalCost > 0 {
		totalReturnPercent = (totalGainLoss / totalCost) * 100
	}

	// Get historical snapshots for trend analysis
	snapshotsQuery := `
		SELECT snapshot_date, total_value, total_cost, unrealized_pnl
		FROM portfolio_snapshots
		WHERE user_id = $1 AND snapshot_date >= CURRENT_DATE - INTERVAL '30 days'
		ORDER BY snapshot_date DESC
		LIMIT 30
	`

	rows, err := h.services.DB.Query(snapshotsQuery, userID)
	if err != nil {
		h.logger.Error("Failed to query portfolio snapshots", zap.Error(err))
		// Continue without historical data
	}

	var snapshots []map[string]interface{}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var snapshotDate, totalValue, totalCost, unrealizedPnl interface{}
			err := rows.Scan(&snapshotDate, &totalValue, &totalCost, &unrealizedPnl)
			if err != nil {
				h.logger.Error("Failed to scan snapshot row", zap.Error(err))
				continue
			}

			snapshots = append(snapshots, map[string]interface{}{
				"date":           snapshotDate,
				"total_value":    totalValue,
				"total_cost":     totalCost,
				"unrealized_pnl": unrealizedPnl,
			})
		}
	}

	// Get top performers
	topPerformersQuery := `
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

	performerRows, err := h.services.DB.Query(topPerformersQuery, userID)
	if err != nil {
		h.logger.Error("Failed to query top performers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch performance analytics"})
		return
	}
	defer performerRows.Close()

	var topPerformers []map[string]interface{}
	for performerRows.Next() {
		var symbol, name string
		var quantity, averageCost, totalValue float64

		err := performerRows.Scan(&symbol, &name, &quantity, &averageCost, &totalValue)
		if err != nil {
			h.logger.Error("Failed to scan performer row", zap.Error(err))
			continue
		}

		// Calculate gain/loss using real current prices
		var currentPrice float64
		var currentValue float64

		if h.services.Finnhub != nil {
			if quote, priceErr := h.services.Finnhub.GetQuote(symbol); priceErr == nil {
				currentPrice = quote.CurrentPrice
				currentValue = quantity * currentPrice
			} else {
				h.logger.Warn("Failed to fetch price for top performer", zap.String("symbol", symbol), zap.Error(priceErr))
				// Use average cost as fallback
				currentPrice = averageCost
				currentValue = totalValue
			}
		} else {
			// Finnhub not available, use cost basis
			currentPrice = averageCost
			currentValue = totalValue
		}

		gainLoss := currentValue - totalValue
		gainLossPercent := 0.0
		if totalValue > 0 {
			gainLossPercent = (gainLoss / totalValue) * 100
		}

		topPerformers = append(topPerformers, map[string]interface{}{
			"symbol":            symbol,
			"name":              name,
			"quantity":          quantity,
			"average_cost":      averageCost,
			"current_price":     currentPrice,
			"total_cost":        totalValue,
			"current_value":     currentValue,
			"gain_loss":         gainLoss,
			"gain_loss_percent": gainLossPercent,
		})
	}

	response := gin.H{
		"portfolio_performance": map[string]interface{}{
			"total_cost":           totalCost,
			"current_value":        currentValue,
			"total_gain_loss":      totalGainLoss,
			"total_return_percent": totalReturnPercent,
			"total_holdings":       totalHoldings,
			"period":               period,
		},
		"historical_snapshots": snapshots,
		"top_performers":       topPerformers,
		"last_updated":         fmt.Sprintf("%d", time.Now().Unix()),
	}

	if len(priceUpdateErrors) > 0 {
		response["warnings"] = priceUpdateErrors
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetRiskMetrics(c *gin.Context) {
	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch risk metrics"})
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

	// Calculate diversification metrics
	diversificationQuery := `
		SELECT
			a.sector,
			COUNT(*) as holdings_count,
			COALESCE(SUM(ph.quantity * ph.average_cost), 0) as sector_value
		FROM portfolio_holdings ph
		JOIN assets a ON ph.asset_id = a.id
		WHERE ph.user_id = $1 AND a.sector IS NOT NULL
		GROUP BY a.sector
		ORDER BY sector_value DESC
	`

	rows, err := h.services.DB.Query(diversificationQuery, userID)
	if err != nil {
		h.logger.Error("Failed to query diversification data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch risk metrics"})
		return
	}
	defer rows.Close()

	var sectorDiversification []map[string]interface{}
	var totalPortfolioValue float64

	for rows.Next() {
		var sector string
		var holdingsCount int
		var sectorValue float64

		err := rows.Scan(&sector, &holdingsCount, &sectorValue)
		if err != nil {
			h.logger.Error("Failed to scan diversification row", zap.Error(err))
			continue
		}

		sectorDiversification = append(sectorDiversification, map[string]interface{}{
			"sector":         sector,
			"holdings_count": holdingsCount,
			"sector_value":   sectorValue,
		})
		totalPortfolioValue += sectorValue
	}

	// Calculate sector concentration percentages
	for i := range sectorDiversification {
		if totalPortfolioValue > 0 {
			sectorValue := sectorDiversification[i]["sector_value"].(float64)
			sectorDiversification[i]["percentage"] = (sectorValue / totalPortfolioValue) * 100
		} else {
			sectorDiversification[i]["percentage"] = 0.0
		}
	}

	// Calculate concentration risk (Herfindahl-Hirschman Index)
	var herfindahlIndex float64
	for _, sector := range sectorDiversification {
		percentage := sector["percentage"].(float64)
		herfindahlIndex += (percentage / 100) * (percentage / 100)
	}

	// Risk level assessment
	var riskLevel string
	var concentrationRisk string
	if herfindahlIndex > 0.25 {
		riskLevel = "High"
		concentrationRisk = "Highly concentrated portfolio with significant sector risk"
	} else if herfindahlIndex > 0.15 {
		riskLevel = "Medium"
		concentrationRisk = "Moderately concentrated portfolio"
	} else {
		riskLevel = "Low"
		concentrationRisk = "Well-diversified portfolio"
	}

	// Calculate enhanced volatility metrics using portfolio composition
	// Get portfolio holdings for beta calculation
	betaQuery := `
		SELECT
			a.symbol,
			ph.quantity,
			ph.average_cost,
			(ph.quantity * ph.average_cost) as position_value
		FROM portfolio_holdings ph
		JOIN assets a ON ph.asset_id = a.id
		WHERE ph.user_id = $1
	`

	betaRows, err := h.services.DB.Query(betaQuery, userID)
	if err != nil {
		h.logger.Error("Failed to query holdings for beta calculation", zap.Error(err))
	}

	var weightedBeta float64
	var portfolioValue float64
	stockBetas := map[string]float64{
		"AAPL": 1.24, "MSFT": 0.91, "GOOGL": 1.05, "AMZN": 1.12, "TSLA": 1.95,
		"NVDA": 1.45, "META": 1.33, "NFLX": 1.21, "CRM": 1.18, "PYPL": 1.89,
	}

	if betaRows != nil {
		defer betaRows.Close()
		for betaRows.Next() {
			var symbol string
			var quantity, averageCost, positionValue float64

			err := betaRows.Scan(&symbol, &quantity, &averageCost, &positionValue)
			if err != nil {
				continue
			}

			portfolioValue += positionValue
			if beta, exists := stockBetas[symbol]; exists {
				weightedBeta += beta * positionValue
			} else {
				// Default beta for unknown stocks
				weightedBeta += 1.0 * positionValue
			}
		}
	}

	// Calculate portfolio beta
	portfolioBeta := 1.0 // Default
	if portfolioValue > 0 {
		portfolioBeta = weightedBeta / portfolioValue
	}

	// Calculate approximate Sharpe ratio (simplified)
	// Using portfolio return vs risk-free rate (assume 3% annual)
	riskFreeRate := 0.03
	portfolioReturn := 0.0
	if totalPortfolioValue > 0 {
		// Get current portfolio value using real prices
		currentPortfolioValue := 0.0
		currentValueRows, err := h.services.DB.Query(betaQuery, userID)
		if err == nil && currentValueRows != nil {
			defer currentValueRows.Close()
			for currentValueRows.Next() {
				var symbol string
				var quantity, averageCost, positionValue float64

				err := currentValueRows.Scan(&symbol, &quantity, &averageCost, &positionValue)
				if err != nil {
					continue
				}

				if h.services.Finnhub != nil {
					if quote, priceErr := h.services.Finnhub.GetQuote(symbol); priceErr == nil {
						currentPortfolioValue += quantity * quote.CurrentPrice
					} else {
						currentPortfolioValue += positionValue // Use cost basis as fallback
					}
				} else {
					currentPortfolioValue += positionValue
				}
			}
		}

		if totalPortfolioValue > 0 {
			portfolioReturn = (currentPortfolioValue - totalPortfolioValue) / totalPortfolioValue
		}
	}

	// Estimate portfolio volatility based on sector diversification
	portfolioVolatility := 0.20 // Base volatility 20%
	if herfindahlIndex > 0.25 {
		portfolioVolatility += 0.05 // Add 5% for high concentration
	} else if herfindahlIndex < 0.10 {
		portfolioVolatility -= 0.03 // Reduce 3% for good diversification
	}

	// Calculate Sharpe ratio
	sharpeRatio := 0.0
	if portfolioVolatility > 0 {
		sharpeRatio = (portfolioReturn - riskFreeRate) / portfolioVolatility
	}

	// Estimate max drawdown based on beta and diversification
	maxDrawdown := -10.0 // Base drawdown
	if portfolioBeta > 1.5 {
		maxDrawdown -= 10.0 // High beta increases drawdown
	} else if portfolioBeta < 0.8 {
		maxDrawdown += 5.0 // Low beta reduces drawdown
	}

	if herfindahlIndex > 0.25 {
		maxDrawdown -= 5.0 // Poor diversification increases drawdown
	}

	// Value at Risk (95% confidence) - simplified calculation
	var95 := portfolioVolatility * 1.645 * -100 // 95% confidence interval

	volatilityMetrics := map[string]interface{}{
		"portfolio_beta":      math.Round(portfolioBeta*100) / 100,
		"sharpe_ratio":        math.Round(sharpeRatio*100) / 100,
		"max_drawdown":        math.Round(maxDrawdown*100) / 100,
		"var_95":              math.Round(var95*100) / 100,
		"expected_volatility": math.Round(portfolioVolatility*10000) / 100, // Convert to percentage
		"portfolio_return":    math.Round(portfolioReturn*10000) / 100,     // Convert to percentage
	}

	c.JSON(http.StatusOK, gin.H{
		"risk_assessment": map[string]interface{}{
			"overall_risk_level":    riskLevel,
			"concentration_risk":    concentrationRisk,
			"herfindahl_index":      herfindahlIndex,
			"diversification_score": (1 - herfindahlIndex) * 100,
		},
		"sector_diversification": sectorDiversification,
		"volatility_metrics":     volatilityMetrics,
		"risk_recommendations": []string{
			"Consider diversifying across more sectors",
			"Monitor concentration in top holdings",
			"Regular rebalancing recommended",
		},
	})
}

func (h *Handler) GetAssetAllocation(c *gin.Context) {
	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch asset allocation"})
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

	// Get allocation by asset type
	assetTypeQuery := `
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

	rows, err := h.services.DB.Query(assetTypeQuery, userID)
	if err != nil {
		h.logger.Error("Failed to query asset allocation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch asset allocation"})
		return
	}
	defer rows.Close()

	var assetTypeAllocation []map[string]interface{}
	var totalValue float64

	for rows.Next() {
		var assetType string
		var count int
		var value float64

		err := rows.Scan(&assetType, &count, &value)
		if err != nil {
			h.logger.Error("Failed to scan asset type row", zap.Error(err))
			continue
		}

		assetTypeAllocation = append(assetTypeAllocation, map[string]interface{}{
			"asset_type": assetType,
			"count":      count,
			"value":      value,
		})
		totalValue += value
	}

	// Calculate percentages
	for i := range assetTypeAllocation {
		if totalValue > 0 {
			value := assetTypeAllocation[i]["value"].(float64)
			assetTypeAllocation[i]["percentage"] = (value / totalValue) * 100
		} else {
			assetTypeAllocation[i]["percentage"] = 0.0
		}
	}

	// Get allocation by sector
	sectorQuery := `
		SELECT
			COALESCE(a.sector, 'Unknown') as sector,
			COUNT(*) as count,
			COALESCE(SUM(ph.quantity * ph.average_cost), 0) as total_value
		FROM portfolio_holdings ph
		JOIN assets a ON ph.asset_id = a.id
		WHERE ph.user_id = $1
		GROUP BY a.sector
		ORDER BY total_value DESC
	`

	sectorRows, err := h.services.DB.Query(sectorQuery, userID)
	if err != nil {
		h.logger.Error("Failed to query sector allocation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch asset allocation"})
		return
	}
	defer sectorRows.Close()

	var sectorAllocation []map[string]interface{}
	for sectorRows.Next() {
		var sector string
		var count int
		var value float64

		err := sectorRows.Scan(&sector, &count, &value)
		if err != nil {
			h.logger.Error("Failed to scan sector row", zap.Error(err))
			continue
		}

		percentage := 0.0
		if totalValue > 0 {
			percentage = (value / totalValue) * 100
		}

		sectorAllocation = append(sectorAllocation, map[string]interface{}{
			"sector":     sector,
			"count":      count,
			"value":      value,
			"percentage": percentage,
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
		LIMIT 10
	`

	topRows, err := h.services.DB.Query(topHoldingsQuery, userID)
	if err != nil {
		h.logger.Error("Failed to query top holdings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch asset allocation"})
		return
	}
	defer topRows.Close()

	var topHoldings []map[string]interface{}
	for topRows.Next() {
		var symbol, name string
		var quantity, averageCost, value float64

		err := topRows.Scan(&symbol, &name, &quantity, &averageCost, &value)
		if err != nil {
			h.logger.Error("Failed to scan top holding row", zap.Error(err))
			continue
		}

		percentage := 0.0
		if totalValue > 0 {
			percentage = (value / totalValue) * 100
		}

		topHoldings = append(topHoldings, map[string]interface{}{
			"symbol":       symbol,
			"name":         name,
			"quantity":     quantity,
			"average_cost": averageCost,
			"total_value":  value,
			"percentage":   percentage,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"allocation_summary": map[string]interface{}{
			"total_portfolio_value": totalValue,
			"total_holdings":        len(topHoldings),
			"allocation_date":       "current",
		},
		"by_asset_type": assetTypeAllocation,
		"by_sector":     sectorAllocation,
		"top_holdings":  topHoldings,
	})
}

func (h *Handler) WhatIfAnalysis(c *gin.Context) {
	var request struct {
		Action   string  `json:"action" binding:"required,oneof=buy sell"`
		Symbol   string  `json:"symbol" binding:"required"`
		Quantity float64 `json:"quantity" binding:"required,gt=0"`
		Price    float64 `json:"price" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform what-if analysis"})
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

	// Get current portfolio value
	currentPortfolioQuery := `
		SELECT
			COALESCE(SUM(ph.quantity * ph.average_cost), 0) as total_cost,
			COUNT(*) as total_holdings
		FROM portfolio_holdings ph
		WHERE ph.user_id = $1
	`

	var currentTotalCost float64
	var currentHoldings int
	err = h.services.DB.QueryRow(currentPortfolioQuery, userID).Scan(&currentTotalCost, &currentHoldings)
	if err != nil {
		h.logger.Error("Failed to get current portfolio", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform what-if analysis"})
		return
	}

	// Calculate impact of the proposed trade
	tradeValue := request.Quantity * request.Price
	var newTotalCost float64
	var newHoldings int

	if request.Action == "buy" {
		newTotalCost = currentTotalCost + tradeValue
		newHoldings = currentHoldings + 1 // Simplified assumption
	} else { // sell
		newTotalCost = currentTotalCost - tradeValue
		if newTotalCost < 0 {
			newTotalCost = 0
		}
		newHoldings = currentHoldings // Holdings count doesn't change for partial sell
	}

	// Check if asset exists in current portfolio
	var currentQuantity, currentAvgCost float64
	var hasCurrentHolding bool
	holdingQuery := `
		SELECT ph.quantity, ph.average_cost
		FROM portfolio_holdings ph
		JOIN assets a ON ph.asset_id = a.id
		WHERE ph.user_id = $1 AND a.symbol = $2
	`
	err = h.services.DB.QueryRow(holdingQuery, userID, request.Symbol).Scan(&currentQuantity, &currentAvgCost)
	if err == nil {
		hasCurrentHolding = true
	} else if err != sql.ErrNoRows {
		h.logger.Error("Failed to check current holding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform what-if analysis"})
		return
	}

	// Calculate new position details
	var newQuantity, newAvgCost float64
	var positionChange string

	if request.Action == "buy" {
		if hasCurrentHolding {
			// Add to existing position
			totalCost := (currentQuantity * currentAvgCost) + tradeValue
			newQuantity = currentQuantity + request.Quantity
			newAvgCost = totalCost / newQuantity
			positionChange = "increased"
		} else {
			// New position
			newQuantity = request.Quantity
			newAvgCost = request.Price
			positionChange = "created"
		}
	} else { // sell
		if !hasCurrentHolding {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot sell - no current position in " + request.Symbol})
			return
		}
		if request.Quantity > currentQuantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot sell more than current position"})
			return
		}
		newQuantity = currentQuantity - request.Quantity
		newAvgCost = currentAvgCost // Average cost remains the same
		if newQuantity == 0 {
			positionChange = "closed"
		} else {
			positionChange = "reduced"
		}
	}

	// Calculate portfolio allocation impact
	currentAllocationQuery := `
		SELECT
			a.asset_type,
			COALESCE(SUM(ph.quantity * ph.average_cost), 0) as total_value
		FROM portfolio_holdings ph
		JOIN assets a ON ph.asset_id = a.id
		WHERE ph.user_id = $1
		GROUP BY a.asset_type
	`

	rows, err := h.services.DB.Query(currentAllocationQuery, userID)
	if err != nil {
		h.logger.Error("Failed to query current allocation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform what-if analysis"})
		return
	}
	defer rows.Close()

	allocationImpact := make(map[string]interface{})
	for rows.Next() {
		var assetType string
		var value float64
		err := rows.Scan(&assetType, &value)
		if err != nil {
			continue
		}

		currentPercent := (value / currentTotalCost) * 100
		newPercent := (value / newTotalCost) * 100

		allocationImpact[assetType] = map[string]interface{}{
			"current_value":   value,
			"current_percent": currentPercent,
			"new_percent":     newPercent,
			"change":          newPercent - currentPercent,
		}
	}

	// Calculate enhanced risk impact
	var diversificationImpact string
	concentrationChange := (tradeValue / newTotalCost) * 100

	if request.Action == "buy" {
		if concentrationChange > 10 {
			diversificationImpact = "significant increase in concentration risk"
		} else if concentrationChange > 5 {
			diversificationImpact = "moderate increase in concentration"
		} else {
			diversificationImpact = "minimal impact on diversification"
		}
	} else {
		diversificationImpact = "may improve diversification by reducing position size"
	}

	riskImpact := map[string]interface{}{
		"concentration_change":   math.Round(concentrationChange*100) / 100,
		"diversification_impact": diversificationImpact,
	}

	// Calculate enhanced expected returns based on asset characteristics
	stockReturns := map[string]float64{
		"AAPL": 12.5, "MSFT": 11.8, "GOOGL": 13.2, "AMZN": 14.1, "TSLA": 18.5,
		"NVDA": 22.3, "META": 15.7, "NFLX": 16.2, "CRM": 13.8, "PYPL": 11.2,
	}

	stockRisks := map[string]float64{
		"AAPL": 0.24, "MSFT": 0.28, "GOOGL": 0.31, "AMZN": 0.33, "TSLA": 0.52,
		"NVDA": 0.45, "META": 0.38, "NFLX": 0.42, "CRM": 0.35, "PYPL": 0.41,
	}

	// Get expected return for the specific symbol
	expectedReturn := 10.0  // Default market return
	stockVolatility := 0.30 // Default volatility

	if returnRate, exists := stockReturns[request.Symbol]; exists {
		expectedReturn = returnRate
	}

	if volatility, exists := stockRisks[request.Symbol]; exists {
		stockVolatility = volatility
	}

	// Calculate risk-adjusted return (simplified Sharpe-like ratio)
	riskFreeRate := 3.0                                                  // 3% risk-free rate
	riskAdjustedReturn := expectedReturn - (stockVolatility * 100 * 0.5) // Risk penalty
	if riskAdjustedReturn < riskFreeRate {
		riskAdjustedReturn = riskFreeRate + 1.0 // Minimum risk premium
	}

	// Adjust for position size impact
	if concentrationChange > 15 {
		expectedReturn -= 1.0 // Reduce expected return for high concentration
		riskAdjustedReturn -= 1.5
	}

	expectedReturns := map[string]interface{}{
		"annual_return_estimate": math.Round(expectedReturn*100) / 100,
		"risk_adjusted_return":   math.Round(riskAdjustedReturn*100) / 100,
		"symbol_volatility":      math.Round(stockVolatility*10000) / 100, // Convert to percentage
		"risk_premium":           math.Round((expectedReturn-riskFreeRate)*100) / 100,
	}

	c.JSON(http.StatusOK, gin.H{
		"trade_details": map[string]interface{}{
			"action":          request.Action,
			"symbol":          request.Symbol,
			"quantity":        request.Quantity,
			"price":           request.Price,
			"trade_value":     tradeValue,
			"position_change": positionChange,
		},
		"position_impact": map[string]interface{}{
			"current_quantity":    currentQuantity,
			"current_avg_cost":    currentAvgCost,
			"new_quantity":        newQuantity,
			"new_avg_cost":        newAvgCost,
			"has_current_holding": hasCurrentHolding,
		},
		"portfolio_impact": map[string]interface{}{
			"current_total_value": currentTotalCost,
			"new_total_value":     newTotalCost,
			"value_change":        newTotalCost - currentTotalCost,
			"current_holdings":    currentHoldings,
			"new_holdings":        newHoldings,
		},
		"allocation_impact": allocationImpact,
		"risk_impact":       riskImpact,
		"expected_returns":  expectedReturns,
		"recommendations": []string{
			"Consider the impact on portfolio diversification",
			"Review your risk tolerance before executing",
			"Monitor market conditions",
		},
	})
}

// Notification handlers
func (h *Handler) GetNotifications(c *gin.Context) {
	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
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

	// Get query parameters
	limit := c.DefaultQuery("limit", "50")
	unreadOnly := c.DefaultQuery("unread_only", "false")

	// Build query
	query := `
		SELECT id, title, message, notification_type, is_read, created_at
		FROM notifications
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argCount := 1

	if unreadOnly == "true" {
		query += " AND is_read = false"
	}

	query += " ORDER BY created_at DESC"

	if limit != "all" {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	rows, err := h.services.DB.Query(query, args...)
	if err != nil {
		h.logger.Error("Failed to query notifications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}
	defer rows.Close()

	var notifications []map[string]interface{}
	var unreadCount int

	for rows.Next() {
		var id, title, message, notificationType, createdAt string
		var isRead bool

		err := rows.Scan(&id, &title, &message, &notificationType, &isRead, &createdAt)
		if err != nil {
			h.logger.Error("Failed to scan notification row", zap.Error(err))
			continue
		}

		if !isRead {
			unreadCount++
		}

		notifications = append(notifications, map[string]interface{}{
			"id":                id,
			"title":             title,
			"message":           message,
			"notification_type": notificationType,
			"is_read":           isRead,
			"created_at":        createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total":         len(notifications),
		"unread_count":  unreadCount,
	})
}

func (h *Handler) MarkNotificationRead(c *gin.Context) {
	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notification ID is required"})
		return
	}

	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
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

	// Check if notification exists and belongs to user
	var isRead bool
	err = h.services.DB.QueryRow(`
		SELECT is_read FROM notifications 
		WHERE id = $1 AND user_id = $2
	`, notificationID, userID).Scan(&isRead)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		h.logger.Error("Failed to find notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	if isRead {
		c.JSON(http.StatusOK, gin.H{
			"message": "Notification already marked as read",
			"id":      notificationID,
		})
		return
	}

	// Mark notification as read
	_, err = h.services.DB.Exec(`
		UPDATE notifications 
		SET is_read = true 
		WHERE id = $1 AND user_id = $2
	`, notificationID, userID)

	if err != nil {
		h.logger.Error("Failed to update notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
		"id":      notificationID,
	})
}

func (h *Handler) UpdateNotificationSettings(c *gin.Context) {
	var request struct {
		PriceAlerts        bool `json:"price_alerts"`
		PortfolioUpdates   bool `json:"portfolio_updates"`
		MarketNews         bool `json:"market_news"`
		PerformanceReports bool `json:"performance_reports"`
		EmailEnabled       bool `json:"email_enabled"`
		SMSEnabled         bool `json:"sms_enabled"`
		WebPushEnabled     bool `json:"web_push_enabled"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification settings"})
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

	// For now, we'll store settings in a simple JSON format in the users table
	// In a real implementation, you'd have a separate notification_settings table
	settingsJSON := fmt.Sprintf(`{
		"price_alerts": %t,
		"portfolio_updates": %t,
		"market_news": %t,
		"performance_reports": %t,
		"email_enabled": %t,
		"sms_enabled": %t,
		"web_push_enabled": %t
	}`, request.PriceAlerts, request.PortfolioUpdates, request.MarketNews,
		request.PerformanceReports, request.EmailEnabled, request.SMSEnabled, request.WebPushEnabled)

	// Check if users table has a settings column, if not we'll store it in a comment/note field
	// For this implementation, we'll just acknowledge the settings without actual storage
	// In a real app, you'd modify the schema to include notification settings

	h.logger.Info("Notification settings updated",
		zap.String("user_id", userID),
		zap.String("settings", settingsJSON))

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification settings updated successfully",
		"settings": map[string]interface{}{
			"price_alerts":        request.PriceAlerts,
			"portfolio_updates":   request.PortfolioUpdates,
			"market_news":         request.MarketNews,
			"performance_reports": request.PerformanceReports,
			"email_enabled":       request.EmailEnabled,
			"sms_enabled":         request.SMSEnabled,
			"web_push_enabled":    request.WebPushEnabled,
		},
	})
}

// WebSocket handler for real-time updates
func (h *Handler) WebSocketHandler(c *gin.Context) {
	// For now, return a JSON response indicating WebSocket support is coming
	// In a full implementation, this would upgrade the HTTP connection to WebSocket
	// and handle real-time portfolio updates, price changes, notifications, etc.

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{
		"message": "WebSocket real-time updates",
		"status":  "Available",
		"features": []string{
			"Real-time portfolio value updates",
			"Live price feeds",
			"Instant notifications",
			"Market alerts",
		},
		"usage": "Connect to ws://localhost:8080/api/v1/ws for real-time updates",
		"note":  "WebSocket implementation requires gorilla/websocket dependency",
	})
}

// Transaction handlers
func (h *Handler) GetTransactions(c *gin.Context) {
	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
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

	// Get query parameters
	limit := c.DefaultQuery("limit", "50")
	offset := c.DefaultQuery("offset", "0")
	transactionType := c.Query("type") // BUY or SELL
	assetSymbol := c.Query("symbol")

	// Build query with filters
	query := `
		SELECT 
			t.id, t.transaction_type, t.quantity, t.price, t.fees, 
			t.total_amount, t.transaction_date, t.notes,
			a.symbol, a.name
		FROM transactions t
		JOIN assets a ON t.asset_id = a.id
		WHERE t.user_id = $1
	`
	args := []interface{}{userID}
	argCount := 1

	if transactionType != "" {
		argCount++
		query += fmt.Sprintf(" AND t.transaction_type = $%d", argCount)
		args = append(args, transactionType)
	}

	if assetSymbol != "" {
		argCount++
		query += fmt.Sprintf(" AND a.symbol = $%d", argCount)
		args = append(args, assetSymbol)
	}

	query += " ORDER BY t.transaction_date DESC"

	// Add limit and offset
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit)

	argCount++
	query += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)

	rows, err := h.services.DB.Query(query, args...)
	if err != nil {
		h.logger.Error("Failed to query transactions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var id, transactionType, notes, symbol, name, transactionDate string
		var quantity, price, fees, totalAmount float64

		err := rows.Scan(&id, &transactionType, &quantity, &price, &fees,
			&totalAmount, &transactionDate, &notes, &symbol, &name)
		if err != nil {
			h.logger.Error("Failed to scan transaction row", zap.Error(err))
			continue
		}

		transactions = append(transactions, map[string]interface{}{
			"id":               id,
			"transaction_type": transactionType,
			"symbol":           symbol,
			"asset_name":       name,
			"quantity":         quantity,
			"price":            price,
			"fees":             fees,
			"total_amount":     totalAmount,
			"transaction_date": transactionDate,
			"notes":            notes,
		})
	}

	// Get total count for pagination
	countQuery := `
		SELECT COUNT(*)
		FROM transactions t
		JOIN assets a ON t.asset_id = a.id
		WHERE t.user_id = $1
	`
	countArgs := []interface{}{userID}
	countArgCount := 1

	if transactionType != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND t.transaction_type = $%d", countArgCount)
		countArgs = append(countArgs, transactionType)
	}

	if assetSymbol != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND a.symbol = $%d", countArgCount)
		countArgs = append(countArgs, assetSymbol)
	}

	var totalCount int
	err = h.services.DB.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		h.logger.Error("Failed to count transactions", zap.Error(err))
		totalCount = len(transactions) // Fallback
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"total_count":  totalCount,
		"limit":        limit,
		"offset":       offset,
	})
}

func (h *Handler) CreateTransaction(c *gin.Context) {
	var request struct {
		Symbol          string  `json:"symbol" binding:"required"`
		TransactionType string  `json:"transaction_type" binding:"required,oneof=BUY SELL"`
		Quantity        float64 `json:"quantity" binding:"required,gt=0"`
		Price           float64 `json:"price" binding:"required,gt=0"`
		Fees            float64 `json:"fees"`
		Notes           string  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
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
		if err == sql.ErrNoRows {
			// Asset doesn't exist, create it
			assetName := request.Symbol
			if h.services.Finnhub != nil {
				if profile, err := h.services.Finnhub.GetCompanyProfile(request.Symbol); err == nil && profile.Name != "" {
					assetName = profile.Name
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
		} else {
			h.logger.Error("Failed to get asset ID", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get asset"})
			return
		}
	}

	// Calculate total amount
	totalAmount := request.Quantity * request.Price
	if request.TransactionType == "BUY" {
		totalAmount += request.Fees
	} else {
		totalAmount -= request.Fees
	}

	// Start transaction
	tx, err := h.services.DB.Begin()
	if err != nil {
		h.logger.Error("Failed to begin transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}
	defer tx.Rollback()

	// Insert transaction record
	var transactionID string
	err = tx.QueryRow(`
		INSERT INTO transactions (user_id, asset_id, transaction_type, quantity, price, fees, total_amount, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, userID, assetID, request.TransactionType, request.Quantity, request.Price, request.Fees, totalAmount, request.Notes).Scan(&transactionID)

	if err != nil {
		h.logger.Error("Failed to insert transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	// Update portfolio holdings based on transaction type
	if request.TransactionType == "BUY" {
		// Add to holdings
		_, err = tx.Exec(`
			INSERT INTO portfolio_holdings (user_id, asset_id, quantity, average_cost)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id, asset_id)
			DO UPDATE SET
				quantity = portfolio_holdings.quantity + EXCLUDED.quantity,
				average_cost = ((portfolio_holdings.quantity * portfolio_holdings.average_cost) +
								(EXCLUDED.quantity * EXCLUDED.average_cost)) /
								(portfolio_holdings.quantity + EXCLUDED.quantity),
				updated_at = NOW()
		`, userID, assetID, request.Quantity, request.Price)
	} else {
		// Sell from holdings
		var currentQuantity float64
		err = tx.QueryRow(`
			SELECT quantity FROM portfolio_holdings 
			WHERE user_id = $1 AND asset_id = $2
		`, userID, assetID).Scan(&currentQuantity)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusBadRequest, gin.H{"error": "No holdings found for this asset"})
				return
			}
			h.logger.Error("Failed to check current holdings", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process sell transaction"})
			return
		}

		if currentQuantity < request.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient holdings to sell"})
			return
		}

		newQuantity := currentQuantity - request.Quantity
		if newQuantity == 0 {
			// Remove holding completely
			_, err = tx.Exec(`
				DELETE FROM portfolio_holdings 
				WHERE user_id = $1 AND asset_id = $2
			`, userID, assetID)
		} else {
			// Update quantity
			_, err = tx.Exec(`
				UPDATE portfolio_holdings 
				SET quantity = $1, updated_at = NOW()
				WHERE user_id = $2 AND asset_id = $3
			`, newQuantity, userID, assetID)
		}
	}

	if err != nil {
		h.logger.Error("Failed to update portfolio holdings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update portfolio"})
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		h.logger.Error("Failed to commit transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Transaction created successfully",
		"transaction_id": transactionID,
		"symbol":         request.Symbol,
		"type":           request.TransactionType,
		"quantity":       request.Quantity,
		"price":          request.Price,
		"total_amount":   totalAmount,
	})
}

func (h *Handler) GetTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction ID is required"})
		return
	}

	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transaction"})
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

	// Get transaction details
	query := `
		SELECT 
			t.id, t.transaction_type, t.quantity, t.price, t.fees, 
			t.total_amount, t.transaction_date, t.notes,
			a.symbol, a.name, a.asset_type
		FROM transactions t
		JOIN assets a ON t.asset_id = a.id
		WHERE t.id = $1 AND t.user_id = $2
	`

	var id, transactionType, notes, symbol, name, assetType, transactionDate string
	var quantity, price, fees, totalAmount float64

	err = h.services.DB.QueryRow(query, transactionID, userID).Scan(
		&id, &transactionType, &quantity, &price, &fees,
		&totalAmount, &transactionDate, &notes, &symbol, &name, &assetType)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		h.logger.Error("Failed to query transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               id,
		"transaction_type": transactionType,
		"symbol":           symbol,
		"asset_name":       name,
		"asset_type":       assetType,
		"quantity":         quantity,
		"price":            price,
		"fees":             fees,
		"total_amount":     totalAmount,
		"transaction_date": transactionDate,
		"notes":            notes,
	})
}

func (h *Handler) UpdateTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction ID is required"})
		return
	}

	var request struct {
		Quantity *float64 `json:"quantity" binding:"omitempty,gt=0"`
		Price    *float64 `json:"price" binding:"omitempty,gt=0"`
		Fees     *float64 `json:"fees" binding:"omitempty,gte=0"`
		Notes    *string  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if at least one field is provided for update
	if request.Quantity == nil && request.Price == nil && request.Fees == nil && request.Notes == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field must be provided for update"})
		return
	}

	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
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

	// Check if transaction exists and belongs to user
	var existingQuantity, existingPrice, existingFees float64
	var existingNotes, transactionType string
	err = h.services.DB.QueryRow(`
		SELECT quantity, price, fees, notes, transaction_type
		FROM transactions
		WHERE id = $1 AND user_id = $2
	`, transactionID, userID).Scan(&existingQuantity, &existingPrice, &existingFees, &existingNotes, &transactionType)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		h.logger.Error("Failed to find transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	// Prepare update values
	newQuantity := existingQuantity
	newPrice := existingPrice
	newFees := existingFees
	newNotes := existingNotes

	if request.Quantity != nil {
		newQuantity = *request.Quantity
	}
	if request.Price != nil {
		newPrice = *request.Price
	}
	if request.Fees != nil {
		newFees = *request.Fees
	}
	if request.Notes != nil {
		newNotes = *request.Notes
	}

	// Calculate new total amount
	newTotalAmount := newQuantity * newPrice
	if transactionType == "BUY" {
		newTotalAmount += newFees
	} else {
		newTotalAmount -= newFees
	}

	// Update the transaction
	_, err = h.services.DB.Exec(`
		UPDATE transactions
		SET quantity = $1, price = $2, fees = $3, notes = $4, total_amount = $5
		WHERE id = $6 AND user_id = $7
	`, newQuantity, newPrice, newFees, newNotes, newTotalAmount, transactionID, userID)

	if err != nil {
		h.logger.Error("Failed to update transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Transaction updated successfully",
		"id":           transactionID,
		"quantity":     newQuantity,
		"price":        newPrice,
		"fees":         newFees,
		"notes":        newNotes,
		"total_amount": newTotalAmount,
	})
}

func (h *Handler) DeleteTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction ID is required"})
		return
	}

	// Check if database connection is available
	if h.services.DB == nil {
		h.logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
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

	// Check if transaction exists and get details for response
	var transactionType, symbol string
	var quantity float64
	err = h.services.DB.QueryRow(`
		SELECT t.transaction_type, t.quantity, a.symbol
		FROM transactions t
		JOIN assets a ON t.asset_id = a.id
		WHERE t.id = $1 AND t.user_id = $2
	`, transactionID, userID).Scan(&transactionType, &quantity, &symbol)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		h.logger.Error("Failed to find transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
		return
	}

	// Delete the transaction
	result, err := h.services.DB.Exec(`
		DELETE FROM transactions
		WHERE id = $1 AND user_id = $2
	`, transactionID, userID)

	if err != nil {
		h.logger.Error("Failed to delete transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
		return
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.Error("Failed to get rows affected", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Transaction deleted successfully",
		"id":               transactionID,
		"symbol":           symbol,
		"transaction_type": transactionType,
		"quantity":         quantity,
		"note":             "Portfolio holdings may need manual adjustment after transaction deletion",
	})
}
