package user

import "time"


type User struct {
	Id int `json:"id"`
	Email string `json:"email"`
	Password string `json:"-"`
	Created_at time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type UserLoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}
