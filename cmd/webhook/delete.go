package webhook

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	repo  string
	force bool
}

func newDeleteCmd() *cobra.Command {
	opts := &deleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a webhook",
		Long: `Delete a webhook from the repository.

By default, asks for confirmation before deleting.
Use --force to skip confirmation.`,
		Example: `  # Delete webhook (with confirmation)
  gf webhook delete abc123

  # Delete webhook without confirmation
  gf webhook delete abc123 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runDelete(opts *deleteOptions, webhookID string) error {
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

	// Check if webhook exists
	webhook, err := client.Webhooks().Get(repo.Owner, repo.Name, webhookID)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("webhook %q not found in %s", webhookID, repo.FullName())
		}
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	// Confirm deletion
	if !opts.force {
		fmt.Printf("Are you sure you want to delete webhook for %s? [y/N]: ", webhook.URL)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Delete webhook
	err = client.Webhooks().Delete(repo.Owner, repo.Name, webhookID)
	if err != nil {
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to delete webhooks in %s", repo.FullName())
		}
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	fmt.Printf("✓ Deleted webhook %s\n", webhookID)
	return nil
}
