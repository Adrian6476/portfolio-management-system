package handlers

import (
	"database/sql"
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

// Test data constants for consistent testing
var (
	testUserID    = "test-user-123"
	testAssetID   = "test-asset-456"
	testHoldingID = "test-holding-789"
	defaultUser   = "default_user"
)

// Helper function to create a test handler with mock database
func createTestHandler(t *testing.T) (*Handler, sqlmock.Sqlmock, func()) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	mockServices := &services.Services{
		DB:      db,
		Finnhub: nil, // We'll set this when needed in specific tests
		Logger:  logger,
	}

	handler := NewHandler(mockServices, logger)

	cleanup := func() {
		db.Close()
	}

	return handler, mock, cleanup
}

// Helper function to create test router with handler
func createTestRouter(handler *Handler, method, path string, handlerFunc gin.HandlerFunc) *gin.Engine {
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

// TestGetPortfolio tests the GetPortfolio handler
func TestGetPortfolio(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   []string
		dbNil          bool
	}{
		{
			name: "successful portfolio fetch",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "quantity", "average_cost", "purchase_date"}).
					AddRow("1", "AAPL", "Apple Inc.", "STOCK", 10.0, 150.0, "2024-01-01").
					AddRow("2", "GOOGL", "Alphabet Inc.", "STOCK", 5.0, 2800.0, "2024-01-02")

				mock.ExpectQuery(`SELECT (.+) FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id JOIN users u ON ph.user_id = u.id WHERE u.username = (.+) ORDER BY ph.created_at DESC`).
					WithArgs(defaultUser).
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"holdings", "total_holdings", "AAPL", "GOOGL"},
		},
		{
			name: "empty portfolio",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "symbol", "name", "asset_type", "quantity", "average_cost", "purchase_date"})
				mock.ExpectQuery(`SELECT (.+) FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id JOIN users u ON ph.user_id = u.id WHERE u.username = (.+) ORDER BY ph.created_at DESC`).
					WithArgs(defaultUser).
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"holdings", "total_holdings"},
		},
		{
			name: "database query error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT (.+) FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id JOIN users u ON ph.user_id = u.id WHERE u.username = (.+) ORDER BY ph.created_at DESC`).
					WithArgs(defaultUser).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to fetch portfolio"},
		},
		{
			name:           "nil database connection",
			setupMock:      func(mock sqlmock.Sqlmock) {},
			dbNil:          true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to fetch portfolio"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mock, cleanup := createTestHandler(t)
			defer cleanup()

			if tt.dbNil {
				handler.services.DB = nil
			} else {
				tt.setupMock(mock)
			}

			router := createTestRouter(handler, "GET", "/portfolio", handler.GetPortfolio)

			req, _ := http.NewRequest("GET", "/portfolio", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			for _, expected := range tt.expectedBody {
				assert.Contains(t, w.Body.String(), expected)
			}

			if !tt.dbNil {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

// TestAddHolding tests the AddHolding handler
func TestAddHolding(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   []string
		dbNil          bool
	}{
		{
			name:        "successful holding addition with existing asset",
			requestBody: `{"symbol": "AAPL", "quantity": 10.0, "average_cost": 150.0}`,
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Asset lookup (exists)
				mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = (.+)`).
					WithArgs("AAPL").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testAssetID))

				// Insert/update holding with cost averaging
				mock.ExpectExec(`INSERT INTO portfolio_holdings \(user_id, asset_id, quantity, average_cost\) VALUES \(.+\) ON CONFLICT \(user_id, asset_id\) DO UPDATE SET (.+)`).
					WithArgs(testUserID, testAssetID, 10.0, 150.0).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   []string{"Holding added successfully", "AAPL", "10", "150"},
		},
		{
			name:        "successful holding addition with new asset creation",
			requestBody: `{"symbol": "TSLA", "quantity": 5.0, "average_cost": 200.0}`,
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Asset lookup (doesn't exist)
				mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = (.+)`).
					WithArgs("TSLA").
					WillReturnError(sql.ErrNoRows)

				// Create new asset (2 parameters: symbol and name, asset_type and currency are hardcoded)
				mock.ExpectQuery(`INSERT INTO assets \(symbol, name, asset_type, currency\) VALUES \(.+\) RETURNING id`).
					WithArgs("TSLA", "TSLA").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testAssetID))

				// Insert holding
				mock.ExpectExec(`INSERT INTO portfolio_holdings \(user_id, asset_id, quantity, average_cost\) VALUES \(.+\) ON CONFLICT \(user_id, asset_id\) DO UPDATE SET (.+)`).
					WithArgs(testUserID, testAssetID, 5.0, 200.0).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   []string{"Holding added successfully", "TSLA", "5", "200"},
		},
		{
			name:           "invalid JSON input",
			requestBody:    `{"symbol": "AAPL", "quantity": -5, "average_cost": 150.0}`, // negative quantity
			setupMock:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{"Field validation", "failed"},
		},
		{
			name:           "missing required fields",
			requestBody:    `{"symbol": "AAPL"}`, // missing quantity and average_cost
			setupMock:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{"Field validation", "required"},
		},
		{
			name:        "user not found error",
			requestBody: `{"symbol": "AAPL", "quantity": 10.0, "average_cost": 150.0}`,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to get user"},
		},
		{
			name:        "asset creation failure",
			requestBody: `{"symbol": "INVALID", "quantity": 10.0, "average_cost": 150.0}`,
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Asset lookup (doesn't exist)
				mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = (.+)`).
					WithArgs("INVALID").
					WillReturnError(sql.ErrNoRows)

				// Asset creation fails
				mock.ExpectQuery(`INSERT INTO assets \(symbol, name, asset_type, currency\) VALUES \(.+\) RETURNING id`).
					WithArgs("INVALID", "INVALID").
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to create asset"},
		},
		{
			name:        "holding insertion failure",
			requestBody: `{"symbol": "AAPL", "quantity": 10.0, "average_cost": 150.0}`,
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Asset lookup (exists)
				mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = (.+)`).
					WithArgs("AAPL").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testAssetID))

				// Insert holding fails
				mock.ExpectExec(`INSERT INTO portfolio_holdings \(user_id, asset_id, quantity, average_cost\) VALUES \(.+\) ON CONFLICT \(user_id, asset_id\) DO UPDATE SET (.+)`).
					WithArgs(testUserID, testAssetID, 10.0, 150.0).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to add holding"},
		},
		{
			name:           "nil database connection",
			requestBody:    `{"symbol": "AAPL", "quantity": 10.0, "average_cost": 150.0}`,
			setupMock:      func(mock sqlmock.Sqlmock) {},
			dbNil:          true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to get user"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mock, cleanup := createTestHandler(t)
			defer cleanup()

			if tt.dbNil {
				handler.services.DB = nil
			} else {
				tt.setupMock(mock)
			}

			router := createTestRouter(handler, "POST", "/portfolio/holdings", handler.AddHolding)

			req, _ := http.NewRequest("POST", "/portfolio/holdings", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			for _, expected := range tt.expectedBody {
				assert.Contains(t, w.Body.String(), expected)
			}

			if !tt.dbNil && tt.expectedStatus != http.StatusBadRequest {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

// TestUpdateHolding tests the UpdateHolding handler
func TestUpdateHolding(t *testing.T) {
	tests := []struct {
		name           string
		holdingID      string
		requestBody    string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   []string
		dbNil          bool
	}{
		{
			name:        "successful quantity update only",
			holdingID:   testHoldingID,
			requestBody: `{"quantity": 15.0}`,
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Check holding exists and get current values
				mock.ExpectQuery(`SELECT ph.quantity, ph.average_cost, a.symbol FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)`).
					WithArgs(testHoldingID, testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"quantity", "average_cost", "symbol"}).AddRow(10.0, 150.0, "AAPL"))

				// Update holding
				mock.ExpectExec(`UPDATE portfolio_holdings SET quantity = (.+), average_cost = (.+), updated_at = NOW\(\) WHERE id = (.+) AND user_id = (.+)`).
					WithArgs(15.0, 150.0, testHoldingID, testUserID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Holding updated successfully", "AAPL", "15"},
		},
		{
			name:        "successful average cost update only",
			holdingID:   testHoldingID,
			requestBody: `{"average_cost": 175.0}`,
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Check holding exists and get current values
				mock.ExpectQuery(`SELECT ph.quantity, ph.average_cost, a.symbol FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)`).
					WithArgs(testHoldingID, testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"quantity", "average_cost", "symbol"}).AddRow(10.0, 150.0, "AAPL"))

				// Update holding
				mock.ExpectExec(`UPDATE portfolio_holdings SET quantity = (.+), average_cost = (.+), updated_at = NOW\(\) WHERE id = (.+) AND user_id = (.+)`).
					WithArgs(10.0, 175.0, testHoldingID, testUserID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Holding updated successfully", "AAPL", "175"},
		},
		{
			name:        "successful update of both fields",
			holdingID:   testHoldingID,
			requestBody: `{"quantity": 20.0, "average_cost": 160.0}`,
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Check holding exists and get current values
				mock.ExpectQuery(`SELECT ph.quantity, ph.average_cost, a.symbol FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)`).
					WithArgs(testHoldingID, testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"quantity", "average_cost", "symbol"}).AddRow(10.0, 150.0, "AAPL"))

				// Update holding
				mock.ExpectExec(`UPDATE portfolio_holdings SET quantity = (.+), average_cost = (.+), updated_at = NOW\(\) WHERE id = (.+) AND user_id = (.+)`).
					WithArgs(20.0, 160.0, testHoldingID, testUserID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Holding updated successfully", "AAPL", "20", "160"},
		},
		{
			name:           "missing holding ID parameter",
			holdingID:      "",
			requestBody:    `{"quantity": 15.0}`,
			setupMock:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusNotFound, // 404 because empty path doesn't match route
			expectedBody:   []string{"404"},
		},
		{
			name:           "no fields provided for update",
			holdingID:      testHoldingID,
			requestBody:    `{}`,
			setupMock:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{"At least one field (quantity or average_cost) must be provided"},
		},
		{
			name:           "invalid negative quantity",
			holdingID:      testHoldingID,
			requestBody:    `{"quantity": -5.0}`,
			setupMock:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{"Field validation", "failed"},
		},
		{
			name:        "holding not found",
			holdingID:   "non-existent-id",
			requestBody: `{"quantity": 15.0}`,
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Check holding exists (not found)
				mock.ExpectQuery(`SELECT ph.quantity, ph.average_cost, a.symbol FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)`).
					WithArgs("non-existent-id", testUserID).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Holding not found"},
		},
		{
			name:        "user authorization failure",
			holdingID:   testHoldingID,
			requestBody: `{"quantity": 15.0}`,
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Check holding exists but belongs to different user
				mock.ExpectQuery(`SELECT ph.quantity, ph.average_cost, a.symbol FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)`).
					WithArgs(testHoldingID, testUserID).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Holding not found"},
		},
		{
			name:        "database update failure",
			holdingID:   testHoldingID,
			requestBody: `{"quantity": 15.0}`,
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Check holding exists and get current values
				mock.ExpectQuery(`SELECT ph.quantity, ph.average_cost, a.symbol FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)`).
					WithArgs(testHoldingID, testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"quantity", "average_cost", "symbol"}).AddRow(10.0, 150.0, "AAPL"))

				// Update fails
				mock.ExpectExec(`UPDATE portfolio_holdings SET quantity = (.+), average_cost = (.+), updated_at = NOW\(\) WHERE id = (.+) AND user_id = (.+)`).
					WithArgs(15.0, 150.0, testHoldingID, testUserID).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to update holding"},
		},
		{
			name:           "nil database connection",
			holdingID:      testHoldingID,
			requestBody:    `{"quantity": 15.0}`,
			setupMock:      func(mock sqlmock.Sqlmock) {},
			dbNil:          true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to update holding"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mock, cleanup := createTestHandler(t)
			defer cleanup()

			if tt.dbNil {
				handler.services.DB = nil
			} else {
				tt.setupMock(mock)
			}

			router := gin.New()
			router.PUT("/portfolio/holdings/:id", handler.UpdateHolding)

			url := "/portfolio/holdings/" + tt.holdingID
			req, _ := http.NewRequest("PUT", url, strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			for _, expected := range tt.expectedBody {
				assert.Contains(t, w.Body.String(), expected)
			}

			if !tt.dbNil && tt.expectedStatus != http.StatusBadRequest {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}
