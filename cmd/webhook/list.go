package webhook

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
		Short: "List webhooks",
		Long:  `List all webhooks in the repository.`,
		Example: `  # List webhooks
  gf webhook list

  # Output as JSON
  gf webhook list --json`,
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

	// Fetch webhooks
	webhooks, err := client.Webhooks().List(repo.Owner, repo.Name)
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	if len(webhooks) == 0 {
		fmt.Printf("No webhooks in %s\n", repo.FullName())
		return nil
	}

	// JSON output
	if opts.json {
		data, err := json.MarshalIndent(webhooks, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Print table
	fmt.Printf("\n%-36s %-8s %-40s %s\n", "ID", "ACTIVE", "URL", "EVENTS")
	fmt.Println(strings.Repeat("-", 100))

	for _, w := range webhooks {
		active := "✗"
		if w.Active {
			active = "✓"
		}

		urlStr := w.URL
		if len(urlStr) > 38 {
			urlStr = urlStr[:38] + "..."
		}

		events := strings.Join(w.Events, ", ")
		if len(events) > 20 {
			events = events[:20] + "..."
		}

		fmt.Printf("%-36s %-8s %-40s %s\n", w.ID, active, urlStr, events)
	}

	return nil
}
