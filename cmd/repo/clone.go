package repo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/josinSbazin/gf/internal/config"
	gitpkg "github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

// cloneTimeout is the maximum time allowed for git clone operation.
// Clone operations can be slow for large repositories over slow connections.
const cloneTimeout = 10 * time.Minute

type cloneOptions struct {
	directory string
	ssh       bool
}

func newCloneCmd() *cobra.Command {
	opts := &cloneOptions{}

	cmd := &cobra.Command{
		Use:   "clone <repository> [<directory>]",
		Short: "Clone a repository",
		Long: `Clone a GitFlic repository to the local machine.

Repository can be specified as:
  - owner/name (uses default host)
  - full URL (https://gitflic.ru/project/owner/name)`,
		Example: `  # Clone repository
  gf repo clone owner/project

  # Clone to specific directory
  gf repo clone owner/project mydir

  # Clone using SSH
  gf repo clone owner/project --ssh`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArg := args[0]
			if len(args) > 1 {
				opts.directory = args[1]
			}
			return runClone(opts, repoArg)
		},
	}

	cmd.Flags().BoolVar(&opts.ssh, "ssh", false, "Clone using SSH instead of HTTPS")

	return cmd
}

func runClone(opts *cloneOptions, repoArg string) error {
	var owner, name, host string
	var cloneURL string

	// Check if it's already a URL
	if strings.HasPrefix(repoArg, "https://") || strings.HasPrefix(repoArg, "git@") {
		// Use URL directly
		cloneURL = repoArg
		// Extract name from URL for directory
		parts := strings.Split(strings.TrimSuffix(repoArg, ".git"), "/")
		name = parts[len(parts)-1]
		if len(parts) >= 2 {
			owner = parts[len(parts)-2]
		}
		host = config.DefaultHost()
	} else {
		// Parse as owner/name
		repo, err := gitpkg.ParseRepoFlag(repoArg, config.DefaultHost())
		if err != nil {
			return fmt.Errorf("invalid repository: %w", err)
		}
		owner = repo.Owner
		name = repo.Name
		host = repo.Host
		cloneURL = "" // Will be built below
	}

	// Build clone URL if not already set
	if cloneURL == "" {
		if opts.ssh {
			cloneURL = fmt.Sprintf("git@%s:%s/%s.git", host, owner, name)
		} else {
			cloneURL = fmt.Sprintf("https://%s/project/%s/%s.git", host, owner, name)
		}
	}

	// Determine target directory
	targetDir := opts.directory
	if targetDir == "" {
		targetDir = name
	}

	// Security: Prevent path traversal attacks
	if strings.Contains(targetDir, "..") || filepath.IsAbs(targetDir) {
		return fmt.Errorf("invalid directory name: %s (path traversal not allowed)", targetDir)
	}
	// Validate directory name doesn't contain path separators when auto-detected
	if opts.directory == "" && (strings.Contains(name, "/") || strings.Contains(name, "\\")) {
		return fmt.Errorf("invalid repository name for directory: %s", name)
	}

	// Check if directory already exists
	if _, err := os.Stat(targetDir); err == nil {
		return fmt.Errorf("directory '%s' already exists", targetDir)
	}

	// Run git clone with timeout and signal handling
	fmt.Printf("Cloning into '%s'...\n", targetDir)

	ctx, cancel := context.WithTimeout(context.Background(), cloneTimeout)
	defer cancel()

	// Handle interrupt signal for clean cancellation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go func() {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
		}
	}()

	gitCmd := exec.CommandContext(ctx, "git", "clone", cloneURL, targetDir)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr

	if err := gitCmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("git clone timed out after %v", cloneTimeout)
		}
		if ctx.Err() == context.Canceled {
			return fmt.Errorf("git clone cancelled")
		}
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Print success message
	absPath, _ := filepath.Abs(targetDir)
	fmt.Printf("\nCloned %s/%s to %s\n", owner, name, absPath)

	return nil
}
