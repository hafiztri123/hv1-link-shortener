package user

import "log/slog"

var EmailAlreadyExists = &EmailAlreadyExistsErr{}
var UserNotFound = &UserNotFoundErr{}
var InvalidCredentials = &InvalidCredentialErr{}
var UnexpectedError = &UnexpectedErr{}

type EmailAlreadyExistsErr struct {
	email string
}

func (e *EmailAlreadyExistsErr) Error() string {
	slog.Error("Email already exists", "email", e.email)
	return "Email already exists"
}

type UserNotFoundErr struct {
	email string
}

func (e *UserNotFoundErr) Error() string {
	slog.Error("User with selected email not found", "email", e.email)
	return "User not found"
}

type InvalidCredentialErr struct{}

func (e *InvalidCredentialErr) Error() string {
	slog.Error("Password not matched with existing user")
	return "Invalid credentials"
}

type UnexpectedErr struct {
	action string
	err    error
}

func (e *UnexpectedErr) Error() string {
	slog.Error("unexpected error has occured", "action", e.action, "error", e.err)
	return "Unexpected error has occured"
}
