package url

import (
	"context"
	"log"
)

type Service struct {
	repo     URLRepository
	idOffset uint64
}

func NewService(repo URLRepository, idOffset uint64) *Service {
	return &Service{repo: repo, idOffset: idOffset}
}

func (s *Service) CreateShortCode(ctx context.Context, longURL string) error {
	id, err := s.repo.Insert(ctx, longURL)

	if err != nil {
		log.Println("Failed to insert URL: %v", err)
		return err
	}

	shortcode := toBase62(uint64(id) + s.idOffset)

	err = s.repo.UpdateShortCode(ctx, id, shortcode)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) FetchLongURL(ctx context.Context, shortCode string) (*string, error) {

	id := fromBase62(shortCode) - s.idOffset

	url, err := s.repo.GetByID(ctx, int64(id))
	if err != nil {
		log.Println("Failed to fetch URL: %v", err)
		return nil, err
	}

	return &url.LongURL, nil
}
