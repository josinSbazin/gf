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
	fmt.Printf("\n%-36s %-50s %s\n", "ID", "URL", "EVENTS")
	fmt.Println(strings.Repeat("-", 100))

	for _, w := range webhooks {
		urlStr := w.URL
		if len(urlStr) > 48 {
			urlStr = urlStr[:48] + "..."
		}

		events := eventsToString(w.Events)
		if len(events) > 30 {
			events = events[:30] + "..."
		}

		fmt.Printf("%-36s %-50s %s\n", w.ID, urlStr, events)
	}

	return nil
}

// eventsToString converts WebhookEvents to a readable string
func eventsToString(e *api.WebhookEvents) string {
	if e == nil {
		return ""
	}
	var events []string
	if e.Push {
		events = append(events, "push")
	}
	if e.MergeRequestCreate || e.MergeRequestUpdate || e.Merge {
		events = append(events, "merge_request")
	}
	if e.IssueCreate || e.IssueUpdate {
		events = append(events, "issue")
	}
	if e.ReleaseCreate || e.ReleaseUpdate || e.ReleaseDelete {
		events = append(events, "release")
	}
	if e.PipelineNew || e.PipelineSuccess || e.PipelineFail {
		events = append(events, "pipeline")
	}
	if e.TagCreate || e.TagDelete {
		events = append(events, "tag")
	}
	if e.BranchCreate || e.BranchUpdate || e.BranchDelete {
		events = append(events, "branch")
	}
	if e.CollaboratorAdd || e.CollaboratorDelete {
		events = append(events, "collaborator")
	}
	if e.DiscussionCreate {
		events = append(events, "discussion")
	}
	return strings.Join(events, ", ")
}
