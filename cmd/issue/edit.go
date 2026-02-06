package issue

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type editOptions struct {
	repo        string
	title       string
	description string
}

func newEditCmd() *cobra.Command {
	opts := &editOptions{}

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit an issue",
		Long: `Edit an existing issue.

You can update the title and description.`,
		Example: `  # Edit issue title
  gf issue edit 42 --title "New title"

  # Edit issue description
  gf issue edit 42 --description "Updated description"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid issue ID: %s", args[0])
			}
			return runEdit(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Issue title")
	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Issue description")

	return cmd
}

func runEdit(opts *editOptions, id int) error {
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

	// Check if issue exists
	_, err = client.Issues().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("issue #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to get issue: %w", err)
	}

	// Build update request
	req := &api.UpdateIssueRequest{}
	hasChanges := false

	if opts.title != "" {
		req.Title = opts.title
		hasChanges = true
	}
	if opts.description != "" {
		req.Description = opts.description
		hasChanges = true
	}

	if !hasChanges {
		return fmt.Errorf("no changes specified. Use --title or --description")
	}

	// Update issue
	issue, err := client.Issues().Update(repo.Owner, repo.Name, id, req)
	if err != nil {
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to edit issues in %s", repo.FullName())
		}
		return fmt.Errorf("failed to update issue: %w", err)
	}

	fmt.Printf("✓ Updated issue #%d\n", issue.LocalID)
	return nil
}
