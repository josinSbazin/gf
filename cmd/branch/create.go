package branch

import (
	"fmt"
	"regexp"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type createOptions struct {
	repo string
	ref  string
}

// branchNameRegex validates branch names (git naming rules)
var branchNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9/_.-]*$`)

// pathTraversalRegex detects path traversal patterns (pre-compiled for performance)
var pathTraversalRegex = regexp.MustCompile(`\.\.`)

func newCreateCmd() *cobra.Command {
	opts := &createOptions{}

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new branch",
		Long: `Create a new branch in the repository.

The branch is created from the specified ref (branch name or commit hash).
If no ref is specified, the default branch is used.`,
		Example: `  # Create branch from default branch
  gf branch create feature/new-feature

  # Create branch from specific branch
  gf branch create hotfix/fix-123 --ref main

  # Create branch from commit hash
  gf branch create release/v1.0 --ref abc123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVar(&opts.ref, "ref", "", "Source branch or commit hash (default: default branch)")

	return cmd
}

func runCreate(opts *createOptions, name string) error {
	// Validate branch name
	if !branchNameRegex.MatchString(name) {
		return fmt.Errorf("invalid branch name: %q\nBranch names must start with alphanumeric and contain only alphanumeric, underscore, slash, dash, or dot", name)
	}

	// Security: prevent path traversal
	if containsPathTraversal(name) {
		return fmt.Errorf("invalid branch name: path traversal not allowed")
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

	// Create branch
	branch, err := client.Branches().Create(repo.Owner, repo.Name, &api.CreateBranchRequest{
		NewBranch:    name,
		OriginBranch: ref,
	})
	if err != nil {
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to create branches in %s", repo.FullName())
		}
		return fmt.Errorf("failed to create branch: %w", err)
	}

	fmt.Printf("✓ Created branch %q from %q\n", branch.Name, ref)
	return nil
}

// containsPathTraversal checks for path traversal patterns
func containsPathTraversal(s string) bool {
	return pathTraversalRegex.MatchString(s)
}
