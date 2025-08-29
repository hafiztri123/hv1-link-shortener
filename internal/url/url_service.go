package url

import (
	"context"
	"fmt"
	"log"
)

type Service struct {
	repo URLRepository
	idOffset uint64
}

func NewService(repo URLRepository, idOffset uint64) *Service {
	return &Service{repo: repo, idOffset: idOffset}
}

func (s *Service) CreateShortURL(ctx context.Context, longURL string) error {
	id, err := s.repo.Insert(ctx, longURL)

	if err != nil {
		log.Fatalf("Failed to insert URL: %v", err)
		return err
	}

	shortUrl := toBase62(uint64(id) + s.idOffset)
	fmt.Println("shortUrl: ", shortUrl)

	err = s.repo.UpdateShortCode(ctx, id, shortUrl)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) FetchLongURL(ctx context.Context, shortURL string) (*string, error) {

	id := fromBase62(shortURL) - s.idOffset

	url, err := s.repo.GetByID(ctx, int64(id))
	if err != nil {
		log.Fatalf("Failed to get URL by ID: %v", err)
	}

	return &url.LongURL, nil
}
