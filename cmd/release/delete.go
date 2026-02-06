package release

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	repo  string
	force bool
}

func newDeleteCmd() *cobra.Command {
	opts := &deleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete <tag>",
		Short: "Delete a release",
		Long: `Delete a release from the repository.

By default, asks for confirmation before deleting.
Use --force to skip confirmation.

Note: This only deletes the release, not the underlying git tag.`,
		Example: `  # Delete release (with confirmation)
  gf release delete v1.0.0

  # Delete release without confirmation
  gf release delete v1.0.0 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runDelete(opts *deleteOptions, tagName string) error {
	// Get repository
	repo, err := git.ResolveRepo(opts.repo, config.DefaultHost())
	if err != nil {
		return fmt.Errorf("could not determine repository: %w\nUse --repo owner/name to specify", err)
	}

	// Load config and create client
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	token, err := cfg.Token()
	if err != nil {
		return fmt.Errorf("not authenticated. Run 'gf auth login' first")
	}

	client := api.NewClient(config.BaseURL(cfg.ActiveHost), token)

	// Check if release exists
	release, err := client.Releases().Get(repo.Owner, repo.Name, tagName)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("release %q not found in %s", tagName, repo.FullName())
		}
		return fmt.Errorf("failed to get release: %w", err)
	}

	// Confirm deletion
	if !opts.force {
		fmt.Printf("Are you sure you want to delete release %q (%s)? [y/N]: ", release.TagName, release.Title)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Delete release
	err = client.Releases().Delete(repo.Owner, repo.Name, tagName)
	if err != nil {
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to delete releases in %s", repo.FullName())
		}
		return fmt.Errorf("failed to delete release: %w", err)
	}

	fmt.Printf("✓ Deleted release %q\n", tagName)
	return nil
}
