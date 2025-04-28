package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Wallet struct {
	ID      uuid.UUID `json:"walletId"`
	Balance int64     `json:"balance"`
}

type Operation struct {
	WalletID      uuid.UUID `json:"walletId"`
	OperationType string    `json:"operationType"`
	Amount        int64     `json:"amount"`
}

var db *sql.DB

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	// Инициализация базы данных
	initDB()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallet", HandleWalletOperation).Methods("POST")
	router.HandleFunc("/api/v1/wallets/{walletId}", GetWalletBalance).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func initDB() {
	var err error
	db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME")))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Error connecting to the database: %q", err)
	}

	log.Println("Successfully connected to the database")
}

func GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletId, err := uuid.Parse(vars["walletId"])
	if err != nil {
		http.Error(w, "Invalid wallet ID", http.StatusBadRequest)
		return
	}

	var balance int64
	err = db.QueryRow("SELECT balance FROM wallets WHERE id = $1", walletId).Scan(&balance)
	if err == sql.ErrNoRows {
		http.Error(w, "Wallet not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(Wallet{ID: walletId, Balance: balance})
}

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
