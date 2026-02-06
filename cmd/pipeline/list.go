package pipeline

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
	limit int
	repo  string
	json  bool
}

func newListCmd() *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pipelines",
		Long:  `List CI/CD pipelines in the current repository.`,
		Example: `  # List recent pipelines
  gf pipeline list

  # List with limit
  gf pipeline list --limit 10

  # Output as JSON
  gf pipeline list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts)
		},
	}

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

	// Fetch pipelines
	pipelines, err := client.Pipelines().List(repo.Owner, repo.Name)
	if err != nil {
		return fmt.Errorf("failed to list pipelines: %w", err)
	}

	if len(pipelines) == 0 {
		fmt.Printf("No pipelines in %s\n", repo.FullName())
		return nil
	}

	// Limit results
	if opts.limit > 0 && len(pipelines) > opts.limit {
		pipelines = pipelines[:opts.limit]
	}

	// JSON output
	if opts.json {
		data, err := json.MarshalIndent(pipelines, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Print table
	fmt.Printf("\n%-6s %-10s %-25s %-10s %-10s %s\n", "ID", "STATUS", "BRANCH", "SHA", "DURATION", "UPDATED")
	fmt.Println(strings.Repeat("-", 80))

	for _, p := range pipelines {
		color := api.StatusColor(p.Status)
		reset := api.ColorReset()
		status := fmt.Sprintf("%s%s %-7s%s", color, api.StatusIcon(p.Status), p.NormalizedStatus(), reset)

		branch := p.Ref
		if len(branch) > 22 {
			branch = branch[:22] + "..."
		}

		duration := output.FormatDuration(p.Duration)
		updated := output.FormatRelativeTime(p.CreatedAt.Time)

		fmt.Printf("#%-5d %-10s %-25s %-10s %-10s %s\n",
			p.LocalID,
			status,
			branch,
			p.SHA(),
			duration,
			updated,
		)
	}

	return nil
}
