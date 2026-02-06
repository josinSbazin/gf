package commit

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/spf13/cobra"
)

const diffTimeout = 2 * time.Minute

// commitHashRegex validates commit hash format (hex string, 7-40 chars)
var commitHashRegex = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)

type diffOptions struct {
	stat     bool
	nameOnly bool
	color    string
}

func newDiffCmd() *cobra.Command {
	opts := &diffOptions{}

	cmd := &cobra.Command{
		Use:   "diff <hash>",
		Short: "View commit diff",
		Long: `View the diff (changes) introduced by a commit.

This uses local git to show the diff. The repository must be cloned locally.`,
		Example: `  # View commit diff
  gf commit diff abc1234

  # Show only stats
  gf commit diff abc1234 --stat

  # Show only file names
  gf commit diff abc1234 --name-only`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiff(opts, args[0])
		},
	}

	cmd.Flags().BoolVar(&opts.stat, "stat", false, "Show diffstat only")
	cmd.Flags().BoolVar(&opts.nameOnly, "name-only", false, "Show only file names")
	cmd.Flags().StringVar(&opts.color, "color", "auto", "Use color: always, never, auto")

	return cmd
}

func runDiff(opts *diffOptions, hash string) error {
	// Validate commit hash format
	if !commitHashRegex.MatchString(hash) {
		return fmt.Errorf("invalid commit hash format: %s", hash)
	}

	ctx, cancel := context.WithTimeout(context.Background(), diffTimeout)
	defer cancel()

	// Build git show command - shows commit message + diff
	args := []string{"show"}

	// Color option
	switch opts.color {
	case "always":
		args = append(args, "--color=always")
	case "never":
		args = append(args, "--color=never")
	default:
		args = append(args, "--color=auto")
	}

	// Output format options
	if opts.stat {
		args = append(args, "--stat")
	} else if opts.nameOnly {
		args = append(args, "--name-only")
	}

	args = append(args, hash)

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("git show timed out")
		}
		return fmt.Errorf("failed to show commit diff: %w\nMake sure you're in a git repository with the commit available", err)
	}

	return nil
}
