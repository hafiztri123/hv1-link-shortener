package auth

import "log/slog"

var ValueNotFound = &ValueNotFoundErr{}

type ValueNotFoundErr struct {
	Action string
}

func (e *ValueNotFoundErr) Error() string {
	slog.Error("Value not found", "error", e.Action)
	return "Unexpected error has occured, please try again"
}

func (e *ValueNotFoundErr) Is(target error) bool {
	_, ok := target.(*ValueNotFoundErr)
	return ok
}

