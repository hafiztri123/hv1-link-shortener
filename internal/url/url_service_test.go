package url

import (
	"context"
	"testing"
)

type MockRepository struct {
	InsertFunc          func(ctx context.Context, longURL string) (int64, error)
	UpdateShortCodeFunc func(ctx context.Context, id int64, shortURL string) error
}

func (m *MockRepository) Insert(ctx context.Context, longURL string) (int64, error) {
	return m.InsertFunc(ctx, longURL)
}

func (m *MockRepository) UpdateShortCode(ctx context.Context, id int64, shortURL string) error {
	return m.UpdateShortCodeFunc(ctx, id, shortURL)
}

func TestCreateShortURL(t *testing.T) {
	mockRepo := &MockRepository{}

	mockRepo.InsertFunc = func(ctx context.Context, longURL string) (int64, error) {
		return 1000, nil
	}

	mockRepo.UpdateShortCodeFunc = func(ctx context.Context, id int64, shortURL string) error {
		return nil
	}

	service := NewService(mockRepo)

	url, err := service.CreateShortURL(context.Background(), "https://example.com")

	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	expectedShortURL := "http://localhost:8080/g8"
	if url.ShortURL != expectedShortURL {
		t.Errorf("expected short URL %s, got: %s", expectedShortURL, url.ShortURL)
	}
}
