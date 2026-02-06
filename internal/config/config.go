package config

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	configDir  = ".gf"
	configFile = "config.json"

	// DefaultHostname is the default GitFlic cloud hostname
	DefaultHostname = "gitflic.ru"
	// DefaultAPIHost is the API subdomain for GitFlic cloud
	DefaultAPIHost = "api.gitflic.ru"
	// DefaultAPIBaseURL is the full API base URL (pre-computed to avoid concatenation)
	DefaultAPIBaseURL = "https://" + DefaultAPIHost
)

var (
	ErrNoToken     = errors.New("no token configured")
	ErrNoHost      = errors.New("no host configured")
	ErrNotLoggedIn = errors.New("not logged in")
	ErrInternalIP  = errors.New("hostname resolves to internal IP address")
)

// internalIPNetworks holds pre-parsed CIDR networks for performance
// Initialized once at package load time
var internalIPNetworks []*net.IPNet

func init() {
	// Pre-parse CIDR ranges for internal IP detection (RFC 1918, RFC 5737, etc.)
	cidrRanges := []string{
		"10.0.0.0/8",     // Class A private
		"172.16.0.0/12",  // Class B private
		"192.168.0.0/16", // Class C private
		"127.0.0.0/8",    // Loopback
		"169.254.0.0/16", // Link-local
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 private
		"fe80::/10",      // IPv6 link-local
	}
	for _, cidr := range cidrRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			internalIPNetworks = append(internalIPNetworks, network)
		}
	}
}

// IsInternalHost checks if hostname resolves to an internal/private IP.
// Returns true for localhost, loopback, and private network addresses.
// This is a security check to prevent accidental credential leakage to internal services.
func IsInternalHost(hostname string) bool {
	// localhost is explicitly internal
	if hostname == "localhost" || strings.HasPrefix(hostname, "localhost:") {
		return true
	}

	// Try to parse as IP address directly
	host := hostname
	if h, _, err := net.SplitHostPort(hostname); err == nil {
		host = h
	}

	ip := net.ParseIP(host)
	if ip == nil {
		// Not a direct IP, try DNS lookup
		ips, err := net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			return false // Can't resolve, assume external
		}
		ip = ips[0]
	}

	// Check against pre-parsed internal network ranges
	for _, network := range internalIPNetworks {
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// Config represents the gf configuration
type Config struct {
	Version    int              `json:"version"`
	ActiveHost string           `json:"active_host"`
	Hosts      map[string]*Host `json:"hosts"`
}

// Host represents a GitFlic host configuration
type Host struct {
	Token    string `json:"token"`
	User     string `json:"user"`
	Protocol string `json:"protocol,omitempty"`
}

// DefaultHost returns the default GitFlic hostname
func DefaultHost() string {
	return DefaultHostname
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, configFile), nil
}

// Load reads the config from disk
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &Config{
				Version:    1,
				ActiveHost: DefaultHost(),
				Hosts:      make(map[string]*Host),
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Hosts == nil {
		cfg.Hosts = make(map[string]*Host)
	}

	return &cfg, nil
}

// Save writes the config to disk
func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	// Create directory with restricted permissions
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Write file with restricted permissions (owner read/write only)
	return os.WriteFile(path, data, 0600)
}

// GetHost returns the host configuration for the given hostname
func (c *Config) GetHost(hostname string) *Host {
	if c.Hosts == nil {
		return nil
	}
	return c.Hosts[hostname]
}

// ActiveHostConfig returns the configuration for the active host
func (c *Config) ActiveHostConfig() *Host {
	if c.ActiveHost == "" {
		c.ActiveHost = DefaultHost()
	}
	return c.GetHost(c.ActiveHost)
}

// Token returns the token for the active host
// Priority: GF_TOKEN env > config file
func (c *Config) Token() (string, error) {
	// Check environment variable first
	if token := os.Getenv("GF_TOKEN"); token != "" {
		return token, nil
	}

	host := c.ActiveHostConfig()
	if host == nil || host.Token == "" {
		return "", ErrNoToken
	}
	return host.Token, nil
}

// SetHost sets the host configuration
func (c *Config) SetHost(hostname string, host *Host) {
	if c.Hosts == nil {
		c.Hosts = make(map[string]*Host)
	}
	c.Hosts[hostname] = host
}

// BaseURL returns the API base URL for the given hostname
func BaseURL(hostname string) string {
	if hostname == DefaultHostname {
		return DefaultAPIBaseURL
	}
	// For self-hosted instances
	return "https://" + hostname + "/rest-api"
}
