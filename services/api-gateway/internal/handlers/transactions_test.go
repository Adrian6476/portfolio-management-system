package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/portfolio-management/api-gateway/internal/services"
)

// TestGetTransactions tests the GetTransactions handler
func TestGetTransactions(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   []string
	}{
		{
			name:        "successful transaction listing with pagination",
			queryParams: "?limit=10&offset=0",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock transactions query
				rows := sqlmock.NewRows([]string{
					"id", "transaction_type", "quantity", "price", "fees",
					"total_amount", "transaction_date", "notes", "symbol", "name",
				}).
					AddRow("tx1", "BUY", 10.0, 150.0, 1.0, 1501.0, "2024-01-01", "Test buy", "AAPL", "Apple Inc.").
					AddRow("tx2", "SELL", 5.0, 160.0, 1.0, 799.0, "2024-01-02", "Test sell", "AAPL", "Apple Inc.")

				mock.ExpectQuery("SELECT (.+) FROM transactions t JOIN assets a ON t.asset_id = a.id WHERE t.user_id = \\$1 ORDER BY t.transaction_date DESC LIMIT \\$2 OFFSET \\$3").
					WithArgs("user1", "10", "0").
					WillReturnRows(rows)

				// Mock count query
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM transactions t JOIN assets a ON t.asset_id = a.id WHERE t.user_id = \\$1").
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"transactions", "total_count", "AAPL", "BUY", "SELL"},
		},
		{
			name:        "transaction filtering by type",
			queryParams: "?type=BUY&limit=10&offset=0",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock filtered transactions query
				rows := sqlmock.NewRows([]string{
					"id", "transaction_type", "quantity", "price", "fees",
					"total_amount", "transaction_date", "notes", "symbol", "name",
				}).
					AddRow("tx1", "BUY", 10.0, 150.0, 1.0, 1501.0, "2024-01-01", "Test buy", "AAPL", "Apple Inc.")

				mock.ExpectQuery("SELECT (.+) FROM transactions t JOIN assets a ON t.asset_id = a.id WHERE t.user_id = \\$1 AND t.transaction_type = \\$2 ORDER BY t.transaction_date DESC LIMIT \\$3 OFFSET \\$4").
					WithArgs("user1", "BUY", "10", "0").
					WillReturnRows(rows)

				// Mock count query
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM transactions t JOIN assets a ON t.asset_id = a.id WHERE t.user_id = \\$1 AND t.transaction_type = \\$2").
					WithArgs("user1", "BUY").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"transactions", "BUY"},
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
			router.GET("/transactions", handler.GetTransactions)

			req, _ := http.NewRequest("GET", "/transactions"+tt.queryParams, nil)
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

func TestGetTransactions_NilDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		DB:     nil, // nil database
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	router := gin.New()
	router.GET("/transactions", handler.GetTransactions)

	req, _ := http.NewRequest("GET", "/transactions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to fetch transactions")
}

// TestCreateTransaction tests the CreateTransaction handler
func TestCreateTransaction_BuyOrder(t *testing.T) {
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

	// Setup mocks for successful buy transaction
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

	mock.ExpectQuery("SELECT id FROM assets WHERE symbol = \\$1").
		WithArgs("AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("asset1"))

	mock.ExpectBegin()

	// total_amount = 10 * 150 + 1 = 1501 for BUY
	mock.ExpectQuery("INSERT INTO transactions \\(user_id, asset_id, transaction_type, quantity, price, fees, total_amount, notes\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6, \\$7, \\$8\\) RETURNING id").
		WithArgs("user1", "asset1", "BUY", 10.0, 150.0, 1.0, 1501.0, "Test buy transaction").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("tx1"))

	mock.ExpectExec("INSERT INTO portfolio_holdings \\(user_id, asset_id, quantity, average_cost\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\) ON CONFLICT \\(user_id, asset_id\\) DO UPDATE SET (.+)").
		WithArgs("user1", "asset1", 10.0, 150.0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	router := gin.New()
	router.POST("/transactions", handler.CreateTransaction)

	requestBody := map[string]interface{}{
		"symbol":           "AAPL",
		"transaction_type": "BUY",
		"quantity":         10.0,
		"price":            150.0,
		"fees":             1.0,
		"notes":            "Test buy transaction",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Transaction created successfully")
	assert.Contains(t, w.Body.String(), "tx1")
	assert.Contains(t, w.Body.String(), "AAPL")
	assert.Contains(t, w.Body.String(), "BUY")
	assert.Contains(t, w.Body.String(), "1501")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTransaction_SellOrder(t *testing.T) {
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

	// Setup mocks for successful sell transaction
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

	mock.ExpectQuery("SELECT id FROM assets WHERE symbol = \\$1").
		WithArgs("AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("asset1"))

	mock.ExpectBegin()

	// total_amount = 5 * 160 - 1 = 799 for SELL
	mock.ExpectQuery("INSERT INTO transactions \\(user_id, asset_id, transaction_type, quantity, price, fees, total_amount, notes\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6, \\$7, \\$8\\) RETURNING id").
		WithArgs("user1", "asset1", "SELL", 5.0, 160.0, 1.0, 799.0, "Test sell transaction").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("tx2"))

	// Mock current holdings check for SELL
	mock.ExpectQuery("SELECT quantity FROM portfolio_holdings WHERE user_id = \\$1 AND asset_id = \\$2").
		WithArgs("user1", "asset1").
		WillReturnRows(sqlmock.NewRows([]string{"quantity"}).AddRow(10.0))

	// Mock portfolio holdings update for SELL (10 - 5 = 5 remaining)
	mock.ExpectExec("UPDATE portfolio_holdings SET quantity = \\$1, updated_at = NOW\\(\\) WHERE user_id = \\$2 AND asset_id = \\$3").
		WithArgs(5.0, "user1", "asset1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	router := gin.New()
	router.POST("/transactions", handler.CreateTransaction)

	requestBody := map[string]interface{}{
		"symbol":           "AAPL",
		"transaction_type": "SELL",
		"quantity":         5.0,
		"price":            160.0,
		"fees":             1.0,
		"notes":            "Test sell transaction",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Transaction created successfully")
	assert.Contains(t, w.Body.String(), "tx2")
	assert.Contains(t, w.Body.String(), "AAPL")
	assert.Contains(t, w.Body.String(), "SELL")
	assert.Contains(t, w.Body.String(), "799")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTransaction_InsufficientHoldings(t *testing.T) {
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

	// Setup mocks for insufficient holdings
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

	mock.ExpectQuery("SELECT id FROM assets WHERE symbol = \\$1").
		WithArgs("AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("asset1"))

	mock.ExpectBegin()

	mock.ExpectQuery("INSERT INTO transactions \\(user_id, asset_id, transaction_type, quantity, price, fees, total_amount, notes\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6, \\$7, \\$8\\) RETURNING id").
		WithArgs("user1", "asset1", "SELL", 15.0, 160.0, 1.0, 2399.0, "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("tx4"))

	// Mock current holdings check (only 10 available)
	mock.ExpectQuery("SELECT quantity FROM portfolio_holdings WHERE user_id = \\$1 AND asset_id = \\$2").
		WithArgs("user1", "asset1").
		WillReturnRows(sqlmock.NewRows([]string{"quantity"}).AddRow(10.0))

	// Expect rollback due to insufficient holdings
	mock.ExpectRollback()

	router := gin.New()
	router.POST("/transactions", handler.CreateTransaction)

	requestBody := map[string]interface{}{
		"symbol":           "AAPL",
		"transaction_type": "SELL",
		"quantity":         15.0, // More than available
		"price":            160.0,
		"fees":             1.0,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Insufficient holdings to sell")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTransaction_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	router := gin.New()
	router.POST("/transactions", handler.CreateTransaction)

	// Test missing required symbol
	requestBody := map[string]interface{}{
		"transaction_type": "BUY",
		"quantity":         10.0,
		"price":            150.0,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Symbol")
}

// TestGetTransaction tests the GetTransaction handler
func TestGetTransaction_Success(t *testing.T) {
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

	// Mock transaction query
	rows := sqlmock.NewRows([]string{
		"id", "transaction_type", "quantity", "price", "fees",
		"total_amount", "transaction_date", "notes", "symbol", "name", "asset_type",
	}).AddRow("tx1", "BUY", 10.0, 150.0, 1.0, 1501.0, "2024-01-01", "Test transaction", "AAPL", "Apple Inc.", "STOCK")

	mock.ExpectQuery("SELECT (.+) FROM transactions t JOIN assets a ON t.asset_id = a.id WHERE t.id = \\$1 AND t.user_id = \\$2").
		WithArgs("tx1", "user1").
		WillReturnRows(rows)

	router := gin.New()
	router.GET("/transactions/:id", handler.GetTransaction)

	req, _ := http.NewRequest("GET", "/transactions/tx1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "tx1")
	assert.Contains(t, w.Body.String(), "BUY")
	assert.Contains(t, w.Body.String(), "AAPL")
	assert.Contains(t, w.Body.String(), "Apple Inc.")
	assert.Contains(t, w.Body.String(), "STOCK")
	assert.Contains(t, w.Body.String(), "1501")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTransaction_NotFound(t *testing.T) {
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

	// Mock transaction query that returns no rows
	mock.ExpectQuery("SELECT (.+) FROM transactions t JOIN assets a ON t.asset_id = a.id WHERE t.id = \\$1 AND t.user_id = \\$2").
		WithArgs("nonexistent", "user1").
		WillReturnError(sql.ErrNoRows)

	router := gin.New()
	router.GET("/transactions/:id", handler.GetTransaction)

	req, _ := http.NewRequest("GET", "/transactions/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Transaction not found")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUpdateTransaction tests the UpdateTransaction handler
func TestUpdateTransaction_Success(t *testing.T) {
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

	// Mock existing transaction query
	mock.ExpectQuery("SELECT quantity, price, fees, notes, transaction_type FROM transactions WHERE id = \\$1 AND user_id = \\$2").
		WithArgs("tx1", "user1").
		WillReturnRows(sqlmock.NewRows([]string{"quantity", "price", "fees", "notes", "transaction_type"}).
			AddRow(10.0, 150.0, 1.0, "Old notes", "BUY"))

	// Mock update query - new total: 15 * 150 + 1 = 2251
	mock.ExpectExec("UPDATE transactions SET quantity = \\$1, price = \\$2, fees = \\$3, notes = \\$4, total_amount = \\$5 WHERE id = \\$6 AND user_id = \\$7").
		WithArgs(15.0, 150.0, 1.0, "Updated notes", 2251.0, "tx1", "user1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	router := gin.New()
	router.PUT("/transactions/:id", handler.UpdateTransaction)

	requestBody := map[string]interface{}{
		"quantity": 15.0,
		"notes":    "Updated notes",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/transactions/tx1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Transaction updated successfully")
	assert.Contains(t, w.Body.String(), "tx1")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDeleteTransaction tests the DeleteTransaction handler
func TestDeleteTransaction_Success(t *testing.T) {
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

	// Mock transaction existence check query
	mock.ExpectQuery("SELECT t.transaction_type, t.quantity, a.symbol FROM transactions t JOIN assets a ON t.asset_id = a.id WHERE t.id = \\$1 AND t.user_id = \\$2").
		WithArgs("tx1", "user1").
		WillReturnRows(sqlmock.NewRows([]string{"transaction_type", "quantity", "symbol"}).AddRow("BUY", 10.0, "AAPL"))

	// Mock delete query
	mock.ExpectExec("DELETE FROM transactions WHERE id = \\$1 AND user_id = \\$2").
		WithArgs("tx1", "user1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	router := gin.New()
	router.DELETE("/transactions/:id", handler.DeleteTransaction)

	req, _ := http.NewRequest("DELETE", "/transactions/tx1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Transaction deleted successfully")
	assert.Contains(t, w.Body.String(), "tx1")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestGetNotifications tests the GetNotifications handler
func TestGetNotifications_Success(t *testing.T) {
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

	// Mock notifications query
	rows := sqlmock.NewRows([]string{
		"id", "title", "message", "notification_type", "is_read", "created_at",
	}).
		AddRow("notif1", "Price Alert", "AAPL reached target", "PRICE_ALERT", false, "2024-01-01").
		AddRow("notif2", "Portfolio Update", "Holdings updated", "PORTFOLIO_UPDATE", true, "2024-01-02")

	mock.ExpectQuery("SELECT id, title, message, notification_type, is_read, created_at FROM notifications WHERE user_id = \\$1 ORDER BY created_at DESC LIMIT \\$2").
		WithArgs("user1", "50").
		WillReturnRows(rows)

	router := gin.New()
	router.GET("/notifications", handler.GetNotifications)

	req, _ := http.NewRequest("GET", "/notifications", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "notifications")
	assert.Contains(t, w.Body.String(), "unread_count")
	assert.Contains(t, w.Body.String(), "Price Alert")
	assert.Contains(t, w.Body.String(), "Portfolio Update")
	assert.Contains(t, w.Body.String(), "1") // 1 unread

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestMarkNotificationRead tests the MarkNotificationRead handler
func TestMarkNotificationRead_Success(t *testing.T) {
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

	// Mock notification check (unread)
	mock.ExpectQuery("SELECT is_read FROM notifications WHERE id = \\$1 AND user_id = \\$2").
		WithArgs("notif1", "user1").
		WillReturnRows(sqlmock.NewRows([]string{"is_read"}).AddRow(false))

	// Mock update query
	mock.ExpectExec("UPDATE notifications SET is_read = true WHERE id = \\$1 AND user_id = \\$2").
		WithArgs("notif1", "user1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	router := gin.New()
	router.PUT("/notifications/:id/read", handler.MarkNotificationRead)

	req, _ := http.NewRequest("PUT", "/notifications/notif1/read", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Notification marked as read")
	assert.Contains(t, w.Body.String(), "notif1")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUpdateNotificationSettings tests the UpdateNotificationSettings handler
func TestUpdateNotificationSettings_Success(t *testing.T) {
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

	router := gin.New()
	router.PUT("/settings/notifications", handler.UpdateNotificationSettings)

	requestBody := map[string]interface{}{
		"price_alerts":        true,
		"portfolio_updates":   true,
		"market_news":         false,
		"performance_reports": true,
		"email_enabled":       true,
		"sms_enabled":         false,
		"web_push_enabled":    true,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/settings/notifications", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Notification settings updated successfully")
	assert.Contains(t, w.Body.String(), "price_alerts")
	assert.Contains(t, w.Body.String(), "portfolio_updates")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestWebSocketHandler tests the WebSocketHandler
func TestWebSocketHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock WebSocket hub
	hub := &services.WebSocketHub{
		Register:   make(chan *services.Client, 1),
		Unregister: make(chan *services.Client, 1),
	}

	mockServices := &services.Services{
		WebSocket: hub,
		Logger:    logger,
	}

	handler := NewHandler(mockServices, logger)

	router := gin.New()
	router.GET("/ws", handler.WebSocketHandler)

	// Test with a regular HTTP request (should fail to upgrade)
	req, _ := http.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 400 because it's not a WebSocket upgrade request
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to upgrade to WebSocket")
}

// TestWebSocketHandler_MissingService tests WebSocket handler without WebSocket service
func TestWebSocketHandler_MissingService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		WebSocket: nil, // No WebSocket service
		Logger:    logger,
	}

	handler := NewHandler(mockServices, logger)

	router := gin.New()
	router.GET("/ws", handler.WebSocketHandler)

	// Test with a regular HTTP request (no upgrade headers)
	req, _ := http.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 400 because WebSocket service is nil
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "WebSocket service not available")
}

// TestUtilityFunctions tests the utility helper functions
func TestGetAssetIDBySymbol_Success(t *testing.T) {
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

	mock.ExpectQuery("SELECT id FROM assets WHERE symbol = \\$1").
		WithArgs("AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("asset1"))

	assetID, err := handler.getAssetIDBySymbol("AAPL")

	assert.NoError(t, err)
	assert.Equal(t, "asset1", assetID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserID_Success(t *testing.T) {
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

	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("default_user").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

	userID, err := handler.getUserID("default_user")

	assert.NoError(t, err)
	assert.Equal(t, "user1", userID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTransaction_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing required symbol",
			requestBody:    map[string]interface{}{"transaction_type": "BUY", "quantity": 10.0, "price": 150.0},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Symbol",
		},
		{
			name:           "invalid transaction type",
			requestBody:    map[string]interface{}{"symbol": "AAPL", "transaction_type": "INVALID", "quantity": 10.0, "price": 150.0},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "TransactionType",
		},
		{
			name:           "negative quantity",
			requestBody:    map[string]interface{}{"symbol": "AAPL", "transaction_type": "BUY", "quantity": -10.0, "price": 150.0},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Quantity",
		},
		{
			name:           "negative price",
			requestBody:    map[string]interface{}{"symbol": "AAPL", "transaction_type": "BUY", "quantity": 10.0, "price": -150.0},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Price",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/transactions", handler.CreateTransaction)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/transactions", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedError)
		})
	}
}

// TestGetTransaction tests the GetTransaction handler
func TestGetTransaction(t *testing.T) {
	tests := []struct {
		name           string
		transactionID  string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   []string
	}{
		{
			name:          "successful transaction retrieval",
			transactionID: "tx1",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock transaction query
				rows := sqlmock.NewRows([]string{
					"id", "transaction_type", "quantity", "price", "fees",
					"total_amount", "transaction_date", "notes", "symbol", "name", "asset_type",
				}).AddRow("tx1", "BUY", 10.0, 150.0, 1.0, 1501.0, "2024-01-01", "Test transaction", "AAPL", "Apple Inc.", "STOCK")

				mock.ExpectQuery("SELECT (.+) FROM transactions t JOIN assets a ON t.asset_id = a.id WHERE t.id = \\$1 AND t.user_id = \\$2").
					WithArgs("tx1", "user1").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"tx1", "BUY", "AAPL", "Apple Inc.", "STOCK", "1501"},
		},
		{
			name:          "transaction not found",
			transactionID: "nonexistent",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock transaction query that returns no rows
				mock.ExpectQuery("SELECT (.+) FROM transactions t JOIN assets a ON t.asset_id = a.id WHERE t.id = \\$1 AND t.user_id = \\$2").
					WithArgs("nonexistent", "user1").
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Transaction not found"},
		},
		{
			name:           "empty transaction ID",
			transactionID:  "",
			setupMock:      func(mock sqlmock.Sqlmock) {}, // No mock needed
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"404 page not found"},
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
			router.GET("/transactions/:id", handler.GetTransaction)

			url := "/transactions/" + tt.transactionID
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

// TestUpdateTransaction tests the UpdateTransaction handler
func TestUpdateTransaction(t *testing.T) {
	tests := []struct {
		name           string
		transactionID  string
		requestBody    map[string]interface{}
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   []string
	}{
		{
			name:          "successful partial update",
			transactionID: "tx1",
			requestBody: map[string]interface{}{
				"quantity": 15.0,
				"notes":    "Updated notes",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock existing transaction query
				mock.ExpectQuery("SELECT quantity, price, fees, notes, transaction_type FROM transactions WHERE id = \\$1 AND user_id = \\$2").
					WithArgs("tx1", "user1").
					WillReturnRows(sqlmock.NewRows([]string{"quantity", "price", "fees", "notes", "transaction_type"}).
						AddRow(10.0, 150.0, 1.0, "Old notes", "BUY"))

				// Mock update query - new total: 15 * 150 + 1 = 2251
				mock.ExpectExec("UPDATE transactions SET quantity = \\$1, price = \\$2, fees = \\$3, notes = \\$4, total_amount = \\$5 WHERE id = \\$6 AND user_id = \\$7").
					WithArgs(15.0, 150.0, 1.0, "Updated notes", 2251.0, "tx1", "user1").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Transaction updated successfully", "tx1"},
		},
		{
			name:          "update all fields",
			transactionID: "tx2",
			requestBody: map[string]interface{}{
				"quantity": 8.0,
				"price":    200.0,
				"fees":     2.0,
				"notes":    "Fully updated transaction",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock existing transaction query
				mock.ExpectQuery("SELECT quantity, price, fees, notes, transaction_type FROM transactions WHERE id = \\$1 AND user_id = \\$2").
					WithArgs("tx2", "user1").
					WillReturnRows(sqlmock.NewRows([]string{"quantity", "price", "fees", "notes", "transaction_type"}).
						AddRow(5.0, 160.0, 1.0, "Old notes", "SELL"))

				// Mock update query - new total for SELL: 8 * 200 - 2 = 1598
				mock.ExpectExec("UPDATE transactions SET quantity = \\$1, price = \\$2, fees = \\$3, notes = \\$4, total_amount = \\$5 WHERE id = \\$6 AND user_id = \\$7").
					WithArgs(8.0, 200.0, 2.0, "Fully updated transaction", 1598.0, "tx2", "user1").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Transaction updated successfully", "tx2"},
		},
		{
			name:          "transaction not found",
			transactionID: "nonexistent",
			requestBody: map[string]interface{}{
				"quantity": 10.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock existing transaction query that returns no rows
				mock.ExpectQuery("SELECT quantity, price, fees, notes, transaction_type FROM transactions WHERE id = \\$1 AND user_id = \\$2").
					WithArgs("nonexistent", "user1").
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Transaction not found"},
		},
		{
			name:           "no fields provided for update",
			transactionID:  "tx1",
			requestBody:    map[string]interface{}{},
			setupMock:      func(mock sqlmock.Sqlmock) {}, // No mock needed
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{"At least one field must be provided for update"},
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
			router.PUT("/transactions/:id", handler.UpdateTransaction)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PUT", "/transactions/"+tt.transactionID, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
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

// TestDeleteTransaction tests the DeleteTransaction handler
func TestDeleteTransaction(t *testing.T) {
	tests := []struct {
		name           string
		transactionID  string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   []string
	}{
		{
			name:          "successful transaction deletion",
			transactionID: "tx1",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock transaction existence check query
				mock.ExpectQuery("SELECT t.transaction_type, t.quantity, a.symbol FROM transactions t JOIN assets a ON t.asset_id = a.id WHERE t.id = \\$1 AND t.user_id = \\$2").
					WithArgs("tx1", "user1").
					WillReturnRows(sqlmock.NewRows([]string{"transaction_type", "quantity", "symbol"}).AddRow("BUY", 10.0, "AAPL"))

				// Mock delete query
				mock.ExpectExec("DELETE FROM transactions WHERE id = \\$1 AND user_id = \\$2").
					WithArgs("tx1", "user1").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Transaction deleted successfully", "tx1"},
		},
		{
			name:          "transaction not found",
			transactionID: "nonexistent",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock transaction existence check query that returns no rows
				mock.ExpectQuery("SELECT t.transaction_type, t.quantity, a.symbol FROM transactions t JOIN assets a ON t.asset_id = a.id WHERE t.id = \\$1 AND t.user_id = \\$2").
					WithArgs("nonexistent", "user1").
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Transaction not found"},
		},
		{
			name:           "empty transaction ID",
			transactionID:  "",
			setupMock:      func(mock sqlmock.Sqlmock) {}, // No mock needed
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"404 page not found"},
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
			router.DELETE("/transactions/:id", handler.DeleteTransaction)

			url := "/transactions/" + tt.transactionID
			req, _ := http.NewRequest("DELETE", url, nil)
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

// TestGetNotifications tests the GetNotifications handler
func TestGetNotifications(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   []string
	}{
		{
			name:        "successful notifications listing",
			queryParams: "?limit=10",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock notifications query
				rows := sqlmock.NewRows([]string{
					"id", "title", "message", "notification_type", "is_read", "created_at",
				}).
					AddRow("notif1", "Price Alert", "AAPL reached target", "PRICE_ALERT", false, "2024-01-01").
					AddRow("notif2", "Portfolio Update", "Holdings updated", "PORTFOLIO_UPDATE", true, "2024-01-02")

				mock.ExpectQuery("SELECT id, title, message, notification_type, is_read, created_at FROM notifications WHERE user_id = \\$1 ORDER BY created_at DESC LIMIT \\$2").
					WithArgs("user1", "10").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"notifications", "unread_count", "Price Alert", "Portfolio Update", "1"}, // 1 unread
		},
		{
			name:        "unread notifications only",
			queryParams: "?unread_only=true&limit=10",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock unread notifications query
				rows := sqlmock.NewRows([]string{
					"id", "title", "message", "notification_type", "is_read", "created_at",
				}).
					AddRow("notif1", "Price Alert", "AAPL reached target", "PRICE_ALERT", false, "2024-01-01")

				mock.ExpectQuery("SELECT id, title, message, notification_type, is_read, created_at FROM notifications WHERE user_id = \\$1 AND is_read = false ORDER BY created_at DESC LIMIT \\$2").
					WithArgs("user1", "10").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"notifications", "Price Alert", "unread_count", "1"},
		},
		{
			name:        "empty notifications list",
			queryParams: "?limit=10",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock empty notifications query
				rows := sqlmock.NewRows([]string{
					"id", "title", "message", "notification_type", "is_read", "created_at",
				})

				mock.ExpectQuery("SELECT id, title, message, notification_type, is_read, created_at FROM notifications WHERE user_id = \\$1 ORDER BY created_at DESC LIMIT \\$2").
					WithArgs("user1", "10").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"notifications", "total", "unread_count", "0"},
		},
		{
			name:        "all notifications without limit",
			queryParams: "?limit=all",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock notifications query without limit
				rows := sqlmock.NewRows([]string{
					"id", "title", "message", "notification_type", "is_read", "created_at",
				}).
					AddRow("notif1", "Alert 1", "Message 1", "PRICE_ALERT", false, "2024-01-01").
					AddRow("notif2", "Alert 2", "Message 2", "PORTFOLIO_UPDATE", false, "2024-01-02")

				mock.ExpectQuery("SELECT id, title, message, notification_type, is_read, created_at FROM notifications WHERE user_id = \\$1 ORDER BY created_at DESC").
					WithArgs("user1").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"notifications", "Alert 1", "Alert 2", "unread_count", "2"},
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
			router.GET("/notifications", handler.GetNotifications)

			req, _ := http.NewRequest("GET", "/notifications"+tt.queryParams, nil)
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

func TestGetNotifications_NilDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		DB:     nil, // nil database
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	router := gin.New()
	router.GET("/notifications", handler.GetNotifications)

	req, _ := http.NewRequest("GET", "/notifications", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to fetch notifications")
}

// TestMarkNotificationRead tests the MarkNotificationRead handler
func TestMarkNotificationRead(t *testing.T) {
	tests := []struct {
		name           string
		notificationID string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   []string
	}{
		{
			name:           "successful notification mark as read",
			notificationID: "notif1",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock notification check (unread)
				mock.ExpectQuery("SELECT is_read FROM notifications WHERE id = \\$1 AND user_id = \\$2").
					WithArgs("notif1", "user1").
					WillReturnRows(sqlmock.NewRows([]string{"is_read"}).AddRow(false))

				// Mock update query
				mock.ExpectExec("UPDATE notifications SET is_read = true WHERE id = \\$1 AND user_id = \\$2").
					WithArgs("notif1", "user1").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Notification marked as read", "notif1"},
		},
		{
			name:           "already read notification",
			notificationID: "notif2",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock notification check (already read)
				mock.ExpectQuery("SELECT is_read FROM notifications WHERE id = \\$1 AND user_id = \\$2").
					WithArgs("notif2", "user1").
					WillReturnRows(sqlmock.NewRows([]string{"is_read"}).AddRow(true))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Notification already marked as read", "notif2"},
		},
		{
			name:           "notification not found",
			notificationID: "nonexistent",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))

				// Mock notification check that returns no rows
				mock.ExpectQuery("SELECT is_read FROM notifications WHERE id = \\$1 AND user_id = \\$2").
					WithArgs("nonexistent", "user1").
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Notification not found"},
		},
		{
			name:           "empty notification ID",
			notificationID: "",
			setupMock:      func(mock sqlmock.Sqlmock) {}, // No mock needed
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{"Notification ID is required"},
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
			router.PUT("/notifications/:id/read", handler.MarkNotificationRead)

			url := "/notifications/" + tt.notificationID + "/read"
			req, _ := http.NewRequest("PUT", url, nil)
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

// TestUpdateNotificationSettings tests the UpdateNotificationSettings handler
func TestUpdateNotificationSettings(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   []string
	}{
		{
			name: "successful settings update",
			requestBody: map[string]interface{}{
				"price_alerts":        true,
				"portfolio_updates":   true,
				"market_news":         false,
				"performance_reports": true,
				"email_enabled":       true,
				"sms_enabled":         false,
				"web_push_enabled":    true,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))
			},
			expectedStatus: http.StatusOK,
			expectedBody: []string{
				"Notification settings updated successfully",
				"price_alerts", "portfolio_updates", "market_news",
				"performance_reports", "email_enabled", "sms_enabled", "web_push_enabled",
			},
		},
		{
			name: "all settings disabled",
			requestBody: map[string]interface{}{
				"price_alerts":        false,
				"portfolio_updates":   false,
				"market_news":         false,
				"performance_reports": false,
				"email_enabled":       false,
				"sms_enabled":         false,
				"web_push_enabled":    false,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock user ID query
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Notification settings updated successfully", "settings"},
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
			router.PUT("/settings/notifications", handler.UpdateNotificationSettings)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PUT", "/settings/notifications", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
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

func TestUpdateNotificationSettings_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockServices := &services.Services{
		Logger: logger,
	}

	handler := NewHandler(mockServices, logger)

	router := gin.New()
	router.PUT("/settings/notifications", handler.UpdateNotificationSettings)

	// Invalid JSON body
	req, _ := http.NewRequest("PUT", "/settings/notifications", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}

// TestWebSocketHandler tests the WebSocketHandler
func TestWebSocketHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	// Create mock WebSocket hub
	hub := &services.WebSocketHub{
		Register:   make(chan *services.Client, 1),
		Unregister: make(chan *services.Client, 1),
	}

	mockServices := &services.Services{
		WebSocket: hub,
		Logger:    logger,
	}

	handler := NewHandler(mockServices, logger)

	router := gin.New()
	router.GET("/ws", handler.WebSocketHandler)

	req, _ := http.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 400 because it's not a WebSocket upgrade request
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to upgrade to WebSocket")
}

// TestUtilityFunctions tests the utility helper functions
func TestGetAssetIDBySymbol(t *testing.T) {
	tests := []struct {
		name          string
		symbol        string
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
		expectedID    string
	}{
		{
			name:   "existing asset",
			symbol: "AAPL",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM assets WHERE symbol = \\$1").
					WithArgs("AAPL").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("asset1"))
			},
			expectedError: false,
			expectedID:    "asset1",
		},
		{
			name:   "non-existent asset",
			symbol: "UNKNOWN",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM assets WHERE symbol = \\$1").
					WithArgs("UNKNOWN").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: true,
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

			assetID, err := handler.getAssetIDBySymbol(tt.symbol)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, assetID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
		expectedID    string
	}{
		{
			name:     "existing user",
			username: "default_user",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("default_user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user1"))
			},
			expectedError: false,
			expectedID:    "user1",
		},
		{
			name:     "non-existent user",
			username: "unknown_user",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
					WithArgs("unknown_user").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: true,
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

			userID, err := handler.getUserID(tt.username)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
