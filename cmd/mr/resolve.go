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

type resolveOptions struct {
	repo       string
	discussion string
}

func newResolveCmd() *cobra.Command {
	opts := &resolveOptions{}

	cmd := &cobra.Command{
		Use:   "resolve <mr-id>",
		Short: "Resolve a discussion on a merge request",
		Long: `Mark a discussion thread as resolved.

Use --discussion to specify the discussion UUID (shown in 'gf mr comments' output).
If a non-root discussion UUID is passed, the root discussion is automatically resolved.`,
		Example: `  # Resolve a discussion
  gf mr resolve 42 --discussion abc12345

  # Short flag
  gf mr resolve 42 -d abc12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid merge request ID: %s", args[0])
			}
			return runResolve(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVarP(&opts.discussion, "discussion", "d", "", "Discussion UUID to resolve")
	_ = cmd.MarkFlagRequired("discussion")

	return cmd
}

func runResolve(opts *resolveOptions, id int) error {
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

	// Resolve discussion
	_, err = client.MergeRequests().ResolveDiscussion(repo.Owner, repo.Name, id, opts.discussion)
	if err != nil {
		return fmt.Errorf("failed to resolve discussion: %w", err)
	}

	fmt.Printf("✓ Resolved discussion on MR #%d\n", mr.LocalID)
	return nil
}
