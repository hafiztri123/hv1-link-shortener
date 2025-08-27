package response

import (
	"encoding/json"
	"net/http"
)

func Success(w http.ResponseWriter, status int, data any) {
	response := map[string]any{
		"status": "success",
		"data":   data,
	}

	writeJSON(w, status, response)
}

func Error(w http.ResponseWriter, status int, message string) {
	response := map[string]any{
		"status":  "error",
		"message": message,
	}

	writeJSON(w, status, response)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
