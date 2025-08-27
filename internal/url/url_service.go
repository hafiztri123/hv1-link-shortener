package url

import (
	"context"
	"database/sql"
)
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateShortURL(ctx context.Context, longURL string) (*URL, error) {
	id, err := s.repo.Insert(ctx, longURL)

	if err != nil {
		return nil, err
	}

	
}
