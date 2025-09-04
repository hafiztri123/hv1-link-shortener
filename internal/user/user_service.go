package user

import (
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req RegisterRequest) error
	Login(ctx context.Context, req LoginRequest) error
}

type Service struct {
	db   *sql.DB
	repo UserRepository
}

func NewService(db *sql.DB, repo UserRepository) *Service {
	return &Service{
		db:   db,
		repo: repo,
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

func (s *Service) Login(ctx context.Context, req LoginRequest) error {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return &InvalidCredentialErr{}
		}

		return &UnexpectedErr{action: "verify the hashed password"}
	}

	return nil
}
