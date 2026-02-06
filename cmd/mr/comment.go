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
	"github.com/josinSbazin/gf/internal/output"
	"github.com/spf13/cobra"
)

type commentOptions struct {
	repo string
	body string
}

func newCommentCmd() *cobra.Command {
	opts := &commentOptions{}

	cmd := &cobra.Command{
		Use:   "comment <id>",
		Short: "Add a comment to a merge request",
		Long: `Add a comment to a merge request.

Without --body flag, opens an interactive prompt for the comment text.`,
		Example: `  # Add comment interactively
  gf mr comment 42

  # Add comment with body
  gf mr comment 42 --body "Looks good to me!"

  # Pipe comment from stdin
  echo "LGTM" | gf mr comment 42 --body -`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid merge request ID: %s", args[0])
			}
			return runComment(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVarP(&opts.body, "body", "b", "", "Comment body (use - to read from stdin)")

	return cmd
}

func runComment(opts *commentOptions, id int) error {
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

	// Get comment body
	body := opts.body
	if body == "-" {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		body = strings.Join(lines, "\n")
	} else if body == "" {
		// Interactive mode
		fmt.Printf("Adding comment to MR #%d: %s\n\n", mr.LocalID, mr.Title)

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Comment: ")
		body, _ = reader.ReadString('\n')
		body = strings.TrimSpace(body)
	}

	if body == "" {
		return fmt.Errorf("comment body cannot be empty")
	}

	// Create discussion
	_, err = client.MergeRequests().CreateDiscussion(repo.Owner, repo.Name, id, &api.CreateDiscussionRequest{
		Message: body,
	})
	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	fmt.Printf("✓ Added comment to MR #%d\n", mr.LocalID)
	return nil
}

func newCommentsCmd() *cobra.Command {
	opts := &struct {
		repo string
	}{}

	cmd := &cobra.Command{
		Use:     "comments <id>",
		Aliases: []string{"discussions"},
		Short:   "List comments on a merge request",
		Long:    `List all comments and discussions on a merge request.`,
		Example: `  # List comments
  gf mr comments 42`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid merge request ID: %s", args[0])
			}
			return runComments(opts.repo, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runComments(repoFlag string, id int) error {
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

	// Get MR info first
	mr, err := client.MergeRequests().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("merge request #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to get merge request: %w", err)
	}

	// Get discussions
	discussions, err := client.MergeRequests().ListDiscussions(repo.Owner, repo.Name, id)
	if err != nil {
		return fmt.Errorf("failed to list comments: %w", err)
	}

	if len(discussions) == 0 {
		fmt.Printf("No comments on MR #%d: %s\n", mr.LocalID, mr.Title)
		return nil
	}

	fmt.Printf("\nComments on MR #%d: %s\n", mr.LocalID, mr.Title)
	fmt.Println(strings.Repeat("─", 60))

	for _, d := range discussions {
		resolved := ""
		if d.Resolved {
			resolved = " [resolved]"
		}

		fmt.Printf("\n@%s • %s%s\n", d.Author.Username, output.FormatRelativeTime(d.CreatedAt), resolved)
		fmt.Printf("%s\n", d.Message)
	}

	fmt.Println()
	return nil
}
