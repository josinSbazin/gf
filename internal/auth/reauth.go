package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"golang.org/x/term"
)

// PromptReauth prompts user for a new token when the current one is invalid.
// Returns a new API client with the fresh token, or error if re-auth fails/is declined.
// This allows inline re-authentication without requiring separate `gf auth login`.
func PromptReauth(hostname string) (*api.Client, error) {
	// Check if we're in an interactive terminal
	if !term.IsTerminal(int(syscall.Stdin)) {
		return nil, fmt.Errorf("token expired or invalid: run 'gf auth login' to re-authenticate")
	}

	fmt.Fprintf(os.Stderr, "\nToken expired or invalid for %s\n", hostname)
	fmt.Fprintf(os.Stderr, "Enter new token (or press Enter to cancel): ")

	// Read token without echo
	tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		// Fallback to regular input if terminal password reading fails
		reader := bufio.NewReader(os.Stdin)
		tokenStr, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read token: %w", err)
		}
		tokenBytes = []byte(strings.TrimSpace(tokenStr))
	}
	fmt.Fprintln(os.Stderr) // newline after hidden input

	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return nil, fmt.Errorf("re-authentication cancelled")
	}

	// Verify new token
	baseURL := config.BaseURL(hostname)
	client := api.NewClient(baseURL, token)

	user, err := client.Users().Me()
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Save to config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	cfg.SetHost(hostname, &config.Host{
		Token:    token,
		User:     user.Username,
		Protocol: "https",
	})

	if err := config.Save(cfg); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Logged in as %s\n\n", user.Username)
	return client, nil
}

// HandleTokenError checks if err is a token error and offers inline re-auth.
// If re-auth succeeds, returns new client; if user declines or error, returns original error.
func HandleTokenError(err error, hostname string) (*api.Client, error) {
	if !api.IsTokenInvalid(err) {
		return nil, err
	}

	client, reAuthErr := PromptReauth(hostname)
	if reAuthErr != nil {
		// Return original error with hint to login
		return nil, err
	}

	return client, nil
}

// RetryWithReauth executes fn and, if it returns a token error, prompts for re-auth
// and retries once. This is the recommended way to wrap API calls that need re-auth support.
func RetryWithReauth[T any](hostname string, fn func() (T, error)) (T, error) {
	result, err := fn()
	if err == nil {
		return result, nil
	}

	// Try re-auth if token is invalid
	if api.IsTokenInvalid(err) {
		if _, reAuthErr := PromptReauth(hostname); reAuthErr == nil {
			// Retry with new token (fn should reload config/client internally)
			return fn()
		}
	}

	return result, err
}
