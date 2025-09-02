package api

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
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

func TestHandleCreateURL(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "success",
			input:      `{"long_url": "https://example.com"}`,
			err:        nil,
			wantStatus: http.StatusCreated,
			wantMsg:    "Success!, Short URL created",
		},
		{
			name:       "invalid request payload",
			input:      `{"failed,"}`,
			err:        nil,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Invalid request payload",
		},
		{
			name:       "invalid url",
			input:      `{"long_url": "example.com"}`,
			err:        nil,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Invalid URL",
		},
		{
			name:       "missing long url",
			input:      `{"long_url": ""}`,
			err:        nil,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "long_url is a required field",
		},
		{
			name:       "failed to create short url",
			input:      `{"long_url": "https://example.com"}`,
			err:        errors.New("Failed to create short URL"),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "Failed to create short URL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &Server{
				urlService: &mockURLService{
					createError: tc.err,
				},
			}

			requestBody := []byte(tc.input)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/url/shorten", bytes.NewBuffer(requestBody))
			rr := httptest.NewRecorder()

			server.handleCreateURL(rr, req)

			assert.Equal(t, rr.Code, tc.wantStatus)
			assert.Contains(t, rr.Body.String(), tc.wantMsg)
		})
	}
}

func TestFetchLongURL(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		wantResult  string
		wantStatus  int
		fetchResult string
		fetchError  error
	}{
		{
			name:        "Success",
			input:       "success",
			wantResult:  "https://example.com",
			wantStatus:  http.StatusMovedPermanently,
			fetchResult: "https://example.com",
			fetchError:  nil,
		},
		{
			name:        "query params short code missing",
			input:       "",
			wantResult:  "",
			wantStatus:  http.StatusBadRequest,
			fetchResult: "",
			fetchError:  nil,
		},
		{
			name:        "service returning error (non database error)",
			input:       "https://example.com",
			wantResult:  "",
			wantStatus:  http.StatusInternalServerError,
			fetchResult: "",
			fetchError:  errors.New("Error"),
		},
		{
			name:       "service returning error (database error)",
			input:      "https://example.com",
			wantResult: "",
			wantStatus: http.StatusNotFound,
			fetchError: sql.ErrNoRows,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &Server{
				urlService: &mockURLService{
					FetchResult: tc.fetchResult,
					FetchError:  tc.fetchError,
				},
			}

			reqCtx := chi.NewRouteContext()
			reqCtx.URLParams.Add("shortCode", tc.input)

			req, _ := http.NewRequest(http.MethodGet, "/api/v1/url", nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, reqCtx))

			rr := httptest.NewRecorder()

			server.handleFetchURL(rr, req)

			assert.Equal(t, rr.Code, tc.wantStatus)
			if tc.wantResult != "" {
				assert.Contains(t, rr.Body.String(), tc.wantResult)

			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	testcases := []struct {
		name               string
		setupMockDb        func(*mockDB)
		setupMockRedis     func(redismock.ClientMock)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "db fails, redis succeeds",
			setupMockDb: func(db *mockDB) {
				db.ShouldFail = true
			},
			setupMockRedis: func(redis redismock.ClientMock) {

			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Database not connected",
		},
		{
			name: "db succeeds, redis fails",
			setupMockDb: func(db *mockDB) {
				db.ShouldFail = false

			},
			setupMockRedis: func(redis redismock.ClientMock) {
				redis.ExpectPing().SetErr(errors.New("mock redis error"))

			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Redis not connected",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			db := &mockDB{}
			tc.setupMockDb(db)

			redis, mockRedis := redismock.NewClientMock()
			tc.setupMockRedis(mockRedis)

			server := &Server{
				db:    db,
				redis: redis,
			}
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
			rr := httptest.NewRecorder()

			server.healthCheckHandler(rr, req)

			assert.Equal(t, rr.Code, tc.expectedStatusCode)
			assert.Contains(t, rr.Body.String(), tc.expectedBody)
		})
	}

}
