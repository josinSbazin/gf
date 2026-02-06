package webhook

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type createOptions struct {
	repo   string
	events []string
	secret string
}

// Available webhook events (short names for CLI)
var availableEvents = []string{
	"push",
	"merge_request",
	"issue",
	"release",
	"pipeline",
	"tag",
	"branch",
	"collaborator",
	"discussion",
}

func newCreateCmd() *cobra.Command {
	opts := &createOptions{}

	cmd := &cobra.Command{
		Use:   "create <url>",
		Short: "Create a webhook",
		Long: fmt.Sprintf(`Create a new webhook in the repository.

Available events: %s

If --secret is not provided, a random secret will be generated.`, strings.Join(availableEvents, ", ")),
		Example: `  # Create webhook for push events
  gf webhook create https://example.com/webhook --events push

  # Create webhook for multiple events
  gf webhook create https://example.com/webhook --events push,merge_request,pipeline

  # Create webhook with secret
  gf webhook create https://example.com/webhook --events push --secret mysecret`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringSliceVarP(&opts.events, "events", "e", []string{"push"}, "Events to trigger webhook (comma-separated)")
	cmd.Flags().StringVarP(&opts.secret, "secret", "s", "", "Webhook secret for signature verification (auto-generated if not provided)")

	return cmd
}

func runCreate(opts *createOptions, webhookURL string) error {
	// Validate URL
	parsedURL, err := url.Parse(webhookURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return fmt.Errorf("invalid URL: must be http or https")
	}

	// Security warning: internal/private URLs may indicate SSRF risk
	if config.IsInternalHost(parsedURL.Host) {
		fmt.Fprintf(os.Stderr, "⚠ Warning: webhook URL points to an internal/private address (%s)\n", parsedURL.Host)
	}

	// Validate events
	for _, event := range opts.events {
		if !isValidEvent(event) {
			return fmt.Errorf("invalid event: %q\nAvailable events: %s", event, strings.Join(availableEvents, ", "))
		}
	}

	// Generate secret if not provided
	secret := opts.secret
	if secret == "" {
		secret = generateSecret()
	}

	// Build events object
	events := buildEventsObject(opts.events)

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

	// Create webhook
	webhook, err := client.Webhooks().Create(repo.Owner, repo.Name, &api.CreateWebhookRequest{
		URL:    webhookURL,
		Secret: secret,
		Events: events,
	})
	if err != nil {
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to create webhooks in %s", repo.FullName())
		}
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	fmt.Printf("✓ Created webhook %s\n", webhook.ID)
	fmt.Printf("  URL: %s\n", webhook.URL)
	if opts.secret == "" {
		fmt.Printf("  Secret: %s\n", secret)
	}
	fmt.Printf("  Events: %s\n", strings.Join(opts.events, ", "))
	return nil
}

func isValidEvent(event string) bool {
	for _, e := range availableEvents {
		if strings.EqualFold(event, e) {
			return true
		}
	}
	return false
}

// generateSecret creates a random 32-character hex secret
func generateSecret() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simple default if crypto/rand fails
		return "gf-webhook-secret"
	}
	return hex.EncodeToString(bytes)
}

// buildEventsObject converts CLI event names to GitFlic API format
func buildEventsObject(events []string) *api.WebhookEvents {
	e := &api.WebhookEvents{}
	for _, event := range events {
		switch strings.ToLower(event) {
		case "push":
			e.Push = true
		case "merge_request":
			e.MergeRequestCreate = true
			e.MergeRequestUpdate = true
			e.Merge = true
		case "issue":
			e.IssueCreate = true
			e.IssueUpdate = true
			e.NewIssueNote = true
		case "release":
			e.ReleaseCreate = true
			e.ReleaseUpdate = true
			e.ReleaseDelete = true
		case "pipeline":
			e.PipelineNew = true
			e.PipelineSuccess = true
			e.PipelineFail = true
		case "tag":
			e.TagCreate = true
			e.TagDelete = true
		case "branch":
			e.BranchCreate = true
			e.BranchUpdate = true
			e.BranchDelete = true
		case "collaborator":
			e.CollaboratorAdd = true
			e.CollaboratorDelete = true
		case "discussion":
			e.DiscussionCreate = true
		}
	}
	return e
}
