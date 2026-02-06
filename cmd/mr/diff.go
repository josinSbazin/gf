package mr

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

const diffTimeout = 2 * time.Minute

type diffOptions struct {
	repo     string
	stat     bool
	nameOnly bool
	color    string
}

func newDiffCmd() *cobra.Command {
	opts := &diffOptions{}

	cmd := &cobra.Command{
		Use:   "diff <id>",
		Short: "Show diff of a merge request",
		Long: `Show the diff between source and target branches of a merge request.

This fetches the latest changes and shows the diff locally using git.`,
		Example: `  # Show diff for MR #42
  gf mr diff 42

  # Show only file statistics
  gf mr diff 42 --stat

  # Show only changed file names
  gf mr diff 42 --name-only`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid merge request ID: %s", args[0])
			}
			return runDiff(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVar(&opts.stat, "stat", false, "Show diffstat instead of patch")
	cmd.Flags().BoolVar(&opts.nameOnly, "name-only", false, "Show only names of changed files")
	cmd.Flags().StringVar(&opts.color, "color", "auto", "Use color: always, never, auto")

	return cmd
}

func runDiff(opts *diffOptions, id int) error {
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

	// Get MR info
	mr, err := client.MergeRequests().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("merge request #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to get merge request: %w", err)
	}

	sourceBranch := mr.SourceBranch.Title
	targetBranch := mr.TargetBranch.Title

	// Validate branch names
	if err := validateBranchName(sourceBranch); err != nil {
		return fmt.Errorf("invalid source branch name: %w", err)
	}
	if err := validateBranchName(targetBranch); err != nil {
		return fmt.Errorf("invalid target branch name: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), diffTimeout)
	defer cancel()

	// Fetch latest changes
	fmt.Fprintf(os.Stderr, "Fetching branches...\n")
	fetchCmd := exec.CommandContext(ctx, "git", "fetch", "origin", sourceBranch, targetBranch)
	fetchCmd.Stderr = os.Stderr
	if err := fetchCmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("git fetch timed out")
		}
		// Continue anyway - branches might already be available locally
	}

	// Build diff command
	diffArgs := []string{"diff"}

	// Color option
	switch opts.color {
	case "always":
		diffArgs = append(diffArgs, "--color=always")
	case "never":
		diffArgs = append(diffArgs, "--color=never")
	default:
		diffArgs = append(diffArgs, "--color=auto")
	}

	// Output format options
	if opts.stat {
		diffArgs = append(diffArgs, "--stat")
	} else if opts.nameOnly {
		diffArgs = append(diffArgs, "--name-only")
	}

	// Three-dot diff: shows changes in source since it diverged from target
	diffArgs = append(diffArgs, fmt.Sprintf("origin/%s...origin/%s", targetBranch, sourceBranch))

	fmt.Fprintf(os.Stderr, "Showing diff: %s → %s\n\n", sourceBranch, targetBranch)

	diffCmd := exec.CommandContext(ctx, "git", diffArgs...)
	diffCmd.Stdout = os.Stdout
	diffCmd.Stderr = os.Stderr

	if err := diffCmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("git diff timed out")
		}
		return fmt.Errorf("failed to show diff: %w", err)
	}

	return nil
}
