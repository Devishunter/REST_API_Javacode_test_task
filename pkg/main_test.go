package pkg

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleWalletOperation(t *testing.T) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"walletId":      "some-valid-uuid",
		"operationType": "DEPOSIT",
		"amount":        1000,
	})

	req, err := http.NewRequest("POST", "/api/v1/wallet", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleWalletOperation) // Используйте импортированную функцию
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Logf("Response body: %s", rr.Body.String())
	}
}

func HandleWalletOperation(w http.ResponseWriter, r *http.Request) {
	var db *sql.DB
	var op Operation
	if err := json.NewDecoder(r.Body).Decode(&op); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	var currentBalance int64
	err := db.QueryRow("SELECT balance FROM wallets WHERE id = $1", op.WalletID).Scan(&currentBalance)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	switch op.OperationType {
	case "DEPOSIT":
		currentBalance += op.Amount
	case "WITHDRAW":
		if currentBalance < op.Amount {
			http.Error(w, "Insufficient funds", http.StatusBadRequest)
			return
		}
		currentBalance -= op.Amount
	default:
		http.Error(w, "Invalid operation type", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO wallets (id, balance) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET balance = $2", op.WalletID, currentBalance)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
