package url

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Service struct {
	repo URLRepository
}

func NewService(repo URLRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateShortURL(ctx context.Context, longURL string) (*URL, error) {
	fmt.Println("longURL: ", longURL)
	id, err := s.repo.Insert(ctx, longURL)

	if err != nil {
		log.Fatalf("Failed to insert URL: %v", err)
		return nil, err
	}

	offset, ok := os.LookupEnv("ID_OFFSET")
	if !ok {
		log.Println("ID_OFFSET Environment variable not set")
	}

	offsetNumber, err := strconv.Atoi(offset)
	if err != nil {
		log.Fatalf("Failed to convert offset to int: %v", err)
	}

	shortUrl := toBase62(uint64(id + int64(offsetNumber)))
	fmt.Println("shortUrl: ", shortUrl)

	err = s.repo.UpdateShortCode(ctx, id, shortUrl)
	if err != nil {
		return nil, err
	}

	return &URL{
		ID:        id,
		LongURL:   longURL,
		ShortURL:  "http://localhost:8080/" + shortUrl, // TODO: changed when in production
		CreatedAt: time.Now(),
	}, nil
}
