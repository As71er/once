package apperrors

import "errors"

var (
	ErrInternal     = errors.New("internal error")
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("resource not found")
	ErrCreation     = errors.New("failed to create resource")
	ErrDeletion     = errors.New("failed to delete resource")
)
