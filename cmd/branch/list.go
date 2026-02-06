package branch

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
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
		Short: "List branches",
		Long:  `List all branches in the repository.`,
		Example: `  # List branches
  gf branch list

  # List branches in specific repo
  gf branch list -R owner/repo

  # Output as JSON
  gf branch list --json`,
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

	// Fetch branches
	branches, err := client.Branches().List(repo.Owner, repo.Name)
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	if len(branches) == 0 {
		fmt.Printf("No branches in %s\n", repo.FullName())
		return nil
	}

	// JSON output
	if opts.json {
		data, err := json.MarshalIndent(branches, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Print table
	fmt.Printf("\n%-40s %-10s %s\n", "BRANCH", "DEFAULT", "COMMIT")
	fmt.Println(strings.Repeat("-", 70))

	for _, b := range branches {
		name := b.Name
		if len(name) > 38 {
			name = name[:38] + "..."
		}

		defaultMark := ""
		if b.IsDefault {
			defaultMark = "*"
		}

		hash := b.Hash
		// If Hash is empty, try to get from LastCommit
		if hash == "" && b.LastCommit != nil {
			hash = b.LastCommit.Hash
		}
		if len(hash) > 7 {
			hash = hash[:7]
		}

		fmt.Printf("%-40s %-10s %s\n", name, defaultMark, hash)
	}

	return nil
}
