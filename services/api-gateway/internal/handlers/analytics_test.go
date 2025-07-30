package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/portfolio-management/api-gateway/internal/services"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// GetAssets handler tests
func TestGetAssets(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   []string
	}{
		{
			name:        "successful assets retrieval",
			queryParams: map[string]string{},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "exchange", "currency", "sector", "created_at"}).
					AddRow("1", "AAPL", "Apple Inc.", "STOCK", "NASDAQ", "USD", "Technology", "2024-01-01T00:00:00Z").
					AddRow("2", "GOOGL", "Alphabet Inc.", "STOCK", "NASDAQ", "USD", "Technology", "2024-01-01T00:00:00Z")

				mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at FROM assets WHERE 1=1 ORDER BY symbol ASC LIMIT \\$1").
					WithArgs("50").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"AAPL", "Apple Inc.", "GOOGL", "Alphabet Inc.", "Technology", "NASDAQ"},
		},
		{
			name:        "assets filtered by type",
			queryParams: map[string]string{"type": "STOCK"},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "exchange", "currency", "sector", "created_at"}).
					AddRow("1", "AAPL", "Apple Inc.", "STOCK", "NASDAQ", "USD", "Technology", "2024-01-01T00:00:00Z")

				mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at FROM assets WHERE 1=1 AND asset_type = \\$1 ORDER BY symbol ASC LIMIT \\$2").
					WithArgs("STOCK", "50").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"AAPL", "STOCK"},
		},
		{
			name:        "assets with search query",
			queryParams: map[string]string{"search": "Apple"},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "exchange", "currency", "sector", "created_at"}).
					AddRow("1", "AAPL", "Apple Inc.", "STOCK", "NASDAQ", "USD", "Technology", "2024-01-01T00:00:00Z")

				mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at FROM assets WHERE 1=1 AND \\(symbol ILIKE \\$1 OR name ILIKE \\$1\\) ORDER BY symbol ASC LIMIT \\$2").
					WithArgs("%Apple%", "50").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"AAPL", "Apple Inc."},
		},
		{
			name:        "assets with custom limit",
			queryParams: map[string]string{"limit": "10"},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "exchange", "currency", "sector", "created_at"}).
					AddRow("1", "AAPL", "Apple Inc.", "STOCK", "NASDAQ", "USD", "Technology", "2024-01-01T00:00:00Z")

				mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at FROM assets WHERE 1=1 ORDER BY symbol ASC LIMIT \\$1").
					WithArgs("10").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"AAPL"},
		},
		{
			name:        "empty assets result",
			queryParams: map[string]string{},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "exchange", "currency", "sector", "created_at"})

				mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at FROM assets WHERE 1=1 ORDER BY symbol ASC LIMIT \\$1").
					WithArgs("50").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"\"total\":0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			logger, _ := zap.NewDevelopment()

			// Create mock database
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			mockServices := &services.Services{
				DB:     db,
				Logger: logger,
			}

			handler := NewHandler(mockServices, logger)

			tt.setupMock(mock)

			router := gin.New()
			router.GET("/assets", handler.GetAssets)

			url := "/assets"
			if len(tt.queryParams) > 0 {
				url += "?"
				first := true
				for key, value := range tt.queryParams {
					if !first {
						url += "&"
					}
					url += key + "=" + value
					first = false
				}
			}

			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			for _, expectedStr := range tt.expectedBody {
				assert.Contains(t, w.Body.String(), expectedStr)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetAssets_NilDB(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		DB:     nil,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	router := gin.New()
	router.GET("/assets", handler.GetAssets)

	req, _ := http.NewRequest("GET", "/assets", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to fetch assets")
}

func TestGetAssets_DatabaseError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at FROM assets").
		WillReturnError(sql.ErrConnDone)

	router := gin.New()
	router.GET("/assets", handler.GetAssets)

	req, _ := http.NewRequest("GET", "/assets", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to fetch assets")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// GetAsset handler tests
func TestGetAsset_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "exchange", "currency", "sector", "created_at", "updated_at"}).
		AddRow("1", "AAPL", "Apple Inc.", "STOCK", "NASDAQ", "USD", "Technology", "2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z")

	mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at, updated_at FROM assets WHERE symbol = \\$1").
		WithArgs("AAPL").
		WillReturnRows(rows)

	// Mock market data query (optional)
	mock.ExpectQuery("SELECT price, change_24h, timestamp FROM market_data WHERE asset_id = \\$1 ORDER BY timestamp DESC LIMIT 1").
		WithArgs("1").
		WillReturnError(sql.ErrNoRows) // No market data available

	router := gin.New()
	router.GET("/assets/:symbol", handler.GetAsset)

	req, _ := http.NewRequest("GET", "/assets/AAPL", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "AAPL")
	assert.Contains(t, w.Body.String(), "Apple Inc.")
	assert.Contains(t, w.Body.String(), "Technology")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAsset_NotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at, updated_at FROM assets WHERE symbol = \\$1").
		WithArgs("NONEXISTENT").
		WillReturnError(sql.ErrNoRows)

	router := gin.New()
	router.GET("/assets/:symbol", handler.GetAsset)

	req, _ := http.NewRequest("GET", "/assets/NONEXISTENT", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Asset not found")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// GetCurrentPrice handler tests
func TestGetCurrentPrice_ServiceUnavailable(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		Finnhub: nil, // No Finnhub service
		Logger:  logger,
	}

	handler := NewHandler(mockServices, logger)

	router := gin.New()
	router.GET("/assets/:symbol/price", handler.GetCurrentPrice)

	req, _ := http.NewRequest("GET", "/assets/AAPL/price", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "Market data service not available")
}

// Mock Finnhub for testing
type Quote struct {
	CurrentPrice  float64
	Change        float64
	PercentChange float64
}

type MockFinnhub struct {
	QuoteData map[string]*Quote
}

func (m *MockFinnhub) GetQuote(symbol string) (*Quote, error) {
	if quote, exists := m.QuoteData[symbol]; exists {
		return quote, nil
	}
	return nil, fmt.Errorf("symbol not found")
}

func TestGetCurrentPrice_EmptySymbol(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	router := gin.New()
	router.GET("/assets/:symbol/price", handler.GetCurrentPrice)

	req, _ := http.NewRequest("GET", "/assets//price", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Symbol is required")
}

// GetPriceHistory handler tests
func TestGetPriceHistory_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Mock asset ID lookup first
	mock.ExpectQuery("SELECT id FROM assets WHERE symbol = \\$1").
		WithArgs("AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	// Mock price history query
	rows := sqlmock.NewRows([]string{"date", "open", "high", "low", "close", "volume"}).
		AddRow("2024-01-01T00:00:00Z", 148.0, 152.0, 147.0, 150.0, 1000000).
		AddRow("2024-01-02T00:00:00Z", 150.0, 154.0, 149.0, 152.0, 1100000)

	mock.ExpectQuery("SELECT (.+) FROM price_history").
		WillReturnRows(rows)

	router := gin.New()
	router.GET("/assets/:symbol/history", handler.GetPriceHistory)

	req, _ := http.NewRequest("GET", "/assets/AAPL/history?period=30d&interval=1d", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "AAPL")
	assert.Contains(t, w.Body.String(), "history")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPriceHistory_DefaultParameters(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Mock asset ID lookup first
	mock.ExpectQuery("SELECT id FROM assets WHERE symbol = \\$1").
		WithArgs("AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	// Mock price history query
	rows := sqlmock.NewRows([]string{"date", "open", "high", "low", "close", "volume"}).
		AddRow("2024-01-01T00:00:00Z", 148.0, 152.0, 147.0, 150.0, 1000000)

	mock.ExpectQuery("SELECT (.+) FROM price_history").
		WillReturnRows(rows)

	router := gin.New()
	router.GET("/assets/:symbol/history", handler.GetPriceHistory)

	req, _ := http.NewRequest("GET", "/assets/AAPL/history", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "AAPL")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// GetPerformanceAnalytics handler tests
func TestGetPerformanceAnalytics_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

	// Mock portfolio totals query
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COUNT\\(\\*\\) as total_holdings FROM portfolio_holdings ph WHERE ph.user_id = \\$1").
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"total_cost", "total_holdings"}).AddRow(3000.0, 2))

	// Mock holdings query for market value calculation
	holdingsRows := sqlmock.NewRows([]string{"symbol", "quantity", "average_cost"}).
		AddRow("AAPL", 10.0, 150.0).
		AddRow("GOOGL", 5.0, 2500.0)

	mock.ExpectQuery("SELECT (.+) FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = \\$1").
		WithArgs("user1").
		WillReturnRows(holdingsRows)

	// Mock snapshots query for historical data
	snapshotsRows := sqlmock.NewRows([]string{"snapshot_date", "total_value", "total_cost", "unrealized_pnl"}).
		AddRow("2024-01-01", 2500.0, 2000.0, 500.0).
		AddRow("2024-01-15", 2750.0, 2000.0, 750.0)

	mock.ExpectQuery("SELECT (.+) FROM portfolio_snapshots WHERE user_id = \\$1").
		WithArgs("user1").
		WillReturnRows(snapshotsRows)

	// Mock top performers query
	topPerformersRows := sqlmock.NewRows([]string{"symbol", "name", "quantity", "average_cost", "total_value"}).
		AddRow("GOOGL", "Alphabet Inc.", 5.0, 2500.0, 12500.0).
		AddRow("AAPL", "Apple Inc.", 10.0, 150.0, 1500.0)

	mock.ExpectQuery("SELECT (.+) FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = \\$1 ORDER BY \\(ph.quantity \\* ph.average_cost\\) DESC LIMIT 5").
		WithArgs("user1").
		WillReturnRows(topPerformersRows)

	router := gin.New()
	router.GET("/analytics/performance", handler.GetPerformanceAnalytics)

	req, _ := http.NewRequest("GET", "/analytics/performance?period=30d", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "portfolio_performance")
	assert.Contains(t, w.Body.String(), "total_value")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPerformanceAnalytics_NilDB(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		DB:     nil,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	router := gin.New()
	router.GET("/analytics/performance", handler.GetPerformanceAnalytics)

	req, _ := http.NewRequest("GET", "/analytics/performance", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to fetch performance analytics")
}

// GetRiskMetrics handler tests
func TestGetRiskMetrics_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

	// Mock portfolio holdings for risk calculations
	holdingsRows := sqlmock.NewRows([]string{"sector", "holdings_count", "sector_value"}).
		AddRow("Technology", 2, 4390.0).
		AddRow("Financial", 1, 750.0)

	mock.ExpectQuery("SELECT a.sector, COUNT\\(\\*\\) as holdings_count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as sector_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = \\$1 AND a.sector IS NOT NULL GROUP BY a.sector ORDER BY sector_value DESC").
		WithArgs("user1").
		WillReturnRows(holdingsRows)

	// Mock additional query for beta calculation
	betaRows := sqlmock.NewRows([]string{"symbol", "quantity", "average_cost", "position_value"}).
		AddRow("AAPL", 10.0, 150.0, 1500.0).
		AddRow("MSFT", 8.0, 300.0, 2400.0).
		AddRow("JPM", 5.0, 140.0, 700.0)

	mock.ExpectQuery("SELECT a.symbol, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as position_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = \\$1").
		WithArgs("user1").
		WillReturnRows(betaRows)

	router := gin.New()
	router.GET("/analytics/risk", handler.GetRiskMetrics)

	req, _ := http.NewRequest("GET", "/analytics/risk", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "diversification")
	assert.Contains(t, w.Body.String(), "sector_diversification")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRiskMetrics_EmptyPortfolio(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

	// Mock empty portfolio holdings
	holdingsRows := sqlmock.NewRows([]string{"sector", "holdings_count", "sector_value"})

	mock.ExpectQuery("SELECT a.sector, COUNT\\(\\*\\) as holdings_count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as sector_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = \\$1 AND a.sector IS NOT NULL GROUP BY a.sector ORDER BY sector_value DESC").
		WithArgs("user1").
		WillReturnRows(holdingsRows)

	// Mock empty beta calculation query
	betaRows := sqlmock.NewRows([]string{"symbol", "quantity", "average_cost", "position_value"})

	mock.ExpectQuery("SELECT a.symbol, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as position_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = \\$1").
		WithArgs("user1").
		WillReturnRows(betaRows)

	router := gin.New()
	router.GET("/analytics/risk", handler.GetRiskMetrics)

	req, _ := http.NewRequest("GET", "/analytics/risk", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "diversification")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// GetAssetAllocation handler tests
func TestGetAssetAllocation_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

	// Mock portfolio holdings for allocation
	assetTypeRows := sqlmock.NewRows([]string{"asset_type", "count", "total_value"}).
		AddRow("STOCK", 2, 45250.0).
		AddRow("CRYPTO", 1, 30000.0)

	mock.ExpectQuery("SELECT a.asset_type, COUNT\\(\\*\\) as count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = \\$1 GROUP BY a.asset_type ORDER BY total_value DESC").
		WithArgs("user1").
		WillReturnRows(assetTypeRows)

	// Mock sector allocation query
	sectorRows := sqlmock.NewRows([]string{"sector", "count", "total_value"}).
		AddRow("Technology", 2, 45250.0).
		AddRow("Cryptocurrency", 1, 30000.0)

	mock.ExpectQuery("SELECT COALESCE\\(a.sector, 'Unknown'\\) as sector, COUNT\\(\\*\\) as count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = \\$1 GROUP BY a.sector ORDER BY total_value DESC").
		WithArgs("user1").
		WillReturnRows(sectorRows)

	// Mock top holdings query
	topHoldingsRows := sqlmock.NewRows([]string{"symbol", "name", "quantity", "average_cost", "total_value"}).
		AddRow("BTC-USD", "Bitcoin", 0.5, 50000.0, 30000.0).
		AddRow("GOOGL", "Alphabet Inc.", 5.0, 2500.0, 13500.0).
		AddRow("AAPL", "Apple Inc.", 10.0, 150.0, 1750.0)

	mock.ExpectQuery("SELECT a.symbol, a.name, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = \\$1 ORDER BY \\(ph.quantity \\* ph.average_cost\\) DESC LIMIT 10").
		WithArgs("user1").
		WillReturnRows(topHoldingsRows)

	router := gin.New()
	router.GET("/analytics/allocation", handler.GetAssetAllocation)

	req, _ := http.NewRequest("GET", "/analytics/allocation", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "by_asset_type")
	assert.Contains(t, w.Body.String(), "by_sector")
	assert.Contains(t, w.Body.String(), "top_holdings")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// WhatIfAnalysis handler tests
func TestWhatIfAnalysis_BuyScenario(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

	// Mock current portfolio totals query
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COUNT\\(\\*\\) as total_holdings FROM portfolio_holdings ph WHERE ph.user_id = \\$1").
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"total_cost", "total_holdings"}).AddRow(3400.0, 2))

	// Mock current holding check for GOOGL
	mock.ExpectQuery("SELECT ph.quantity, ph.average_cost FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = \\$1 AND a.symbol = \\$2").
		WithArgs("user1", "GOOGL").
		WillReturnError(sql.ErrNoRows) // No existing GOOGL position

	// Mock allocation query
	allocationRows := sqlmock.NewRows([]string{"asset_type", "total_value"}).
		AddRow("STOCK", 3400.0)

	mock.ExpectQuery("SELECT a.asset_type, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = \\$1 GROUP BY a.asset_type").
		WithArgs("user1").
		WillReturnRows(allocationRows)

	router := gin.New()
	router.POST("/analytics/what-if", handler.WhatIfAnalysis)

	requestBody := map[string]interface{}{
		"action":   "buy",
		"symbol":   "GOOGL",
		"quantity": 5.0,
		"price":    2700.0,
	}
	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/analytics/what-if", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "trade_details")
	assert.Contains(t, w.Body.String(), "allocation_impact")
	assert.Contains(t, w.Body.String(), "risk_impact")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWhatIfAnalysis_SellScenario(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

	// Mock current portfolio totals query
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COUNT\\(\\*\\) as total_holdings FROM portfolio_holdings ph WHERE ph.user_id = \\$1").
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"total_cost", "total_holdings"}).AddRow(3400.0, 2))

	// Mock current holding check for AAPL (existing position)
	mock.ExpectQuery("SELECT ph.quantity, ph.average_cost FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = \\$1 AND a.symbol = \\$2").
		WithArgs("user1", "AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"quantity", "average_cost"}).AddRow(10.0, 150.0))

	// Mock allocation query
	allocationRows := sqlmock.NewRows([]string{"asset_type", "total_value"}).
		AddRow("STOCK", 3400.0)

	mock.ExpectQuery("SELECT a.asset_type, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = \\$1 GROUP BY a.asset_type").
		WithArgs("user1").
		WillReturnRows(allocationRows)

	router := gin.New()
	router.POST("/analytics/what-if", handler.WhatIfAnalysis)

	requestBody := map[string]interface{}{
		"action":   "sell",
		"symbol":   "AAPL",
		"quantity": 5.0,
		"price":    175.0,
	}
	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/analytics/what-if", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "trade_details")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWhatIfAnalysis_ValidationError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	router := gin.New()
	router.POST("/analytics/what-if", handler.WhatIfAnalysis)

	// Missing required fields
	requestBody := map[string]interface{}{
		"symbol": "AAPL",
	}
	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/analytics/what-if", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}

func TestWhatIfAnalysis_InvalidJSON(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	router := gin.New()
	router.POST("/analytics/what-if", handler.WhatIfAnalysis)

	req, _ := http.NewRequest("POST", "/analytics/what-if", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid character")
}
