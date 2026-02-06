package tag

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

// isHexString checks if a string is a valid hexadecimal string
func isHexString(s string) bool {
	for _, c := range strings.ToLower(s) {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

type createOptions struct {
	repo    string
	ref     string
	message string
}

// tagNameRegex validates tag names
var tagNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9/_.-]*$`)

// pathTraversalRegex detects path traversal patterns (pre-compiled for performance)
var pathTraversalRegex = regexp.MustCompile(`\.\.`)

func newCreateCmd() *cobra.Command {
	opts := &createOptions{}

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new tag",
		Long: `Create a new tag in the repository.

The tag is created at the specified ref (branch name or commit hash).
If no ref is specified, the default branch HEAD is used.

Use --message to create an annotated tag.`,
		Example: `  # Create lightweight tag at default branch
  gf tag create v1.0.0

  # Create tag at specific branch
  gf tag create v1.0.0 --ref main

  # Create annotated tag with message
  gf tag create v1.0.0 --message "Release version 1.0.0"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVar(&opts.ref, "ref", "", "Branch or commit to tag (default: default branch)")
	cmd.Flags().StringVarP(&opts.message, "message", "m", "", "Tag message (for annotated tag)")

	return cmd
}

func runCreate(opts *createOptions, name string) error {
	// Validate tag name
	if !tagNameRegex.MatchString(name) {
		return fmt.Errorf("invalid tag name: %q\nTag names must start with alphanumeric and contain only alphanumeric, underscore, slash, dash, or dot", name)
	}

	// Security: prevent path traversal
	if pathTraversalRegex.MatchString(name) {
		return fmt.Errorf("invalid tag name: path traversal not allowed")
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

	// Get ref (default branch if not specified)
	ref := opts.ref
	if ref == "" {
		defaultBranch, err := client.Branches().GetDefault(repo.Owner, repo.Name)
		if err != nil {
			return fmt.Errorf("failed to get default branch: %w", err)
		}
		ref = defaultBranch.Name
	}

	// Create tag
	// Determine if ref is a commit hash or branch name
	req := &api.CreateTagRequest{
		TagName: name,
		Message: opts.message,
	}
	// Simple heuristic: 40-char hex string = commit hash
	if len(ref) == 40 && isHexString(ref) {
		req.CommitID = ref
	} else {
		req.BranchName = ref
	}

	tag, err := client.Tags().Create(repo.Owner, repo.Name, req)
	if err != nil {
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to create tags in %s", repo.FullName())
		}
		return fmt.Errorf("failed to create tag: %w", err)
	}

	if opts.message != "" {
		fmt.Printf("✓ Created annotated tag %q at %s\n", tag.Name, ref)
	} else {
		fmt.Printf("✓ Created tag %q at %s\n", tag.Name, ref)
	}
	return nil
}
