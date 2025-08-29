package url

import (
	"context"
	"errors"
	"testing"
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

func TestCreateShortURL(t *testing.T) {
	testCases := []struct {
		name        string
		longURL     string
		setupMock   func(*MockRepository)
		expectedErr error
	}{
		{
			name:    "success",
			longURL: "https://example.com/success",
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
			name:    "database insert fails",
			longURL: "https://example.com/failure",
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
			Service := NewService(MockRepository, 1000)

			err := Service.CreateShortCode(context.Background(), tc.longURL)

			if (err != nil && tc.expectedErr == nil) || (err == nil && tc.expectedErr != nil) || (err != nil && tc.expectedErr != nil && err.Error() != tc.expectedErr.Error()) {
				t.Errorf("unexpected error: got %v want %v", err, tc.expectedErr)
			}

		})
	}
}
