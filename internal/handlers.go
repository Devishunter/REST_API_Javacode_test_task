package internal

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// Wallet представляет структуру кошелька
type Wallet struct {
	ID      uuid.UUID `json:"walletId"`
	Balance int64     `json:"balance"`
}

// Operation представляет структуру операции
type Operation struct {
	WalletID      uuid.UUID `json:"walletId"`
	OperationType string    `json:"operationType"`
	Amount        int64     `json:"amount"`
}

var db *sql.DB // Предполагается, что подключение к базе данных передается или инициализируется

// HandleWalletOperation обрабатывает операции DEPOSIT и WITHDRAW
func HandleWalletOperation(w http.ResponseWriter, r *http.Request) {
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

// GetWalletBalance возвращает баланс кошелька по его ID
func GetWalletBalance(w http.ResponseWriter, r *http.Request, walletId uuid.UUID) {
	var balance int64
	err := db.QueryRow("SELECT balance FROM wallets WHERE id = $1", walletId).Scan(&balance)
	if err == sql.ErrNoRows {
		http.Error(w, "Wallet not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(Wallet{ID: walletId, Balance: balance})
}
