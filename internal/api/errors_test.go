package api

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name   string
		err    *APIError
		want   string
	}{
		{
			name: "with message",
			err:  &APIError{StatusCode: 404, Message: "not found"},
			want: "API error 404: not found",
		},
		{
			name: "without message",
			err:  &APIError{StatusCode: 500},
			want: "API error 500",
		},
		{
			name: "zero status",
			err:  &APIError{},
			want: "API error 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"ErrNotFound", ErrNotFound, true},
		{"APIError 404", &APIError{StatusCode: 404}, true},
		{"APIError 500", &APIError{StatusCode: 500}, false},
		{"other error", errors.New("something"), false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"ErrUnauthorized", ErrUnauthorized, true},
		{"APIError 401", &APIError{StatusCode: 401}, true},
		{"APIError 403", &APIError{StatusCode: 403}, false},
		{"other error", errors.New("something"), false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnauthorized(tt.err); got != tt.want {
				t.Errorf("IsUnauthorized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsForbidden(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"ErrForbidden", ErrForbidden, true},
		{"APIError 403", &APIError{StatusCode: 403}, true},
		{"APIError 401", &APIError{StatusCode: 401}, false},
		{"other error", errors.New("something"), false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsForbidden(tt.err); got != tt.want {
				t.Errorf("IsForbidden() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorsAreErrors(t *testing.T) {
	// Verify sentinel errors implement error interface
	var _ error = ErrNotFound
	var _ error = ErrUnauthorized
	var _ error = ErrForbidden
	var _ error = ErrNetwork
	var _ error = &APIError{}
}
