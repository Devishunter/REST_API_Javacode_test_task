package pkg

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Devishunter/REST_API_JAVACODE_TEST_TASK/internal"
)

func TestHandleWalletOperation(t *testing.T) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"walletId":      "some-uuid",
		"operationType": "DEPOSIT",
		"amount":        1000,
	})

	req, err := http.NewRequest("POST", "/api/v1/wallet", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(internal.HandleWalletOperation) // Используйте импортированную функцию
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
