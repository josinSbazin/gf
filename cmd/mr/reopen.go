package mr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type reopenOptions struct {
	repo string
}

func newReopenCmd() *cobra.Command {
	opts := &reopenOptions{}

	cmd := &cobra.Command{
		Use:   "reopen <id>",
		Short: "Reopen a closed merge request",
		Long:  `Reopen a merge request that was closed without merging.`,
		Example: `  # Reopen MR #42
  gf mr reopen 42`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid merge request ID: %s", args[0])
			}
			return runReopen(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runReopen(opts *reopenOptions, id int) error {
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

	// Get MR info first
	mr, err := client.MergeRequests().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("merge request #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to get merge request: %w", err)
	}

	if mr.State() == "merged" {
		return fmt.Errorf("merge request #%d is already merged, cannot reopen", id)
	}

	if mr.State() == "open" {
		return fmt.Errorf("merge request #%d is already open", id)
	}

	// Reopen MR
	if err := client.MergeRequests().Reopen(repo.Owner, repo.Name, id); err != nil {
		return fmt.Errorf("failed to reopen merge request: %w", err)
	}

	fmt.Printf("✓ Reopened merge request #%d: %s\n", mr.LocalID, mr.Title)
	return nil
}
