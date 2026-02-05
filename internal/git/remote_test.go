package git

import (
	"os"
	"testing"
)

func TestRepository_FullName(t *testing.T) {
	tests := []struct {
		repo *Repository
		want string
	}{
		{&Repository{Owner: "uply-dev", Name: "backend"}, "uply-dev/backend"},
		{&Repository{Owner: "user", Name: "repo"}, "user/repo"},
		{&Repository{Owner: "", Name: ""}, "/"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.repo.FullName(); got != tt.want {
				t.Errorf("FullName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseRepoString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Repository
		wantErr bool
	}{
		{
			name:  "owner/repo",
			input: "uply-dev/backend",
			want:  &Repository{Host: "gitflic.ru", Owner: "uply-dev", Name: "backend"},
		},
		{
			name:  "host/owner/repo",
			input: "git.company.com/org/project",
			want:  &Repository{Host: "git.company.com", Owner: "org", Name: "project"},
		},
		{
			name:    "single part",
			input:   "just-repo",
			wantErr: true,
		},
		{
			name:    "too many parts",
			input:   "a/b/c/d",
			wantErr: true,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRepoString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != tt.want.Host {
				t.Errorf("Host = %q, want %q", got.Host, tt.want.Host)
			}
			if got.Owner != tt.want.Owner {
				t.Errorf("Owner = %q, want %q", got.Owner, tt.want.Owner)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
		})
	}
}

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    *Repository
		wantErr bool
	}{
		{
			name: "https with project path",
			url:  "https://gitflic.ru/project/uply-dev/backend.git",
			want: &Repository{Host: "gitflic.ru", Owner: "uply-dev", Name: "backend"},
		},
		{
			name: "https with project path no .git",
			url:  "https://gitflic.ru/project/uply-dev/backend",
			want: &Repository{Host: "gitflic.ru", Owner: "uply-dev", Name: "backend"},
		},
		{
			name: "https alternative format",
			url:  "https://gitflic.ru/owner/repo.git",
			want: &Repository{Host: "gitflic.ru", Owner: "owner", Name: "repo"},
		},
		{
			name: "git SSH format",
			url:  "git@gitflic.ru:uply-dev/backend.git",
			want: &Repository{Host: "gitflic.ru", Owner: "uply-dev", Name: "backend"},
		},
		{
			name: "git SSH format no .git",
			url:  "git@gitflic.ru:owner/repo",
			want: &Repository{Host: "gitflic.ru", Owner: "owner", Name: "repo"},
		},
		{
			name: "ssh:// format",
			url:  "ssh://git@gitflic.ru/owner/repo.git",
			want: &Repository{Host: "gitflic.ru", Owner: "owner", Name: "repo"},
		},
		{
			name: "self-hosted https",
			url:  "https://git.company.com/project/org/myproject.git",
			want: &Repository{Host: "git.company.com", Owner: "org", Name: "myproject"},
		},
		{
			name: "self-hosted ssh",
			url:  "git@git.company.com:org/myproject.git",
			want: &Repository{Host: "git.company.com", Owner: "org", Name: "myproject"},
		},
		{
			name:    "invalid url",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "github url",
			url:     "https://github.com/user/repo.git",
			want:    &Repository{Host: "github.com", Owner: "user", Name: "repo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRemoteURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != tt.want.Host {
				t.Errorf("Host = %q, want %q", got.Host, tt.want.Host)
			}
			if got.Owner != tt.want.Owner {
				t.Errorf("Owner = %q, want %q", got.Owner, tt.want.Owner)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
		})
	}
}

func TestDetectRepo_FromEnv(t *testing.T) {
	// Set environment variable
	os.Setenv("GF_REPO", "test-owner/test-repo")
	defer os.Unsetenv("GF_REPO")

	repo, err := DetectRepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.Owner != "test-owner" {
		t.Errorf("Owner = %q, want test-owner", repo.Owner)
	}
	if repo.Name != "test-repo" {
		t.Errorf("Name = %q, want test-repo", repo.Name)
	}
	if repo.Host != "gitflic.ru" {
		t.Errorf("Host = %q, want gitflic.ru", repo.Host)
	}
}

func TestDetectRepo_FromEnvWithHost(t *testing.T) {
	// Set environment variable with host
	os.Setenv("GF_REPO", "git.example.com/org/project")
	defer os.Unsetenv("GF_REPO")

	repo, err := DetectRepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.Host != "git.example.com" {
		t.Errorf("Host = %q, want git.example.com", repo.Host)
	}
	if repo.Owner != "org" {
		t.Errorf("Owner = %q, want org", repo.Owner)
	}
	if repo.Name != "project" {
		t.Errorf("Name = %q, want project", repo.Name)
	}
}

func TestErrors(t *testing.T) {
	// Verify error messages
	if ErrNotGitRepo.Error() == "" {
		t.Error("ErrNotGitRepo has empty message")
	}
	if ErrNoRemote.Error() == "" {
		t.Error("ErrNoRemote has empty message")
	}
}
