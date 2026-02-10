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

type replyOptions struct {
	repo       string
	body       string
	discussion string
}

func newReplyCmd() *cobra.Command {
	opts := &replyOptions{}

	cmd := &cobra.Command{
		Use:   "reply <mr-id>",
		Short: "Reply to a discussion on a merge request",
		Long: `Reply to an existing discussion thread on a merge request.

Use --discussion to specify the discussion UUID (shown in 'gf mr comments' output).`,
		Example: `  # Reply to a discussion
  gf mr reply 42 --discussion abc12345 --body "Fixed in latest commit"

  # Pipe reply from stdin
  echo "Done" | gf mr reply 42 --discussion abc12345 --body -`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid merge request ID: %s", args[0])
			}
			return runReply(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVarP(&opts.body, "body", "b", "", "Reply body (use - to read from stdin)")
	cmd.Flags().StringVarP(&opts.discussion, "discussion", "d", "", "Discussion UUID to reply to")
	_ = cmd.MarkFlagRequired("discussion")

	return cmd
}

func runReply(opts *replyOptions, id int) error {
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

	// Get reply body
	body := opts.body
	if body == "-" {
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		body = strings.Join(lines, "\n")
	} else if body == "" {
		fmt.Printf("Replying to discussion on MR #%d: %s\n\n", mr.LocalID, mr.Title)

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Reply: ")
		body, _ = reader.ReadString('\n')
		body = strings.TrimSpace(body)
	}

	if body == "" {
		return fmt.Errorf("reply body cannot be empty")
	}

	// Reply to discussion
	_, err = client.MergeRequests().ReplyDiscussion(repo.Owner, repo.Name, id, &api.ReplyDiscussionRequest{
		DiscussionUUID: opts.discussion,
		Message:        body,
	})
	if err != nil {
		return fmt.Errorf("failed to reply: %w", err)
	}

	fmt.Printf("✓ Replied to discussion on MR #%d\n", mr.LocalID)
	return nil
}
