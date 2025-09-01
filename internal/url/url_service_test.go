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
	InsertFunc          func(ctx context.Context, longURL string) (int64, error)
	UpdateShortCodeFunc func(ctx context.Context, id int64, shortCode string) error
	GetByIDFunc         func(ctx context.Context, id int64) (*URL, error)
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

func TestCreateShortCode(t *testing.T) {
	testCases := []struct {
		name          string
		expectedInput string
		setupMock     func(*MockRepository)
		expectedErr   error
	}{
		{
			name:          "success",
			expectedInput: "https://example.com/success",
			setupMock: func(mock *MockRepository) {
				mock.InsertFunc = func(ctx context.Context, longURL string) (int64, error) {
					return 123, nil
				}

				mock.UpdateShortCodeFunc = func(ctx context.Context, id int64, shortCode string) error {
					return nil
				}
			},
			expectedErr: nil,
		},
		{
			name:          "database insert fails",
			expectedInput: "https://example.com/failure",
			setupMock: func(mock *MockRepository) {
				mock.InsertFunc = func(ctx context.Context, longURL string) (int64, error) {
					return 0, errors.New("database error")
				}
			},
			expectedErr: errors.New("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			MockRepository := &MockRepository{}
			tc.setupMock(MockRepository)
			Service := NewService(MockRepository, nil, 1000)

			err := Service.CreateShortCode(context.Background(), tc.expectedInput)

			if (err != nil && tc.expectedErr == nil) || (err == nil && tc.expectedErr != nil) || (err != nil && tc.expectedErr != nil && err.Error() != tc.expectedErr.Error()) {
				t.Errorf("unexpected error: got %v want %v", err, tc.expectedErr)
			}

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
				redisMock.ExpectDel("lock:g8").SetVal(1)
				redisMock.ExpectSet("url:g8", "https://db.com", 1*time.Hour).SetVal("OK")
				repoMock.GetByIDFunc = func(ctx context.Context, id int64) (*URL, error) {
					return &URL{LongURL: "https://db.com"}, nil
				}
			},
			expectedURL: "https://db.com",
			expectedErr: nil,
		},

		{
			name: "cache miss, lock not acquired, timeout",
			shortCode: "g8",
			setupMock: func(repoMock *MockRepository, redisMock redismock.ClientMock) {
				redisMock.ExpectGet("url:g8").SetErr(redis.Nil)
				redisMock.ExpectSetNX("lock:g8", "1", 10*time.Second).SetVal(false)
				redisMock.ExpectGet("url:g8").SetErr(redis.Nil)
				redisMock.ExpectGet("url:g8").SetErr(redis.Nil)

				repoMock.GetByIDFunc = func(ctx context.Context, id int64) (*URL, error) {
					return &URL{LongURL: "https://db-fallback.com" }, nil
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
