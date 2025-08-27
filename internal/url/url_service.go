package url

import (
	"context"
)
type Service struct {
	repo URLRepository
}

func NewService(repo URLRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateShortURL(ctx context.Context, longURL string) (*URL, error) {
	id, err := s.repo.Insert(ctx, longURL)

	if err != nil {
		return nil, err
	}

	shortUrl := toBase62(uint64(id))

	err = s.repo.UpdateShortCode(ctx, id, shortUrl)
	if err != nil {
		return nil, err
	}

	return &URL{
		ID: id,
		LongURL: longURL,
		ShortURL: "http://localhost:8080/" + shortUrl, // TODO: changed when in production
	}, nil
}
