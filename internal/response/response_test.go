package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestImage(t *testing.T) {
	rr := httptest.NewRecorder()
	Success(rr, "Success", http.StatusOK, []byte("Hello world"))

	assert.Equal(t, "image/png", rr.Header().Get("Content-Type"))
}

func TestWriteJSONWithoutData(t *testing.T) {
	rr := httptest.NewRecorder()
	Success(rr, "Success", http.StatusOK)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestWriteIMGError(t *testing.T) {
	// Create a ResponseRecorder that will fail on Write
	rr := &FailingResponseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}

	// This should trigger the error path in writeIMG
	Success(rr, "Success", http.StatusOK, []byte("image data"))

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "image/png", rr.Header().Get("Content-Type"))
}

// FailingResponseRecorder simulates a write failure
type FailingResponseRecorder struct {
	*httptest.ResponseRecorder
}

func (f *FailingResponseRecorder) Write(p []byte) (n int, err error) {
	// Return an error to trigger the error path in writeIMG
	return 0, assert.AnError
}

func TestWriteJSONError(t *testing.T) {
	// Test with data that will cause json encoding to fail
	rr := httptest.NewRecorder()

	// Use a channel which can't be marshaled to JSON to trigger the error path
	ch := make(chan int)
	Success(rr, "Success", http.StatusOK, ch)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
}
