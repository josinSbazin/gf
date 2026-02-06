package issue

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/josinSbazin/gf/internal/output"
	"github.com/spf13/cobra"
)

const (
	maxTitleLen = 55
	tableWidth  = 100
)

type listOptions struct {
	state string
	limit int
	repo  string
	json  bool
}

func newListCmd() *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues",
		Long:  `List issues in the current repository.`,
		Example: `  # List open issues
  gf issue list

  # List all issues
  gf issue list --state all

  # List closed issues
  gf issue list --state closed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.state, "state", "s", "open", "Filter by state: open, closed, all")
	cmd.Flags().IntVarP(&opts.limit, "limit", "L", 30, "Maximum number of results")
	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output as JSON")

	return cmd
}

func runList(opts *listOptions) error {
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

	// Fetch issues
	issues, err := client.Issues().List(repo.Owner, repo.Name, &api.IssueListOptions{
		State: opts.state,
	})
	if err != nil {
		return fmt.Errorf("failed to list issues: %w", err)
	}

	// Apply limit
	if opts.limit > 0 && len(issues) > opts.limit {
		issues = issues[:opts.limit]
	}

	if len(issues) == 0 {
		if opts.json {
			fmt.Println("[]")
			return nil
		}
		fmt.Printf("No %s issues in %s\n", opts.state, repo.FullName())
		return nil
	}

	// JSON output
	if opts.json {
		data, err := json.MarshalIndent(issues, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Print header
	fmt.Printf("\nShowing %d issues in %s\n\n", len(issues), repo.FullName())

	// Print table
	fmt.Printf("%-6s %-58s %-12s %s\n", "ID", "TITLE", "AUTHOR", "UPDATED")
	fmt.Println(strings.Repeat("-", tableWidth))

	for _, issue := range issues {
		title := issue.Title
		if len(title) > maxTitleLen {
			title = title[:maxTitleLen] + "..."
		}

		// State with color
		state := issue.State()
		color := api.IssueStateColor(state)
		reset := api.ColorReset()
		stateIcon := "○"
		if state == "open" {
			stateIcon = "●"
		} else if state == "closed" {
			stateIcon = "✗"
		}

		updated := output.FormatRelativeTime(issue.UpdatedAt.Time)

		fmt.Printf("%s%s%s #%-4d %-56s @%-11s %s\n",
			color, stateIcon, reset,
			issue.LocalID,
			title,
			issue.Author.Username,
			updated,
		)
	}

	return nil
}
