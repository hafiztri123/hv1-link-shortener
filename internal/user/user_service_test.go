package user

import (
	"context"
	"hafiztri123/app-link-shortener/internal/auth"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockRepository struct {
	getByEmailResult *User
	getByEmailErr    error
	insertErr        error
}

func (m *mockRepository) Insert(ctx context.Context, email string, password string) error {
	return m.insertErr
}

func (m *mockRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	return m.getByEmailResult, m.getByEmailErr
}

type mockJWT struct {
	token string
	err   error
}

func (m *mockJWT) GenerateToken(userID int64, email string) (string, error) {
	return m.token, m.err
}

func (m *mockJWT) ValidateToken(tokenString string) (*auth.Claims, error) {
	return nil, nil
}

func TestRegister(t *testing.T) {

	data := &User{
		Id:         1,
		Email:      "example@yahoo.com",
		Password:   "example",
		Created_at: time.Now(),
	}

	testCases := []struct {
		name             string
		getByEmailResult *User
		getByEmailErr    error
		insertErr        error
		wantErr          error
	}{
		{
			name:             "success",
			getByEmailResult: data,
			getByEmailErr:    nil,
			insertErr:        nil,
			wantErr:          nil,
		},
		{
			name:             "get by email error",
			getByEmailResult: data,
			insertErr:        EmailAlreadyExists,
			wantErr:          EmailAlreadyExists,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mockRepository{
				getByEmailResult: tc.getByEmailResult,
				getByEmailErr:    tc.getByEmailErr,
				insertErr:        tc.insertErr,
			}

			srv := NewService(nil, mockRepo, nil)

			err := srv.Register(context.Background(), RegisterRequest{
				Email:    tc.getByEmailResult.Email,
				Password: tc.getByEmailResult.Password,
			})

			assert.ErrorIs(t, tc.wantErr, err)
		})
	}
}

func TestLogin(t *testing.T) {
	password := "admin"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		request LoginRequest

		getByEmailResult User
		getByEmailErr    error
		wantErr          bool
	}{
		{
			name: "success",
			request: LoginRequest{
				Email:    "example@mail.com",
				Password: password,
			},
			getByEmailResult: User{
				Id:         1,
				Email:      "example@mail.com",
				Password:   string(hashedPassword),
				Created_at: time.Now(),
			},
			getByEmailErr: nil,
			wantErr:       false,
		},
		{
			name: "invalid password",
			request: LoginRequest{
				Email:    "example@mail.com",
				Password: "invalid",
			},
			getByEmailResult: User{
				Id:         1,
				Email:      "example@mail.com",
				Password:   string(hashedPassword),
				Created_at: time.Now(),
			},
			getByEmailErr: nil,
			wantErr:       true,
		},
		{
			name: "not using hashed password",
			request: LoginRequest{
				Email:    "example@mail.com",
				Password: password,
			},
			getByEmailResult: User{
				Id:         1,
				Email:      "example@mail.com",
				Password:   password,
				Created_at: time.Now(),
			},
			getByEmailErr: nil,
			wantErr:       true,
		},
		{
			name: "get by email error",
			request: LoginRequest{
				Email:    "example@mail.com",
				Password: password,
			},
			getByEmailResult: User{},
			getByEmailErr:    UserNotFound,
			wantErr:          true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mockRepository{
				getByEmailResult: &tc.getByEmailResult,
				getByEmailErr:    tc.getByEmailErr,
			}

			mockJwt := &mockJWT{
				token: "token",
				err:   nil,
			}

			srv := NewService(nil, mockRepo, mockJwt)

			_, err := srv.Login(context.Background(), tc.request)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}
