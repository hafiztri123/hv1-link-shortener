package user

import (
	"context"
	"database/sql"
	"errors"
	"hafiztri123/app-link-shortener/internal/auth"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req RegisterRequest) error
	Login(ctx context.Context, req LoginRequest) (string, error)
}

type Service struct {
	db   *sql.DB
	repo UserRepository
	jwt auth.JWT
}

func NewService(db *sql.DB, repo UserRepository, jwt auth.JWT) *Service {
	return &Service{
		db:   db,
		repo: repo,
		jwt: jwt,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return &UnexpectedErr{action: "hashing password", err: err}
	}

	err = s.repo.Insert(ctx, req.Email, string(hashedPassword))
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (string, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", &InvalidCredentialErr{}
		}

		return "", &UnexpectedErr{action: "verify the hashed password"}
	}

	token, err := s.jwt.GenerateToken(int64(user.Id), user.Email)

	if err != nil {
		return "", err
	}

	return token, nil
}
