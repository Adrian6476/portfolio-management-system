package services

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MarketUpdater handles periodic market data updates and WebSocket broadcasting
type MarketUpdater struct {
	db        *sql.DB
	finnhub   *FinnhubClient
	websocket *WebSocketHub
	logger    *zap.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewMarketUpdater creates a new market updater service
func NewMarketUpdater(db *sql.DB, finnhub *FinnhubClient, websocket *WebSocketHub, logger *zap.Logger) *MarketUpdater {
	ctx, cancel := context.WithCancel(context.Background())
	return &MarketUpdater{
		db:        db,
		finnhub:   finnhub,
		websocket: websocket,
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start begins the periodic market data updates
func (m *MarketUpdater) Start() {
	if m.finnhub == nil {
		m.logger.Warn("Finnhub client not available, market updater will not start")
		return
	}

	m.logger.Info("Starting market data updater")

	// Start price update routine (every 30 seconds)
	m.wg.Add(1)
	go m.priceUpdateRoutine(30 * time.Second)

	// Start portfolio update routine (every 60 seconds)
	m.wg.Add(1)
	go m.portfolioUpdateRoutine(60 * time.Second)
}

// Stop gracefully shuts down the market updater
func (m *MarketUpdater) Stop() {
	m.logger.Info("Stopping market data updater")
	m.cancel()
	m.wg.Wait()
	m.logger.Info("Market data updater stopped")
}

// priceUpdateRoutine periodically updates prices for all portfolio holdings
func (m *MarketUpdater) priceUpdateRoutine(interval time.Duration) {
	defer m.wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.updatePortfolioPrices()
		}
	}
}

// portfolioUpdateRoutine periodically broadcasts portfolio summaries
func (m *MarketUpdater) portfolioUpdateRoutine(interval time.Duration) {
	defer m.wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.broadcastPortfolioUpdates()
		}
	}
}

// updatePortfolioPrices fetches current prices for all portfolio holdings and broadcasts updates
func (m *MarketUpdater) updatePortfolioPrices() {
	// Get all unique symbols from portfolio holdings
	query := `
		SELECT DISTINCT a.symbol
		FROM portfolio_holdings ph
		JOIN assets a ON ph.asset_id = a.id
		WHERE ph.quantity > 0
	`

	rows, err := m.db.Query(query)
	if err != nil {
		m.logger.Error("Failed to query portfolio symbols", zap.Error(err))
		return
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			m.logger.Error("Failed to scan symbol", zap.Error(err))
			continue
		}
		symbols = append(symbols, symbol)
	}

	// Update prices for each symbol
	for _, symbol := range symbols {
		m.updateAndBroadcastPrice(symbol)
		// Small delay to avoid hitting API rate limits
		time.Sleep(100 * time.Millisecond)
	}

	m.logger.Info("Updated prices for portfolio holdings", zap.Int("symbols_count", len(symbols)))
}

// updateAndBroadcastPrice fetches and broadcasts price update for a specific symbol
func (m *MarketUpdater) updateAndBroadcastPrice(symbol string) {
	quote, err := m.finnhub.GetQuote(symbol)
	if err != nil {
		m.logger.Error("Failed to get quote for symbol",
			zap.String("symbol", symbol), zap.Error(err))
		return
	}

	// Store price in database (optional - for historical tracking)
	_, err = m.db.Exec(`
		INSERT INTO market_data (asset_id, price, change_24h, timestamp)
		SELECT a.id, $2, $3, NOW()
		FROM assets a WHERE a.symbol = $1
		ON CONFLICT (asset_id) DO UPDATE SET
			price = EXCLUDED.price,
			change_24h = EXCLUDED.change_24h,
			timestamp = EXCLUDED.timestamp
	`, symbol, quote.CurrentPrice, quote.Change)

	if err != nil {
		m.logger.Warn("Failed to store market data",
			zap.String("symbol", symbol), zap.Error(err))
		// Continue with broadcast even if storage fails
	}

	// Broadcast price update via WebSocket
	if m.websocket != nil {
		update := PriceUpdate{
			Symbol:        symbol,
			CurrentPrice:  quote.CurrentPrice,
			Change:        quote.Change,
			ChangePercent: quote.PercentChange,
			High:          quote.HighPriceOfDay,
			Low:           quote.LowPriceOfDay,
		}

		m.websocket.BroadcastPriceUpdate(update)
	}
}

// broadcastPortfolioUpdates calculates and broadcasts portfolio summaries for all users
func (m *MarketUpdater) broadcastPortfolioUpdates() {
	// Get all users with portfolio holdings
	query := `
		SELECT DISTINCT u.username, u.id
		FROM users u
		JOIN portfolio_holdings ph ON u.id = ph.user_id
		WHERE ph.quantity > 0
	`

	rows, err := m.db.Query(query)
	if err != nil {
		m.logger.Error("Failed to query users with portfolios", zap.Error(err))
		return
	}
	defer rows.Close()

	userCount := 0
	for rows.Next() {
		var username, userID string
		if err := rows.Scan(&username, &userID); err != nil {
			m.logger.Error("Failed to scan user", zap.Error(err))
			continue
		}

		m.calculateAndBroadcastPortfolioUpdate(userID)
		userCount++
	}

	if userCount > 0 {
		m.logger.Info("Broadcasted portfolio updates", zap.Int("users_count", userCount))
	}
}

// calculateAndBroadcastPortfolioUpdate calculates and broadcasts portfolio update for a specific user
func (m *MarketUpdater) calculateAndBroadcastPortfolioUpdate(userID string) {
	query := `
		SELECT
			a.symbol,
			ph.quantity,
			ph.average_cost
		FROM portfolio_holdings ph
		JOIN assets a ON ph.asset_id = a.id
		WHERE ph.user_id = $1 AND ph.quantity > 0
	`

	rows, err := m.db.Query(query, userID)
	if err != nil {
		m.logger.Error("Failed to query user portfolio",
			zap.String("user_id", userID), zap.Error(err))
		return
	}
	defer rows.Close()

	var totalValue, totalCost, totalGainLoss float64
	holdingCount := 0

	for rows.Next() {
		var symbol string
		var quantity, averageCost float64

		if err := rows.Scan(&symbol, &quantity, &averageCost); err != nil {
			m.logger.Error("Failed to scan portfolio holding", zap.Error(err))
			continue
		}

		holdingCount++
		costBasis := quantity * averageCost
		totalCost += costBasis

		// Get current price
		currentPrice := averageCost // fallback
		if quote, priceErr := m.finnhub.GetQuote(symbol); priceErr == nil {
			currentPrice = quote.CurrentPrice
		}

		currentValue := quantity * currentPrice
		totalValue += currentValue
		totalGainLoss += (currentValue - costBasis)
	}

	if holdingCount == 0 {
		return
	}

	// Calculate percentages
	unrealizedGainLossPercent := 0.0
	if totalCost > 0 {
		unrealizedGainLossPercent = (totalGainLoss / totalCost) * 100
	}

	// Simplified daily change calculation
	dailyChange := totalGainLoss * 0.1
	dailyChangePercent := 0.0
	if totalValue > 0 {
		dailyChangePercent = (dailyChange / totalValue) * 100
	}

	// Broadcast portfolio update
	if m.websocket != nil {
		update := PortfolioUpdate{
			TotalValue:                totalValue,
			DailyChange:               dailyChange,
			DailyChangePercent:        dailyChangePercent,
			UnrealizedGainLoss:        totalGainLoss,
			UnrealizedGainLossPercent: unrealizedGainLossPercent,
		}

		m.websocket.BroadcastPortfolioUpdate(update)
	}
}
