package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultHost(t *testing.T) {
	if got := DefaultHost(); got != "gitflic.ru" {
		t.Errorf("DefaultHost() = %q, want gitflic.ru", got)
	}
}

func TestBaseURL(t *testing.T) {
	tests := []struct {
		hostname string
		want     string
	}{
		{"gitflic.ru", "https://api.gitflic.ru"},
		{"git.company.com", "https://git.company.com/rest-api"},
		{"self-hosted.io", "https://self-hosted.io/rest-api"},
	}

	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			if got := BaseURL(tt.hostname); got != tt.want {
				t.Errorf("BaseURL(%q) = %q, want %q", tt.hostname, got, tt.want)
			}
		})
	}
}

func TestConfig_GetHost(t *testing.T) {
	cfg := &Config{
		Hosts: map[string]*Host{
			"gitflic.ru": {Token: "token1", User: "user1"},
		},
	}

	// Test existing host
	host := cfg.GetHost("gitflic.ru")
	if host == nil {
		t.Fatal("GetHost returned nil for existing host")
	}
	if host.Token != "token1" {
		t.Errorf("Token = %q, want token1", host.Token)
	}

	// Test non-existing host
	if cfg.GetHost("unknown") != nil {
		t.Error("GetHost should return nil for unknown host")
	}

	// Test nil hosts map
	cfg.Hosts = nil
	if cfg.GetHost("gitflic.ru") != nil {
		t.Error("GetHost should return nil when Hosts is nil")
	}
}

func TestConfig_SetHost(t *testing.T) {
	cfg := &Config{}

	// Test setting on nil hosts map
	cfg.SetHost("gitflic.ru", &Host{Token: "test"})
	if cfg.Hosts == nil {
		t.Fatal("Hosts should be initialized")
	}
	if cfg.Hosts["gitflic.ru"] == nil {
		t.Fatal("Host not set")
	}
	if cfg.Hosts["gitflic.ru"].Token != "test" {
		t.Errorf("Token = %q, want test", cfg.Hosts["gitflic.ru"].Token)
	}

	// Test overwriting
	cfg.SetHost("gitflic.ru", &Host{Token: "new"})
	if cfg.Hosts["gitflic.ru"].Token != "new" {
		t.Errorf("Token = %q, want new", cfg.Hosts["gitflic.ru"].Token)
	}
}

func TestConfig_Token(t *testing.T) {
	// Test with valid token
	cfg := &Config{
		ActiveHost: "gitflic.ru",
		Hosts: map[string]*Host{
			"gitflic.ru": {Token: "my-token"},
		},
	}
	token, err := cfg.Token()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "my-token" {
		t.Errorf("Token() = %q, want my-token", token)
	}

	// Test with empty token
	cfg.Hosts["gitflic.ru"].Token = ""
	_, err = cfg.Token()
	if err != ErrNoToken {
		t.Errorf("expected ErrNoToken, got %v", err)
	}

	// Test with no host
	cfg.ActiveHost = "unknown"
	_, err = cfg.Token()
	if err != ErrNoToken {
		t.Errorf("expected ErrNoToken, got %v", err)
	}
}

func TestConfig_Token_EnvOverride(t *testing.T) {
	// Test GF_TOKEN environment variable takes priority
	cfg := &Config{
		ActiveHost: "gitflic.ru",
		Hosts: map[string]*Host{
			"gitflic.ru": {Token: "config-token"},
		},
	}

	// Set environment variable
	origEnv := os.Getenv("GF_TOKEN")
	os.Setenv("GF_TOKEN", "env-token")
	defer os.Setenv("GF_TOKEN", origEnv)

	token, err := cfg.Token()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "env-token" {
		t.Errorf("Token() = %q, want env-token (from GF_TOKEN)", token)
	}

	// Unset env, should fall back to config
	os.Unsetenv("GF_TOKEN")
	token, err = cfg.Token()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "config-token" {
		t.Errorf("Token() = %q, want config-token", token)
	}
}

func TestConfig_ActiveHostConfig(t *testing.T) {
	cfg := &Config{
		ActiveHost: "gitflic.ru",
		Hosts: map[string]*Host{
			"gitflic.ru": {Token: "token"},
		},
	}

	host := cfg.ActiveHostConfig()
	if host == nil {
		t.Fatal("ActiveHostConfig returned nil")
	}
	if host.Token != "token" {
		t.Errorf("Token = %q", host.Token)
	}

	// Test with empty active host (should default)
	cfg.ActiveHost = ""
	host = cfg.ActiveHostConfig()
	if host == nil {
		t.Fatal("ActiveHostConfig returned nil with empty ActiveHost")
	}
}

func TestConfig_JSONSerialization(t *testing.T) {
	cfg := &Config{
		Version:    1,
		ActiveHost: "gitflic.ru",
		Hosts: map[string]*Host{
			"gitflic.ru": {
				Token:    "test-token",
				User:     "testuser",
				Protocol: "https",
			},
		},
	}

	// Marshal
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal
	var cfg2 Config
	if err := json.Unmarshal(data, &cfg2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg2.Version != 1 {
		t.Errorf("Version = %d", cfg2.Version)
	}
	if cfg2.ActiveHost != "gitflic.ru" {
		t.Errorf("ActiveHost = %q", cfg2.ActiveHost)
	}
	if cfg2.Hosts["gitflic.ru"] == nil {
		t.Fatal("Host not deserialized")
	}
	if cfg2.Hosts["gitflic.ru"].Token != "test-token" {
		t.Errorf("Token = %q", cfg2.Hosts["gitflic.ru"].Token)
	}
}

func TestLoadSave(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "gf-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for test (both Unix and Windows)
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	// Test Save
	cfg := &Config{
		Version:    1,
		ActiveHost: "gitflic.ru",
		Hosts: map[string]*Host{
			"gitflic.ru": {Token: "test-token", User: "testuser"},
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	path := filepath.Join(tmpDir, ".gf", "config.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}
	// On Unix, check permissions (skip on Windows)
	if os.PathSeparator == '/' {
		if info.Mode().Perm() != 0600 {
			t.Errorf("file permissions = %o, want 0600", info.Mode().Perm())
		}
	}

	// Test Load
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.ActiveHost != "gitflic.ru" {
		t.Errorf("ActiveHost = %q", loaded.ActiveHost)
	}
	if loaded.Hosts["gitflic.ru"] == nil || loaded.Hosts["gitflic.ru"].Token != "test-token" {
		t.Error("Token not loaded correctly")
	}
}

func TestLoad_NoFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "gf-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for test (both Unix and Windows)
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	// Load should return empty config when file doesn't exist
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("cfg is nil")
	}
	if cfg.Version != 1 {
		t.Errorf("Version = %d, want 1", cfg.Version)
	}
	if cfg.ActiveHost != "gitflic.ru" {
		t.Errorf("ActiveHost = %q, want gitflic.ru", cfg.ActiveHost)
	}
	if cfg.Hosts == nil {
		t.Error("Hosts is nil")
	}
}
