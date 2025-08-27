package api

import (
	"encoding/json"
	"hafiztri123/app-link-shortener/internal/response"
	"hafiztri123/app-link-shortener/internal/url"
	"net/http"
)

func (s *Server) handleCreateURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req url.CreateURLRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.LongURL == "" {
		response.Error(w, http.StatusBadRequest, "long_url is a required field")
		return
	}

	url, err := s.urlService.CreateShortURL(r.Context(), req.LongURL)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create short URL")
		return
	}

	response.Success(w, http.StatusCreated, url)
}
