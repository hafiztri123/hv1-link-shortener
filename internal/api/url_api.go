package api

import (
	"encoding/json"
	"hafiztri123/app-link-shortener/internal/response"
	"hafiztri123/app-link-shortener/internal/url"
	"net/http"
)



func (s *Server) handleCreateURL (w http.ResponseWriter, r *http.Request) {
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
	


}

