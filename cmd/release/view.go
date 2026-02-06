package release

import (
	"encoding/json"
	"fmt"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/browser"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/josinSbazin/gf/internal/output"
	"github.com/spf13/cobra"
)

type viewOptions struct {
	repo string
	json bool
	web  bool
}

func newViewCmd() *cobra.Command {
	opts := &viewOptions{}

	cmd := &cobra.Command{
		Use:   "view <tag>",
		Short: "View a release",
		Long:  `View details of a specific release.`,
		Example: `  # View release v1.0.0
  gf release view v1.0.0

  # View release as JSON
  gf release view v1.0.0 --json

  # Open in browser
  gf release view v1.0.0 --web`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output as JSON")
	cmd.Flags().BoolVarP(&opts.web, "web", "w", false, "Open in browser")

	return cmd
}

func runView(opts *viewOptions, tagName string) error {
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

	// Fetch release
	release, err := client.Releases().Get(repo.Owner, repo.Name, tagName)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("release '%s' not found in %s", tagName, repo.FullName())
		}
		return fmt.Errorf("failed to get release: %w", err)
	}

	// JSON output
	if opts.json {
		data, err := json.MarshalIndent(release, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Print release details
	fmt.Println()
	fmt.Printf("%s\n", release.Title)
	fmt.Printf("Tag: %s\n", release.TagName)

	releaseType := "Release"
	if release.IsDraft {
		releaseType = "Draft"
	} else if release.IsPrerelease {
		releaseType = "Pre-release"
	}
	fmt.Printf("Type: %s\n", releaseType)

	if !release.PublishedAt.IsZero() {
		fmt.Printf("Published: %s\n", output.FormatRelativeTime(release.PublishedAt))
	} else {
		fmt.Printf("Created: %s\n", output.FormatRelativeTime(release.CreatedAt))
	}

	if release.Author.Username != "" {
		fmt.Printf("Author: @%s\n", release.Author.Username)
	}

	if release.Description != "" {
		fmt.Println()
		fmt.Println("---")
		fmt.Println(release.Description)
	}

	fmt.Println()

	// URL - GitFlic uses release UUID in web URLs, not tag name
	releaseURL := fmt.Sprintf("https://%s/project/%s/%s/release/%s",
		repo.Host, repo.Owner, repo.Name, release.ID)

	if opts.web {
		fmt.Printf("Opening %s in browser...\n", releaseURL)
		return browser.Open(releaseURL)
	}

	fmt.Printf("View in browser: %s\n", releaseURL)

	return nil
}
