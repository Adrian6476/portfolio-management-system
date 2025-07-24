package handlers

import (
	"bytes"
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

func TestHandler_HealthCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create a mock services struct (we don't need real connections for this test)
	mockServices := &services.Services{
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	// Create a test request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
	assert.Contains(t, w.Body.String(), "api-gateway")
}

func TestHandler_GetPortfolio(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Set up expected query and result
	rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "quantity", "average_cost", "purchase_date"}).
		AddRow("1", "AAPL", "Apple Inc.", "STOCK", 10.0, 150.0, "2024-01-01").
		AddRow("2", "GOOGL", "Alphabet Inc.", "STOCK", 5.0, 2800.0, "2024-01-02")

	mock.ExpectQuery("SELECT (.+) FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id JOIN users u ON ph.user_id = u.id WHERE u.username = (.+) ORDER BY ph.created_at DESC").
		WithArgs("default_user").
		WillReturnRows(rows)

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.GET("/portfolio", handler.GetPortfolio)

	// Create a test request
	req, _ := http.NewRequest("GET", "/portfolio", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "holdings")
	assert.Contains(t, w.Body.String(), "total_holdings")
	assert.Contains(t, w.Body.String(), "AAPL")
	assert.Contains(t, w.Body.String(), "GOOGL")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetPortfolio_NilDB(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create a mock services struct with nil DB (simulating connection failure)
	mockServices := &services.Services{
		DB:     nil, // This will cause the handler to return an error
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.GET("/portfolio", handler.GetPortfolio)

	// Create a test request
	req, _ := http.NewRequest("GET", "/portfolio", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results - should return 500 error due to nil DB
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to fetch portfolio")
}

func TestHandler_AddHolding(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Set up expected queries
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	mock.ExpectQuery("SELECT id FROM assets WHERE symbol = (.+)").
		WithArgs("AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("asset-123"))

	mock.ExpectExec("INSERT INTO portfolio_holdings (.+) ON CONFLICT (.+) DO UPDATE SET (.+)").
		WithArgs("user-123", "asset-123", 10.0, 150.0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.POST("/portfolio/holdings", handler.AddHolding)

	// Create a test request with JSON body
	requestBody := `{"symbol": "AAPL", "quantity": 10.0, "average_cost": 150.0}`
	req, _ := http.NewRequest("POST", "/portfolio/holdings", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Holding added successfully")
	assert.Contains(t, w.Body.String(), "AAPL")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_AddHolding_NilDB(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create a mock services struct with nil DB (simulating connection failure)
	mockServices := &services.Services{
		DB:     nil, // This will cause the handler to return an error
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.POST("/holdings", handler.AddHolding)

	// Test with valid JSON payload
	payload := `{"symbol": "AAPL", "quantity": 10, "average_cost": 150.50}`
	req, _ := http.NewRequest("POST", "/holdings", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results - should return 500 error due to nil DB
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to get user")
}

func TestHandler_AddHolding_InvalidJSON(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.POST("/holdings", handler.AddHolding)

	// Test with invalid JSON payload
	payload := `{"symbol": "AAPL", "quantity": -10}` // negative quantity should fail validation
	req, _ := http.NewRequest("POST", "/holdings", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results - should return 400 error due to validation failure
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_UpdateHolding(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Set up expected queries
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	mock.ExpectQuery("SELECT ph.quantity, ph.average_cost, a.symbol FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)").
		WithArgs("holding-123", "user-123").
		WillReturnRows(sqlmock.NewRows([]string{"quantity", "average_cost", "symbol"}).AddRow(10.0, 150.0, "AAPL"))

	mock.ExpectExec("UPDATE portfolio_holdings SET quantity = (.+), average_cost = (.+), updated_at = NOW\\(\\) WHERE id = (.+) AND user_id = (.+)").
		WithArgs(15.0, 160.0, "holding-123", "user-123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.PUT("/portfolio/holdings/:id", handler.UpdateHolding)

	// Create a test request with JSON body
	requestBody := `{"quantity": 15.0, "average_cost": 160.0}`
	req, _ := http.NewRequest("PUT", "/portfolio/holdings/holding-123", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Holding updated successfully")
	assert.Contains(t, w.Body.String(), "AAPL")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_UpdateHolding_MissingID(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.PUT("/portfolio/holdings/:id", handler.UpdateHolding)

	// Create a test request without ID
	requestBody := `{"quantity": 15.0}`
	req, _ := http.NewRequest("PUT", "/portfolio/holdings/", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results - should return 404 because route doesn't match
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_UpdateHolding_NoFields(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.PUT("/portfolio/holdings/:id", handler.UpdateHolding)

	// Create a test request with empty JSON
	requestBody := `{}`
	req, _ := http.NewRequest("PUT", "/portfolio/holdings/holding-123", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results - should return 400 error
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "At least one field")
}

func TestHandler_UpdateHolding_NotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Set up expected queries
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	mock.ExpectQuery("SELECT ph.quantity, ph.average_cost, a.symbol FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)").
		WithArgs("nonexistent-holding", "user-123").
		WillReturnError(sqlmock.ErrCancelled)

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.PUT("/portfolio/holdings/:id", handler.UpdateHolding)

	// Create a test request
	requestBody := `{"quantity": 15.0}`
	req, _ := http.NewRequest("PUT", "/portfolio/holdings/nonexistent-holding", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results - should return 404 error
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Holding not found")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_RemoveHolding(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Set up expected queries
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	mock.ExpectQuery("SELECT a.symbol, ph.quantity FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)").
		WithArgs("holding-123", "user-123").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "quantity"}).AddRow("AAPL", 10.0))

	mock.ExpectExec("DELETE FROM portfolio_holdings WHERE id = (.+) AND user_id = (.+)").
		WithArgs("holding-123", "user-123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.DELETE("/portfolio/holdings/:id", handler.RemoveHolding)

	// Create a test request
	req, _ := http.NewRequest("DELETE", "/portfolio/holdings/holding-123", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Holding removed successfully")
	assert.Contains(t, w.Body.String(), "AAPL")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_RemoveHolding_MissingID(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.DELETE("/portfolio/holdings/:id", handler.RemoveHolding)

	// Create a test request without ID
	req, _ := http.NewRequest("DELETE", "/portfolio/holdings/", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results - should return 404 because route doesn't match
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_RemoveHolding_NotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Set up expected queries
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	mock.ExpectQuery("SELECT a.symbol, ph.quantity FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)").
		WithArgs("nonexistent-holding", "user-123").
		WillReturnError(sqlmock.ErrCancelled)

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.DELETE("/portfolio/holdings/:id", handler.RemoveHolding)

	// Create a test request
	req, _ := http.NewRequest("DELETE", "/portfolio/holdings/nonexistent-holding", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results - should return 404 error
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Holding not found")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetPortfolioSummary(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Set up expected queries
	mock.ExpectQuery("SELECT id FROM users WHERE username = (.+)").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-123"))

	// Portfolio summary query
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) as total_holdings, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_cost, COALESCE\\(SUM\\(ph.quantity\\), 0\\) as total_shares FROM portfolio_holdings ph WHERE ph.user_id = (.+)").
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{"total_holdings", "total_cost", "total_shares"}).AddRow(2, 15500.0, 15.0))

	// Asset allocation query
	allocationRows := sqlmock.NewRows([]string{"asset_type", "count", "total_value"}).
		AddRow("STOCK", 2, 15500.0)
	mock.ExpectQuery("SELECT a.asset_type, COUNT\\(\\*\\) as count, COALESCE\\(SUM\\(ph.quantity \\* ph.average_cost\\), 0\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) GROUP BY a.asset_type ORDER BY total_value DESC").
		WithArgs("user-123").
		WillReturnRows(allocationRows)

	// Top holdings query
	topHoldingsRows := sqlmock.NewRows([]string{"symbol", "name", "quantity", "average_cost", "total_value"}).
		AddRow("AAPL", "Apple Inc.", 10.0, 150.0, 1500.0).
		AddRow("GOOGL", "Alphabet Inc.", 5.0, 2800.0, 14000.0)
	mock.ExpectQuery("SELECT a.symbol, a.name, ph.quantity, ph.average_cost, \\(ph.quantity \\* ph.average_cost\\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.user_id = (.+) ORDER BY \\(ph.quantity \\* ph.average_cost\\) DESC LIMIT 5").
		WithArgs("user-123").
		WillReturnRows(topHoldingsRows)

	mockServices := &services.Services{
		DB:     db,
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.GET("/portfolio/summary", handler.GetPortfolioSummary)

	// Create a test request
	req, _ := http.NewRequest("GET", "/portfolio/summary", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "summary")
	assert.Contains(t, w.Body.String(), "asset_allocation")
	assert.Contains(t, w.Body.String(), "top_holdings")
	assert.Contains(t, w.Body.String(), "AAPL")
	assert.Contains(t, w.Body.String(), "GOOGL")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetPortfolioSummary_NilDB(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create a mock services struct with nil DB (simulating connection failure)
	mockServices := &services.Services{
		DB:     nil, // This will cause the handler to return an error
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	// Create a test router
	router := gin.New()
	router.GET("/portfolio/summary", handler.GetPortfolioSummary)

	// Create a test request
	req, _ := http.NewRequest("GET", "/portfolio/summary", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the results - should return 500 error due to nil DB
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to fetch portfolio summary")
}
