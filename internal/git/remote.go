package git

import (
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	ErrNotGitRepo = errors.New("not a git repository (or any of the parent directories)")
	ErrNoRemote   = errors.New("could not determine repository from git remotes")
)

// Repository represents a GitFlic repository
type Repository struct {
	Host  string
	Owner string
	Name  string
}

// FullName returns "owner/name"
func (r *Repository) FullName() string {
	return r.Owner + "/" + r.Name
}

// DetectRepo determines the repository from git remotes or environment
func DetectRepo() (*Repository, error) {
	// Check environment variable first
	if repo := os.Getenv("GF_REPO"); repo != "" {
		return parseRepoString(repo)
	}

	// Try to get from git remote
	output, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return nil, ErrNotGitRepo
	}

	return parseRemoteURL(strings.TrimSpace(string(output)))
}

// parseRepoString parses "owner/repo" or "host/owner/repo" format
func parseRepoString(s string) (*Repository, error) {
	parts := strings.Split(s, "/")
	switch len(parts) {
	case 2:
		return &Repository{
			Host:  "gitflic.ru",
			Owner: parts[0],
			Name:  parts[1],
		}, nil
	case 3:
		return &Repository{
			Host:  parts[0],
			Owner: parts[1],
			Name:  parts[2],
		}, nil
	default:
		return nil, errors.New("invalid repository format, expected owner/repo or host/owner/repo")
	}
}

// parseRemoteURL parses various git remote URL formats
func parseRemoteURL(url string) (*Repository, error) {
	patterns := []struct {
		re   *regexp.Regexp
		host int
		owner int
		name int
	}{
		// https://gitflic.ru/project/owner/repo.git
		{
			re:    regexp.MustCompile(`https?://([^/]+)/project/([^/]+)/([^/]+?)(?:\.git)?$`),
			host:  1,
			owner: 2,
			name:  3,
		},
		// https://gitflic.ru/owner/repo.git (alternative format)
		{
			re:    regexp.MustCompile(`https?://([^/]+)/([^/]+)/([^/]+?)(?:\.git)?$`),
			host:  1,
			owner: 2,
			name:  3,
		},
		// git@gitflic.ru:owner/repo.git
		{
			re:    regexp.MustCompile(`git@([^:]+):([^/]+)/([^/]+?)(?:\.git)?$`),
			host:  1,
			owner: 2,
			name:  3,
		},
		// ssh://git@gitflic.ru/owner/repo.git
		{
			re:    regexp.MustCompile(`ssh://git@([^/]+)/([^/]+)/([^/]+?)(?:\.git)?$`),
			host:  1,
			owner: 2,
			name:  3,
		},
	}

	for _, p := range patterns {
		matches := p.re.FindStringSubmatch(url)
		if matches != nil {
			return &Repository{
				Host:  matches[p.host],
				Owner: matches[p.owner],
				Name:  strings.TrimSuffix(matches[p.name], ".git"),
			}, nil
		}
	}

	return nil, ErrNoRemote
}

// CurrentBranch returns the current git branch
func CurrentBranch() (string, error) {
	output, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", ErrNotGitRepo
	}
	return strings.TrimSpace(string(output)), nil
}

// DefaultBranch returns the default branch (main or master)
func DefaultBranch() (string, error) {
	// Try to get from remote HEAD
	output, err := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD").Output()
	if err == nil {
		ref := strings.TrimSpace(string(output))
		// refs/remotes/origin/main -> main
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1], nil
		}
	}

	// Fallback: check if main or master exists
	for _, branch := range []string{"main", "master"} {
		if err := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+branch).Run(); err == nil {
			return branch, nil
		}
	}

	return "main", nil // default fallback
}
