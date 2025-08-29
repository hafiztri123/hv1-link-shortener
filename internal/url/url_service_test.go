package url

import (
	"context"
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
