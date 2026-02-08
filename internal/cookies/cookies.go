package cookies

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const cookiesFile = "cookies.json"

// PersistentCookie represents a cookie for JSON storage
type PersistentCookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path"`
	Expires  time.Time `json:"expires,omitempty"`
	Secure   bool      `json:"secure,omitempty"`
	HttpOnly bool      `json:"httpOnly,omitempty"`
}

// Store manages persistent cookie storage
type Store struct {
	jar      *cookiejar.Jar
	path     string
	mu       sync.Mutex
	modified bool
}

// NewStore creates a new cookie store with persistence
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(home, ".gf", cookiesFile)

	jar, _ := cookiejar.New(nil)
	store := &Store{
		jar:  jar,
		path: path,
	}

	// Load existing cookies
	store.load()

	return store, nil
}

// Jar returns the underlying cookie jar for http.Client
func (s *Store) Jar() http.CookieJar {
	return s.jar
}

// Save persists cookies to disk
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.modified {
		return nil
	}

	// Get cookies for gitflic.ru
	u, _ := url.Parse("https://gitflic.ru")
	cookies := s.jar.Cookies(u)

	if len(cookies) == 0 {
		return nil
	}

	// Convert to persistable format
	var persistent []PersistentCookie
	for _, c := range cookies {
		persistent = append(persistent, PersistentCookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   "gitflic.ru",
			Path:     c.Path,
			Expires:  c.Expires,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
		})
	}

	data, err := json.MarshalIndent(persistent, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}

	s.modified = false
	return os.WriteFile(s.path, data, 0600)
}

// load reads cookies from disk
func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var persistent []PersistentCookie
	if err := json.Unmarshal(data, &persistent); err != nil {
		return err
	}

	// Convert to http.Cookie and set in jar
	var cookies []*http.Cookie
	for _, p := range persistent {
		// Skip expired cookies
		if !p.Expires.IsZero() && p.Expires.Before(time.Now()) {
			continue
		}
		cookies = append(cookies, &http.Cookie{
			Name:     p.Name,
			Value:    p.Value,
			Domain:   p.Domain,
			Path:     p.Path,
			Expires:  p.Expires,
			Secure:   p.Secure,
			HttpOnly: p.HttpOnly,
		})
	}

	if len(cookies) > 0 {
		u, _ := url.Parse("https://gitflic.ru")
		s.jar.SetCookies(u, cookies)
	}

	return nil
}

// MarkModified marks the store as having new cookies to save
func (s *Store) MarkModified() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modified = true
}

// Clear removes all cookies and the cookie file
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create fresh jar
	s.jar, _ = cookiejar.New(nil)
	s.modified = false

	// Remove file
	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
