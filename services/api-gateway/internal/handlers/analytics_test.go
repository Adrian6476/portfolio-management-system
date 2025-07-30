package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/portfolio-management/api-gateway/internal/services"
)

// MockFinnhubClient implements a mock Finnhub client for testing
type MockFinnhubClient struct {
	quotes   map[string]*services.FinnhubQuote
	profiles map[string]*services.FinnhubProfile
	errors   map[string]error
}

func NewMockFinnhubClient() *MockFinnhubClient {
	return &MockFinnhubClient{
		quotes:   make(map[string]*services.FinnhubQuote),
		profiles: make(map[string]*services.FinnhubProfile),
		errors:   make(map[string]error),
	}
}

func (m *MockFinnhubClient) SetQuote(symbol string, quote *services.FinnhubQuote) {
	m.quotes[symbol] = quote
}

func (m *MockFinnhubClient) SetProfile(symbol string, profile *services.FinnhubProfile) {
	m.profiles[symbol] = profile
}

func (m *MockFinnhubClient) SetError(symbol string, err error) {
	m.errors[symbol] = err
}

func (m *MockFinnhubClient) GetQuote(symbol string) (*services.FinnhubQuote, error) {
	if err, exists := m.errors[symbol]; exists {
		return nil, err
	}
	if quote, exists := m.quotes[symbol]; exists {
		return quote, nil
	}
	return nil, fmt.Errorf("quote not found for symbol: %s", symbol)
}

func (m *MockFinnhubClient) GetCompanyProfile(symbol string) (*services.FinnhubProfile, error) {
	if err, exists := m.errors[symbol]; exists {
		return nil, err
	}
	if profile, exists := m.profiles[symbol]; exists {
		return profile, nil
	}
	return nil, fmt.Errorf("profile not found for symbol: %s", symbol)
}

// Test helper functions
func setupTestHandler(db *sql.DB, finnhub *MockFinnhubClient) *Handler {
	logger, _ := zap.NewDevelopment()
	mockServices := &services.Services{
		DB:      db,
		Finnhub: finnhub,
		Logger:  logger,
	}
	return NewHandler(mockServices, logger)
}

func setupTestRouter(handler *Handler, method, path string, handlerFunc gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	switch method {
	case "GET":
		router.GET(path, handlerFunc)
	case "POST":
		router.POST(path, handlerFunc)
	case "PUT":
		router.PUT(path, handlerFunc)
	case "DELETE":
		router.DELETE(path, handlerFunc)
	}
	return router
}

// =============================================================================
// Market Data Handler Tests
// =============================================================================

func TestHandler_GetAssets_Success(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock successful query
	rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "exchange", "currency", "sector", "created_at"}).
		AddRow("1", "AAPL", "Apple Inc.", "STOCK", "NASDAQ", "USD", "Technology", "2024-01-01").
		AddRow("2", "GOOGL", "Alphabet Inc.", "STOCK", "NASDAQ", "USD", "Technology", "2024-01-02").
		AddRow("3", "MSFT", "Microsoft Corp.", "STOCK", "NASDAQ", "USD", "Technology", "2024-01-03")

	mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at FROM assets WHERE 1=1 ORDER BY symbol ASC LIMIT (.+)").
		WithArgs("50").
		WillReturnRows(rows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/assets", handler.GetAssets)

	// Test request
	req, _ := http.NewRequest("GET", "/assets", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "assets")
	assert.Contains(t, w.Body.String(), "AAPL")
	assert.Contains(t, w.Body.String(), "GOOGL")
	assert.Contains(t, w.Body.String(), "MSFT")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetAssets_WithFilters(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock query with type filter
	rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "exchange", "currency", "sector", "created_at"}).
		AddRow("1", "AAPL", "Apple Inc.", "STOCK", "NASDAQ", "USD", "Technology", "2024-01-01")

	mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at FROM assets WHERE 1=1 AND asset_type = (.+) ORDER BY symbol ASC LIMIT (.+)").
		WithArgs("STOCK", "50").
		WillReturnRows(rows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/assets", handler.GetAssets)

	// Test request with type filter
	req, _ := http.NewRequest("GET", "/assets?type=STOCK", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "AAPL")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetAssets_WithSearch(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock query with search filter
	rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "exchange", "currency", "sector", "created_at"}).
		AddRow("1", "AAPL", "Apple Inc.", "STOCK", "NASDAQ", "USD", "Technology", "2024-01-01")

	mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at FROM assets WHERE 1=1 AND \\(symbol ILIKE (.+) OR name ILIKE (.+)\\) ORDER BY symbol ASC LIMIT (.+)").
		WithArgs("%Apple%", "50").
		WillReturnRows(rows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/assets", handler.GetAssets)

	// Test request with search filter
	req, _ := http.NewRequest("GET", "/assets?search=Apple", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "AAPL")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetAssets_EmptyResults(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock empty result
	rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "exchange", "currency", "sector", "created_at"})
	mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at FROM assets WHERE 1=1 ORDER BY symbol ASC LIMIT (.+)").
		WithArgs("50").
		WillReturnRows(rows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/assets", handler.GetAssets)

	// Test request
	req, _ := http.NewRequest("GET", "/assets", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total":0`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetAssets_DatabaseError(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock database error
	mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at FROM assets WHERE 1=1 ORDER BY symbol ASC LIMIT (.+)").
		WithArgs("50").
		WillReturnError(fmt.Errorf("database connection failed"))

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/assets", handler.GetAssets)

	// Test request
	req, _ := http.NewRequest("GET", "/assets", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to fetch assets")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetAssets_NilDB(t *testing.T) {
	handler := setupTestHandler(nil, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/assets", handler.GetAssets)

	// Test request
	req, _ := http.NewRequest("GET", "/assets", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to fetch assets")
}

func TestHandler_GetAsset_Success(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock asset query
	mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at, updated_at FROM assets WHERE symbol = (.+)").
		WithArgs("AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "exchange", "currency", "sector", "created_at", "updated_at"}).
			AddRow("1", "AAPL", "Apple Inc.", "STOCK", "NASDAQ", "USD", "Technology", "2024-01-01", "2024-01-01"))

	// Mock market data query
	mock.ExpectQuery("SELECT price, change_24h, timestamp FROM market_data WHERE asset_id = (.+) ORDER BY timestamp DESC LIMIT 1").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"price", "change_24h", "timestamp"}).
			AddRow(150.50, 2.50, "2024-01-01 10:00:00"))

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/assets/:symbol", handler.GetAsset)

	// Test request
	req, _ := http.NewRequest("GET", "/assets/AAPL", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "AAPL")
	assert.Contains(t, w.Body.String(), "Apple Inc.")
	assert.Contains(t, w.Body.String(), "current_price")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetAsset_NotFound(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock asset not found
	mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at, updated_at FROM assets WHERE symbol = (.+)").
		WithArgs("INVALID").
		WillReturnError(sql.ErrNoRows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/assets/:symbol", handler.GetAsset)

	// Test request
	req, _ := http.NewRequest("GET", "/assets/INVALID", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Asset not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetAsset_MissingSymbol(t *testing.T) {
	handler := setupTestHandler(nil, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/assets/:symbol", handler.GetAsset)

	// Test request with empty symbol
	req, _ := http.NewRequest("GET", "/assets/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code) // Route doesn't match
}

func TestHandler_GetCurrentPrice_Success(t *testing.T) {
	// Setup mock Finnhub client
	mockFinnhub := NewMockFinnhubClient()
	mockFinnhub.SetQuote("AAPL", &services.FinnhubQuote{
		CurrentPrice:       150.50,
		Change:             2.50,
		PercentChange:      1.69,
		HighPriceOfDay:     152.00,
		LowPriceOfDay:      148.00,
		OpenPriceOfDay:     149.00,
		PreviousClosePrice: 148.00,
		Timestamp:          1640995200,
	})

	handler := setupTestHandler(nil, mockFinnhub)
	router := setupTestRouter(handler, "GET", "/price/:symbol", handler.GetCurrentPrice)

	// Test request
	req, _ := http.NewRequest("GET", "/price/AAPL", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "AAPL")
	assert.Contains(t, w.Body.String(), "150.5")
	assert.Contains(t, w.Body.String(), "current_price")
	assert.Contains(t, w.Body.String(), "change")
	assert.Contains(t, w.Body.String(), "change_percent")
}

func TestHandler_GetCurrentPrice_FinnhubError(t *testing.T) {
	// Setup mock Finnhub client with error
	mockFinnhub := NewMockFinnhubClient()
	mockFinnhub.SetError("AAPL", fmt.Errorf("API rate limit exceeded"))

	handler := setupTestHandler(nil, mockFinnhub)
	router := setupTestRouter(handler, "GET", "/price/:symbol", handler.GetCurrentPrice)

	// Test request
	req, _ := http.NewRequest("GET", "/price/AAPL", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to fetch current price")
}

func TestHandler_GetCurrentPrice_NoFinnhub(t *testing.T) {
	handler := setupTestHandler(nil, nil)
	router := setupTestRouter(handler, "GET", "/price/:symbol", handler.GetCurrentPrice)

	// Test request
	req, _ := http.NewRequest("GET", "/price/AAPL", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "Market data service not available")
}

func TestHandler_GetCurrentPrice_MissingSymbol(t *testing.T) {
	handler := setupTestHandler(nil, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/price/:symbol", handler.GetCurrentPrice)

	// Test request with empty symbol
	req, _ := http.NewRequest("GET", "/price/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code) // Route doesn't match
}

func TestHandler_GetPriceHistory_Success(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock asset ID query
	mock.ExpectQuery("SELECT id FROM assets WHERE symbol = (.+)").
		WithArgs("AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	// Mock price history query
	historyRows := sqlmock.NewRows([]string{"date", "open_price", "high_price", "low_price", "close_price", "volume"}).
		AddRow("2024-01-03", 150.00, 152.00, 148.00, 151.00, 1000000).
		AddRow("2024-01-02", 148.00, 150.00, 147.00, 149.00, 1200000).
		AddRow("2024-01-01", 147.00, 149.00, 146.00, 148.00, 1100000)

	mock.ExpectQuery("SELECT date, open_price, high_price, low_price, close_price, volume FROM price_history WHERE asset_id = (.+) AND date >= CURRENT_DATE - INTERVAL '30 days' ORDER BY date DESC LIMIT (.+)").
		WithArgs("1", "100").
		WillReturnRows(historyRows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/price-history/:symbol", handler.GetPriceHistory)

	// Test request
	req, _ := http.NewRequest("GET", "/price-history/AAPL", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "AAPL")
	assert.Contains(t, w.Body.String(), "price_history")
	assert.Contains(t, w.Body.String(), "total_points")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetPriceHistory_AssetNotFound(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock asset not found
	mock.ExpectQuery("SELECT id FROM assets WHERE symbol = (.+)").
		WithArgs("INVALID").
		WillReturnError(sql.ErrNoRows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/price-history/:symbol", handler.GetPriceHistory)

	// Test request
	req, _ := http.NewRequest("GET", "/price-history/INVALID", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Asset not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetPriceHistory_WithPeriodFilter(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock asset ID query
	mock.ExpectQuery("SELECT id FROM assets WHERE symbol = (.+)").
		WithArgs("AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	// Mock price history query with 7d period
	historyRows := sqlmock.NewRows([]string{"date", "open_price", "high_price", "low_price", "close_price", "volume"}).
		AddRow("2024-01-07", 150.00, 152.00, 148.00, 151.00, 1000000)

	mock.ExpectQuery("SELECT date, open_price, high_price, low_price, close_price, volume FROM price_history WHERE asset_id = (.+) AND date >= CURRENT_DATE - INTERVAL '7 days' ORDER BY date DESC LIMIT (.+)").
		WithArgs("1", "100").
		WillReturnRows(historyRows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/price-history/:symbol", handler.GetPriceHistory)

	// Test request with period filter
	req, _ := http.NewRequest("GET", "/price-history/AAPL?period=7d", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"period":"7d"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetPriceHistory_EmptyHistory(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock asset ID query
	mock.ExpectQuery("SELECT id FROM assets WHERE symbol = (.+)").
		WithArgs("AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	// Mock empty price history
	historyRows := sqlmock.NewRows([]string{"date", "open_price", "high_price", "low_price", "close_price", "volume"})
	mock.ExpectQuery("SELECT date, open_price, high_price, low_price, close_price, volume FROM price_history WHERE asset_id = (.+) AND date >= CURRENT_DATE - INTERVAL '30 days' ORDER BY date DESC LIMIT (.+)").
		WithArgs("1", "100").
		WillReturnRows(historyRows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/price-history/:symbol", handler.GetPriceHistory)

	// Test request
	req, _ := http.NewRequest("GET", "/price-history/AAPL", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total_points":0`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// =============================================================================
// Analytics Handler Tests
// =============================================================================

func TestHandler_GetPerformanceAnalytics_Success(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Setup mock Finnhub client
	mockFinnhub := NewMockFinnhubClient()
	mockFinnhub.SetQuote("AAPL", &services.FinnhubQuote{CurrentPrice: 155.00})
	mockFinnhub.SetQuote("GOOGL", &services.FinnhubQuote{CurrentPrice: 2850.00})

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	// Mock portfolio totals query
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COUNT\\(\\*\\) as total_holdings FROM portfolio_holdings ph WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"total_cost", "total_holdings"}).AddRow(15500.0, 2))

	// Mock holdings query for market value calculation
	holdingsRows := sqlmock.NewRows([]string{"symbol", "quantity", "average_cost"}).
		AddRow("AAPL", 10.0, 150.0).
		AddRow("GOOGL", 5.0, 2800.0)
	mock.ExpectQuery("SELECT a.symbol, ph.quantity, ph.average_cost FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(holdingsRows)

	// Mock snapshots query
	snapshotRows := sqlmock.NewRows([]string{"snapshot_date", "total_value", "total_cost", "unrealized_pnl"}).
		AddRow("2024-01-01", 15000.0, 15500.0, -500.0).
		AddRow("2024-01-02", 15800.0, 15500.0, 300.0)
	mock.ExpectQuery("SELECT snapshot_date, total_value, total_cost, unrealized_pnl FROM portfolio_snapshots WHERE user_id = (.+) AND snapshot_date >= CURRENT_DATE - INTERVAL '30 days' ORDER BY snapshot_date DESC LIMIT 30").
		WithArgs("user-123").
		WillReturnRows(snapshotRows)

	// Mock top performers query
	performerRows := sqlmock.NewRows([]string{"symbol", "name", "quantity", "average_cost", "total_value"}).
		AddRow("GOOGL", "Alphabet Inc.", 5.0, 2800.0, 14000.0).
		AddRow("AAPL", "Apple Inc.", 10.0, 150.0, 1500.0)
	mock.ExpectQuery("SELECT a.symbol, a.name, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) ORDER BY \\(ph.quantity \\* ph.average_cost\\) DESC LIMIT 5").
		WithArgs("user-123").
		WillReturnRows(performerRows)

	handler := setupTestHandler(db, mockFinnhub)
	router := setupTestRouter(handler, "GET", "/analytics/performance", handler.GetPerformanceAnalytics)

	// Test request
	req, _ := http.NewRequest("GET", "/analytics/performance", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "portfolio_performance")
	assert.Contains(t, w.Body.String(), "historical_snapshots")
	assert.Contains(t, w.Body.String(), "top_performers")
	assert.Contains(t, w.Body.String(), "total_cost")
	assert.Contains(t, w.Body.String(), "current_value")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetPerformanceAnalytics_WithPriceErrors(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Setup mock Finnhub client with errors
	mockFinnhub := NewMockFinnhubClient()
	mockFinnhub.SetError("AAPL", fmt.Errorf("API error"))
	mockFinnhub.SetQuote("GOOGL", &services.FinnhubQuote{CurrentPrice: 2850.00})

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	// Mock portfolio totals query
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COUNT\\(\\*\\) as total_holdings FROM portfolio_holdings ph WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"total_cost", "total_holdings"}).AddRow(15500.0, 2))

	// Mock holdings query
	holdingsRows := sqlmock.NewRows([]string{"symbol", "quantity", "average_cost"}).
		AddRow("AAPL", 10.0, 150.0).
		AddRow("GOOGL", 5.0, 2800.0)
	mock.ExpectQuery("SELECT a.symbol, ph.quantity, ph.average_cost FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(holdingsRows)

	// Mock empty
	// Mock empty snapshots query
	snapshotRows := sqlmock.NewRows([]string{"snapshot_date", "total_value", "total_cost", "unrealized_pnl"})
	mock.ExpectQuery("SELECT snapshot_date, total_value, total_cost, unrealized_pnl FROM portfolio_snapshots WHERE user_id = (.+) AND snapshot_date >= CURRENT_DATE - INTERVAL '30 days' ORDER BY snapshot_date DESC LIMIT 30").
		WithArgs("user-123").
		WillReturnRows(snapshotRows)

	// Mock top performers query
	performerRows := sqlmock.NewRows([]string{"symbol", "name", "quantity", "average_cost", "total_value"}).
		AddRow("GOOGL", "Alphabet Inc.", 5.0, 2800.0, 14000.0).
		AddRow("AAPL", "Apple Inc.", 10.0, 150.0, 1500.0)
	mock.ExpectQuery("SELECT a.symbol, a.name, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) ORDER BY \\(ph.quantity \\* ph.average_cost\\) DESC LIMIT 5").
		WithArgs("user-123").
		WillReturnRows(performerRows)

	handler := setupTestHandler(db, mockFinnhub)
	router := setupTestRouter(handler, "GET", "/analytics/performance", handler.GetPerformanceAnalytics)

	// Test request
	req, _ := http.NewRequest("GET", "/analytics/performance", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "warnings")
	assert.Contains(t, w.Body.String(), "Could not fetch price for AAPL")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetPerformanceAnalytics_NilDB(t *testing.T) {
	handler := setupTestHandler(nil, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/analytics/performance", handler.GetPerformanceAnalytics)

	// Test request
	req, _ := http.NewRequest("GET", "/analytics/performance", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to fetch performance analytics")
}

func TestHandler_GetRiskMetrics_Success(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Setup mock Finnhub client
	mockFinnhub := NewMockFinnhubClient()
	mockFinnhub.SetQuote("AAPL", &services.FinnhubQuote{CurrentPrice: 155.00})
	mockFinnhub.SetQuote("GOOGL", &services.FinnhubQuote{CurrentPrice: 2850.00})

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	// Mock diversification query
	diversificationRows := sqlmock.NewRows([]string{"sector", "holdings_count", "sector_value"}).
		AddRow("Technology", 2, 15500.0).
		AddRow("Healthcare", 1, 5000.0)
	mock.ExpectQuery("SELECT a.sector, COUNT\\(\\*\\) as holdings_count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as sector_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) AND a.sector IS NOT NULL GROUP BY a.sector ORDER BY sector_value DESC").
		WithArgs("user-123").
		WillReturnRows(diversificationRows)

	// Mock beta calculation query
	betaRows := sqlmock.NewRows([]string{"symbol", "quantity", "average_cost", "position_value"}).
		AddRow("AAPL", 10.0, 150.0, 1500.0).
		AddRow("GOOGL", 5.0, 2800.0, 14000.0)
	mock.ExpectQuery("SELECT a.symbol, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as position_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(betaRows)

	// Mock current value calculation query (duplicate for current portfolio value calculation)
	currentValueRows := sqlmock.NewRows([]string{"symbol", "quantity", "average_cost", "position_value"}).
		AddRow("AAPL", 10.0, 150.0, 1500.0).
		AddRow("GOOGL", 5.0, 2800.0, 14000.0)
	mock.ExpectQuery("SELECT a.symbol, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as position_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(currentValueRows)

	handler := setupTestHandler(db, mockFinnhub)
	router := setupTestRouter(handler, "GET", "/analytics/risk", handler.GetRiskMetrics)

	// Test request
	req, _ := http.NewRequest("GET", "/analytics/risk", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "risk_assessment")
	assert.Contains(t, w.Body.String(), "sector_diversification")
	assert.Contains(t, w.Body.String(), "volatility_metrics")
	assert.Contains(t, w.Body.String(), "herfindahl_index")
	assert.Contains(t, w.Body.String(), "portfolio_beta")
	assert.Contains(t, w.Body.String(), "sharpe_ratio")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetRiskMetrics_HighConcentration(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	// Mock high concentration in single sector
	diversificationRows := sqlmock.NewRows([]string{"sector", "holdings_count", "sector_value"}).
		AddRow("Technology", 5, 95000.0).
		AddRow("Healthcare", 1, 5000.0)
	mock.ExpectQuery("SELECT a.sector, COUNT\\(\\*\\) as holdings_count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as sector_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) AND a.sector IS NOT NULL GROUP BY a.sector ORDER BY sector_value DESC").
		WithArgs("user-123").
		WillReturnRows(diversificationRows)

	// Mock beta calculation query
	betaRows := sqlmock.NewRows([]string{"symbol", "quantity", "average_cost", "position_value"})
	mock.ExpectQuery("SELECT a.symbol, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as position_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(betaRows)

	// Mock current value calculation query
	currentValueRows := sqlmock.NewRows([]string{"symbol", "quantity", "average_cost", "position_value"})
	mock.ExpectQuery("SELECT a.symbol, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as position_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(currentValueRows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/analytics/risk", handler.GetRiskMetrics)

	// Test request
	req, _ := http.NewRequest("GET", "/analytics/risk", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "High")
	assert.Contains(t, w.Body.String(), "concentrated portfolio")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetRiskMetrics_DatabaseError(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	// Mock database error
	mock.ExpectQuery("SELECT a.sector, COUNT\\(\\*\\) as holdings_count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as sector_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) AND a.sector IS NOT NULL GROUP BY a.sector ORDER BY sector_value DESC").
		WithArgs("user-123").
		WillReturnError(fmt.Errorf("database connection failed"))

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/analytics/risk", handler.GetRiskMetrics)

	// Test request
	req, _ := http.NewRequest("GET", "/analytics/risk", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to fetch risk metrics")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetAssetAllocation_Success(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	// Mock asset type allocation query
	assetTypeRows := sqlmock.NewRows([]string{"asset_type", "count", "total_value"}).
		AddRow("STOCK", 3, 15500.0).
		AddRow("BOND", 1, 5000.0)
	mock.ExpectQuery("SELECT a.asset_type, COUNT\\(\\*\\) as count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) GROUP BY a.asset_type ORDER BY total_value DESC").
		WithArgs("user-123").
		WillReturnRows(assetTypeRows)

	// Mock sector allocation query
	sectorRows := sqlmock.NewRows([]string{"sector", "count", "total_value"}).
		AddRow("Technology", 2, 12000.0).
		AddRow("Healthcare", 1, 3500.0).
		AddRow("Unknown", 1, 5000.0)
	mock.ExpectQuery("SELECT COALESCE\\(a.sector, 'Unknown'\\) as sector, COUNT\\(\\*\\) as count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) GROUP BY a.sector ORDER BY total_value DESC").
		WithArgs("user-123").
		WillReturnRows(sectorRows)

	// Mock top holdings query
	topHoldingsRows := sqlmock.NewRows([]string{"symbol", "name", "quantity", "average_cost", "total_value"}).
		AddRow("GOOGL", "Alphabet Inc.", 5.0, 2800.0, 14000.0).
		AddRow("AAPL", "Apple Inc.", 10.0, 150.0, 1500.0)
	mock.ExpectQuery("SELECT a.symbol, a.name, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) ORDER BY \\(ph.quantity \\* ph.average_cost\\) DESC LIMIT 10").
		WithArgs("user-123").
		WillReturnRows(topHoldingsRows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/analytics/allocation", handler.GetAssetAllocation)

	// Test request
	req, _ := http.NewRequest("GET", "/analytics/allocation", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "allocation_summary")
	assert.Contains(t, w.Body.String(), "by_asset_type")
	assert.Contains(t, w.Body.String(), "by_sector")
	assert.Contains(t, w.Body.String(), "top_holdings")
	assert.Contains(t, w.Body.String(), "STOCK")
	assert.Contains(t, w.Body.String(), "Technology")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetAssetAllocation_EmptyPortfolio(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	// Mock empty asset type allocation
	assetTypeRows := sqlmock.NewRows([]string{"asset_type", "count", "total_value"})
	mock.ExpectQuery("SELECT a.asset_type, COUNT\\(\\*\\) as count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) GROUP BY a.asset_type ORDER BY total_value DESC").
		WithArgs("user-123").
		WillReturnRows(assetTypeRows)

	// Mock empty sector allocation
	sectorRows := sqlmock.NewRows([]string{"sector", "count", "total_value"})
	mock.ExpectQuery("SELECT COALESCE\\(a.sector, 'Unknown'\\) as sector, COUNT\\(\\*\\) as count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) GROUP BY a.sector ORDER BY total_value DESC").
		WithArgs("user-123").
		WillReturnRows(sectorRows)

	// Mock empty top holdings
	topHoldingsRows := sqlmock.NewRows([]string{"symbol", "name", "quantity", "average_cost", "total_value"})
	mock.ExpectQuery("SELECT a.symbol, a.name, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) ORDER BY \\(ph.quantity \\* ph.average_cost\\) DESC LIMIT 10").
		WithArgs("user-123").
		WillReturnRows(topHoldingsRows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "GET", "/analytics/allocation", handler.GetAssetAllocation)

	// Test request
	req, _ := http.NewRequest("GET", "/analytics/allocation", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total_portfolio_value":0`)
	assert.Contains(t, w.Body.String(), `"total_holdings":0`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_WhatIfAnalysis_BuySuccess(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	// Mock current portfolio query
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COUNT\\(\\*\\) as total_holdings FROM portfolio_holdings ph WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"total_cost", "total_holdings"}).AddRow(15500.0, 2))

	// Mock current holding check (no existing position)
	mock.ExpectQuery("SELECT ph.quantity, ph.average_cost FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) AND a.symbol = (.+)").
		WithArgs("user-123", "TSLA").
		WillReturnError(sql.ErrNoRows)

	// Mock current allocation query
	allocationRows := sqlmock.NewRows([]string{"asset_type", "total_value"}).
		AddRow("STOCK", 15500.0)
	mock.ExpectQuery("SELECT a.asset_type, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) GROUP BY a.asset_type").
		WithArgs("user-123").
		WillReturnRows(allocationRows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "POST", "/analytics/what-if", handler.WhatIfAnalysis)

	// Test request
	requestBody := `{
		"action": "buy",
		"symbol": "TSLA",
		"quantity": 5.0,
		"price": 800.0
	}`
	req, _ := http.NewRequest("POST", "/analytics/what-if", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "trade_details")
	assert.Contains(t, w.Body.String(), "position_impact")
	assert.Contains(t, w.Body.String(), "portfolio_impact")
	assert.Contains(t, w.Body.String(), "allocation_impact")
	assert.Contains(t, w.Body.String(), "risk_impact")
	assert.Contains(t, w.Body.String(), "expected_returns")
	assert.Contains(t, w.Body.String(), "TSLA")
	assert.Contains(t, w.Body.String(), "buy")
	assert.Contains(t, w.Body.String(), "created")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_WhatIfAnalysis_SellSuccess(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	// Mock current portfolio query
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COUNT\\(\\*\\) as total_holdings FROM portfolio_holdings ph WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"total_cost", "total_holdings"}).AddRow(15500.0, 2))

	// Mock current holding check (existing position)
	mock.ExpectQuery("SELECT ph.quantity, ph.average_cost FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) AND a.symbol = (.+)").
		WithArgs("user-123", "AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"quantity", "average_cost"}).AddRow(10.0, 150.0))

	// Mock current allocation query
	allocationRows := sqlmock.NewRows([]string{"asset_type", "total_value"}).
		AddRow("STOCK", 15500.0)
	mock.ExpectQuery("SELECT a.asset_type, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) GROUP BY a.asset_type").
		WithArgs("user-123").
		WillReturnRows(allocationRows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "POST", "/analytics/what-if", handler.WhatIfAnalysis)

	// Test request
	requestBody := `{
		"action": "sell",
		"symbol": "AAPL",
		"quantity": 5.0,
		"price": 160.0
	}`
	req, _ := http.NewRequest("POST", "/analytics/what-if", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sell")
	assert.Contains(t, w.Body.String(), "reduced")
	assert.Contains(t, w.Body.String(), "AAPL")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_WhatIfAnalysis_SellInsufficientHoldings(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	// Mock current portfolio query
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COUNT\\(\\*\\) as total_holdings FROM portfolio_holdings ph WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"total_cost", "total_holdings"}).AddRow(15500.0, 2))

	// Mock current holding check (insufficient quantity)
	mock.ExpectQuery("SELECT ph.quantity, ph.average_cost FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) AND a.symbol = (.+)").
		WithArgs("user-123", "AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"quantity", "average_cost"}).AddRow(5.0, 150.0))

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "POST", "/analytics/what-if", handler.WhatIfAnalysis)

	// Test request
	requestBody := `{
		"action": "sell",
		"symbol": "AAPL",
		"quantity": 10.0,
		"price": 160.0
	}`
	req, _ := http.NewRequest("POST", "/analytics/what-if", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Cannot sell more than current position")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_WhatIfAnalysis_SellNoPosition(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	// Mock current portfolio query
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COUNT\\(\\*\\) as total_holdings FROM portfolio_holdings ph WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"total_cost", "total_holdings"}).AddRow(15500.0, 2))

	// Mock current holding check (no position)
	mock.ExpectQuery("SELECT ph.quantity, ph.average_cost FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) AND a.symbol = (.+)").
		WithArgs("user-123", "TSLA").
		WillReturnError(sql.ErrNoRows)

	handler := setupTestHandler(db, NewMockFinnhubClient())
	router := setupTestRouter(handler, "POST", "/analytics/what-if", handler.WhatIfAnalysis)

	// Test request
	requestBody := `{
		"action": "sell",
		"symbol": "TSLA",
		"quantity": 5.0,
		"price": 800.0
	}`
	req, _ := http.NewRequest("POST", "/analytics/what-if", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Cannot sell - no current position in TSLA")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_WhatIfAnalysis_InvalidJSON(t *testing.T) {
	handler := setupTestHandler(nil, NewMockFinnhubClient())
	router := setupTestRouter(handler, "POST", "/analytics/what-if", handler.WhatIfAnalysis)

	// Test request with invalid JSON
	requestBody := `{
		"action": "invalid_action",
		"symbol": "AAPL",
		"quantity": -5.0,
		"price": 160.0
	}`
	req, _ := http.NewRequest("POST", "/analytics/what-if", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_WhatIfAnalysis_HighConcentrationRisk(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	// Mock current portfolio query (small portfolio)
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COUNT\\(\\*\\) as total_holdings FROM portfolio_holdings ph WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"total_cost", "total_holdings"}).AddRow(5000.0, 1))
// Mock current holding check (no existing position)
mock.ExpectQuery("SELECT ph.quantity, ph.average_cost FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) AND a.symbol = (.+)").
	WithArgs("user-123", "NVDA").
	WillReturnError(sql.ErrNoRows)

// Mock current allocation query
allocationRows := sqlmock.NewRows([]string{"asset_type", "total_value"}).
	AddRow("STOCK", 5000.0)
mock.ExpectQuery("SELECT a.asset_type, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) GROUP BY a.asset_type").
	WithArgs("user-123").
	WillReturnRows(allocationRows)

handler := setupTestHandler(db, NewMockFinnhubClient())
router := setupTestRouter(handler, "POST", "/analytics/what-if", handler.WhatIfAnalysis)

// Test request with large trade that creates high concentration
requestBody := `{
	"action": "buy",
	"symbol": "NVDA",
	"quantity": 10.0,
	"price": 900.0
}`
req, _ := http.NewRequest("POST", "/analytics/what-if", strings.NewReader(requestBody))
req.Header.Set("Content-Type", "application/json")
w := httptest.NewRecorder()
router.ServeHTTP(w, req)

// Assertions
assert.Equal(t, http.StatusOK, w.Code)
assert.Contains(t, w.Body.String(), "significant increase in concentration risk")
assert.Contains(t, w.Body.String(), "NVDA")
assert.Contains(t, w.Body.String(), "22.3") // NVDA expected return
assert.NoError(t, mock.ExpectationsWereMet())
}

// =============================================================================
// Additional Edge Case Tests
// =============================================================================

func TestHandler_GetRiskMetrics_WellDiversified(t *testing.T) {
// Setup mock database
db, mock, err := sqlmock.New()
assert.NoError(t, err)
defer db.Close()

// Mock user ID query
mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
	WithArgs("default_user").
	WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

// Mock well-diversified portfolio
diversificationRows := sqlmock.NewRows([]string{"sector", "holdings_count", "sector_value"}).
	AddRow("Technology", 2, 5000.0).
	AddRow("Healthcare", 2, 5000.0).
	AddRow("Finance", 2, 5000.0).
	AddRow("Energy", 2, 5000.0)
mock.ExpectQuery("SELECT a.sector, COUNT\\(\\*\\) as holdings_count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as sector_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) AND a.sector IS NOT NULL GROUP BY a.sector ORDER BY sector_value DESC").
	WithArgs("user-123").
	WillReturnRows(diversificationRows)

// Mock beta calculation query
betaRows := sqlmock.NewRows([]string{"symbol", "quantity", "average_cost", "position_value"})
mock.ExpectQuery("SELECT a.symbol, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as position_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+)").
	WithArgs("user-123").
	WillReturnRows(betaRows)

// Mock current value calculation query
currentValueRows := sqlmock.NewRows([]string{"symbol", "quantity", "average_cost", "position_value"})
mock.ExpectQuery("SELECT a.symbol, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as position_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+)").
	WithArgs("user-123").
	WillReturnRows(currentValueRows)

handler := setupTestHandler(db, NewMockFinnhubClient())
router := setupTestRouter(handler, "GET", "/analytics/risk", handler.GetRiskMetrics)

// Test request
req, _ := http.NewRequest("GET", "/analytics/risk", nil)
w := httptest.NewRecorder()
router.ServeHTTP(w, req)

// Assertions
assert.Equal(t, http.StatusOK, w.Code)
assert.Contains(t, w.Body.String(), "Low")
assert.Contains(t, w.Body.String(), "Well-diversified portfolio")
assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_WhatIfAnalysis_KnownStockReturns(t *testing.T) {
// Setup mock database
db, mock, err := sqlmock.New()
assert.NoError(t, err)
defer db.Close()

// Mock user ID query
mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
	WithArgs("default_user").
	WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

// Mock current portfolio query
mock.ExpectQuery("SELECT COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COUNT\\(\\*\\) as total_holdings FROM portfolio_holdings ph WHERE ph.user_id = (.+)").
	WithArgs("user-123").
	WillReturnRows(sqlmock.NewRows([]string{"total_cost", "total_holdings"}).AddRow(50000.0, 5))

// Mock current holding check (no existing position)
mock.ExpectQuery("SELECT ph.quantity, ph.average_cost FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) AND a.symbol = (.+)").
	WithArgs("user-123", "AAPL").
	WillReturnError(sql.ErrNoRows)

// Mock current allocation query
allocationRows := sqlmock.NewRows([]string{"asset_type", "total_value"}).
	AddRow("STOCK", 50000.0)
mock.ExpectQuery("SELECT a.asset_type, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) GROUP BY a.asset_type").
	WithArgs("user-123").
	WillReturnRows(allocationRows)

handler := setupTestHandler(db, NewMockFinnhubClient())
router := setupTestRouter(handler, "POST", "/analytics/what-if", handler.WhatIfAnalysis)

// Test request with AAPL (known stock with specific expected return)
requestBody := `{
	"action": "buy",
	"symbol": "AAPL",
	"quantity": 10.0,
	"price": 150.0
}`
req, _ := http.NewRequest("POST", "/analytics/what-if", strings.NewReader(requestBody))
req.Header.Set("Content-Type", "application/json")
w := httptest.NewRecorder()
router.ServeHTTP(w, req)

// Assertions
assert.Equal(t, http.StatusOK, w.Code)
assert.Contains(t, w.Body.String(), "12.5") // AAPL expected return
assert.Contains(t, w.Body.String(), "0.24") // AAPL volatility
assert.Contains(t, w.Body.String(), "minimal impact on diversification")
assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetPerformanceAnalytics_NoFinnhub(t *testing.T) {
// Setup mock database
db, mock, err := sqlmock.New()
assert.NoError(t, err)
defer db.Close()

// Mock user ID query
mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
	WithArgs("default_user").
	WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

// Mock portfolio totals query
mock.ExpectQuery("SELECT COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COUNT\\(\\*\\) as total_holdings FROM portfolio_holdings ph WHERE ph.user_id = (.+)").
	WithArgs("user-123").
	WillReturnRows(sqlmock.NewRows([]string{"total_cost", "total_holdings"}).AddRow(15500.0, 2))

// Mock holdings query for market value calculation
holdingsRows := sqlmock.NewRows([]string{"symbol", "quantity", "average_cost"}).
	AddRow("AAPL", 10.0, 150.0).
	AddRow("GOOGL", 5.0, 2800.0)
mock.ExpectQuery("SELECT a.symbol, ph.quantity, ph.average_cost FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+)").
	WithArgs("user-123").
	WillReturnRows(holdingsRows)

// Mock empty snapshots query
snapshotRows := sqlmock.NewRows([]string{"snapshot_date", "total_value", "total_cost", "unrealized_pnl"})
mock.ExpectQuery("SELECT snapshot_date, total_value, total_cost, unrealized_pnl FROM portfolio_snapshots WHERE user_id = (.+) AND snapshot_date >= CURRENT_DATE - INTERVAL '30 days' ORDER BY snapshot_date DESC LIMIT 30").
	WithArgs("user-123").
	WillReturnRows(snapshotRows)

// Mock top performers query
performerRows := sqlmock.NewRows([]string{"symbol", "name", "quantity", "average_cost", "total_value"}).
	AddRow("GOOGL", "Alphabet Inc.", 5.0, 2800.0, 14000.0).
	AddRow("AAPL", "Apple Inc.", 10.0, 150.0, 1500.0)
mock.ExpectQuery("SELECT a.symbol, a.name, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) ORDER BY \\(ph.quantity \\* ph.average_cost\\) DESC LIMIT 5").
	WithArgs("user-123").
	WillReturnRows(performerRows)

// Test with no Finnhub client
handler := setupTestHandler(db, nil)
router := setupTestRouter(handler, "GET", "/analytics/performance", handler.GetPerformanceAnalytics)

// Test request
req, _ := http.NewRequest("GET", "/analytics/performance", nil)
w := httptest.NewRecorder()
router.ServeHTTP(w, req)

// Assertions - should still work but use cost basis instead of real prices
assert.Equal(t, http.StatusOK, w.Code)
assert.Contains(t, w.Body.String(), "portfolio_performance")
assert.Contains(t, w.Body.String(), "15500") // Should equal cost basis when no real prices
assert.NoError(t, mock.ExpectationsWereMet())
}

// =============================================================================
// Benchmark Tests for Performance
// =============================================================================

func BenchmarkHandler_GetAssets(b *testing.B) {
// Setup mock database
db, mock, err := sqlmock.New()
if err != nil {
	b.Fatal(err)
}
defer db.Close()

// Mock successful query
rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "exchange", "currency", "sector", "created_at"}).
	AddRow("1", "AAPL", "Apple Inc.", "STOCK", "NASDAQ", "USD", "Technology", "2024-01-01")

for i := 0; i < b.N; i++ {
	mock.ExpectQuery("SELECT id, symbol, name, asset_type, exchange, currency, sector, created_at FROM assets WHERE 1=1 ORDER BY symbol ASC LIMIT (.+)").
		WithArgs("50").
		WillReturnRows(rows)
}

handler := setupTestHandler(db, NewMockFinnhubClient())
router := setupTestRouter(handler, "GET", "/assets", handler.GetAssets)

b.ResetTimer()
for i := 0; i < b.N; i++ {
	req, _ := http.NewRequest("GET", "/assets", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
}
}

func BenchmarkHandler_GetRiskMetrics(b *testing.B) {
// Setup mock database
db, mock, err := sqlmock.New()
if err != nil {
	b.Fatal(err)
}
defer db.Close()

// Setup repeated mocks for benchmark
for i := 0; i < b.N; i++ {
	// Mock user ID query
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	// Mock diversification query
	diversificationRows := sqlmock.NewRows([]string{"sector", "holdings_count", "sector_value"}).
		AddRow("Technology", 2, 15500.0)
	mock.ExpectQuery("SELECT a.sector, COUNT\\(\\*\\) as holdings_count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as sector_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) AND a.sector IS NOT NULL GROUP BY a.sector ORDER BY sector_value DESC").
		WithArgs("user-123").
		WillReturnRows(diversificationRows)

	// Mock beta calculation queries
	betaRows := sqlmock.NewRows([]string{"symbol", "quantity", "average_cost", "position_value"}).
		AddRow("AAPL", 10.0, 150.0, 1500.0)
	mock.ExpectQuery("SELECT a.symbol, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as position_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(betaRows)

	currentValueRows := sqlmock.NewRows([]string{"symbol", "quantity", "average_cost", "position_value"}).
		AddRow("AAPL", 10.0, 150.0, 1500.0)
	mock.ExpectQuery("SELECT a.symbol, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as position_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(currentValueRows)
}

handler := setupTestHandler(db, NewMockFinnhubClient())
router := setupTestRouter(handler, "GET", "/analytics/risk", handler.GetRiskMetrics)

b.ResetTimer()
for i := 0; i < b.N; i++ {
	req, _ := http.NewRequest("GET", "/analytics/risk", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
}
}
		