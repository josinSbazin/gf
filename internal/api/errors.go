package api

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorized = errors.New("unauthorized: run 'gf auth login' to authenticate")
	ErrForbidden    = errors.New("forbidden: you don't have permission to access this resource")
	ErrNotFound     = errors.New("not found")
	ErrNetwork      = errors.New("network error: check your connection")
)

// APIError represents an error from the GitFlic API
type APIError struct {
	StatusCode int
	Message    string
	RequestID  string
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
