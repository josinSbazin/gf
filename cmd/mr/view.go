package mr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/browser"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type viewOptions struct {
	repo string
	web  bool
}

func newViewCmd() *cobra.Command {
	opts := &viewOptions{}

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a merge request",
		Long:  `View details of a merge request.`,
		Example: `  # View merge request #12
  gf mr view 12

  # Open in browser
  gf mr view 12 --web`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid merge request ID: %s", args[0])
			}
			return runView(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVarP(&opts.web, "web", "w", false, "Open in browser")

	return cmd
}

func runView(opts *viewOptions, id int) error {
	// Get repository
	var repo *git.Repository
	var err error

	if opts.repo != "" {
		parts := strings.Split(opts.repo, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format, expected owner/name")
		}
		repo = &git.Repository{
			Host:  config.DefaultHost(),
			Owner: parts[0],
			Name:  parts[1],
		}
	} else {
		repo, err = git.DetectRepo()
		if err != nil {
			return fmt.Errorf("could not determine repository: %w", err)
		}
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

	// Fetch merge request
	mr, err := client.MergeRequests().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("merge request #%d not found", id)
		}
		return fmt.Errorf("failed to get merge request: %w", err)
	}

	// Print details
	stateIcon := "○"
	stateText := mr.State()
	switch mr.State() {
	case "open":
		stateIcon = "●"
		stateText = "Open"
	case "merged":
		stateIcon = "✓"
		stateText = "Merged"
	case "closed":
		stateIcon = "✗"
		stateText = "Closed"
	}

	fmt.Printf("\n%s %s #%d\n", mr.Title, stateIcon, mr.LocalID)
	fmt.Printf("%s • @%s wants to merge %s into %s\n\n",
		stateText, mr.Author.Username, mr.SourceBranch.Title, mr.TargetBranch.Title)

	if mr.Description != "" {
		fmt.Println(mr.Description)
		fmt.Println()
	}

	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	if mr.HasConflicts {
		fmt.Println("⚠ This merge request has conflicts")
	}

	fmt.Printf("Created:  %s\n", formatRelativeTime(mr.CreatedAt))
	fmt.Printf("Updated:  %s\n", formatRelativeTime(mr.UpdatedAt))

	fmt.Println()

	// URL
	url := fmt.Sprintf("https://%s/project/%s/%s/merge-request/%d",
		repo.Host, repo.Owner, repo.Name, mr.LocalID)

	if opts.web {
		fmt.Printf("Opening %s in browser...\n", url)
		return browser.Open(url)
	}

	fmt.Printf("View in browser: %s\n", url)

	return nil
}
