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

type readyOptions struct {
	repo string
}

func newReadyCmd() *cobra.Command {
	opts := &readyOptions{}

	cmd := &cobra.Command{
		Use:   "ready <id>",
		Short: "Mark a draft merge request as ready for review",
		Long:  `Remove the draft status from a merge request, marking it as ready for review.`,
		Example: `  # Mark MR #42 as ready
  gf mr ready 42`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid merge request ID: %s", args[0])
			}
			return runReady(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runReady(opts *readyOptions, id int) error {
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

	if mr.State() != "open" {
		return fmt.Errorf("merge request #%d is %s, cannot mark as ready", id, mr.State())
	}

	// Update MR to remove draft status
	isDraft := false
	_, err = client.MergeRequests().Update(repo.Owner, repo.Name, id, &api.UpdateMRRequest{
		IsDraft: &isDraft,
	})
	if err != nil {
		return fmt.Errorf("failed to mark merge request as ready: %w", err)
	}

	fmt.Printf("✓ Merge request #%d is now ready for review\n", mr.LocalID)
	return nil
}
