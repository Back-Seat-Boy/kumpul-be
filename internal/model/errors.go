package model

import "errors"

// Common errors
var (
	// Not found errors
	ErrUserNotFound    = errors.New("user not found")
	ErrSessionNotFound = errors.New("session not found")

	// Validation errors
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidGoogleID = errors.New("invalid google id")
	ErrInvalidSession  = errors.New("invalid session")

	// Auth errors
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrInvalidState     = errors.New("invalid state parameter")
	ErrInvalidCode      = errors.New("invalid authorization code")
	ErrSessionExpired   = errors.New("session expired")

	// Database errors
	ErrDuplicateEmail   = errors.New("email already exists")
	ErrDuplicateGoogleID = errors.New("google account already linked")
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError
func NewAppError(err error, statusCode int, message string) *AppError {
	return &AppError{
		Err:        err,
		StatusCode: statusCode,
		Message:    message,
	}
}
