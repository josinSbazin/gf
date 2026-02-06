package branch

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
	repo    string
	force   bool
}

func newDeleteCmd() *cobra.Command {
	opts := &deleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a branch",
		Long: `Delete a branch from the repository.

By default, asks for confirmation before deleting.
Use --force to skip confirmation.`,
		Example: `  # Delete branch (with confirmation)
  gf branch delete feature/old-feature

  # Delete branch without confirmation
  gf branch delete feature/old-feature --force`,
		Args:   cobra.ExactArgs(1),
		Hidden: true, // GitFlic API returns 405 Method Not Allowed
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

	// Check if branch exists and is not default
	branch, err := client.Branches().Get(repo.Owner, repo.Name, name)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("branch %q not found in %s", name, repo.FullName())
		}
		return fmt.Errorf("failed to get branch: %w", err)
	}

	if branch.IsDefault {
		return fmt.Errorf("cannot delete the default branch %q", name)
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

	// Delete branch
	err = client.Branches().Delete(repo.Owner, repo.Name, name)
	if err != nil {
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to delete branches in %s", repo.FullName())
		}
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	fmt.Printf("✓ Deleted branch %q\n", name)
	return nil
}
