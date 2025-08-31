package url

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type Service struct {
	repo     URLRepository
	redis    *redis.Client
	idOffset uint64
}

func NewService(repo URLRepository, redis *redis.Client, idOffset uint64) *Service {
	return &Service{repo: repo, redis: redis, idOffset: idOffset}
}

func (s *Service) CreateShortCode(ctx context.Context, longURL string) error {
	id, err := s.repo.Insert(ctx, longURL)

	if err != nil {
		log.Printf("Failed to insert URL: %v", err)
		return err
	}

	shortcode := toBase62(uint64(id) + s.idOffset)

	err = s.repo.UpdateShortCode(ctx, id, shortcode)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) FetchLongURL(ctx context.Context, shortCode string) (string, error) {

	cachedUrl, err := s.redis.Get(ctx, "url:"+shortCode).Result()
	if err == nil {
		return cachedUrl, nil
	}

	id := fromBase62(shortCode) - s.idOffset

	url, err := s.repo.GetByID(ctx, int64(id))
	if err != nil {
		log.Printf("Failed to fetch URL: %v", err)
		return "", err
	}

	err = s.redis.Set(ctx, "url:"+shortCode, url.LongURL, 1*time.Hour).Err()
	if err != nil {
		log.Printf("Failed to cache URL: %v", err)
	}

	return url.LongURL, nil
}
