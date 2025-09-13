package url

import (
	"context"
	"hafiztri123/app-link-shortener/internal/auth"
	"log/slog"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/skip2/go-qrcode"
)

type CreateShortCodeBulkResult struct {
	LongURL   string `json:"long_url"`
	ShortCode string `json:"short_code"`
}

type URLService interface {
	CreateShortCode(context.Context, string) (string, error)
	CreateShortCode_Bulk(context.Context, []string) ([]CreateShortCodeBulkResult, error)
	FetchLongURL(context.Context, string) (string, error)
	FetchUserURLHistory(context.Context, int64) ([]*URL, error)
	GenerateQRCode(string) ([]byte, error)
}

type Service struct {
	repo     URLRepository
	redis    *redis.Client
	idOffset uint64
}

func NewService(repo URLRepository, redis *redis.Client, idOffset uint64) *Service {
	return &Service{repo: repo, redis: redis, idOffset: idOffset}
}

func (s *Service) CreateShortCode(ctx context.Context, longURL string) (string, error) {
	user, _ := auth.GetUserFromContext(ctx)

	var userID *int64
	if user != nil {
		userID = &user.UserID
	}

	shortCode, err := s.repo.FindOrCreateShortCode(ctx, longURL, s.idOffset, userID)
	if err != nil {
		slog.Error("Failed to find or create short code", "error", err, "url", longURL)
		return "", err
	}

	return shortCode, nil
}

func (s *Service) FetchLongURL(ctx context.Context, shortCode string) (string, error) {
	id := FromBase62(shortCode) - s.idOffset

	cacheKey := "url:" + shortCode
	lockKey := "lock:" + shortCode

	cachedUrl, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		return cachedUrl, nil
	}

	if err != redis.Nil {
		slog.Warn("Cache is missing", "error", err, "key", cacheKey)
	}

	lockAcquired, err := s.redis.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
	if err != nil {
		slog.Warn("Redis SetNX for lock failed", "error", err, "key", lockKey)
		return s.getIDFromDatabase(ctx, int64(id))
	}

	if lockAcquired {
		defer s.redis.Del(ctx, lockKey)
		longUrl, err := s.getIDFromDatabase(ctx, int64(id))
		if err != nil {
			slog.Error("Database failed", "error", err, "id", id)
			return "", err
		}

		err = s.redis.Set(ctx, cacheKey, longUrl, 1*time.Hour).Err()

		if err != nil {
			slog.Warn("Redis failed to cache", "error", err, "cacheKey", cacheKey)
		}

		return longUrl, nil
	}

	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return s.getIDFromDatabase(ctx, int64(id))
		case <-ticker.C:
			cachedUrl, err := s.redis.Get(ctx, cacheKey).Result()
			if err == nil {
				return cachedUrl, nil
			}
		}
	}

}

func (s *Service) FetchUserURLHistory(ctx context.Context, userId int64) ([]*URL, error) {

	urls, err := s.repo.GetByUserID_Bulk(ctx, userId)

	if err != nil {
		return nil, err
	}

	return urls, nil

}

func (s *Service) GenerateQRCode(url string) ([]byte, error) {

	qrBytes, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		slog.Error("Failed to generate qr code", "error", err)
		return nil, err
	}

	return qrBytes, nil
}

func (s *Service) CreateShortCode_Bulk(ctx context.Context, longUrl []string) ([]CreateShortCodeBulkResult, error) {
	user, _ := auth.GetUserFromContext(ctx)

	var userID *int64
	if user != nil {
		userID = &user.UserID
	}

	results, err := s.repo.FindOrCreateShortCode_Bulk(ctx, longUrl, s.idOffset, userID)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *Service) getIDFromDatabase(ctx context.Context, id int64) (string, error) {
	url, err := s.repo.GetByID(ctx, id)

	if err != nil {
		return "", err
	}

	return url.LongURL, nil

}
