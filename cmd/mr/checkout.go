package mr

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

const gitCommandTimeout = 2 * time.Minute

// validBranchNameRegex validates git branch names
// Disallows: starting with -, containing .., control chars, spaces, ~, ^, :, \, *, ?, [
var validBranchNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9/_\.]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)

// validateBranchName checks if a branch name is safe for git operations
func validateBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("branch name cannot start with '-'")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("branch name cannot contain '..'")
	}
	if !validBranchNameRegex.MatchString(name) {
		return fmt.Errorf("invalid branch name: %s", name)
	}
	return nil
}

type checkoutOptions struct {
	repo   string
	branch string
	force  bool
}

func newCheckoutCmd() *cobra.Command {
	opts := &checkoutOptions{}

	cmd := &cobra.Command{
		Use:   "checkout <id>",
		Short: "Check out a merge request locally",
		Long: `Check out the source branch of a merge request locally.

This fetches the branch from the remote and switches to it.`,
		Example: `  # Checkout MR #42
  gf mr checkout 42

  # Checkout to a specific local branch name
  gf mr checkout 42 --branch my-review

  # Force checkout (discard local changes)
  gf mr checkout 42 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid merge request ID: %s", args[0])
			}
			return runCheckout(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVarP(&opts.branch, "branch", "b", "", "Local branch name (default: source branch name)")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Force checkout (discard local changes)")

	return cmd
}

func runCheckout(opts *checkoutOptions, id int) error {
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

	// Get MR details
	mr, err := client.MergeRequests().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("merge request #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to get merge request: %w", err)
	}

	// Determine branch names
	remoteBranch := mr.SourceBranch.Title
	localBranch := opts.branch
	if localBranch == "" {
		localBranch = remoteBranch
	}

	// Security: Validate branch names to prevent command injection
	if err := validateBranchName(remoteBranch); err != nil {
		return fmt.Errorf("invalid remote branch name from API: %w", err)
	}
	if err := validateBranchName(localBranch); err != nil {
		return fmt.Errorf("invalid local branch name: %w", err)
	}

	fmt.Printf("Checking out MR #%d: %s\n", mr.LocalID, mr.Title)
	fmt.Printf("Source branch: %s\n", remoteBranch)

	// Create context with timeout for git operations
	ctx, cancel := context.WithTimeout(context.Background(), gitCommandTimeout)
	defer cancel()

	// Fetch the branch
	fmt.Println("Fetching from remote...")
	fetchCmd := exec.CommandContext(ctx, "git", "fetch", "origin", remoteBranch)
	fetchCmd.Stdout = os.Stdout
	fetchCmd.Stderr = os.Stderr
	if err := fetchCmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("git fetch timed out after %v", gitCommandTimeout)
		}
		return fmt.Errorf("failed to fetch branch: %w", err)
	}

	// Checkout the branch
	var checkoutArgs []string
	if opts.force {
		checkoutArgs = []string{"checkout", "-f"}
	} else {
		checkoutArgs = []string{"checkout"}
	}

	// Check if branch exists locally
	checkBranchCmd := exec.Command("git", "rev-parse", "--verify", localBranch)
	branchExists := checkBranchCmd.Run() == nil

	if branchExists {
		// Branch exists, just checkout
		checkoutArgs = append(checkoutArgs, localBranch)
	} else {
		// Create new branch tracking remote
		checkoutArgs = append(checkoutArgs, "-b", localBranch, "origin/"+remoteBranch)
	}

	fmt.Printf("Switching to branch '%s'...\n", localBranch)
	checkoutCmd := exec.CommandContext(ctx, "git", checkoutArgs...)
	checkoutCmd.Stdout = os.Stdout
	checkoutCmd.Stderr = os.Stderr
	if err := checkoutCmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("git checkout timed out after %v", gitCommandTimeout)
		}
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	fmt.Printf("\n✓ Checked out MR #%d on branch '%s'\n", mr.LocalID, localBranch)
	return nil
}
