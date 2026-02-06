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
		Short: "Add a comment to an issue",
		Long: `Add a comment to an issue.

Without --body flag, opens an interactive prompt for the comment text.`,
		Example: `  # Add comment interactively
  gf issue comment 42

  # Add comment with body
  gf issue comment 42 --body "Thanks for reporting!"

  # Pipe comment from stdin
  echo "Fixed in v1.2" | gf issue comment 42 --body -`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid issue ID: %s", args[0])
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

	// Get issue info first
	issue, err := client.Issues().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("issue #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to get issue: %w", err)
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
		fmt.Printf("Adding comment to issue #%d: %s\n\n", issue.LocalID, issue.Title)

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Comment: ")
		body, _ = reader.ReadString('\n')
		body = strings.TrimSpace(body)
	}

	if body == "" {
		return fmt.Errorf("comment body cannot be empty")
	}

	// Create comment
	_, err = client.Issues().CreateComment(repo.Owner, repo.Name, id, body)
	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	fmt.Printf("✓ Added comment to issue #%d\n", issue.LocalID)
	return nil
}

func newCommentsCmd() *cobra.Command {
	opts := &struct {
		repo string
	}{}

	cmd := &cobra.Command{
		Use:   "comments <id>",
		Short: "List comments on an issue",
		Long:  `List all comments on an issue.`,
		Example: `  # List comments
  gf issue comments 42`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid issue ID: %s", args[0])
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

	// Get issue info first
	issue, err := client.Issues().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("issue #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to get issue: %w", err)
	}

	// Get comments
	comments, err := client.Issues().ListComments(repo.Owner, repo.Name, id)
	if err != nil {
		return fmt.Errorf("failed to list comments: %w", err)
	}

	if len(comments) == 0 {
		fmt.Printf("No comments on issue #%d: %s\n", issue.LocalID, issue.Title)
		return nil
	}

	fmt.Printf("\nComments on issue #%d: %s\n", issue.LocalID, issue.Title)
	fmt.Println(strings.Repeat("─", 60))

	for _, c := range comments {
		fmt.Printf("\n@%s • %s\n", c.Author.Username, output.FormatRelativeTime(c.CreatedAt.Time))
		fmt.Printf("%s\n", c.Note)
	}

	fmt.Println()
	return nil
}
