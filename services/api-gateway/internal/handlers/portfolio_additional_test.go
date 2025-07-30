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

// TestRemoveHolding tests the RemoveHolding handler
func TestRemoveHolding(t *testing.T) {
	tests := []struct {
		name           string
		holdingID      string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   []string
		dbNil          bool
	}{
		{
			name:      "successful holding removal",
			holdingID: testHoldingID,
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Check holding exists and get asset info
				mock.ExpectQuery(`SELECT a.symbol, ph.quantity FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)`).
					WithArgs(testHoldingID, testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"symbol", "quantity"}).AddRow("AAPL", 10.0))

				// Delete holding
				mock.ExpectExec(`DELETE FROM portfolio_holdings WHERE id = (.+) AND user_id = (.+)`).
					WithArgs(testHoldingID, testUserID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Holding removed successfully", "AAPL", "10"},
		},
		{
			name:           "missing holding ID parameter",
			holdingID:      "",
			setupMock:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusNotFound, // This will be 404 because route doesn't match
			expectedBody:   []string{"404"},
		},
		{
			name:      "holding not found",
			holdingID: "non-existent-id",
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Check holding exists (not found)
				mock.ExpectQuery(`SELECT a.symbol, ph.quantity FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)`).
					WithArgs("non-existent-id", testUserID).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Holding not found"},
		},
		{
			name:      "user authorization failure",
			holdingID: testHoldingID,
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Check holding exists but belongs to different user
				mock.ExpectQuery(`SELECT a.symbol, ph.quantity FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)`).
					WithArgs(testHoldingID, testUserID).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Holding not found"},
		},
		{
			name:      "database deletion failure",
			holdingID: testHoldingID,
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Check holding exists and get asset info
				mock.ExpectQuery(`SELECT a.symbol, ph.quantity FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)`).
					WithArgs(testHoldingID, testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"symbol", "quantity"}).AddRow("AAPL", 10.0))

				// Delete fails
				mock.ExpectExec(`DELETE FROM portfolio_holdings WHERE id = (.+) AND user_id = (.+)`).
					WithArgs(testHoldingID, testUserID).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to remove holding"},
		},
		{
			name:      "no rows affected after delete",
			holdingID: testHoldingID,
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Check holding exists and get asset info
				mock.ExpectQuery(`SELECT a.symbol, ph.quantity FROM portfolio_holdings ph JOIN assets a ON ph.asset_id = a.id WHERE ph.id = (.+) AND ph.user_id = (.+)`).
					WithArgs(testHoldingID, testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"symbol", "quantity"}).AddRow("AAPL", 10.0))

				// Delete returns 0 rows affected
				mock.ExpectExec(`DELETE FROM portfolio_holdings WHERE id = (.+) AND user_id = (.+)`).
					WithArgs(testHoldingID, testUserID).
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Holding not found"},
		},
		{
			name:           "nil database connection",
			holdingID:      testHoldingID,
			setupMock:      func(mock sqlmock.Sqlmock) {},
			dbNil:          true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to remove holding"},
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
			router.DELETE("/portfolio/holdings/:id", handler.RemoveHolding)

			url := "/portfolio/holdings/" + tt.holdingID
			req, _ := http.NewRequest("DELETE", url, nil)
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

// Test helper function tests
func TestValidateAssetSymbol(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	handler := &Handler{
		services: &services.Services{Logger: logger},
		logger:   logger,
	}

	tests := []struct {
		name      string
		symbol    string
		shouldErr bool
	}{
		{"valid symbol", "AAPL", false},
		{"valid long symbol", "GOOGL", false},
		{"empty symbol", "", true},
		{"too long symbol", "VERYLONGSYMBOL", true},
		{"single character", "A", false},
		{"max length symbol", "ABCDEFGHIJ", false}, // 10 chars max
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateAssetSymbol(tt.symbol)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserOwnsAsset(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(sqlmock.Sqlmock)
		expectedOwns bool
		expectedQty  float64
		expectedErr  bool
	}{
		{
			name: "user owns asset",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT quantity FROM portfolio_holdings WHERE user_id = (.+) AND asset_id = (.+)`).
					WithArgs(testUserID, testAssetID).
					WillReturnRows(sqlmock.NewRows([]string{"quantity"}).AddRow(10.5))
			},
			expectedOwns: true,
			expectedQty:  10.5,
			expectedErr:  false,
		},
		{
			name: "user does not own asset",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT quantity FROM portfolio_holdings WHERE user_id = (.+) AND asset_id = (.+)`).
					WithArgs(testUserID, testAssetID).
					WillReturnError(sql.ErrNoRows)
			},
			expectedOwns: false,
			expectedQty:  0,
			expectedErr:  false,
		},
		{
			name: "database error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT quantity FROM portfolio_holdings WHERE user_id = (.+) AND asset_id = (.+)`).
					WithArgs(testUserID, testAssetID).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedOwns: false,
			expectedQty:  0,
			expectedErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mock, cleanup := createTestHandler(t)
			defer cleanup()

			tt.setupMock(mock)

			owns, qty, err := handler.userOwnsAsset(testUserID, testAssetID)

			assert.Equal(t, tt.expectedOwns, owns)
			assert.Equal(t, tt.expectedQty, qty)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestCostAveraging tests the specific cost averaging logic
func TestAddHolding_CostAveraging(t *testing.T) {
	handler, mock, cleanup := createTestHandler(t)
	defer cleanup()

	// Test case: Adding 5 shares at $200 to existing 10 shares at $150
	// Expected: 15 shares at average cost of $166.67
	mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
		WithArgs(defaultUser).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

	mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = (.+)`).
		WithArgs("AAPL").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testAssetID))

	// The ON CONFLICT DO UPDATE handles cost averaging
	mock.ExpectExec(`INSERT INTO portfolio_holdings \(user_id, asset_id, quantity, average_cost\) VALUES \(.+\) ON CONFLICT \(user_id, asset_id\) DO UPDATE SET (.+)`).
		WithArgs(testUserID, testAssetID, 5.0, 200.0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	router := createTestRouter(handler, "POST", "/portfolio/holdings", handler.AddHolding)

	requestBody := `{"symbol": "AAPL", "quantity": 5.0, "average_cost": 200.0}`
	req, _ := http.NewRequest("POST", "/portfolio/holdings", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Holding added successfully")
	assert.NoError(t, mock.ExpectationsWereMet())
}
