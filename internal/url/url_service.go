package url

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
)

type Service struct {
	repo URLRepository
}

func NewService(repo URLRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateShortURL(ctx context.Context, longURL string) error {
	id, err := s.repo.Insert(ctx, longURL)

	if err != nil {
		log.Fatalf("Failed to insert URL: %v", err)
		return err
	}

	offset, ok := os.LookupEnv("ID_OFFSET")
	if !ok {
		log.Fatal("ID_OFFSET Environment variable not set")
	}

	offsetNumber, err := strconv.Atoi(offset)
	if err != nil {
		log.Fatalf("Failed to convert offset to int: %v", err)
	}

	shortUrl := toBase62(uint64(id + int64(offsetNumber)))
	fmt.Println("shortUrl: ", shortUrl)

	err = s.repo.UpdateShortCode(ctx, id, shortUrl)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) FetchLongURL(ctx context.Context, shortURL string) (*string, error) {
	offset, ok := os.LookupEnv("ID_OFFSET")
	if !ok {
		log.Fatal("ID_OFFSET Environment variable not set")
	}

	offsetNumber, err := strconv.Atoi(offset)
	if err != nil {
		log.Fatalf("Failed to convert offset to int: %v", err)
	}

	id := fromBase62(shortURL) - uint64(offsetNumber)

	url, err := s.repo.GetByID(ctx, int64(id))
	if err != nil {
		log.Fatalf("Failed to get URL by ID: %v", err)
	}

	return &url.LongURL, nil
}
