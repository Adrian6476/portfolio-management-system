package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetTransactions(t *testing.T) {
	t.Run("Success - with pagination", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		// Setup mock expectations
		rows := sqlmock.NewRows([]string{"id", "user_id", "asset_id", "type", "quantity", "price", "timestamp"}).
			AddRow(1, 1, 1, "BUY", 10, 100.50, "2025-01-01T00:00:00Z").
			AddRow(2, 1, 2, "SELL", 5, 200.75, "2025-01-02T00:00:00Z")

		mock.ExpectQuery(`SELECT .* FROM transactions WHERE user_id = \$1 ORDER BY timestamp DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(1, 10, 0).
			WillReturnRows(rows)

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/transactions?limit=10&offset=0", nil)
		c.Set("userID", int64(1))

		// Call handler
		GetTransactions(db)(c)

		// Verify
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - invalid limit parameter", func(t *testing.T) {
		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/transactions?limit=invalid", nil)
		c.Set("userID", int64(1))

		GetTransactions(db)(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Error - database error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(`SELECT .* FROM transactions`).
			WillReturnError(sql.ErrConnDone)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/transactions", nil)
		c.Set("userID", int64(1))

		GetTransactions(db)(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - filter by transaction type", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "user_id", "asset_id", "type", "quantity", "price", "timestamp"}).
			AddRow(1, 1, 1, "BUY", 10, 100.50, "2025-01-01T00:00:00Z")

		mock.ExpectQuery(`SELECT .* FROM transactions WHERE user_id = \$1 AND type = \$2`).
			WithArgs(1, "BUY").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/transactions?type=BUY", nil)
		c.Set("userID", int64(1))

		GetTransactions(db)(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - filter by asset symbol", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		// First mock the asset lookup
		mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = \$1`).
			WithArgs("AAPL").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Then mock the transaction query
		rows := sqlmock.NewRows([]string{"id", "user_id", "asset_id", "type", "quantity", "price", "timestamp"}).
			AddRow(1, 1, 1, "BUY", 10, 100.50, "2025-01-01T00:00:00Z")

		mock.ExpectQuery(`SELECT .* FROM transactions WHERE user_id = \$1 AND asset_id = \$2`).
			WithArgs(1, 1).
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/transactions?symbol=AAPL", nil)
		c.Set("userID", int64(1))

		GetTransactions(db)(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - invalid transaction type", func(t *testing.T) {
		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/transactions?type=INVALID", nil)
		c.Set("userID", int64(1))

		GetTransactions(db)(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Error - asset not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = \$1`).
			WithArgs("UNKNOWN").
			WillReturnError(sql.ErrNoRows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/transactions?symbol=UNKNOWN", nil)
		c.Set("userID", int64(1))

		GetTransactions(db)(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCreateTransaction(t *testing.T) {
	t.Run("Success - BUY transaction", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		// Mock asset lookup
		mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = \$1`).
			WithArgs("AAPL").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Mock balance check
		mock.ExpectQuery(`SELECT balance FROM accounts WHERE user_id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(10000.00))

		// Mock transaction begin
		mock.ExpectBegin()
		// Mock transaction insert
		mock.ExpectExec(`INSERT INTO transactions`).
			WithArgs(1, 1, "BUY", 10, 150.00, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		// Mock balance update
		mock.ExpectExec(`UPDATE accounts SET balance = balance - \$1 WHERE user_id = \$2`).
			WithArgs(1500.00, 1).
			WillReturnResult(sqlmock.NewResult(0, 1))
		// Mock transaction commit
		mock.ExpectCommit()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/transactions", strings.NewReader(`{
			"symbol": "AAPL",
			"type": "BUY",
			"quantity": 10,
			"price": 150.00
		}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", int64(1))

		CreateTransaction(db)(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - insufficient funds", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		// Mock asset lookup
		mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = \$1`).
			WithArgs("AAPL").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Mock balance check
		mock.ExpectQuery(`SELECT balance FROM accounts WHERE user_id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.00))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/transactions", strings.NewReader(`{
			"symbol": "AAPL",
			"type": "BUY",
			"quantity": 10,
			"price": 150.00
		}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", int64(1))

		CreateTransaction(db)(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "insufficient funds")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - invalid transaction type", func(t *testing.T) {
		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/transactions", strings.NewReader(`{
			"symbol": "AAPL",
			"type": "INVALID",
			"quantity": 10,
			"price": 150.00
		}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", int64(1))

		CreateTransaction(db)(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Error - negative quantity", func(t *testing.T) {
		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/transactions", strings.NewReader(`{
			"symbol": "AAPL",
			"type": "BUY",
			"quantity": -10,
			"price": 150.00
		}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", int64(1))

		CreateTransaction(db)(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Success - SELL transaction", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		// Mock asset lookup
		mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = \$1`).
			WithArgs("AAPL").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Mock position check
		mock.ExpectQuery(`SELECT quantity FROM positions WHERE user_id = \$1 AND asset_id = \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"quantity"}).AddRow(20))

		// Mock transaction begin
		mock.ExpectBegin()
		// Mock transaction insert
		mock.ExpectExec(`INSERT INTO transactions`).
			WithArgs(1, 1, "SELL", 10, 150.00, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		// Mock balance update
		mock.ExpectExec(`UPDATE accounts SET balance = balance + \$1 WHERE user_id = \$2`).
			WithArgs(1500.00, 1).
			WillReturnResult(sqlmock.NewResult(0, 1))
		// Mock position update
		mock.ExpectExec(`UPDATE positions SET quantity = quantity - \$1 WHERE user_id = \$2 AND asset_id = \$3`).
			WithArgs(10, 1, 1).
			WillReturnResult(sqlmock.NewResult(0, 1))
		// Mock transaction commit
		mock.ExpectCommit()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/transactions", strings.NewReader(`{
			"symbol": "AAPL",
			"type": "SELL",
			"quantity": 10,
			"price": 150.00
		}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", int64(1))

		CreateTransaction(db)(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - insufficient position quantity", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		// Mock asset lookup
		mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = \$1`).
			WithArgs("AAPL").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Mock position check
		mock.ExpectQuery(`SELECT quantity FROM positions WHERE user_id = \$1 AND asset_id = \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"quantity"}).AddRow(5))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/transactions", strings.NewReader(`{
			"symbol": "AAPL",
			"type": "SELL",
			"quantity": 10,
			"price": 150.00
		}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", int64(1))

		CreateTransaction(db)(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "insufficient quantity")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetTransaction(t *testing.T) {
	t.Run("Success - retrieve transaction", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "user_id", "asset_id", "type", "quantity", "price", "timestamp"}).
			AddRow(1, 1, 1, "BUY", 10, 100.50, "2025-01-01T00:00:00Z")

		mock.ExpectQuery(`SELECT .* FROM transactions WHERE id = \$1 AND user_id = \$2`).
			WithArgs(1, 1).
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/transactions/1", nil)
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		GetTransaction(db)(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - transaction not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(`SELECT .* FROM transactions WHERE id = \$1 AND user_id = \$2`).
			WithArgs(999, 1).
			WillReturnError(sql.ErrNoRows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/transactions/999", nil)
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "999"}}

		GetTransaction(db)(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - unauthorized access", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		// Transaction exists but belongs to different user
		rows := sqlmock.NewRows([]string{"id", "user_id"}).
			AddRow(1, 2)

		mock.ExpectQuery(`SELECT .* FROM transactions WHERE id = \$1 AND user_id = \$2`).
			WithArgs(1, 1).
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/transactions/1", nil)
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		GetTransaction(db)(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - invalid transaction ID", func(t *testing.T) {
		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/transactions/invalid", nil)
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "invalid"}}

		GetTransaction(db)(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateTransaction(t *testing.T) {
	t.Run("Success - partial update", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		// Mock transaction retrieval
		mock.ExpectQuery(`SELECT .* FROM transactions WHERE id = \$1 AND user_id = \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "asset_id", "type", "quantity", "price"}).
				AddRow(1, 1, 1, "BUY", 10, 100.00))

		// Mock update
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE transactions SET quantity = \$1, price = \$2 WHERE id = \$3`).
			WithArgs(15, 100.00, 1).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PATCH", "/transactions/1", strings.NewReader(`{
			"quantity": 15
		}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		UpdateTransaction(db)(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - transaction not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(`SELECT .* FROM transactions WHERE id = \$1 AND user_id = \$2`).
			WithArgs(999, 1).
			WillReturnError(sql.ErrNoRows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PATCH", "/transactions/999", strings.NewReader(`{
			"quantity": 15
		}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "999"}}

		UpdateTransaction(db)(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - invalid quantity", func(t *testing.T) {
		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PATCH", "/transactions/1", strings.NewReader(`{
			"quantity": -5
		}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		UpdateTransaction(db)(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Success - recalculate balance on price change", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		// Mock transaction retrieval
		mock.ExpectQuery(`SELECT .* FROM transactions WHERE id = \$1 AND user_id = \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "asset_id", "type", "quantity", "price"}).
				AddRow(1, 1, 1, "BUY", 10, 100.00))

		// Mock balance adjustment
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE accounts SET balance = balance \+ \$1 WHERE user_id = \$2`).
			WithArgs(100.00, 1). // 10 * (100 - 90)
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`UPDATE transactions SET price = \$1 WHERE id = \$2`).
			WithArgs(90.00, 1).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PATCH", "/transactions/1", strings.NewReader(`{
			"price": 90.00
		}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		UpdateTransaction(db)(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - unauthorized access", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		// Transaction exists but belongs to different user
		mock.ExpectQuery(`SELECT .* FROM transactions WHERE id = \$1 AND user_id = \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).
				AddRow(1, 2))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PATCH", "/transactions/1", strings.NewReader(`{
			"quantity": 15
		}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		UpdateTransaction(db)(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteTransaction(t *testing.T) {
	t.Run("Success - delete transaction", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		// Mock transaction retrieval
		mock.ExpectQuery(`SELECT .* FROM transactions WHERE id = \$1 AND user_id = \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "asset_id", "type", "quantity", "price"}).
				AddRow(1, 1, 1, "BUY", 10, 100.00))

		// Mock deletion
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM transactions WHERE id = \$1`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))
		// Mock balance adjustment for BUY transaction
		mock.ExpectExec(`UPDATE accounts SET balance = balance \+ \$1 WHERE user_id = \$2`).
			WithArgs(1000.00, 1). // 10 * 100.00
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/transactions/1", nil)
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		DeleteTransaction(db)(c)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - transaction not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(`SELECT .* FROM transactions WHERE id = \$1 AND user_id = \$2`).
			WithArgs(999, 1).
			WillReturnError(sql.ErrNoRows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/transactions/999", nil)
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "999"}}

		DeleteTransaction(db)(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - database failure", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		// Mock transaction retrieval
		mock.ExpectQuery(`SELECT .* FROM transactions WHERE id = \$1 AND user_id = \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "asset_id", "type", "quantity", "price"}).
				AddRow(1, 1, 1, "BUY", 10, 100.00))

		// Mock failed deletion
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM transactions WHERE id = \$1`).
			WithArgs(1).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/transactions/1", nil)
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		DeleteTransaction(db)(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetNotifications(t *testing.T) {
	t.Run("Success - with pagination", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "user_id", "type", "message", "is_read", "created_at"}).
			AddRow(1, 1, "TRADE", "Your BUY order for AAPL was executed", false, "2025-01-01T00:00:00Z").
			AddRow(2, 1, "SYSTEM", "System maintenance scheduled", false, "2025-01-02T00:00:00Z")

		mock.ExpectQuery(`SELECT .* FROM notifications WHERE user_id = \$1 ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(1, 10, 0).
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/notifications?limit=10&offset=0", nil)
		c.Set("userID", int64(1))

		GetNotifications(db)(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - filter by type", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "user_id", "type", "message", "is_read", "created_at"}).
			AddRow(1, 1, "TRADE", "Your BUY order for AAPL was executed", false, "2025-01-01T00:00:00Z")

		mock.ExpectQuery(`SELECT .* FROM notifications WHERE user_id = \$1 AND type = \$2`).
			WithArgs(1, "TRADE").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/notifications?type=TRADE", nil)
		c.Set("userID", int64(1))

		GetNotifications(db)(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - invalid limit parameter", func(t *testing.T) {
		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/notifications?limit=invalid", nil)
		c.Set("userID", int64(1))

		GetNotifications(db)(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Error - database error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(`SELECT .* FROM notifications`).
			WillReturnError(sql.ErrConnDone)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/notifications", nil)
		c.Set("userID", int64(1))

		GetNotifications(db)(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMarkNotificationRead(t *testing.T) {
	t.Run("Success - mark notification as read", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		// Mock notification ownership check
		mock.ExpectQuery(`SELECT user_id FROM notifications WHERE id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

		// Mock update
		mock.ExpectExec(`UPDATE notifications SET is_read = true WHERE id = \$1`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PATCH", "/notifications/1/read", nil)
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		MarkNotificationRead(db)(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - notification not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(`SELECT user_id FROM notifications WHERE id = \$1`).
			WithArgs(999).
			WillReturnError(sql.ErrNoRows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PATCH", "/notifications/999/read", nil)
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "999"}}

		MarkNotificationRead(db)(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - unauthorized access", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		// Notification belongs to different user
		mock.ExpectQuery(`SELECT user_id FROM notifications WHERE id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(2))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PATCH", "/notifications/1/read", nil)
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		MarkNotificationRead(db)(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - database failure", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(`SELECT user_id FROM notifications WHERE id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

		mock.ExpectExec(`UPDATE notifications SET is_read = true WHERE id = \$1`).
			WithArgs(1).
			WillReturnError(sql.ErrConnDone)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PATCH", "/notifications/1/read", nil)
		c.Set("userID", int64(1))
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		MarkNotificationRead(db)(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetAssetIDBySymbol(t *testing.T) {
	t.Run("Success - get asset ID", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = \$1`).
			WithArgs("AAPL").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		assetID, err := getAssetIDBySymbol(db, "AAPL")

		assert.Equal(t, int64(1), assetID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - asset not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = \$1`).
			WithArgs("UNKNOWN").
			WillReturnError(sql.ErrNoRows)

		assetID, err := getAssetIDBySymbol(db, "UNKNOWN")

		assert.Equal(t, int64(0), assetID)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - database failure", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock DB: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(`SELECT id FROM assets WHERE symbol = \$1`).
			WithArgs("AAPL").
			WillReturnError(sql.ErrConnDone)

		assetID, err := getAssetIDBySymbol(db, "AAPL")

		assert.Equal(t, int64(0), assetID)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetUserID(t *testing.T) {
	t.Run("Success - get user ID from context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("userID", int64(1))

		userID, err := getUserID(c)

		assert.Equal(t, int64(1), userID)
		assert.NoError(t, err)
	})

	t.Run("Error - user ID not set", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		userID, err := getUserID(c)

		assert.Equal(t, int64(0), userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID not found")
	})

	t.Run("Error - invalid user ID type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("userID", "invalid")

		userID, err := getUserID(c)

		assert.Equal(t, int64(0), userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user ID type")
	})
}