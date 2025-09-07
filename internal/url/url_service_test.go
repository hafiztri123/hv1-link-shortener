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
	InsertFunc                func(context.Context, string) (int64, error)
	UpdateShortCodeFunc       func(context.Context, int64, string) error
	GetByIDFunc               func(context.Context, int64) (*URL, error)
	FindOrCreateShortCodeFunc func(context.Context, string, uint64, *int64) (string, error)
	GetByUserIDBulkFunc       func(context.Context, int64) ([]*URL, error)
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

func (m *MockRepository) GetByUserIDBulk(ctx context.Context, userId int64) ([]*URL, error) {
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
			got, err := service.CreateShortCode(context.Background(), tc.longUrl, nil)

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
