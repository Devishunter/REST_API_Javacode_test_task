package main

import (
	"log"
	"net/http"
	"os"

	"github.com/REST_API_JAVACODE_TEST_TASK/internal"
	"github.com/gorilla/mux"
)

func main() {
	// Инициализация базы данных
	initDB()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallet", internal.HandleWalletOperation).Methods("POST")
	router.HandleFunc("/api/v1/wallets/{walletId}", internal.GetWalletBalance).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
