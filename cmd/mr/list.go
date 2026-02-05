package mr

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

type listOptions struct {
	state  string
	limit  int
	repo   string
	json   bool
}

func newListCmd() *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List merge requests",
		Long:  `List merge requests in the current repository.`,
		Example: `  # List open merge requests
  gf mr list

  # List all merge requests
  gf mr list --state all

  # List merged merge requests
  gf mr list --state merged`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.state, "state", "s", "open", "Filter by state: open, merged, closed, all")
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

	// Fetch merge requests
	mrs, err := client.MergeRequests().List(repo.Owner, repo.Name, &api.MRListOptions{
		State: opts.state,
	})
	if err != nil {
		return fmt.Errorf("failed to list merge requests: %w", err)
	}

	// Apply limit
	if opts.limit > 0 && len(mrs) > opts.limit {
		mrs = mrs[:opts.limit]
	}

	if len(mrs) == 0 {
		if opts.json {
			fmt.Println("[]")
			return nil
		}
		fmt.Printf("No %s merge requests in %s\n", opts.state, repo.FullName())
		return nil
	}

	// JSON output
	if opts.json {
		data, err := json.MarshalIndent(mrs, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Print header
	fmt.Printf("\nShowing %d merge requests in %s\n\n", len(mrs), repo.FullName())

	// Print table
	fmt.Printf("%-6s %-50s %-20s %-12s %s\n", "ID", "TITLE", "BRANCH", "AUTHOR", "UPDATED")
	fmt.Println(strings.Repeat("-", 100))

	for _, mr := range mrs {
		title := mr.Title
		if len(title) > 47 {
			title = title[:47] + "..."
		}

		branch := mr.SourceBranch.Title
		if len(branch) > 17 {
			branch = branch[:17] + "..."
		}

		updated := output.FormatRelativeTime(mr.UpdatedAt)

		fmt.Printf("#%-5d %-50s %-20s @%-11s %s\n",
			mr.LocalID,
			title,
			branch,
			mr.Author.Username,
			updated,
		)
	}

	return nil
}
