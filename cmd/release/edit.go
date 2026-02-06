package release

import (
	"fmt"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type editOptions struct {
	repo         string
	title        string
	description  string
	draft        bool
	noDraft      bool
	prerelease   bool
	noPrerelease bool
}

func newEditCmd() *cobra.Command {
	opts := &editOptions{}

	cmd := &cobra.Command{
		Use:   "edit <tag>",
		Short: "Edit a release",
		Long: `Edit an existing release.

You can update the title, description, draft status, and prerelease status.`,
		Example: `  # Edit release title
  gf release edit v1.0.0 --title "Version 1.0.0 - Stable"

  # Mark as prerelease
  gf release edit v1.0.0 --prerelease

  # Remove draft status
  gf release edit v1.0.0 --no-draft`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEdit(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Release title")
	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Release description")
	cmd.Flags().BoolVar(&opts.draft, "draft", false, "Mark as draft")
	cmd.Flags().BoolVar(&opts.noDraft, "no-draft", false, "Remove draft status")
	cmd.Flags().BoolVar(&opts.prerelease, "prerelease", false, "Mark as prerelease")
	cmd.Flags().BoolVar(&opts.noPrerelease, "no-prerelease", false, "Remove prerelease status")

	return cmd
}

func runEdit(opts *editOptions, tagName string) error {
	// Check conflicting flags
	if opts.draft && opts.noDraft {
		return fmt.Errorf("cannot use both --draft and --no-draft")
	}
	if opts.prerelease && opts.noPrerelease {
		return fmt.Errorf("cannot use both --prerelease and --no-prerelease")
	}

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

	// Check if release exists
	_, err = client.Releases().Get(repo.Owner, repo.Name, tagName)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("release %q not found in %s", tagName, repo.FullName())
		}
		return fmt.Errorf("failed to get release: %w", err)
	}

	// Build update request - tagName is required by API
	req := &api.UpdateReleaseRequest{
		TagName: tagName, // Required by GitFlic API
	}
	hasChanges := false

	if opts.title != "" {
		req.Title = opts.title
		hasChanges = true
	}
	if opts.description != "" {
		req.Description = opts.description
		hasChanges = true
	}
	if opts.draft {
		val := true
		req.IsDraft = &val
		hasChanges = true
	}
	if opts.noDraft {
		val := false
		req.IsDraft = &val
		hasChanges = true
	}
	if opts.prerelease {
		val := true
		req.IsPrerelease = &val
		hasChanges = true
	}
	if opts.noPrerelease {
		val := false
		req.IsPrerelease = &val
		hasChanges = true
	}

	if !hasChanges {
		return fmt.Errorf("no changes specified. Use --title, --description, --draft, --no-draft, --prerelease, or --no-prerelease")
	}

	// Update release
	release, err := client.Releases().Update(repo.Owner, repo.Name, tagName, req)
	if err != nil {
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to edit releases in %s", repo.FullName())
		}
		return fmt.Errorf("failed to update release: %w", err)
	}

	fmt.Printf("✓ Updated release %q\n", release.TagName)
	return nil
}
