package release

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
	maxTitleLen = 40
	tableWidth  = 90
)

type listOptions struct {
	repo  string
	limit int
	json  bool
}

func newListCmd() *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List releases",
		Long:  `List releases in the current repository.`,
		Example: `  # List releases
  gf release list

  # List releases for a specific repo
  gf release list --repo owner/name`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().IntVarP(&opts.limit, "limit", "L", 30, "Maximum number of results")
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

	// Fetch releases
	releases, total, err := client.Releases().List(repo.Owner, repo.Name, nil)
	if err != nil {
		if api.IsNotFound(err) {
			fmt.Printf("No releases in %s\n", repo.FullName())
			return nil
		}
		return fmt.Errorf("failed to list releases: %w", err)
	}

	// Apply limit
	if opts.limit > 0 && len(releases) > opts.limit {
		releases = releases[:opts.limit]
	}

	if len(releases) == 0 {
		if opts.json {
			fmt.Println("[]")
			return nil
		}
		fmt.Printf("No releases in %s\n", repo.FullName())
		return nil
	}

	// JSON output
	if opts.json {
		data, err := json.MarshalIndent(releases, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Print header
	fmt.Printf("\nShowing %d of %d releases in %s\n\n", len(releases), total, repo.FullName())

	// Print table
	fmt.Printf("%-15s %-42s %-10s %s\n", "TAG", "TITLE", "TYPE", "PUBLISHED")
	fmt.Println(strings.Repeat("-", tableWidth))

	for _, rel := range releases {
		title := rel.Title
		if len(title) > maxTitleLen {
			title = title[:maxTitleLen] + "..."
		}

		releaseType := "release"
		if rel.IsDraft {
			releaseType = "draft"
		} else if rel.IsPrerelease {
			releaseType = "pre"
		}

		published := output.FormatRelativeTime(rel.PublishedAt)
		if rel.PublishedAt.IsZero() {
			published = output.FormatRelativeTime(rel.CreatedAt)
		}

		fmt.Printf("%-15s %-42s %-10s %s\n",
			rel.TagName,
			title,
			releaseType,
			published,
		)
	}

	return nil
}
