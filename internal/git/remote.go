package git

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// gitTimeout is the timeout for local git operations.
// Local git commands (rev-parse, symbolic-ref) should complete almost instantly,
// but we set a reasonable timeout to prevent hanging on edge cases.
const gitTimeout = 10 * time.Second

// runGit executes a git command with timeout and returns the output
func runGit(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", errors.New("git command timed out")
		}
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// runGitCheck executes a git command with timeout and returns success/failure
func runGitCheck(args ...string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	return cmd.Run() == nil
}

var (
	ErrNotGitRepo    = errors.New("not a git repository (or any of the parent directories)")
	ErrNoRemote      = errors.New("could not determine repository from git remotes")
	ErrInvalidName   = errors.New("invalid owner or repository name")
)

// validNameRegex validates owner/project names to prevent path traversal
var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9_.]*$`)

// validHostRegex validates hostname format
var validHostRegex = regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9.]*[a-zA-Z0-9]$`)

// ValidateName checks if owner/project name is safe
func ValidateName(name string) error {
	if name == "" || name == "." || name == ".." {
		return ErrInvalidName
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return ErrInvalidName
	}
	if !validNameRegex.MatchString(name) {
		return ErrInvalidName
	}
	return nil
}

// ValidateHost checks if hostname is valid
func ValidateHost(host string) error {
	if host == "" || host == "." || host == ".." {
		return errors.New("invalid hostname")
	}
	if strings.HasPrefix(host, "-") || strings.HasPrefix(host, ".") {
		return errors.New("invalid hostname")
	}
	// Must contain a dot (domain) or be "localhost"
	if !strings.Contains(host, ".") && host != "localhost" {
		return errors.New("invalid hostname")
	}
	if !validHostRegex.MatchString(host) {
		return errors.New("invalid hostname")
	}
	return nil
}

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
		// Use ParseRepoFlag with default host for GF_REPO parsing
		return ParseRepoFlag(repo, "gitflic.ru")
	}

	// Try to get from git remote
	output, err := runGit("remote", "get-url", "origin")
	if err != nil {
		return nil, ErrNotGitRepo
	}

	return parseRemoteURL(output)
}

// Pre-compiled regexes for remote URL parsing (performance optimization)
var remoteURLPatterns = []struct {
	re    *regexp.Regexp
	host  int
	owner int
	name  int
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

// parseRemoteURL parses various git remote URL formats
func parseRemoteURL(url string) (*Repository, error) {
	for _, p := range remoteURLPatterns {
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
	output, err := runGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", ErrNotGitRepo
	}
	return output, nil
}

// ResolveRepo resolves repository from --repo flag or git remote detection
// This is the single entry point for all commands to get repository info
func ResolveRepo(repoFlag string, defaultHost string) (*Repository, error) {
	if repoFlag != "" {
		return ParseRepoFlag(repoFlag, defaultHost)
	}
	return DetectRepo()
}

// ParseRepoFlag parses --repo flag value with validation
func ParseRepoFlag(repoFlag string, defaultHost string) (*Repository, error) {
	if defaultHost == "" {
		defaultHost = "gitflic.ru"
	}

	parts := strings.Split(repoFlag, "/")
	switch len(parts) {
	case 2:
		if err := ValidateName(parts[0]); err != nil {
			return nil, errors.New("invalid owner name")
		}
		if err := ValidateName(parts[1]); err != nil {
			return nil, errors.New("invalid repository name")
		}
		return &Repository{
			Host:  defaultHost,
			Owner: parts[0],
			Name:  parts[1],
		}, nil
	case 3:
		if err := ValidateHost(parts[0]); err != nil {
			return nil, errors.New("invalid hostname")
		}
		if err := ValidateName(parts[1]); err != nil {
			return nil, errors.New("invalid owner name")
		}
		if err := ValidateName(parts[2]); err != nil {
			return nil, errors.New("invalid repository name")
		}
		return &Repository{
			Host:  parts[0],
			Owner: parts[1],
			Name:  parts[2],
		}, nil
	default:
		return nil, errors.New("invalid repository format, expected owner/repo or host/owner/repo")
	}
}

// DefaultBranch returns the default branch (main or master)
func DefaultBranch() (string, error) {
	// Try to get from remote HEAD
	output, err := runGit("symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		// refs/remotes/origin/main -> main
		parts := strings.Split(output, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1], nil
		}
	}

	// Fallback: check if main or master exists
	for _, branch := range []string{"main", "master"} {
		if runGitCheck("rev-parse", "--verify", "refs/heads/"+branch) {
			return branch, nil
		}
	}

	return "main", nil // default fallback
}
