package handlers

import (
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
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/portfolio-management/api-gateway/internal/services"
)

// MockFinnhubClient implements a mock Finnhub client for testing
type MockFinnhubClient struct {
	shouldFail bool
	profile    *services.FinnhubProfile
	quote      *services.FinnhubQuote
}

func (m *MockFinnhubClient) GetCompanyProfile(symbol string) (*services.FinnhubProfile, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	if m.profile != nil {
		return m.profile, nil
	}
	return &services.FinnhubProfile{
		Name:   "Test Company",
		Ticker: symbol,
	}, nil
}

func (m *MockFinnhubClient) GetQuote(symbol string) (*services.FinnhubQuote, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	if m.quote != nil {
		return m.quote, nil
	}
	return &services.FinnhubQuote{
		CurrentPrice:       150.0,
		Change:             5.0,
		PercentChange:      3.45,
		HighPriceOfDay:     155.0,
		LowPriceOfDay:      148.0,
		OpenPriceOfDay:     149.0,
		PreviousClosePrice: 145.0,
		Timestamp:          1640995200,
	}, nil
}

func setupTestHandler(db *sql.DB, finnhub *MockFinnhubClient) *Handler {
	logger, _ := zap.NewDevelopment()
	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}
	// Set Finnhub interface if provided
	if finnhub != nil {
		mockServices.Finnhub = finnhub
	}
	return NewHandler(mockServices, logger)
}

// TestGetPortfolio tests the GetPortfolio handler
func TestGetPortfolio(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Returns portfolio holdings", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "quantity", "average_cost", "purchase_date"}).
			AddRow("1", "AAPL", "Apple Inc.", "STOCK", 10.0, 150.0, "2024-01-01").
			AddRow("2", "GOOGL", "Alphabet Inc.", "STOCK", 5.0, 2800.0, "2024-01-02")

		mock.ExpectQuery("SELECT (.+) FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id JOIN users u ON ph.user_id = u.id WHERE u.username = (.+) ORDER BY ph.created_at DESC").
			WithArgs("default_user").
			WillReturnRows(rows)

		handler := setupTestHandler(db, nil)
		router := gin.New()
		router.GET("/portfolio", handler.GetPortfolio)

		req, _ := http.NewRequest("GET", "/portfolio", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "holdings")
		assert.Contains(t, response, "total_holdings")
		assert.Equal(t, float64(2), response["total_holdings"])
		
		holdings := response["holdings"].([]interface{})
		assert.Len(t, holdings, 2)
		
		firstHolding := holdings[0].(map[string]interface{})
		assert.Equal(t, "AAPL", firstHolding["symbol"])
		assert.Equal(t, "Apple Inc.", firstHolding["name"])
		assert.Equal(t, float64(10), firstHolding["quantity"])
		assert.Equal(t, float64(150), firstHolding["average_cost"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Empty portfolio", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "quantity", "average_cost", "purchase_date"})

		mock.ExpectQuery("SELECT (.+) FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id JOIN users u ON ph.user_id = u.id WHERE u.username = (.+) ORDER BY ph.created_at DESC").
			WithArgs("default_user").
			WillReturnRows(rows)

		handler := setupTestHandler(db, nil)
		router := gin.New()
		router.GET("/portfolio", handler.GetPortfolio)

		req, _ := http.NewRequest("GET", "/portfolio", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, float64(0), response["total_holdings"])
		holdings := response["holdings"]
		assert.Nil(t, holdings)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Database connection nil", func(t *testing.T) {
		handler := setupTestHandler(nil, nil)
		router := gin.New()
		router.GET("/portfolio", handler.GetPortfolio)

		req, _ := http.NewRequest("GET", "/portfolio", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch portfolio")
	})

	t.Run("Error - SQL query fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT (.+) FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id JOIN users u ON ph.user_id = u.id WHERE u.username = (.+) ORDER BY ph.created_at DESC").
			WithArgs("default_user").
			WillReturnError(sql.ErrConnDone)

		handler := setupTestHandler(db, nil)
		router := gin.New()
		router.GET("/portfolio", handler.GetPortfolio)

		req, _ := http.NewRequest("GET", "/portfolio", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch portfolio")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetPortfolioSummary tests the GetPortfolioSummary handler
func TestGetPortfolioSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Returns complete portfolio summary", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Mock user ID query
		mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
			WithArgs("default_user").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

		// Mock portfolio summary query
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) as total_holdings, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COALESCE\\(SUM\\(ph.quantity\\), 0\\) as total_shares FROM portfolio_holdings ph WHERE ph.user_id = (.+)").
			WithArgs("user-123").
			WillReturnRows(sqlmock.NewRows([]string{"total_holdings", "total_cost", "total_shares"}).AddRow(2, 15500.0, 15.0))

		// Mock asset allocation query
		allocationRows := sqlmock.NewRows([]string{"asset_type", "count", "total_value"}).
			AddRow("STOCK", 2, 15500.0)
		mock.ExpectQuery("SELECT a.asset_type, COUNT\\(\\*\\) as count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) GROUP BY a.asset_type ORDER BY total_value DESC").
			WithArgs("user-123").
			WillReturnRows(allocationRows)

		// Mock top holdings query
		topHoldingsRows := sqlmock.NewRows([]string{"symbol", "name", "quantity", "average_cost", "total_value"}).
			AddRow("GOOGL", "Alphabet Inc.", 5.0, 2800.0, 14000.0).
			AddRow("AAPL", "Apple Inc.", 10.0, 150.0, 1500.0)
		mock.ExpectQuery("SELECT a.symbol, a.name, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) ORDER BY \\(ph.quantity \\* ph.average_cost\\) DESC LIMIT 5").
			WithArgs("user-123").
			WillReturnRows(topHoldingsRows)

		handler := setupTestHandler(db, nil)
		router := gin.New()
		router.GET("/portfolio/summary", handler.GetPortfolioSummary)

		req, _ := http.NewRequest("GET", "/portfolio/summary", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		// Check summary section
		summary := response["summary"].(map[string]interface{})
		assert.Equal(t, float64(2), summary["total_holdings"])
		assert.Equal(t, float64(15500), summary["total_cost"])
		assert.Equal(t, float64(15), summary["total_shares"])
		
		// Check asset allocation
		allocations := response["asset_allocation"].([]interface{})
		assert.Len(t, allocations, 1)
		allocation := allocations[0].(map[string]interface{})
		assert.Equal(t, "STOCK", allocation["asset_type"])
		assert.Equal(t, float64(100), allocation["percentage"]) // 15500/15500 * 100
		
		// Check top holdings
		topHoldings := response["top_holdings"].([]interface{})
		assert.Len(t, topHoldings, 2)
		topHolding := topHoldings[0].(map[string]interface{})
		assert.Equal(t, "GOOGL", topHolding["symbol"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Empty portfolio summary", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Mock user ID query
		mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
			WithArgs("default_user").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

		// Mock portfolio summary query - empty portfolio
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) as total_holdings, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COALESCE\\(SUM\\(ph.quantity\\), 0\\) as total_shares FROM portfolio_holdings ph WHERE ph.user_id = (.+)").
			WithArgs("user-123").
			WillReturnRows(sqlmock.NewRows([]string{"total_holdings", "total_cost", "total_shares"}).AddRow(0, 0.0, 0.0))

		// Mock empty asset allocation query
		allocationRows := sqlmock.NewRows([]string{"asset_type", "count", "total_value"})
		mock.ExpectQuery("SELECT a.asset_type, COUNT\\(\\*\\) as count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) GROUP BY a.asset_type ORDER BY total_value DESC").
			WithArgs("user-123").
			WillReturnRows(allocationRows)

		// Mock empty top holdings query
		topHoldingsRows := sqlmock.NewRows([]string{"symbol", "name", "quantity", "average_cost", "total_value"})
		mock.ExpectQuery("SELECT a.symbol, a.name, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) ORDER BY \\(ph.quantity \\* ph.average_cost\\) DESC LIMIT 5").
			WithArgs("user-123").
			WillReturnRows(topHoldingsRows)

		handler := setupTestHandler(db, nil)
		router := gin.New()
		router.GET("/portfolio/summary", handler.GetPortfolioSummary)

		req, _ := http.NewRequest("GET", "/portfolio/summary", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		summary := response["summary"].(map[string]interface{})
		assert.Equal(t, float64(0), summary["total_holdings"])
		assert.Equal(t, float64(0), summary["total_cost"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Database connection nil", func(t *testing.T) {
		handler := setupTestHandler(nil, nil)
		router := gin.New()
		router.GET("/portfolio/summary", handler.GetPortfolioSummary)

		req, _ := http.NewRequest("GET", "/portfolio/summary", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch portfolio summary")
	})

	t.Run("Error - User not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
			WithArgs("default_user").
			WillReturnError(sql.ErrNoRows)

		handler := setupTestHandler(db, nil)
		router := gin.New()
		router.GET("/portfolio/summary", handler.GetPortfolioSummary)

		req, _ := http.NewRequest("GET", "/portfolio/summary", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to get user")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetPortfolioPerformance tests the GetPortfolioPerformance handler
func TestGetPortfolioPerformance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Error - Database connection nil", func(t *testing.T) {
		handler := setupTestHandler(nil, nil)
		router := gin.New()
		router.GET("/portfolio/performance", handler.GetPortfolioPerformance)

		req, _ := http.NewRequest("GET", "/portfolio/performance", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch portfolio performance")
	})

	t.Run("Success - Returns portfolio performance", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockFinnhub := &MockFinnhubClient{
			quote: &services.FinnhubQuote{
				CurrentPrice:       160.0,
				Change:             10.0,
				PercentChange:      6.67,
				HighPriceOfDay:     165.0,
				LowPriceOfDay:      155.0,
				OpenPriceOfDay:     158.0,
				PreviousClosePrice: 150.0,
				Timestamp:          1640995200,
			},
		}

		// Mock user ID query
		mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
			WithArgs("default_user").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

		// Mock holdings query
		holdingsRows := sqlmock.NewRows([]string{"id", "symbol", "name", "quantity", "average_cost", "purchase_date", "cost_basis"}).
			AddRow("1", "AAPL", "Apple Inc.", 10.0, 150.0, "2024-01-01", 1500.0)
		mock.ExpectQuery("SELECT (.+) FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) ORDER BY cost_basis DESC").
			WithArgs("user-123").
			WillReturnRows(holdingsRows)

		// Mock snapshots query
		mock.ExpectQuery("SELECT snapshot_date, total_value, total_cost, unrealized_pnl FROM portfolio_snapshots WHERE user_id = (.+) AND snapshot_date >= CURRENT_DATE - INTERVAL (.+) ORDER BY snapshot_date ASC").
			WithArgs("user-123").
			WillReturnRows(sqlmock.NewRows([]string{"snapshot_date", "total_value", "total_cost", "unrealized_pnl"}))

		// Mock snapshot creation
		mock.ExpectExec("INSERT INTO portfolio_snapshots (.+)").
			WillReturnResult(sqlmock.NewResult(1, 1))

		handler := setupTestHandler(db, mockFinnhub)
		router := gin.New()
		router.GET("/portfolio/performance", handler.GetPortfolioPerformance)

		req, _ := http.NewRequest("GET", "/portfolio/performance", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "performance_summary")
		assert.Contains(t, response, "holdings_performance")
		assert.Contains(t, response, "period")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestAddHolding tests the AddHolding handler
func TestAddHolding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Add holding with existing asset", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Mock user ID query
		mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
			WithArgs("default_user").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

		// Mock existing asset query
		mock.ExpectQuery("SELECT id FROM assets WHERE symbol = (.+)").
			WithArgs("AAPL").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("asset-123"))

		// Mock insert/update holding
		mock.ExpectExec("INSERT INTO portfolio_holdings (.+) ON CONFLICT (.+) DO UPDATE SET (.+)").
			WithArgs("user-123", "asset-123", 10.0, 150.0).
			WillReturnResult(sqlmock.NewResult(1, 1))

		handler := setupTestHandler(db, nil)
		router := gin.New()
		router.POST("/portfolio/holdings", handler.AddHolding)

		requestBody := `{"symbol": "AAPL", "quantity": 10.0, "average_cost": 150.0}`
		req, _ := http.NewRequest("POST", "/portfolio/holdings", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Holding added successfully")
		assert.Contains(t, w.Body.String(), "AAPL")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Add holding with new asset creation via Finnhub", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockFinnhub := &MockFinnhubClient{
			profile: &services.FinnhubProfile{
				Name:   "Apple Inc.",
				Ticker: "AAPL",
			},
		}

		// Mock user ID query
		mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
			WithArgs("default_user").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

		// Mock asset not found
		mock.ExpectQuery("SELECT id FROM assets WHERE symbol = (.+)").
			WithArgs("AAPL").
			WillReturnError(sql.ErrNoRows)

		// Mock asset creation
		mock.ExpectQuery("INSERT INTO assets (.+) RETURNING id").
			WithArgs("AAPL", "Apple Inc.").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("asset-123"))

		// Mock insert holding
		mock.ExpectExec("INSERT INTO portfolio_holdings (.+) ON CONFLICT (.+) DO UPDATE SET (.+)").
			WithArgs("user-123", "asset-123", 10.0, 150.0).
			WillReturnResult(sqlmock.NewResult(1, 1))

		handler := setupTestHandler(db, mockFinnhub)
		router := gin.New()
		router.POST("/portfolio/holdings", handler.AddHolding)

		requestBody := `{"symbol": "AAPL", "quantity": 10.0, "average_cost": 150.0}`
		req, _ := http.NewRequest("POST", "/portfolio/holdings", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Holding added successfully")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Invalid JSON input", func(t *testing.T) {
		handler := setupTestHandler(nil, nil)
		router := gin.New()
		router.POST("/portfolio/holdings", handler.AddHolding)

		requestBody := `{"symbol": "AAPL", "quantity": -10.0}` // negative quantity
		req, _ := http.NewRequest("POST", "/portfolio/holdings", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Error - Database connection nil", func(t *testing.T) {
		handler := setupTestHandler(nil, nil)
		router := gin.New()
		router.POST("/portfolio/holdings", handler.AddHolding)

		requestBody := `{"symbol": "AAPL", "quantity": 10.0, "average_cost": 150.0}`
		req, _ := http.NewRequest("POST", "/portfolio/holdings", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to get user")
	})
}

// TestAddHolding_CostAveraging tests the cost averaging logic specifically
func TestAddHolding_CostAveraging(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Cost averaging calculation in SQL", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Mock user ID query
		mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
			WithArgs("default_user").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

		// Mock existing asset query
		mock.ExpectQuery("SELECT id FROM assets WHERE symbol = (.+)").
			WithArgs("AAPL").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("asset-123"))

		// Mock insert/update holding with cost averaging logic
		mock.ExpectExec("INSERT INTO portfolio_holdings \\(user_id, asset_id, quantity, average_cost\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\) ON CONFLICT \\(user_id, asset_id\\) DO UPDATE SET quantity = portfolio_holdings.quantity \\+ EXCLUDED.quantity, average_cost = \\(\\(portfolio_holdings.quantity \\* portfolio_holdings.average_cost\\) \\+ \\(EXCLUDED.quantity \\* EXCLUDED.average_cost\\)\\) / \\(portfolio_holdings.quantity \\+ EXCLUDED.quantity\\), updated_at = NOW\\(\\)").
			WithArgs("user-123", "asset-123", 5.0, 160.0).
			WillReturnResult(sqlmock.NewResult(1, 1))

		handler := setupTestHandler(db, nil)
		router := gin.New()
		router.POST("/portfolio/holdings", handler.AddHolding)

		// Adding 5 shares at $160 to existing position
		requestBody := `{"symbol": "AAPL", "quantity": 5.0, "average_cost": 160.0}`
		req, _ := http.NewRequest("POST", "/portfolio/holdings", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Holding added successfully")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestUpdateHolding tests the UpdateHolding handler
func TestUpdateHolding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Update quantity only", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Mock user ID query
		mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
			WithArgs("default_user").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

		// Mock existing holding query
		mock.ExpectQuery("SELECT ph.quantity, ph.average_cost, a.symbol FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)").
			WithArgs("holding-123", "user-123").
			WillReturnRows(sqlmock.NewRows([]string{"quantity", "average_cost", "symbol"}).AddRow(10.0, 150.0, "AAPL"))

		// Mock update holding
		mock.ExpectExec("UPDATE portfolio_holdings SET quantity = (.+), average_cost = (.+), updated_at = NOW\\(\\) WHERE id = (.+) AND user_id = (.+)").
			WithArgs(15.0, 150.0, "holding-123", "user-123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		handler := setupTestHandler(db, nil)
		router := gin.New()
		router.PUT("/portfolio/holdings/:id", handler.UpdateHolding)

		requestBody := `{"quantity": 15.0}`
		req, _ := http.NewRequest("PUT", "/portfolio/holdings/holding-123", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Holding updated successfully")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - No fields provided for update", func(t *testing.T) {
		handler := setupTestHandler(nil, nil)
		router := gin.New()
		router.PUT("/portfolio/holdings/:id", handler.UpdateHolding)

		requestBody := `{}`
		req, _ := http.NewRequest("PUT", "/portfolio/holdings/holding-123", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "At least one field")
	})
}

// TestRemoveHolding tests the RemoveHolding handler
func TestRemoveHolding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Remove holding", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Mock user ID query
		mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
			WithArgs("default_user").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

		// Mock holding existence and ownership check
		mock.ExpectQuery("SELECT a.symbol, ph.quantity FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)").
			WithArgs("holding-123", "user-123").
			WillReturnRows(sqlmock.NewRows([]string{"symbol", "quantity"}).AddRow("AAPL", 10.0))

		// Mock delete holding
		mock.ExpectExec("DELETE FROM portfolio_holdings WHERE id = (.+) AND user_id = (.+)").
			WithArgs("holding-123", "user-123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		handler := setupTestHandler(db, nil)
		router := gin.New()
		router.DELETE("/portfolio/holdings/:id", handler.RemoveHolding)

		req, _ := http.NewRequest("DELETE", "/portfolio/holdings/holding-123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Holding removed successfully")
		assert.Contains(t, w.Body.String(), "AAPL")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Non-existent holding", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Mock user ID query
		mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
			WithArgs("default_user").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

		// Mock non-existent holding
		mock.ExpectQuery("SELECT a.symbol, ph.quantity FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)").
			WithArgs("nonexistent-holding", "user-123").
			WillReturnError(sql.ErrNoRows)

		handler := setupTestHandler(db, nil)
		router := gin.New()
		router.DELETE("/portfolio/holdings/:id", handler.RemoveHolding)

		req, _ := http.NewRequest("DELETE", "/portfolio/holdings/nonexistent-holding", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Holding not found")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetCurrentPrice tests the GetCurrentPrice handler with Finnhub integration
func TestGetCurrentPrice(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Get current price from Finnhub", func(t *testing.T) {
		mockFinnhub := &MockFinnhubClient{
			quote: &services.FinnhubQuote{
				CurrentPrice:       150.25,
				Change:             2.50,
				PercentChange:      1.69,
				HighPriceOfDay:     152.00,
				LowPriceOfDay:      148.50,
				OpenPriceOfDay:     149.00,
				PreviousClosePrice: 147.75,
				Timestamp:          1640995200,
			},
		}

		handler := setupTestHandler(nil, mockFinnhub)
		router := gin.New()
		router.GET("/market/prices/:symbol", handler.GetCurrentPrice)

		req, _ := http.NewRequest("GET", "/market/prices/AAPL", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "AAPL", response["symbol"])
		assert.Equal(t, 150.25, response["current_price"])
		assert.Equal(t, 2.50, response["change"])
		assert.Equal(t, 1.69, response["change_percent"])
	})

	t.Run("Error - Finnhub service not available", func(t *testing.T) {
		handler := setupTestHandler(nil, nil) // No Finnhub client
		router := gin.New()
		router.GET("/market/prices/:symbol", handler.GetCurrentPrice)

		req, _ := http.NewRequest("GET", "/market/prices/AAPL", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "Market data service not available")
	})

	t.Run("Error - Finnhub API failure", func(t *testing.T) {
		mockFinnhub := &MockFinnhubClient{shouldFail: true}

		handler := setupTestHandler(nil, mockFinnhub)
		router := gin.New()
		router.GET("/market/prices/:symbol", handler.GetCurrentPrice)

		req, _ := http.NewRequest("GET", "/market/prices/AAPL", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch current price")
	})
}

// TestCreateSampleData tests the CreateSampleData helper function
func TestCreateSampleData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Create sample data", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Mock user ID query
		mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
			WithArgs("default_user").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

		// Mock transaction begin
		mock.ExpectBegin()

		// Mock asset insertions (multiple assets)
		for i := 0; i < 8; i++ { // 8 sample assets
			mock.ExpectExec("INSERT INTO assets (.+) ON CONFLICT (.+) DO NOTHING").
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		// Mock clear existing holdings
		mock.ExpectExec("DELETE FROM portfolio_holdings WHERE user_id = (.+)").
			WithArgs("user-123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock sample holdings insertions (5 holdings)
		for i := 0; i < 5; i++ {
			mock.ExpectQuery("SELECT id FROM assets WHERE symbol = (.+)").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(fmt.Sprintf("asset-%d", i)))
			mock.ExpectExec("INSERT INTO portfolio_holdings (.+)").
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		// Mock clear existing transactions
		mock.ExpectExec("DELETE FROM transactions WHERE user_id = (.+)").
			WithArgs("user-123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock sample transactions insertions (7 transactions)
		for i := 0; i < 7; i++ {
			mock.ExpectQuery("SELECT id FROM assets WHERE symbol = (.+)").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(fmt.Sprintf("asset-%d", i)))
			mock.ExpectExec("INSERT INTO transactions (.+)").
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		// Mock clear existing notifications
		mock.ExpectExec("DELETE FROM notifications WHERE user_id = (.+)").
			WithArgs("user-123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock sample notifications insertions (5 notifications)
		for i := 0; i < 5; i++ {
			mock.ExpectExec("INSERT INTO notifications (.+)").
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		// Mock transaction commit
		mock.ExpectCommit()

		handler := setupTestHandler(db, nil)

		err = handler.CreateSampleData()
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Database connection nil", func(t *testing.T) {
		handler := setupTestHandler(nil, nil)

		err := handler.CreateSampleData()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection is nil")
	})
}

// TestValidateAssetSymbol tests the validateAssetSymbol helper function
func TestValidateAssetSymbol(t *testing.T) {
	handler := setupTestHandler(nil, nil)

	t.Run("Success - Valid symbol", func(t *testing.T) {
		err := handler.validateAssetSymbol("AAPL")
		assert.NoError(t, err)
	})

	t.Run("Error - Empty symbol", func(t *testing.T) {
		err := handler.validateAssetSymbol("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid symbol length")
	})

	t.Run("Error - Symbol too long", func(t *testing.T) {
		err := handler.validateAssetSymbol("ABCDEFGHIJK") // 11 characters
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid symbol length")
	})
}

// TestUserOwnsAsset tests the userOwnsAsset helper function
func TestUserOwnsAsset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - User owns asset", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT quantity FROM portfolio_holdings WHERE user_id = (.+) AND asset_id = (.+)").
			WithArgs("user-123", "asset-456").
			WillReturnRows(sqlmock.NewRows([]string{"quantity"}).AddRow(10.5))

		handler := setupTestHandler(db, nil)

		owns, quantity, err := handler.userOwnsAsset("user-123", "asset-456")
		assert.NoError(t, err)
		assert.True(t, owns)
		assert.Equal(t, 10.5, quantity)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - User does not own asset", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT quantity FROM portfolio_holdings WHERE user_id = (.+) AND asset_id = (.+)").
			WithArgs("user-123", "asset-456").
			WillReturnError(sql.ErrNoRows)

		handler := setupTestHandler(db, nil)

		owns, quantity, err := handler.userOwnsAsset("user-123", "asset-456")
		assert.NoError(t, err)
		assert.False(t, owns)
		assert.Equal(t, 0.0, quantity)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetAssetIDBySymbol tests the getAssetIDBySymbol helper function
func TestGetAssetIDBySymbol(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Asset found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT id FROM assets WHERE symbol = (.+)").
			WithArgs("AAPL").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("asset-123"))

		handler := setupTestHandler(db, nil)

		assetID, err := handler.getAssetIDBySymbol("AAPL")
		assert.NoError(t, err)
		assert.Equal(t, "asset-123", assetID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Asset not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT id FROM assets WHERE symbol = (.+)").
			WithArgs("NONEXISTENT").
			WillReturnError(sql.ErrNoRows)

		handler := setupTestHandler(db, nil)

		assetID, err := handler.getAssetIDBySymbol("NONEXISTENT")
		assert.Error(t, err)
		assert.Equal(t, "", assetID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetUserID tests the getUserID helper function
func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - User found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
			WithArgs("default_user").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

		handler := setupTestHandler(db, nil)

		userID, err := handler.getUserID("default_user")
		assert.NoError(t, err)
		assert.Equal(t, "user-123", userID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - User not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
			WithArgs("nonexistent_user").
			WillReturnError(sql.ErrNoRows)

		handler := setupTestHandler(db, nil)

		userID, err := handler.getUserID("nonexistent_user")
		assert.Error(t, err)
		assert.Equal(t, "", userID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestPortfolioIntegration tests integration scenarios
func TestPortfolioIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Integration - Add holding then get portfolio", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Mock for AddHolding
		mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
			WithArgs("default_user").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

		mock.ExpectQuery("SELECT id FROM assets WHERE symbol = (.+)").
			WithArgs("AAPL").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("asset-123"))

		mock.ExpectExec("INSERT INTO portfolio_holdings (.+) ON CONFLICT (.+) DO UPDATE SET (.+)").
			WithArgs("user-123", "asset-123", 10.0, 150.0).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock for GetPortfolio
		rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "quantity", "average_cost", "purchase_date"}).
			AddRow("1", "AAPL", "Apple Inc.", "STOCK", 10.0, 150.0, "2024-01-01")

		mock.ExpectQuery("SELECT (.+) FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id JOIN users u ON ph.user_id = u.id WHERE u.username = (.+) ORDER BY ph.created_at DESC").
			WithArgs("default_user").
			WillReturnRows(rows)

		handler := setupTestHandler(db, nil)
		router := gin.New()
		router.POST("/portfolio/holdings", handler.AddHolding)
		router.GET("/portfolio", handler.GetPortfolio)

		// First add a holding
		requestBody := `{"symbol": "AAPL", "quantity": 10.0, "average_cost": 150.0}`
		req, _ := http.NewRequest("POST", "/portfolio/holdings", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Then get the portfolio
		req, _ = http.NewRequest("GET", "/portfolio", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "AAPL")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Benchmark tests for performance
func BenchmarkGetPortfolio(b *testing.B) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(b, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "quantity", "average_cost", "purchase_date"}).
		AddRow("1", "AAPL", "Apple Inc.", "STOCK", 10.0, 150.0, "2024-01-01")

	for i := 0; i < b.N; i++ {
		mock.ExpectQuery("SELECT (.+) FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id JOIN users u ON ph.user_id = u.id WHERE u.username = (.+) ORDER BY ph.created_at DESC").
			WithArgs("default_user").
			WillReturnRows(rows)
	}

	handler := setupTestHandler(db, nil)
	router := gin.New()
	router.GET("/portfolio", handler.GetPortfolio)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/portfolio", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}