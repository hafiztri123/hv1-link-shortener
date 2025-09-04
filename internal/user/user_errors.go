package user

import "log/slog"

type EmailAlreadyExistsErr struct {
	email string
}

func (e *EmailAlreadyExistsErr)  Error() string {
	slog.Error("Email already exists", e.email)
	return "Email already exists"
}


type UserNotFoundErr struct {
	email string
}

func (e *UserNotFoundErr) Error() string {
	slog.Error("User with selected email not found", "email", e.email)
	return "User not found"
}
