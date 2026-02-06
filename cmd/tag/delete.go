package tag

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
		Short: "Delete a tag",
		Long: `Delete a tag from the repository.

By default, asks for confirmation before deleting.
Use --force to skip confirmation.

Note: Uses 'git push --delete' because GitFlic REST API
does not support tag deletion.`,
		Example: `  # Delete tag (with confirmation)
  gf tag delete v1.0.0

  # Delete tag without confirmation
  gf tag delete v1.0.0 --force

  # Specify remote explicitly
  gf tag delete v1.0.0 --remote origin`,
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

	// Validate tag via API if possible
	repo, _ := git.ResolveRepo(opts.repo, config.DefaultHost())
	if repo != nil {
		cfg, err := config.Load()
		if err == nil {
			token, err := cfg.Token()
			if err == nil {
				client := api.NewClient(config.BaseURL(cfg.ActiveHost), token)
				_, err := client.Tags().Get(repo.Owner, repo.Name, name)
				if err != nil {
					if api.IsNotFound(err) {
						return fmt.Errorf("tag %q not found in %s", name, repo.FullName())
					}
					// Non-fatal: continue with git
				}
			}
		}
	}

	// Confirm deletion
	if !opts.force {
		fmt.Printf("Are you sure you want to delete tag %q? [y/N]: ", name)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Delete via git (API not supported)
	fmt.Fprintf(os.Stderr, "Note: GitFlic API does not support tag deletion, using git\n")

	ctx, cancel := context.WithTimeout(context.Background(), deleteTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "push", "--delete", remoteName, "refs/tags/"+name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("git push timed out")
		}
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	fmt.Printf("✓ Deleted tag %q\n", name)
	return nil
}
