package webhook

import (
	"fmt"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "test <id>",
		Short: "Test a webhook",
		Long:  `Send a test payload to a webhook endpoint.`,
		Example: `  # Test webhook
  gf webhook test abc123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTest(repo, args[0])
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runTest(repoFlag string, webhookID string) error {
	// Get repository
	repo, err := git.ResolveRepo(repoFlag, config.DefaultHost())
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

	// Get webhook to show URL
	webhook, err := client.Webhooks().Get(repo.Owner, repo.Name, webhookID)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("webhook %q not found in %s", webhookID, repo.FullName())
		}
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	// Test webhook
	fmt.Printf("Sending test payload to %s...\n", webhook.URL)
	err = client.Webhooks().Test(repo.Owner, repo.Name, webhookID)
	if err != nil {
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to test webhooks in %s", repo.FullName())
		}
		return fmt.Errorf("failed to test webhook: %w", err)
	}

	fmt.Printf("✓ Test payload sent to webhook\n")
	return nil
}
