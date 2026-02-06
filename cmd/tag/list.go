package tag

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
	repo string
	json bool
}

func newListCmd() *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tags",
		Long:  `List all tags in the repository.`,
		Example: `  # List tags
  gf tag list

  # List tags in specific repo
  gf tag list -R owner/repo

  # Output as JSON
  gf tag list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts)
		},
	}

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

	// Fetch tags
	tags, err := client.Tags().List(repo.Owner, repo.Name)
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	if len(tags) == 0 {
		fmt.Printf("No tags in %s\n", repo.FullName())
		return nil
	}

	// JSON output
	if opts.json {
		data, err := json.MarshalIndent(tags, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Print table
	fmt.Printf("\n%-30s %-10s %s\n", "TAG", "COMMIT", "DATE")
	fmt.Println(strings.Repeat("-", 60))

	for _, t := range tags {
		name := t.Name
		if len(name) > 28 {
			name = name[:28] + "..."
		}

		// Use ObjectID or CommitID for the hash
		hash := t.ObjectID
		if hash == "" {
			hash = t.CommitID
		}
		if len(hash) > 7 {
			hash = hash[:7]
		}

		// Get date from PersonIdent or CreatedAt
		date := ""
		if t.PersonIdent != nil && !t.PersonIdent.When.IsZero() {
			date = output.FormatRelativeTime(t.PersonIdent.When)
		} else if !t.CreatedAt.IsZero() {
			date = output.FormatRelativeTime(t.CreatedAt)
		}

		fmt.Printf("%-30s %-10s %s\n", name, hash, date)
	}

	return nil
}
