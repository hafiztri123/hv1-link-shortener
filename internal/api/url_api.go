package api

import (
	"encoding/json"
	"hafiztri123/app-link-shortener/internal/response"
	"hafiztri123/app-link-shortener/internal/url"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleCreateURL(w http.ResponseWriter, r *http.Request) {

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

	err = s.urlService.CreateShortCode(r.Context(), req.LongURL)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create short URL")
		return
	}

	response.Success(w, "Success!, Short URL created", http.StatusCreated)
}

func (s *Server) handleFetchURL(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	if shortCode == "" {
		response.Error(w, http.StatusBadRequest, "short_url is a required field")
		return
	}

	longURL, err := s.urlService.FetchLongURL(r.Context(), shortCode)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch long URL")
		return
	}

	response.Success(w, "Success!, Long URL fetched", http.StatusOK, longURL)

}
