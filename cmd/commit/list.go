package commit

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
	repo   string
	ref    string
	limit  int
	json   bool
}

func newListCmd() *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List commits",
		Long:  `List commits in the repository.`,
		Example: `  # List commits on default branch
  gf commit list

  # List commits on specific branch
  gf commit list --ref develop

  # List with limit
  gf commit list --limit 10

  # Output as JSON
  gf commit list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVar(&opts.ref, "ref", "", "Branch or tag name")
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

	// Fetch commits
	commits, err := client.Commits().List(repo.Owner, repo.Name, &api.CommitListOptions{
		Ref:     opts.ref,
		PerPage: opts.limit,
	})
	if err != nil {
		return fmt.Errorf("failed to list commits: %w", err)
	}

	if len(commits) == 0 {
		fmt.Printf("No commits in %s\n", repo.FullName())
		return nil
	}

	// Limit results
	if opts.limit > 0 && len(commits) > opts.limit {
		commits = commits[:opts.limit]
	}

	// JSON output
	if opts.json {
		data, err := json.MarshalIndent(commits, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Print table
	fmt.Printf("\n%-10s %-20s %-40s %s\n", "HASH", "AUTHOR", "MESSAGE", "DATE")
	fmt.Println(strings.Repeat("-", 90))

	for _, c := range commits {
		hash := c.Hash
		if len(hash) > 7 {
			hash = hash[:7]
		}

		author := c.AuthorName
		if len(author) > 18 {
			author = author[:18] + "..."
		}

		message := strings.Split(c.Message, "\n")[0] // First line only
		if len(message) > 38 {
			message = message[:38] + "..."
		}

		date := output.FormatRelativeTime(c.CreatedAt)

		fmt.Printf("%-10s %-20s %-40s %s\n", hash, author, message, date)
	}

	return nil
}
