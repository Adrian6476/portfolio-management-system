package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// TestGetPortfolioSummary tests the GetPortfolioSummary handler
func TestGetPortfolioSummary(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   []string
		dbNil          bool
	}{
		{
			name: "successful portfolio summary",
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Portfolio summary query
				mock.ExpectQuery(`SELECT COUNT\(\*\) as total_holdings, COALESCE\(SUM\(ph\.quantity \* ph\.average_cost\), 0\) as total_cost, COALESCE\(SUM\(ph\.quantity\), 0\) as total_shares FROM portfolio_holdings ph WHERE ph\.user_id = (.+)`).
					WithArgs(testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"total_holdings", "total_cost", "total_shares"}).
						AddRow(3, 15000.0, 50.0))

				// Asset allocation query
				mock.ExpectQuery(`SELECT a\.asset_type, COUNT\(\*\) as count, COALESCE\(SUM\(ph\.quantity \* ph\.average_cost\), 0\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph\.asset_id = a\.id WHERE ph\.user_id = (.+) GROUP BY a\.asset_type ORDER BY total_value DESC`).
					WithArgs(testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"asset_type", "count", "total_value"}).
						AddRow("STOCK", 2, 12000.0).
						AddRow("ETF", 1, 3000.0))

				// Top holdings query
				mock.ExpectQuery(`SELECT a\.symbol, a\.name, ph\.quantity, ph\.average_cost, \(ph\.quantity \* ph\.average_cost\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph\.asset_id = a\.id WHERE ph\.user_id = (.+) ORDER BY \(ph\.quantity \* ph\.average_cost\) DESC LIMIT 5`).
					WithArgs(testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"symbol", "name", "quantity", "average_cost", "total_value"}).
						AddRow("AAPL", "Apple Inc.", 10.0, 800.0, 8000.0).
						AddRow("GOOGL", "Alphabet Inc.", 2.0, 2000.0, 4000.0).
						AddRow("SPY", "SPDR S&P 500 ETF", 10.0, 300.0, 3000.0))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"summary", "asset_allocation", "top_holdings", "total_holdings", "AAPL", "GOOGL", "SPY"},
		},
		{
			name: "empty portfolio summary",
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Portfolio summary query - empty portfolio
				mock.ExpectQuery(`SELECT COUNT\(\*\) as total_holdings, COALESCE\(SUM\(ph\.quantity \* ph\.average_cost\), 0\) as total_cost, COALESCE\(SUM\(ph\.quantity\), 0\) as total_shares FROM portfolio_holdings ph WHERE ph\.user_id = (.+)`).
					WithArgs(testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"total_holdings", "total_cost", "total_shares"}).
						AddRow(0, 0.0, 0.0))

				// Asset allocation query - empty
				mock.ExpectQuery(`SELECT a\.asset_type, COUNT\(\*\) as count, COALESCE\(SUM\(ph\.quantity \* ph\.average_cost\), 0\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph\.asset_id = a\.id WHERE ph\.user_id = (.+) GROUP BY a\.asset_type ORDER BY total_value DESC`).
					WithArgs(testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"asset_type", "count", "total_value"}))

				// Top holdings query - empty
				mock.ExpectQuery(`SELECT a\.symbol, a\.name, ph\.quantity, ph\.average_cost, \(ph\.quantity \* ph\.average_cost\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph\.asset_id = a\.id WHERE ph\.user_id = (.+) ORDER BY \(ph\.quantity \* ph\.average_cost\) DESC LIMIT 5`).
					WithArgs(testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"symbol", "name", "quantity", "average_cost", "total_value"}))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"summary", "asset_allocation", "top_holdings", `"total_holdings":0`},
		},
		{
			name: "user not found error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to get user"},
		},
		{
			name: "portfolio summary query error",
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Portfolio summary query fails
				mock.ExpectQuery(`SELECT COUNT\(\*\) as total_holdings, COALESCE\(SUM\(ph\.quantity \* ph\.average_cost\), 0\) as total_cost, COALESCE\(SUM\(ph\.quantity\), 0\) as total_shares FROM portfolio_holdings ph WHERE ph\.user_id = (.+)`).
					WithArgs(testUserID).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to fetch portfolio summary"},
		},
		{
			name: "asset allocation query error",
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Portfolio summary query
				mock.ExpectQuery(`SELECT COUNT\(\*\) as total_holdings, COALESCE\(SUM\(ph\.quantity \* ph\.average_cost\), 0\) as total_cost, COALESCE\(SUM\(ph\.quantity\), 0\) as total_shares FROM portfolio_holdings ph WHERE ph\.user_id = (.+)`).
					WithArgs(testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"total_holdings", "total_cost", "total_shares"}).
						AddRow(3, 15000.0, 50.0))

				// Asset allocation query fails
				mock.ExpectQuery(`SELECT a\.asset_type, COUNT\(\*\) as count, COALESCE\(SUM\(ph\.quantity \* ph\.average_cost\), 0\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph\.asset_id = a\.id WHERE ph\.user_id = (.+) GROUP BY a\.asset_type ORDER BY total_value DESC`).
					WithArgs(testUserID).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to fetch portfolio summary"},
		},
		{
			name: "top holdings query error",
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Portfolio summary query
				mock.ExpectQuery(`SELECT COUNT\(\*\) as total_holdings, COALESCE\(SUM\(ph\.quantity \* ph\.average_cost\), 0\) as total_cost, COALESCE\(SUM\(ph\.quantity\), 0\) as total_shares FROM portfolio_holdings ph WHERE ph\.user_id = (.+)`).
					WithArgs(testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"total_holdings", "total_cost", "total_shares"}).
						AddRow(3, 15000.0, 50.0))

				// Asset allocation query
				mock.ExpectQuery(`SELECT a\.asset_type, COUNT\(\*\) as count, COALESCE\(SUM\(ph\.quantity \* ph\.average_cost\), 0\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph\.asset_id = a\.id WHERE ph\.user_id = (.+) GROUP BY a\.asset_type ORDER BY total_value DESC`).
					WithArgs(testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"asset_type", "count", "total_value"}).
						AddRow("STOCK", 2, 12000.0))

				// Top holdings query fails
				mock.ExpectQuery(`SELECT a\.symbol, a\.name, ph\.quantity, ph\.average_cost, \(ph\.quantity \* ph\.average_cost\) as total_value FROM portfolio_holdings ph JOIN assets a ON ph\.asset_id = a\.id WHERE ph\.user_id = (.+) ORDER BY \(ph\.quantity \* ph\.average_cost\) DESC LIMIT 5`).
					WithArgs(testUserID).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to fetch portfolio summary"},
		},
		{
			name:           "nil database connection",
			setupMock:      func(mock sqlmock.Sqlmock) {},
			dbNil:          true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to fetch portfolio summary"},
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

			router := createTestRouter(handler, "GET", "/portfolio/summary", handler.GetPortfolioSummary)

			req, _ := http.NewRequest("GET", "/portfolio/summary", nil)
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

// TestGetPortfolioPerformance tests the GetPortfolioPerformance handler
func TestGetPortfolioPerformance(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   []string
		dbNil          bool
	}{
		{
			name: "successful portfolio performance",
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Portfolio holdings query for performance calculation
				mock.ExpectQuery(`SELECT ph\.id, a\.symbol, a\.name, ph\.quantity, ph\.average_cost, ph\.purchase_date, \(ph\.quantity \* ph\.average_cost\) as cost_basis FROM portfolio_holdings ph JOIN assets a ON ph\.asset_id = a\.id WHERE ph\.user_id = (.+) ORDER BY cost_basis DESC`).
					WithArgs(testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "symbol", "name", "quantity", "average_cost", "purchase_date", "cost_basis"}).
						AddRow("1", "AAPL", "Apple Inc.", 10.0, 150.0, "2024-01-01", 1500.0).
						AddRow("2", "GOOGL", "Alphabet Inc.", 5.0, 2800.0, "2024-01-02", 14000.0))

				// Historical snapshots query (mocked to return empty for now)
				mock.ExpectQuery(`SELECT snapshot_date, total_value, total_cost, unrealized_pnl FROM portfolio_snapshots WHERE user_id = (.+) AND snapshot_date >= CURRENT_DATE - INTERVAL (.+) ORDER BY snapshot_date ASC`).
					WithArgs(testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"snapshot_date", "total_value", "total_cost", "unrealized_pnl"}))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"performance", "total_return", "holdings_performance"},
		},
		{
			name: "empty portfolio performance",
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Empty portfolio holdings
				mock.ExpectQuery(`SELECT ph\.id, a\.symbol, a\.name, ph\.quantity, ph\.average_cost, ph\.purchase_date, \(ph\.quantity \* ph\.average_cost\) as cost_basis FROM portfolio_holdings ph JOIN assets a ON ph\.asset_id = a\.id WHERE ph\.user_id = (.+) ORDER BY cost_basis DESC`).
					WithArgs(testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "symbol", "name", "quantity", "average_cost", "purchase_date", "cost_basis"}))

				// Historical snapshots query (empty result)
				mock.ExpectQuery(`SELECT snapshot_date, total_value, total_cost, unrealized_pnl FROM portfolio_snapshots WHERE user_id = (.+) AND snapshot_date >= CURRENT_DATE - INTERVAL (.+) ORDER BY snapshot_date ASC`).
					WithArgs(testUserID).
					WillReturnRows(sqlmock.NewRows([]string{"snapshot_date", "total_value", "total_cost", "unrealized_pnl"}))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"performance", "total_return", "holdings_performance"},
		},
		{
			name: "user not found error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to get user"},
		},
		{
			name: "portfolio holdings query error",
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Portfolio holdings query fails
				mock.ExpectQuery(`SELECT ph\.id, a\.symbol, a\.name, ph\.quantity, ph\.average_cost, ph\.purchase_date, \(ph\.quantity \* ph\.average_cost\) as cost_basis FROM portfolio_holdings ph JOIN assets a ON ph\.asset_id = a\.id WHERE ph\.user_id = (.+) ORDER BY cost_basis DESC`).
					WithArgs(testUserID).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to fetch portfolio performance"},
		},
		{
			name:           "nil database connection",
			setupMock:      func(mock sqlmock.Sqlmock) {},
			dbNil:          true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to fetch portfolio performance"},
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

			router := createTestRouter(handler, "GET", "/portfolio/performance", handler.GetPortfolioPerformance)

			url := "/portfolio/performance"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}
			req, _ := http.NewRequest("GET", url, nil)
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

// TestCreateSampleData tests the CreateSampleData utility function
func TestCreateSampleData(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(sqlmock.Sqlmock)
		expectedErr bool
		dbNil       bool
	}{
		{
			name: "successful sample data creation",
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Begin transaction
				mock.ExpectBegin()

				// Expect multiple INSERT statements for assets
				for i := 0; i < 8; i++ { // 8 sample assets
					mock.ExpectExec(`INSERT INTO assets \(symbol, name, asset_type, exchange, currency, sector\) VALUES \(.+\) ON CONFLICT \(symbol\) DO NOTHING`).
						WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
				}

				// Clear existing holdings first
				mock.ExpectExec(`DELETE FROM portfolio_holdings WHERE user_id = (.+)`).
					WithArgs(testUserID).
					WillReturnResult(sqlmock.NewResult(0, 0))

				// Interleaved asset lookups and holdings insertions (5 holdings: AAPL, GOOGL, MSFT, TSLA, AMZN)
				holdings := []string{"AAPL", "GOOGL", "MSFT", "TSLA", "AMZN"}
				for i, symbol := range holdings {
					// Asset ID lookup for this holding
					mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = (.+)`).
						WithArgs(symbol).
						WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(fmt.Sprintf("asset-%d", i+1)))

					// Holdings insertion for this holding
					mock.ExpectExec(`INSERT INTO portfolio_holdings \(user_id, asset_id, quantity, average_cost\) VALUES \(.+\)`).
						WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
				}

				// Clear existing transactions
				mock.ExpectExec(`DELETE FROM transactions WHERE user_id = (.+)`).
					WithArgs(testUserID).
					WillReturnResult(sqlmock.NewResult(0, 0))

				// Interleaved asset lookups and transactions insertions (7 transactions with some symbols repeating)
				transactions := []string{"AAPL", "AAPL", "GOOGL", "MSFT", "MSFT", "TSLA", "AMZN"}
				for i, symbol := range transactions {
					// Asset ID lookup for this transaction
					mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = (.+)`).
						WithArgs(symbol).
						WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(fmt.Sprintf("asset-%d", (i%5)+1)))

					// Transaction insertion for this transaction
					mock.ExpectExec(`INSERT INTO transactions \(user_id, asset_id, transaction_type, quantity, price, fees, total_amount, notes, transaction_date\) VALUES \(.+\)`).
						WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
				}

				// Clear existing notifications
				mock.ExpectExec(`DELETE FROM notifications WHERE user_id = (.+)`).
					WithArgs(testUserID).
					WillReturnResult(sqlmock.NewResult(0, 0))

				// Notifications insertion
				for i := 0; i < 5; i++ { // 5 sample notifications
					mock.ExpectExec(`INSERT INTO notifications \(user_id, title, message, notification_type, created_at\) VALUES \(.+\)`).
						WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
				}

				// Commit transaction
				mock.ExpectCommit()
			},
			expectedErr: false,
		},
		{
			name: "user not found error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: true,
		},
		{
			name: "transaction begin failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				// User ID lookup
				mock.ExpectQuery(`SELECT id FROM users WHERE username = (.+)`).
					WithArgs(defaultUser).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))

				// Begin transaction fails
				mock.ExpectBegin().WillReturnError(fmt.Errorf("transaction error"))
			},
			expectedErr: true,
		},
		{
			name:        "nil database connection",
			setupMock:   func(mock sqlmock.Sqlmock) {},
			dbNil:       true,
			expectedErr: true,
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

			err := handler.CreateSampleData()

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if !tt.dbNil {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}
