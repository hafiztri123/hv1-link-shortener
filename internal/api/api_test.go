package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

type mockDB struct {
	ShouldFail bool
}

type mockURLService struct {
	createError error
	FetchResult string
	FetchError  error
}

type mockRedis struct {
	ShouldFail bool
}

func (m *mockRedis) Ping(ctx context.Context) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	if m.ShouldFail {
		cmd.SetErr(errors.New("FAIL: unable to ping mock redis"))
	}

	return cmd

}

func (m *mockDB) Ping() error {
	if m.ShouldFail {
		return errors.New("FAIL: unable to ping mock database")
	}
	return nil
}

func (m *mockURLService) CreateShortCode(ctx context.Context, longURL string) error {
	return m.createError
}

func (m *mockURLService) FetchLongURL(ctx context.Context, shortCode string) (string, error) {
	return m.FetchResult, m.FetchError
}

func TestHandleCreateURL_Success(t *testing.T) {

	server := &Server{
		urlService: &mockURLService{
			createError: nil,
		},
	}

	requestBody := []byte(`{"long_url": "https://example.com"}`)
	req, _ := http.NewRequest("POST", "/api/v1/url/shorten", bytes.NewBuffer(requestBody))
	rr := httptest.NewRecorder()

	server.handleCreateURL(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Body.String(), "Success!, Short URL created")
}

func TestHandleCreateURL_ServiceFailure(t *testing.T) {
	server := &Server{
		urlService: &mockURLService{
			createError: errors.New("Failed to create short URL"),
		},
	}

	requestBody := []byte(`{"long_url": "https://example.com"}`)
	req, _ := http.NewRequest("POST", "/api/v1/url/shorten", bytes.NewBuffer(requestBody))
	rr := httptest.NewRecorder()

	server.handleCreateURL(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandleCreateURL_InvalidJSON(t *testing.T) {

	server := &Server{}

	requestBody := []byte(`{"long_url": "https://example.com`)
	req, _ := http.NewRequest("POST", "/api/v1/url/shorten", bytes.NewBuffer(requestBody))
	rr := httptest.NewRecorder()

	server.handleCreateURL(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestFetchURL_Success(t *testing.T) {
	server := &Server{
		urlService: &mockURLService{
			FetchResult: "https://example.com",
			FetchError:  nil,
		},
	}

	reqCtx := chi.NewRouteContext()
	reqCtx.URLParams.Add("shortCode", "example")

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/url/example", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, reqCtx))

	rr := httptest.NewRecorder()

	server.handleFetchURL(rr, req)

	assert.Equal(t, http.StatusMovedPermanently, rr.Code)
	assert.Contains(t, rr.Body.String(), "https://example.com")
}

func TestFetchURL_Failure(t *testing.T) {
	server := &Server{
		urlService: &mockURLService{
			FetchResult: "",
			FetchError:  errors.New("Failed to fetch long URL"),
		},
	}

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/url/example", nil)
	rr := httptest.NewRecorder()

	server.handleFetchURL(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHealthCheck(t *testing.T) {
	testcases := []struct {
		name               string
		dbShouldFail       bool
		redisShouldFail    bool
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "db fails, redis succeeds",
			dbShouldFail:       true,
			redisShouldFail:    false,
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Database not connected",
		},
		{
			name:               "db succeeds, redis fails",
			dbShouldFail:       false,
			redisShouldFail:    true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Redis not connected",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			server := &Server{
				db: &mockDB{
					ShouldFail: tc.dbShouldFail,
				},
				redis: &mockRedis{
					ShouldFail: tc.redisShouldFail,
				},
			}

			req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
			rr := httptest.NewRecorder()

			server.healthCheckHandler(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedBody)
		})
	}

}
