package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// New creates a new AppError
func New(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Predefined errors
var (
	ErrInvalidCredentials = New(http.StatusUnauthorized, "Invalid credentials", nil)
	ErrUserNotFound       = New(http.StatusNotFound, "User not found", nil)
	ErrUserAlreadyExists  = New(http.StatusConflict, "User already exists", nil)
	ErrInvalidToken       = New(http.StatusUnauthorized, "Invalid token", nil)
	ErrForbidden          = New(http.StatusForbidden, "Access forbidden", nil)
	ErrValidation         = New(http.StatusBadRequest, "Validation error", nil)
	ErrInternal           = New(http.StatusInternalServerError, "Internal server error", nil)
)
