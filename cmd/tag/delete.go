package tag

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
		Use:   "delete <name>",
		Short: "Delete a tag",
		Long: `Delete a tag from the repository.

By default, asks for confirmation before deleting.
Use --force to skip confirmation.`,
		Example: `  # Delete tag (with confirmation)
  gf tag delete v1.0.0

  # Delete tag without confirmation
  gf tag delete v1.0.0 --force`,
		Args:   cobra.ExactArgs(1),
		Hidden: true, // GitFlic API has no delete endpoint for tags
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runDelete(opts *deleteOptions, name string) error {
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

	// Check if tag exists
	_, err = client.Tags().Get(repo.Owner, repo.Name, name)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("tag %q not found in %s", name, repo.FullName())
		}
		return fmt.Errorf("failed to get tag: %w", err)
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

	// Delete tag
	err = client.Tags().Delete(repo.Owner, repo.Name, name)
	if err != nil {
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to delete tags in %s", repo.FullName())
		}
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	fmt.Printf("✓ Deleted tag %q\n", name)
	return nil
}
