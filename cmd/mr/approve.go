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

type approveOptions struct {
	repo string
}

func newApproveCmd() *cobra.Command {
	opts := &approveOptions{}

	cmd := &cobra.Command{
		Use:   "approve <id>",
		Short: "Approve a merge request",
		Long:  `Approve a merge request for merging.`,
		Example: `  # Approve MR #42
  gf mr approve 42`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid merge request ID: %s", args[0])
			}
			return runApprove(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runApprove(opts *approveOptions, id int) error {
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

	// Approve MR
	if err := client.MergeRequests().Approve(repo.Owner, repo.Name, id); err != nil {
		return fmt.Errorf("failed to approve merge request: %w", err)
	}

	fmt.Printf("✓ Approved merge request #%d: %s\n", mr.LocalID, mr.Title)
	return nil
}
