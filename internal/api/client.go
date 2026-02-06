package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/josinSbazin/gf/internal/version"
)

const (
	maxRetries    = 3
	retryBaseWait = 500 * time.Millisecond
)

// Client is the GitFlic API client
type Client struct {
	BaseURL      string
	Token        string
	httpClient   *http.Client
	cookiesMu    sync.Mutex
	cookiesReady atomic.Bool
}

// NewClient creates a new API client with cookie jar for DDoS Guard support
func NewClient(baseURL, token string) *Client {
	// cookiejar.New with nil options cannot return an error (Go 1.x behavior)
	jar, _ := cookiejar.New(nil)
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
	}
}

// warmupCookies visits the main GitFlic site to obtain DDoS Guard cookies.
// This is required because api.gitflic.ru is protected by DDoS Guard which
// blocks requests without valid __ddg* cookies.
// Uses double-checked locking pattern for performance.
func (c *Client) warmupCookies(ctx context.Context) error {
	// Fast path: check without lock
	if c.cookiesReady.Load() {
		return nil
	}

	// Slow path: acquire lock and check again
	c.cookiesMu.Lock()
	defer c.cookiesMu.Unlock()

	// Double-check after acquiring lock
	if c.cookiesReady.Load() {
		return nil
	}

	// Extract hostname from BaseURL to visit the main site
	// api.gitflic.ru -> gitflic.ru
	mainSiteURL := c.getMainSiteURL()
	if mainSiteURL == "" {
		c.cookiesReady.Store(true) // Skip for non-gitflic hosts
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mainSiteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create warmup request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to warmup cookies: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	c.cookiesReady.Store(true)

	if os.Getenv("GF_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[DEBUG] Warmed up DDoS Guard cookies from %s\n", mainSiteURL)
	}

	return nil
}

// getMainSiteURL returns the main site URL for cookie warmup
func (c *Client) getMainSiteURL() string {
	parsed, err := url.Parse(c.BaseURL)
	if err != nil {
		return ""
	}

	host := parsed.Host
	// api.gitflic.ru -> gitflic.ru
	if strings.HasPrefix(host, "api.") {
		host = strings.TrimPrefix(host, "api.")
	}

	// Only warmup for gitflic.ru (cloud version has DDoS Guard)
	if !strings.Contains(host, "gitflic.ru") {
		return ""
	}

	return "https://" + host + "/"
}

// resetCookies clears cookie state to force re-warmup on next request
func (c *Client) resetCookies() {
	c.cookiesMu.Lock()
	defer c.cookiesMu.Unlock()
	c.cookiesReady.Store(false)
	// cookiejar.New with nil options cannot return an error (Go 1.x behavior)
	jar, _ := cookiejar.New(nil)
	c.httpClient.Jar = jar
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
func (c *Client) doRequest(ctx context.Context, method, urlStr string, bodyData []byte, out any) error {
	// Warmup cookies for DDoS Guard (only for gitflic.ru)
	if err := c.warmupCookies(ctx); err != nil {
		if os.Getenv("GF_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[DEBUG] Cookie warmup failed: %v\n", err)
		}
		// Continue anyway, might work without cookies
	}

	var bodyReader io.Reader
	if bodyData != nil {
		bodyReader = bytes.NewReader(bodyData)
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	// Use browser-like User-Agent for DDoS Guard compatibility
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; gf-cli/"+version.Version+")")
	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}

	// Debug mode: print request details
	if os.Getenv("GF_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s %s\n", method, urlStr)
		if bodyData != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] Request body: %s\n", string(bodyData))
		}
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
		// Read body for error handling
		bodyBytes, _ := io.ReadAll(resp.Body)

		// Debug mode: print response details on error
		if os.Getenv("GF_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[DEBUG] Response status: %d\n", resp.StatusCode)
			fmt.Fprintf(os.Stderr, "[DEBUG] Response body: %s\n", string(bodyBytes))
		}

		// Check if this is a DDoS Guard block (403 with AuthenticationException in body)
		if resp.StatusCode == http.StatusForbidden && strings.Contains(string(bodyBytes), "AuthenticationException") {
			// Reset cookies and return special error
			c.resetCookies()
			return ErrDDoSGuardBlock
		}

		// Reset body for handleError
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
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

// UploadFile uploads a file using multipart form data
func (c *Client) UploadFile(path, fieldName, fileName string, fileData io.Reader, out any) error {
	return c.UploadFileWithContext(context.Background(), path, fieldName, fileName, fileData, out)
}

// UploadFileWithContext uploads a file with context support
func (c *Client) UploadFileWithContext(ctx context.Context, path, fieldName, fileName string, fileData io.Reader, out any) error {
	// Warmup cookies for DDoS Guard
	if err := c.warmupCookies(ctx); err != nil {
		if os.Getenv("GF_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[DEBUG] Cookie warmup failed: %v\n", err)
		}
	}

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, fileData); err != nil {
		return fmt.Errorf("failed to write file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	urlStr := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; gf-cli/"+version.Version+")")
	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
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

// DownloadFile downloads a file and returns the response body
func (c *Client) DownloadFile(path string) (io.ReadCloser, string, error) {
	return c.DownloadFileWithContext(context.Background(), path)
}

// DownloadFileWithContext downloads a file with context support
func (c *Client) DownloadFileWithContext(ctx context.Context, path string) (io.ReadCloser, string, error) {
	// Warmup cookies for DDoS Guard
	if err := c.warmupCookies(ctx); err != nil {
		if os.Getenv("GF_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[DEBUG] Cookie warmup failed: %v\n", err)
		}
	}

	urlStr := c.BaseURL + path

	// Debug mode: print request details
	if os.Getenv("GF_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[DEBUG] GET (download) %s\n", urlStr)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; gf-cli/"+version.Version+")")
	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrNetwork, err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, "", c.handleError(resp)
	}

	// Extract filename from Content-Disposition header if available
	fileName := ""
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			fileName = params["filename"]
		}
	}

	return resp.Body, fileName, nil
}

// rawRequest performs a request without 403 diagnosis (to avoid recursion)
func (c *Client) rawRequest(method, path string, body, out any) error {
	var bodyData []byte
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyData = data
	}

	url := c.BaseURL + path
	return c.doRawRequest(context.Background(), method, url, bodyData, out)
}

// doRawRequest performs a single HTTP request without 403 diagnosis
func (c *Client) doRawRequest(ctx context.Context, method, url string, bodyData []byte, out any) error {
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
		// Simple error handling without DiagnoseForbidden
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return ErrUnauthorized
		case http.StatusForbidden:
			return ErrForbidden
		case http.StatusNotFound:
			return ErrNotFound
		default:
			return &APIError{StatusCode: resp.StatusCode}
		}
	}

	if out != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) handleError(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		// Diagnose whether this is a token issue or permission issue
		return c.DiagnoseForbidden()
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

// ValidateToken checks if the current token is valid by calling /user/me
// Returns nil if token is valid, ErrTokenInvalid if expired/invalid,
// or other errors for network issues
func (c *Client) ValidateToken() error {
	var user struct {
		ID string `json:"id"`
	}
	// Use rawRequest to avoid infinite recursion via handleError->DiagnoseForbidden->ValidateToken
	err := c.rawRequest(http.MethodGet, "/user/me", nil, &user)
	if err != nil {
		if IsForbidden(err) || IsUnauthorized(err) {
			return ErrTokenInvalid
		}
		return err
	}
	return nil
}

// DiagnoseForbidden checks if a 403 error is due to an invalid token or lack of permissions.
// Call this when you receive a 403 to get a more specific error message.
// Returns ErrTokenInvalid if the token is invalid/expired, or ErrForbidden if the token
// is valid but the user lacks permissions for the specific resource.
func (c *Client) DiagnoseForbidden() error {
	if err := c.ValidateToken(); err != nil {
		if IsNetworkError(err) {
			return ErrForbidden // Can't determine, return original error
		}
		return ErrTokenInvalid
	}
	return ErrForbidden // Token valid, but no permission to resource
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

// Issues returns the issue service
func (c *Client) Issues() *IssueService {
	return &IssueService{client: c}
}

// Releases returns the release service
func (c *Client) Releases() *ReleaseService {
	return &ReleaseService{client: c}
}

// Branches returns the branch service
func (c *Client) Branches() *BranchService {
	return &BranchService{client: c}
}

// Tags returns the tag service
func (c *Client) Tags() *TagService {
	return &TagService{client: c}
}

// Commits returns the commit service
func (c *Client) Commits() *CommitService {
	return &CommitService{client: c}
}

// Files returns the file service
func (c *Client) Files() *FileService {
	return &FileService{client: c}
}

// Webhooks returns the webhook service
func (c *Client) Webhooks() *WebhookService {
	return &WebhookService{client: c}
}
