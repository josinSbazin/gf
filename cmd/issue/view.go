package issue

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/browser"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/josinSbazin/gf/internal/output"
	"github.com/spf13/cobra"
)

type viewOptions struct {
	repo string
	json bool
	web  bool
}

func newViewCmd() *cobra.Command {
	opts := &viewOptions{}

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View an issue",
		Long:  `Display the details of an issue.`,
		Example: `  # View issue #42
  gf issue view 42

  # View issue in JSON format
  gf issue view 42 --json

  # Open in browser
  gf issue view 42 --web`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Support both "42" and "#42" formats
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid issue ID: %s", args[0])
			}
			return runView(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output as JSON")
	cmd.Flags().BoolVarP(&opts.web, "web", "w", false, "Open in browser")

	return cmd
}

func runView(opts *viewOptions, id int) error {
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

	// Fetch issue
	issue, err := client.Issues().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("issue #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to get issue: %w", err)
	}

	// JSON output
	if opts.json {
		data, err := json.MarshalIndent(issue, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Print issue details
	fmt.Printf("\n#%d %s\n", issue.LocalID, issue.Title)
	fmt.Printf("Status: %s\n", issue.Status.Title)
	fmt.Printf("Author: @%s\n", issue.Author.Username)
	fmt.Printf("Created: %s\n", output.FormatRelativeTime(issue.CreatedAt.Time))
	fmt.Printf("Updated: %s\n", output.FormatRelativeTime(issue.UpdatedAt.Time))

	if issue.Description != "" {
		fmt.Printf("\n--- Description ---\n%s\n", issue.Description)
	}

	fmt.Println()

	// URL
	issueURL := fmt.Sprintf("https://%s/project/%s/%s/issue/%d",
		repo.Host, repo.Owner, repo.Name, issue.LocalID)

	if opts.web {
		fmt.Printf("Opening %s in browser...\n", issueURL)
		return browser.Open(issueURL)
	}

	fmt.Printf("View in browser: %s\n", issueURL)

	return nil
}
