package mr

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

type reviewOptions struct {
	repo    string
	body    string
	approve bool
}

func newReviewCmd() *cobra.Command {
	opts := &reviewOptions{}

	cmd := &cobra.Command{
		Use:   "review <id>",
		Short: "Review a merge request (approve + comment)",
		Long: `Submit a review on a merge request.

Combines approval and commenting into a single command.
Use --approve to approve the MR along with the comment.`,
		Example: `  # Approve with comment
  gf mr review 42 --approve --body "LGTM!"

  # Just leave a review comment (no approval)
  gf mr review 42 --body "Needs changes, see inline comments"

  # Approve without comment
  gf mr review 42 --approve

  # Pipe review from stdin
  echo "Ship it" | gf mr review 42 --approve --body -`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid merge request ID: %s", args[0])
			}
			return runReview(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVarP(&opts.body, "body", "b", "", "Review comment body (use - to read from stdin)")
	cmd.Flags().BoolVarP(&opts.approve, "approve", "a", false, "Approve the merge request")

	return cmd
}

func runReview(opts *reviewOptions, id int) error {
	if !opts.approve && opts.body == "" {
		return fmt.Errorf("specify --approve and/or --body for review")
	}

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

	// Get body from stdin if needed
	body := opts.body
	if body == "-" {
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		body = strings.Join(lines, "\n")
	}

	// Add comment if body is provided
	if body != "" {
		_, err = client.MergeRequests().CreateDiscussion(repo.Owner, repo.Name, id, &api.CreateDiscussionRequest{
			Message: body,
		})
		if err != nil {
			return fmt.Errorf("failed to add review comment: %w", err)
		}
		fmt.Printf("✓ Added review comment to MR #%d\n", mr.LocalID)
	}

	// Approve if requested
	if opts.approve {
		if err := client.MergeRequests().Approve(repo.Owner, repo.Name, id); err != nil {
			return fmt.Errorf("failed to approve merge request: %w", err)
		}
		fmt.Printf("✓ Approved MR #%d: %s\n", mr.LocalID, mr.Title)
	}

	return nil
}
