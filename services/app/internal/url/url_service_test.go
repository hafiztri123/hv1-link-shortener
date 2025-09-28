package url

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

type MockRepository struct {
	InsertFunc                    func(context.Context, string) (int64, error)
	UpdateShortCodeFunc           func(context.Context, int64, string) error
	GetByIDFunc                   func(context.Context, int64) (*URL, error)
	FindOrCreateShortCodeFunc     func(context.Context, string, uint64, *int64) (string, error)
	GetByUserIDBulkFunc           func(context.Context, int64) ([]*URL, error)
	FindOrCreateShortCodeBulkFunc func(context.Context, []string, uint64, *int64) ([]CreateShortCodeBulkResult, error)
}

func (m *MockRepository) Insert(ctx context.Context, longURL string) (int64, error) {
	return m.InsertFunc(ctx, longURL)
}

func (m *MockRepository) UpdateShortCode(ctx context.Context, id int64, shortCode string) error {
	return m.UpdateShortCodeFunc(ctx, id, shortCode)
}

func (m *MockRepository) GetByID(ctx context.Context, id int64) (*URL, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *MockRepository) FindOrCreateShortCode(ctx context.Context, longURL string, idOffset uint64, userId *int64) (string, error) {
	return m.FindOrCreateShortCodeFunc(ctx, longURL, idOffset, nil)
}

func (m *MockRepository) FindOrCreateShortCode_Bulk(ctx context.Context, longURLs []string, idOffset uint64, userId *int64) ([]CreateShortCodeBulkResult, error) {
	return m.FindOrCreateShortCodeBulkFunc(ctx, longURLs, idOffset, nil)
}

func (m *MockRepository) GetByUserID_Bulk(ctx context.Context, userId int64) ([]*URL, error) {
	return m.GetByUserIDBulkFunc(ctx, userId)
}

func TestCreateShortcode(t *testing.T) {
	testCases := []struct {
		name      string
		longUrl   string
		setupMock func(*MockRepository)
		want      string
		wantErr   error
	}{
		{
			name:    "success",
			longUrl: "https://example.com/success",
			setupMock: func(mock *MockRepository) {
				mock.FindOrCreateShortCodeFunc = func(ctx context.Context, longURL string, idOffset uint64, userId *int64) (string, error) {
					return "success", nil
				}
			},
			want:    "success",
			wantErr: nil,
		},

		{
			name:    "database error",
			longUrl: "https://example.com/failure",
			setupMock: func(mock *MockRepository) {
				mock.FindOrCreateShortCodeFunc = func(ctx context.Context, longURL string, idOffset uint64, userId *int64) (string, error) {
					return "", errors.New("database error")
				}
			},
			want:    "",
			wantErr: errors.New("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			MockRepository := &MockRepository{}
			tc.setupMock(MockRepository)
			service := &Service{repo: MockRepository, idOffset: 1000}
			got, err := service.CreateShortCode(context.Background(), tc.longUrl)

			if tc.wantErr != nil {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}

}

func TestFetchLongURL(t *testing.T) {
	testCases := []struct {
		name        string
		shortCode   string
		setupMock   func(repoMock *MockRepository, redisMock redismock.ClientMock)
		expectedURL string
		expectedErr error
	}{
		{
			name:      "success - cache hit",
			shortCode: "g8",
			setupMock: func(mock *MockRepository, redisMock redismock.ClientMock) {
				redisMock.ExpectGet("url:g8").SetVal("https://cached.com")
			},
			expectedURL: "https://cached.com",
			expectedErr: nil,
		},

		{
			name:      "database cache miss, lock acquired",
			shortCode: "g8",
			setupMock: func(repoMock *MockRepository, redisMock redismock.ClientMock) {
				redisMock.ExpectGet("url:g8").SetErr(redis.Nil)
				redisMock.ExpectSetNX("lock:g8", "1", 10*time.Second).SetVal(true)
				redisMock.ExpectSet("url:g8", "https://db.com", 1*time.Hour).SetVal("OK")
				redisMock.ExpectDel("lock:g8").SetVal(1)
				repoMock.GetByIDFunc = func(ctx context.Context, id int64) (*URL, error) {
					return &URL{LongURL: "https://db.com"}, nil
				}
			},
			expectedURL: "https://db.com",
			expectedErr: nil,
		},

		{
			name:      "cache miss, lock not acquired, timeout",
			shortCode: "g8",
			setupMock: func(repoMock *MockRepository, redisMock redismock.ClientMock) {
				redisMock.ExpectGet("url:g8").SetErr(redis.Nil)
				redisMock.ExpectSetNX("lock:g8", "1", 10*time.Second).SetVal(false)
				redisMock.ExpectGet("url:g8").SetErr(redis.Nil)
				redisMock.ExpectGet("url:g8").SetErr(redis.Nil)

				repoMock.GetByIDFunc = func(ctx context.Context, id int64) (*URL, error) {
					return &URL{LongURL: "https://db-fallback.com"}, nil
				}
			},
			expectedURL: "https://db-fallback.com",
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			redisClient, redisMock := redismock.NewClientMock()
			mockRepository := &MockRepository{}
			tc.setupMock(mockRepository, redisMock)
			service := NewService(mockRepository, redisClient, 0)

			longURL, err := service.FetchLongURL(context.Background(), tc.shortCode)

			assert.Equal(t, tc.expectedURL, longURL)
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFetchUserURLHistory(t *testing.T) {
	testCases := []struct {
		name string
		urls []*URL
		err  error
	}{
		{
			name: "success",
			urls: []*URL{
				{LongURL: "https://example.com/1"},
				{LongURL: "https://example.com/2"},
				{LongURL: "https://example.com/3"},
			},
			err: nil,
		},
		{
			name: "database error",
			urls: nil,
			err:  errors.New("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			MockRepository := &MockRepository{}
			MockRepository.GetByUserIDBulkFunc = func(ctx context.Context, userID int64) ([]*URL, error) {
				return tc.urls, tc.err
			}

			service := NewService(MockRepository, nil, 0)

			urls, err := service.FetchUserURLHistory(context.Background(), 1)

			assert.Equal(t, tc.urls, urls)
			if tc.err != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateQRCode(t *testing.T) {
	testCases := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "success",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "invalid url",
			url:     "",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		service := NewService(nil, nil, 0)
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.GenerateQRCode(tc.url)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestGenerateShortCodeBulk(t *testing.T) {
	testCases := []struct {
		name    string
		input   []string
		result  []CreateShortCodeBulkResult
		err     error
		wantErr bool
	}{
		{
			name: "success",
			input: []string{
				"https://example.com/1",
				"https://example.com/2",
				"https://example.com/3",
			},
			result: []CreateShortCodeBulkResult{
				{ShortCode: "1", LongURL: "https://example.com/1"},
				{ShortCode: "2", LongURL: "https://example.com/2"},
				{ShortCode: "3", LongURL: "https://example.com/3"},
			},
			err:     nil,
			wantErr: false,
		},

		{
			name: "invalid url",
			input: []string{
				"",
				"https://example.com/2",
				"https://example.com/3",
			},
			result:  nil,
			err:     errors.New("invalid url"),
			wantErr: true,
		},
	}

	redis, _ := redismock.NewClientMock()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := &MockRepository{
				FindOrCreateShortCodeBulkFunc: func(ctx context.Context, s []string, u uint64, i *int64) ([]CreateShortCodeBulkResult, error) {
					return tc.result, tc.err
				},
			}

			service := NewService(mockRepository, redis, 0)
			result, err := service.CreateShortCode_Bulk(context.Background(), tc.input)

			assert.Equal(t, tc.result, result)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func TestFetchLongURL_CacheError(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()
	mockRepository := &MockRepository{}

	// Mock redis get to return an error (not redis.Nil)
	redisMock.ExpectGet("url:g8").SetErr(errors.New("redis connection error"))

	// Mock database fallback to also return an error
	mockRepository.GetByIDFunc = func(ctx context.Context, id int64) (*URL, error) {
		return nil, errors.New("database error")
	}

	service := NewService(mockRepository, redisClient, 0)

	_, err := service.FetchLongURL(context.Background(), "g8")
	assert.Error(t, err)
}

func TestFetchLongURL_DatabaseError(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()
	mockRepository := &MockRepository{}

	// Mock redis get to return redis.Nil (cache miss)
	redisMock.ExpectGet("url:g8").SetErr(redis.Nil)

	// Mock database to return an error
	mockRepository.GetByIDFunc = func(ctx context.Context, id int64) (*URL, error) {
		return nil, errors.New("database error")
	}

	service := NewService(mockRepository, redisClient, 0)

	_, err := service.FetchLongURL(context.Background(), "g8")
	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
}
