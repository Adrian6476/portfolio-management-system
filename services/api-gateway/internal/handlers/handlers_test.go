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
