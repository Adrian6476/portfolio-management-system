package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

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
