package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/josinSbazin/gf/internal/version"
)

const (
	maxRetries    = 3
	retryBaseWait = 500 * time.Millisecond
)

// Client is the GitFlic API client
type Client struct {
	BaseURL    string
	Token      string
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// REST performs an HTTP request and decodes the JSON response
func (c *Client) REST(method, path string, body, out any) error {
	return c.RESTWithContext(context.Background(), method, path, body, out)
}

// RESTWithContext performs an HTTP request with context support for cancellation
// Includes automatic retry with exponential backoff for network errors
func (c *Client) RESTWithContext(ctx context.Context, method, path string, body, out any) error {
	var bodyData []byte
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyData = data
	}

	url := c.BaseURL + path
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Wait before retry (skip on first attempt)
		if attempt > 0 {
			wait := retryBaseWait * time.Duration(1<<(attempt-1)) // exponential backoff
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		err := c.doRequest(ctx, method, url, bodyData, out)
		if err == nil {
			return nil
		}

		// Only retry on network errors, not HTTP errors
		if !isNetworkError(err) {
			return err
		}

		lastErr = err
	}

	return lastErr
}

// doRequest performs a single HTTP request
func (c *Client) doRequest(ctx context.Context, method, url string, bodyData []byte, out any) error {
	var bodyReader io.Reader
	if bodyData != nil {
		bodyReader = bytes.NewReader(bodyData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "gf-cli/"+version.Version)
	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.Canceled {
			return context.Canceled
		}
		if ctx.Err() == context.DeadlineExceeded {
			return context.DeadlineExceeded
		}
		return fmt.Errorf("%w: %v", ErrNetwork, err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return c.handleError(resp)
	}

	if out != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// isNetworkError returns true if the error is a retryable network error
func isNetworkError(err error) bool {
	return errors.Is(err, ErrNetwork)
}

// Get performs a GET request
func (c *Client) Get(path string, out any) error {
	return c.REST(http.MethodGet, path, nil, out)
}

// GetWithContext performs a GET request with context
func (c *Client) GetWithContext(ctx context.Context, path string, out any) error {
	return c.RESTWithContext(ctx, http.MethodGet, path, nil, out)
}

// Post performs a POST request
func (c *Client) Post(path string, body, out any) error {
	return c.REST(http.MethodPost, path, body, out)
}

// PostWithContext performs a POST request with context
func (c *Client) PostWithContext(ctx context.Context, path string, body, out any) error {
	return c.RESTWithContext(ctx, http.MethodPost, path, body, out)
}

// Put performs a PUT request
func (c *Client) Put(path string, body, out any) error {
	return c.REST(http.MethodPut, path, body, out)
}

// Delete performs a DELETE request
func (c *Client) Delete(path string) error {
	return c.REST(http.MethodDelete, path, nil, nil)
}

func (c *Client) handleError(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	default:
		// Try to parse error message from response
		var errResp struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			msg := errResp.Message
			if msg == "" {
				msg = errResp.Error
			}
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    msg,
			}
		}
		return &APIError{
			StatusCode: resp.StatusCode,
		}
	}
}

// MergeRequests returns the merge request service
func (c *Client) MergeRequests() *MergeRequestService {
	return &MergeRequestService{client: c}
}

// Pipelines returns the pipeline service
func (c *Client) Pipelines() *PipelineService {
	return &PipelineService{client: c}
}

// Projects returns the project service
func (c *Client) Projects() *ProjectService {
	return &ProjectService{client: c}
}

// Users returns the user service
func (c *Client) Users() *UserService {
	return &UserService{client: c}
}
