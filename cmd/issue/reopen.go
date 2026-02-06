package issue

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

func newReopenCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "reopen <id>",
		Short: "Reopen a closed issue",
		Long:  `Reopen a previously closed issue.`,
		Example: `  # Reopen issue
  gf issue reopen 42`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid issue ID: %s", args[0])
			}
			return runReopen(repo, id)
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runReopen(repoFlag string, id int) error {
	// Get repository
	repo, err := git.ResolveRepo(repoFlag, config.DefaultHost())
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

	if issue.State() == "open" {
		fmt.Printf("Issue #%d is already open\n", id)
		return nil
	}

	// Reopen issue
	err = client.Issues().Reopen(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to reopen issues in %s", repo.FullName())
		}
		return fmt.Errorf("failed to reopen issue: %w", err)
	}

	fmt.Printf("✓ Reopened issue #%d\n", id)
	return nil
}
