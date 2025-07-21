package handlers

import (
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
	
	mockServices := &services.Services{
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
	assert.Contains(t, w.Body.String(), "Coming Soon")
}