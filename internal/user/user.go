package user

import "time"


type User struct {
	id int `json:"id"`
	email string `json:"email"`
	password string `json:"-"`
	created_at time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	email string `json:"email"`
	password string `json:"password"`
}
