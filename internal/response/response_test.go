package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuccess(t *testing.T) {
	rr := httptest.NewRecorder()

	testData := "Hello world"

	Success(rr, "Success", http.StatusOK, testData)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedContentType := "application/json"
	if ctype := rr.Header().Get("Content-Type"); ctype != expectedContentType {
		t.Errorf("handler return wrong header content-type: got %v want %v", ctype, expectedContentType)
	}

	var responseBody map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if status, _ := responseBody["status"].(string); status != "success" {
		t.Errorf("expected status 'success', got: %v", status)
	}

}

func TestError(t *testing.T) {
	rr := httptest.NewRecorder()

	Error(rr, http.StatusInternalServerError, "Error")

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	expectedContentType := "application/json"
	if ctype := rr.Header().Get("Content-Type"); ctype != expectedContentType {
		t.Errorf("handler return wrong header content-type: got %v want %v", ctype, expectedContentType)
	}

	var responseBody map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if status, _ := responseBody["status"].(string); status != "error" {
		t.Errorf("expected status 'error', got: %v", status)
	}
}
