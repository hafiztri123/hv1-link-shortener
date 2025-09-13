package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"hafiztri123/app-link-shortener/internal/auth"
	"hafiztri123/app-link-shortener/internal/url"
	"hafiztri123/app-link-shortener/internal/user"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

type mockDB struct {
	ShouldFail bool
}

type mockURLService struct {
	createResult         string
	createError          error
	createBulkResult     []url.CreateShortCodeBulkResult
	createBulkError      error
	FetchResult          string
	FetchError           error
	GenerateQRCodeResult []byte
	GenerateQRCodeError  error
	FetchListResult      any
	FetchListResultError error
}

type mockUserService struct {
	token string
	err   error
}

func (m *mockDB) Ping() error {
	if m.ShouldFail {
		return errors.New("FAIL: unable to ping mock database")
	}
	return nil
}

func (m *mockURLService) CreateShortCode(ctx context.Context, longURL string) (string, error) {
	return m.createResult, m.createError
}

func (m *mockURLService) FetchLongURL(ctx context.Context, shortCode string) (string, error) {
	return m.FetchResult, m.FetchError
}

func (m *mockURLService) GenerateQRCode(url string) ([]byte, error) {
	return m.GenerateQRCodeResult, m.GenerateQRCodeError
}

func (m *mockURLService) FetchUserURLHistory(ctx context.Context, userId int64) ([]*url.URL, error) {
	return m.FetchListResult.([]*url.URL), m.FetchListResultError

}

func (m *mockUserService) Register(ctx context.Context, req user.RegisterRequest) error {
	return m.err
}

func (m *mockUserService) Login(ctx context.Context, req user.LoginRequest) (string, error) {
	return m.token, m.err
}

func (m *mockURLService) CreateShortCode_Bulk(ctx context.Context, longURLs []string) ([]url.CreateShortCodeBulkResult, error) {
	return m.createBulkResult, m.createBulkError
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
			wantStatus: http.StatusOK,
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

func TestRegister(t *testing.T) {
	validRequestBody := `{"email": "example@mail.com", "password": "example"}`
	testCases := []struct {
		name           string
		input          string
		registerErr    error
		wantStatusCode int
	}{
		{
			name:           "success",
			input:          validRequestBody,
			registerErr:    nil,
			wantStatusCode: http.StatusCreated,
		},

		{
			name:           "bad request payload",
			input:          `{"invalid": "invalid}`,
			registerErr:    nil,
			wantStatusCode: http.StatusBadRequest,
		},

		{
			name:           "invalid credentials",
			input:          validRequestBody,
			registerErr:    user.InvalidCredentials,
			wantStatusCode: http.StatusUnauthorized,
		},

		{
			name:           "user not found",
			input:          validRequestBody,
			registerErr:    user.UserNotFound,
			wantStatusCode: http.StatusNotFound,
		},

		{
			name:           "email alredy exists",
			input:          validRequestBody,
			registerErr:    user.EmailAlreadyExists,
			wantStatusCode: http.StatusConflict,
		},

		{
			name:           "unexpected error",
			input:          validRequestBody,
			registerErr:    user.UnexpectedError,
			wantStatusCode: http.StatusInternalServerError,
		},

		{
			name:           "internal server error",
			input:          validRequestBody,
			registerErr:    errors.New("example"),
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &Server{
				userService: &mockUserService{
					err: tc.registerErr,
				},
			}

			requestBody := []byte(tc.input)

			rrl := httptest.NewRequest(http.MethodPost, "/api/v1/user/register", bytes.NewBuffer(requestBody))
			rr := httptest.NewRecorder()
			server.handleRegister(rr, rrl)

			assert.Equal(t, tc.wantStatusCode, rr.Code)

		})
	}
}

func TestLogin(t *testing.T) {
	validRequestBody := `{"email": "example@mail.com", "password": "example"}`
	testCases := []struct {
		name           string
		input          string
		token          string
		registerErr    error
		wantStatusCode int
	}{
		{
			name:           "success",
			input:          validRequestBody,
			token:          "token",
			registerErr:    nil,
			wantStatusCode: http.StatusOK,
		},

		{
			name:           "bad request payload",
			input:          `{"invalid": "invalid}`,
			token:          "",
			registerErr:    nil,
			wantStatusCode: http.StatusBadRequest,
		},

		{
			name:           "invalid credentials",
			input:          validRequestBody,
			registerErr:    user.InvalidCredentials,
			wantStatusCode: http.StatusUnauthorized,
		},

		{
			name:           "unexpected error",
			input:          validRequestBody,
			registerErr:    user.UnexpectedError,
			wantStatusCode: http.StatusInternalServerError,
		},

		{
			name:           "internal server error",
			input:          validRequestBody,
			registerErr:    errors.New("example"),
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &Server{
				userService: &mockUserService{
					token: tc.token,
					err:   tc.registerErr,
				},
			}

			requestBody := []byte(tc.input)

			rrl := httptest.NewRequest(http.MethodPost, "/api/v1/user/login", bytes.NewBuffer(requestBody))
			rr := httptest.NewRecorder()
			server.handleLogin(rr, rrl)

			assert.Equal(t, tc.wantStatusCode, rr.Code)
			if tc.wantStatusCode == http.StatusOK {
				assert.Contains(t, rr.Body.String(), tc.token)
			}

		})
	}
}

func TestHandleFetchUserURLHistory(t *testing.T) {
	testCases := []struct {
		name           string
		fetchResult    []*url.URL
		fetchError     error
		emptyClaims    bool
		wantStatusCode int
	}{
		{
			name: "success",
			fetchResult: []*url.URL{
				{LongURL: "https://example.com/1"},
				{LongURL: "https://example.com/2"},
				{LongURL: "https://example.com/3"},
			},
			fetchError:     nil,
			emptyClaims:    false,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "empty Claims",
			fetchResult:    nil,
			fetchError:     nil,
			emptyClaims:    true,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "service error",
			fetchResult:    nil,
			fetchError:     errors.New("example"),
			emptyClaims:    false,
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mockURLService{
				FetchListResult:      tc.fetchResult,
				FetchListResultError: tc.fetchError,
			}

			server := &Server{
				urlService: mockService,
			}

			rrl := httptest.NewRequest(http.MethodGet, "/api/v1/user/history", nil)

			if !tc.emptyClaims {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "example@mail.com",
				}
				ctx := context.WithValue(rrl.Context(), auth.UserContextKey, claims)
				rrl = rrl.WithContext(ctx)

			}

			rr := httptest.NewRecorder()
			server.handleFetchUserURLHistory(rr, rrl)

			assert.Equal(t, tc.wantStatusCode, rr.Code)
		})
	}
}

func TestHandleGenerateQR(t *testing.T) {
	testCases := []struct {
		name              string
		shortcode         string
		wantHeaders       map[string]string
		setMockUrlService func(*mockURLService)
		wantStatusCode    int
	}{
		{
			name:      "success",
			shortcode: "example",
			setMockUrlService: func(mu *mockURLService) {
				mu.FetchResult = "https://example.com"
				mu.FetchError = nil
				mu.GenerateQRCodeResult = []byte("example")
				mu.GenerateQRCodeError = nil
			},
			wantHeaders: map[string]string{
				"Content-Type": "image/png",
			},
			wantStatusCode: http.StatusOK,
		},

		{
			name:      "missing short code",
			shortcode: "",
			setMockUrlService: func(mu *mockURLService) {
				mu.FetchResult = "https://example.com"
				mu.FetchError = nil
				mu.GenerateQRCodeResult = []byte("example")
				mu.GenerateQRCodeError = nil
			},
			wantStatusCode: http.StatusBadRequest,
		},

		{
			name:      "fetch long url error (no rows)",
			shortcode: "example",
			setMockUrlService: func(mu *mockURLService) {
				mu.FetchResult = ""
				mu.FetchError = sql.ErrNoRows
				mu.GenerateQRCodeResult = []byte("example")
				mu.GenerateQRCodeError = nil
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:      "fetch long url error (non database error)",
			shortcode: "example",
			setMockUrlService: func(mu *mockURLService) {
				mu.FetchResult = ""
				mu.FetchError = errors.New("example")
				mu.GenerateQRCodeResult = []byte("example")
				mu.GenerateQRCodeError = nil
			},
			wantStatusCode: http.StatusInternalServerError,
		},

		{
			name:      "generate qr error",
			shortcode: "example",
			setMockUrlService: func(mu *mockURLService) {
				mu.FetchResult = "https://example.com"
				mu.FetchError = nil
				mu.GenerateQRCodeResult = nil
				mu.GenerateQRCodeError = errors.New("example")
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUrlService := &mockURLService{}
			tc.setMockUrlService(mockUrlService)
			server := NewServer(nil, nil, mockUrlService, nil, nil)

			reqCtx := chi.NewRouteContext()

			if tc.shortcode != "" {
				reqCtx.URLParams.Add("shortCode", tc.shortcode)
			}

			rrl := httptest.NewRequest(http.MethodGet, "/api/v1/url/example/qr", nil)

			rrl = rrl.WithContext(context.WithValue(rrl.Context(), chi.RouteCtxKey, reqCtx))

			rr := httptest.NewRecorder()

			server.handleGenerateQR(rr, rrl)

			assert.Equal(t, tc.wantStatusCode, rr.Code)
			for k, v := range tc.wantHeaders {
				assert.Equal(t, v, rr.Header().Get(k))
			}
		})
	}
}

func TestHandleCreateURL_Bulk(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		mockResult     []url.CreateShortCodeBulkResult
		mockError      error
		wantStatus     int
		wantMsg        string
		wantResultSize int
	}{
		{
			name:  "success",
			input: `{"long_urls": ["https://example.com/1", "https://example.com/2"]}`,
			mockResult: []url.CreateShortCodeBulkResult{
				{LongURL: "https://example.com/1", ShortCode: "abc1"},
				{LongURL: "https://example.com/2", ShortCode: "abc2"},
			},
			mockError:      nil,
			wantStatus:     http.StatusOK,
			wantMsg:        "Success! Bulk short URLs created",
			wantResultSize: 2,
		},
		{
			name:       "invalid request payload",
			input:      `{"failed,"}`,
			mockResult: nil,
			mockError:  nil,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Invalid request payload",
		},
		{
			name:       "empty long_urls array",
			input:      `{"long_urls": []}`,
			mockResult: nil,
			mockError:  nil,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "long_urls array cannot be empty",
		},
		{
			name:       "empty URL in array",
			input:      `{"long_urls": ["https://example.com", ""]}`,
			mockResult: nil,
			mockError:  nil,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "long_urls[1] cannot be empty",
		},
		{
			name:       "invalid URL in array",
			input:      `{"long_urls": ["https://example.com", "invalid-url"]}`,
			mockResult: nil,
			mockError:  nil,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "long_urls[1] is not a valid URL",
		},
		{
			name:       "service error",
			input:      `{"long_urls": ["https://example.com"]}`,
			mockResult: nil,
			mockError:  errors.New("database error"),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "Failed to create short URLs",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockURLService := &mockURLService{
				createBulkResult: tc.mockResult,
				createBulkError:  tc.mockError,
			}

			server := &Server{
				urlService: mockURLService,
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/url/shorten/bulk", bytes.NewBuffer([]byte(tc.input)))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			server.handleCreateURL_Bulk(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)

			if tc.wantMsg != "" {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tc.wantMsg, response["message"])
			}

			if tc.wantResultSize > 0 {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)

				data, ok := response["data"].([]interface{})
				assert.True(t, ok)
				assert.Equal(t, tc.wantResultSize, len(data))
			}
		})
	}
}
