package issue

import (
	"fmt"
	"strconv"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type closeOptions struct {
	repo string
}

func newCloseCmd() *cobra.Command {
	opts := &closeOptions{}

	cmd := &cobra.Command{
		Use:   "close <id>",
		Short: "Close an issue",
		Long:  `Close an issue.`,
		Example: `  # Close issue #42
  gf issue close 42`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid issue ID: %s", args[0])
			}
			return runClose(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runClose(opts *closeOptions, id int) error {
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

	// Close issue
	if err := client.Issues().Close(repo.Owner, repo.Name, id); err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("issue #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to close issue: %w", err)
	}

	fmt.Printf("Closed issue #%d in %s\n", id, repo.FullName())
	return nil
}
