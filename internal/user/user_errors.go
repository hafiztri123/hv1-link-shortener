package user

import "log/slog"

type EmailAlreadyExistsErr struct {
	email string
}

func (e *EmailAlreadyExistsErr)  Error() string {
	slog.Error("Email already exists", e.email)
	return "Email already exists"
}
