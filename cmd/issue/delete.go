package issue

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
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
		Use:   "delete <id>",
		Short: "Delete an issue",
		Long: `Delete an issue from the repository.

By default, asks for confirmation before deleting.
Use --force to skip confirmation.

Warning: This action cannot be undone.`,
		Example: `  # Delete issue (with confirmation)
  gf issue delete 42

  # Delete issue without confirmation
  gf issue delete 42 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid issue ID: %s", args[0])
			}
			return runDeleteIssue(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runDeleteIssue(opts *deleteOptions, id int) error {
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

	// Check if issue exists
	issue, err := client.Issues().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("issue #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to get issue: %w", err)
	}

	// Confirm deletion
	if !opts.force {
		fmt.Printf("Are you sure you want to delete issue #%d: %s? [y/N]: ", issue.LocalID, issue.Title)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Delete issue
	err = client.Issues().Delete(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsMethodNotAllowed(err) {
			return fmt.Errorf("issue deletion is not supported by GitFlic API\nUse the web interface: https://%s/%s/%s/issue/%d",
				cfg.ActiveHost, repo.Owner, repo.Name, id)
		}
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to delete issues in %s", repo.FullName())
		}
		return fmt.Errorf("failed to delete issue: %w", err)
	}

	fmt.Printf("✓ Deleted issue #%d\n", id)
	return nil
}
