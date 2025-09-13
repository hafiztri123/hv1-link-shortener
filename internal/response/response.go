package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ListResponse[T any] struct {
	Data  []T
	Count int
}

func Success(w http.ResponseWriter, message string, status int, data ...any) {
	if len(data) > 0 && data[0] != nil {
		if byteSlice, ok := data[0].([]byte); ok {
			writeIMG(w, status, byteSlice)
			return
		}

		response := map[string]any{
			"status":  "success",
			"message": message,
			"data":    data[0],
		}

		writeJSON(w, status, response)
		return
	}

	response := map[string]any{
		"status":  "success",
		"message": message,
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
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		slog.Error("Failed to write json", "error", err)
	}
}

func writeIMG(w http.ResponseWriter, status int, v []byte) {
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(status)
	if _, err := w.Write(v); err != nil {
		slog.Error("Failed to write image", "error", err)
	}
}
