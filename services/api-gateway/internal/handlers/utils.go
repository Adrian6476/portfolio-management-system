package handlers

import (
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// CreateSampleData creates sample portfolio data for testing
func (h *Handler) CreateSampleData() error {
	if h.services.DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Get user ID
	var userID string
	err := h.services.DB.QueryRow("SELECT id FROM users WHERE username = $1", "default_user").Scan(&userID)
	if err != nil {
		h.logger.Error("Failed to get user ID", zap.Error(err))
		return err
	}

	// Sample assets to add
	sampleAssets := []struct {
		Symbol    string
		Name      string
		AssetType string
		Sector    string
	}{
		{"AAPL", "Apple Inc.", "STOCK", "Technology"},
		{"GOOGL", "Alphabet Inc.", "STOCK", "Technology"},
		{"MSFT", "Microsoft Corporation", "STOCK", "Technology"},
		{"TSLA", "Tesla, Inc.", "STOCK", "Automotive"},
		{"AMZN", "Amazon.com, Inc.", "STOCK", "E-commerce"},
		{"NVDA", "NVIDIA Corporation", "STOCK", "Technology"},
		{"JPM", "JPMorgan Chase & Co.", "STOCK", "Financial"},
		{"JNJ", "Johnson & Johnson", "STOCK", "Healthcare"},
	}

	// Sample holdings
	sampleHoldings := []struct {
		Symbol      string
		Quantity    float64
		AverageCost float64
	}{
		{"AAPL", 10.0, 175.50},
		{"GOOGL", 5.0, 2800.75},
		{"MSFT", 15.0, 380.25},
		{"TSLA", 8.0, 220.00},
		{"AMZN", 3.0, 3100.50},
	}

	// Sample transactions
	sampleTransactions := []struct {
		Symbol          string
		TransactionType string
		Quantity        float64
		Price           float64
		Fees            float64
		Notes           string
		DaysAgo         int
	}{
		{"AAPL", "BUY", 5.0, 170.00, 2.50, "Initial purchase", 30},
		{"AAPL", "BUY", 5.0, 181.00, 2.50, "Dollar cost averaging", 15},
		{"GOOGL", "BUY", 5.0, 2800.75, 5.00, "Growth investment", 25},
		{"MSFT", "BUY", 10.0, 375.00, 3.00, "Tech diversification", 20},
		{"MSFT", "BUY", 5.0, 390.75, 2.00, "Additional purchase", 10},
		{"TSLA", "BUY", 8.0, 220.00, 4.00, "EV exposure", 18},
		{"AMZN", "BUY", 3.0, 3100.50, 7.50, "E-commerce play", 22},
	}

	// Sample notifications
	sampleNotifications := []struct {
		Title            string
		Message          string
		NotificationType string
		DaysAgo          int
	}{
		{"Portfolio Update", "Your portfolio value increased by 2.5% today", "PORTFOLIO_UPDATE", 1},
		{"Price Alert", "AAPL reached your target price of $180", "PRICE_ALERT", 2},
		{"Market News", "Tech sector showing strong performance this week", "MARKET_NEWS", 3},
		{"Performance Report", "Monthly portfolio performance report is ready", "PERFORMANCE_REPORT", 7},
		{"Risk Alert", "Your portfolio concentration in tech sector is high", "RISK_ALERT", 5},
	}

	// Start transaction
	tx, err := h.services.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert assets (ignore conflicts)
	for _, asset := range sampleAssets {
		_, err = tx.Exec(`
			INSERT INTO assets (symbol, name, asset_type, exchange, currency, sector)
			VALUES ($1, $2, $3, 'NASDAQ', 'USD', $4)
			ON CONFLICT (symbol) DO NOTHING
		`, asset.Symbol, asset.Name, asset.AssetType, asset.Sector)
		if err != nil {
			h.logger.Error("Failed to insert sample asset", zap.String("symbol", asset.Symbol), zap.Error(err))
			return err
		}
	}

	// Clear existing holdings for clean sample data
	_, err = tx.Exec("DELETE FROM portfolio_holdings WHERE user_id = $1", userID)
	if err != nil {
		h.logger.Error("Failed to clear existing holdings", zap.Error(err))
		return err
	}

	// Insert sample holdings
	for _, holding := range sampleHoldings {
		var assetID string
		err = tx.QueryRow("SELECT id FROM assets WHERE symbol = $1", holding.Symbol).Scan(&assetID)
		if err != nil {
			h.logger.Error("Failed to get asset ID for holding", zap.String("symbol", holding.Symbol), zap.Error(err))
			continue
		}

		_, err = tx.Exec(`
			INSERT INTO portfolio_holdings (user_id, asset_id, quantity, average_cost)
			VALUES ($1, $2, $3, $4)
		`, userID, assetID, holding.Quantity, holding.AverageCost)
		if err != nil {
			h.logger.Error("Failed to insert sample holding", zap.String("symbol", holding.Symbol), zap.Error(err))
			return err
		}
	}

	// Clear existing transactions
	_, err = tx.Exec("DELETE FROM transactions WHERE user_id = $1", userID)
	if err != nil {
		h.logger.Error("Failed to clear existing transactions", zap.Error(err))
		return err
	}

	// Insert sample transactions
	for _, transaction := range sampleTransactions {
		var assetID string
		err = tx.QueryRow("SELECT id FROM assets WHERE symbol = $1", transaction.Symbol).Scan(&assetID)
		if err != nil {
			h.logger.Error("Failed to get asset ID for transaction", zap.String("symbol", transaction.Symbol), zap.Error(err))
			continue
		}

		totalAmount := transaction.Quantity * transaction.Price
		if transaction.TransactionType == "BUY" {
			totalAmount += transaction.Fees
		} else {
			totalAmount -= transaction.Fees
		}

		transactionDate := time.Now().AddDate(0, 0, -transaction.DaysAgo)

		_, err = tx.Exec(`
			INSERT INTO transactions (user_id, asset_id, transaction_type, quantity, price, fees, total_amount, notes, transaction_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, userID, assetID, transaction.TransactionType, transaction.Quantity, transaction.Price,
			transaction.Fees, totalAmount, transaction.Notes, transactionDate)
		if err != nil {
			h.logger.Error("Failed to insert sample transaction", zap.String("symbol", transaction.Symbol), zap.Error(err))
			return err
		}
	}

	// Clear existing notifications
	_, err = tx.Exec("DELETE FROM notifications WHERE user_id = $1", userID)
	if err != nil {
		h.logger.Error("Failed to clear existing notifications", zap.Error(err))
		return err
	}

	// Insert sample notifications
	for _, notification := range sampleNotifications {
		createdAt := time.Now().AddDate(0, 0, -notification.DaysAgo)

		_, err = tx.Exec(`
			INSERT INTO notifications (user_id, title, message, notification_type, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, userID, notification.Title, notification.Message, notification.NotificationType, createdAt)
		if err != nil {
			h.logger.Error("Failed to insert sample notification", zap.String("title", notification.Title), zap.Error(err))
			return err
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		h.logger.Error("Failed to commit sample data transaction", zap.Error(err))
		return err
	}

	h.logger.Info("Sample data created successfully")
	return nil
}

// Helper function to validate asset symbol
func (h *Handler) validateAssetSymbol(symbol string) error {
	if len(symbol) < 1 || len(symbol) > 10 {
		return fmt.Errorf("invalid symbol length")
	}
	// Add more validation as needed
	return nil
}

// Helper function to check if user owns an asset
func (h *Handler) userOwnsAsset(userID, assetID string) (bool, float64, error) {
	var quantity float64
	err := h.services.DB.QueryRow(`
		SELECT quantity FROM portfolio_holdings 
		WHERE user_id = $1 AND asset_id = $2
	`, userID, assetID).Scan(&quantity)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, 0, nil
		}
		return false, 0, err
	}

	return true, quantity, nil
}

// Helper function to get asset ID by symbol
func (h *Handler) getAssetIDBySymbol(symbol string) (string, error) {
	var assetID string
	err := h.services.DB.QueryRow("SELECT id FROM assets WHERE symbol = $1", symbol).Scan(&assetID)
	return assetID, err
}

// Helper function to get user ID by username
func (h *Handler) getUserID(username string) (string, error) {
	var userID string
	err := h.services.DB.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&userID)
	return userID, err
}
