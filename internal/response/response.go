package response

import "net/http"

func Success(w http.ResponseWriter, status int, data any) {
	response := map[string]any {
		"status": "success",
		
	}
}
