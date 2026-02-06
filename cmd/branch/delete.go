package branch

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

const deleteTimeout = 60 * time.Second

type deleteOptions struct {
	repo   string
	force  bool
	remote string
}

func newDeleteCmd() *cobra.Command {
	opts := &deleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a branch",
		Long: `Delete a branch from the repository.

By default, asks for confirmation before deleting.
Use --force to skip confirmation.

Note: Uses 'git push --delete' because GitFlic REST API
does not support branch deletion.`,
		Example: `  # Delete branch (with confirmation)
  gf branch delete feature/old-feature

  # Delete branch without confirmation
  gf branch delete feature/old-feature --force

  # Specify remote explicitly
  gf branch delete feature/old-feature --remote origin`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name) - for validation")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&opts.remote, "remote", "", "Git remote name (default: auto-detect)")

	return cmd
}

func runDelete(opts *deleteOptions, name string) error {
	// Find remote
	remoteName := opts.remote
	if remoteName == "" {
		remote, err := git.FindGitflicRemote()
		if err != nil {
			return fmt.Errorf("could not find GitFlic remote: %w\nUse --remote to specify", err)
		}
		remoteName = remote
	}

	// Validate branch via API if possible
	repo, _ := git.ResolveRepo(opts.repo, config.DefaultHost())
	if repo != nil {
		cfg, err := config.Load()
		if err == nil {
			token, err := cfg.Token()
			if err == nil {
				client := api.NewClient(config.BaseURL(cfg.ActiveHost), token)
				branch, err := client.Branches().Get(repo.Owner, repo.Name, name)
				if err != nil {
					if api.IsNotFound(err) {
						return fmt.Errorf("branch %q not found in %s", name, repo.FullName())
					}
					// Non-fatal: continue with git
				} else if branch.IsDefault {
					return fmt.Errorf("cannot delete the default branch %q", name)
				}
			}
		}
	}

	// Confirm deletion
	if !opts.force {
		fmt.Printf("Are you sure you want to delete branch %q? [y/N]: ", name)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Delete via git (API not supported)
	fmt.Fprintf(os.Stderr, "Note: GitFlic API does not support branch deletion, using git\n")

	ctx, cancel := context.WithTimeout(context.Background(), deleteTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "push", "--delete", remoteName, name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("git push timed out")
		}
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	fmt.Printf("✓ Deleted branch %q\n", name)
	return nil
}
