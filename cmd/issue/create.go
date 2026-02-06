package issue

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type createOptions struct {
	repo        string
	title       string
	description string
	quiet       bool
}

func newCreateCmd() *cobra.Command {
	opts := &createOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue",
		Long:  `Create a new issue in the repository.`,
		Example: `  # Create issue interactively
  gf issue create

  # Create issue with title
  gf issue create --title "Bug: login fails"

  # Create issue with title and description
  gf issue create --title "Feature request" --body "Add dark mode support"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Issue title")
	cmd.Flags().StringVarP(&opts.description, "body", "b", "", "Issue description (required by GitFlic, auto-filled if empty)")
	cmd.Flags().BoolVarP(&opts.quiet, "quiet", "q", false, "Output only the issue number")

	return cmd
}

func runCreate(opts *createOptions) error {
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

	// Interactive mode if title not provided
	title := opts.title
	description := opts.description

	if title == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Issue title: ")
		title, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read title: %w", err)
		}
		title = strings.TrimSpace(title)

		if title == "" {
			return fmt.Errorf("title cannot be empty")
		}

		if description == "" {
			fmt.Print("Description (optional, press Enter to skip): ")
			description, _ = reader.ReadString('\n')
			description = strings.TrimSpace(description)
		}
	}

	// GitFlic requires non-empty description
	if description == "" {
		description = "No description provided"
	}

	// Create issue
	issue, err := client.Issues().Create(repo.Owner, repo.Name, &api.CreateIssueRequest{
		Title:       title,
		Description: description,
	})
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	// Quiet mode - output only ID
	if opts.quiet {
		fmt.Printf("%d\n", issue.LocalID)
		return nil
	}

	fmt.Printf("Created issue #%d: %s\n", issue.LocalID, issue.Title)
	fmt.Printf("View at: https://%s/project/%s/%s/issue/%d\n",
		cfg.ActiveHost, repo.Owner, repo.Name, issue.LocalID)

	return nil
}
