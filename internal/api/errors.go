package api

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorized   = errors.New("unauthorized: run 'gf auth login' to authenticate")
	ErrForbidden      = errors.New("forbidden: you don't have permission to access this resource")
	ErrTokenInvalid   = errors.New("token expired or invalid: run 'gf auth login' to re-authenticate")
	ErrNotFound       = errors.New("not found")
	ErrNetwork        = errors.New("network error: check your connection")
	ErrDDoSGuardBlock = errors.New("blocked by DDoS protection: retrying with fresh cookies")
)

// APIError represents an error from the GitFlic API
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("API error %d", e.StatusCode)
}

// IsNotFound returns true if the error is a 404
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return errors.Is(err, ErrNotFound)
}

// IsUnauthorized returns true if the error is a 401
func IsUnauthorized(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401
	}
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden returns true if the error is a 403
func IsForbidden(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 403
	}
	return errors.Is(err, ErrForbidden)
}

// ExitError is returned when a command wants to exit with a specific code
// This allows proper cleanup via defer statements
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.Code)
}

// NewExitError creates an ExitError with the given code
func NewExitError(code int) *ExitError {
	return &ExitError{Code: code}
}

// GetExitCode returns the exit code if err is an ExitError, otherwise 1
func GetExitCode(err error) int {
	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return exitErr.Code
	}
	return 1
}

// IsExitError returns true if the error is an ExitError
func IsExitError(err error) bool {
	var exitErr *ExitError
	return errors.As(err, &exitErr)
}

// IsNetworkError returns true if the error is a retryable network error
func IsNetworkError(err error) bool {
	return errors.Is(err, ErrNetwork)
}

// IsTokenInvalid returns true if the error indicates an invalid/expired token
func IsTokenInvalid(err error) bool {
	return errors.Is(err, ErrTokenInvalid)
}
