package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"hafiztri123/app-link-shortener/internal/auth"
	"hafiztri123/app-link-shortener/internal/response"
	"hafiztri123/app-link-shortener/internal/url"
	"hafiztri123/app-link-shortener/internal/user"
	"hafiztri123/app-link-shortener/internal/utils"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler interface {
	healthCheckHandler(http.ResponseWriter, *http.Request)
	handleCreateURL(http.ResponseWriter, *http.Request)
	handleFetchURL(http.ResponseWriter, *http.Request)
	handleFetchUserURLHistory(http.ResponseWriter, *http.Request)
}

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	err := s.db.Ping()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err != nil {
		http.Error(w, "Database not connected", http.StatusInternalServerError)
		log.Printf("Database health check failed: %v", err)
		return
	}

	err = s.redis.Ping(ctx).Err()
	if err != nil {
		http.Error(w, "Redis not connected", http.StatusInternalServerError)
		log.Printf("Redis health check failed: %v", err)
		return

	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "DB and Redis is connected")
}

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

	ok := utils.IsValidURL(req.LongURL)
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid URL")
		return
	}

	shortcode, err := s.urlService.CreateShortCode(r.Context(), req.LongURL)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create short URL")
		return
	}

	response.Success(w, "Success!, Short URL created", http.StatusOK, url.CreateURLResponse{ShortCode: shortcode})
}

func (s *Server) handleFetchURL(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	if shortCode == "" {
		response.Error(w, http.StatusBadRequest, "short_url is a required field")
		return
	}

	longURL, err := s.urlService.FetchLongURL(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.Error(w, http.StatusNotFound, "Short URL not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch long URL")
		return
	}

	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req user.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
	}

	err = s.userService.Register(r.Context(), req)
	if err != nil {
		switch err.(type) {
		case *user.InvalidCredentialErr:
			response.Error(w, http.StatusUnauthorized, err.Error())
			return
		case *user.UserNotFoundErr:
			response.Error(w, http.StatusNotFound, err.Error())
			return
		case *user.EmailAlreadyExistsErr:
			response.Error(w, http.StatusConflict, err.Error())
		case *user.UnexpectedErr:
			response.Error(w, http.StatusInternalServerError, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "Something has occured, please try again later")
			return
		}
	}

	response.Success(w, "Account created", http.StatusCreated)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req user.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
	}

	token, err := s.userService.Login(r.Context(), req)
	if err != nil {
		switch err.(type) {
		case *user.InvalidCredentialErr:
			response.Error(w, http.StatusUnauthorized, err.Error())
			return
		case *user.UnexpectedErr:
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		default:
			response.Error(w, http.StatusInternalServerError, "something has occured, please try again later")
			return

		}
	}

	response.Success(w, "Success", http.StatusOK, user.LoginResponse{Token: token} )
}

func (s *Server) handleFetchUserURLHistory(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "not authorized")
		return
	}

	urls, err := s.urlService.FetchUserURLHistory(r.Context(), claims.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Unexpected error has occured, please try again later")
		return
	}

	response.Success(w, "success fetching user url history", http.StatusOK, response.ListResponse[*url.URL]{
		Data:  urls,
		Count: len(urls),
	})
}
