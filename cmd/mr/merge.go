package mr

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type mergeOptions struct {
	squash       bool
	deleteBranch bool
	yes          bool
	repo         string
}

func newMergeCmd() *cobra.Command {
	opts := &mergeOptions{}

	cmd := &cobra.Command{
		Use:   "merge <id>",
		Short: "Merge a merge request",
		Long:  `Merge a merge request.`,
		Example: `  # Select MR to merge interactively
  gf mr merge

  # Merge specific MR
  gf mr merge 12

  # Merge with squash
  gf mr merge 12 --squash

  # Merge without confirmation
  gf mr merge 12 --yes`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id int
			if len(args) > 0 {
				var err error
				id, err = strconv.Atoi(strings.TrimPrefix(args[0], "#"))
				if err != nil {
					return fmt.Errorf("invalid merge request ID: %s", args[0])
				}
			}
			return runMerge(opts, id)
		},
	}

	cmd.Flags().BoolVar(&opts.squash, "squash", false, "Squash commits when merging")
	cmd.Flags().BoolVarP(&opts.deleteBranch, "delete-branch", "d", false, "Delete source branch after merge")
	cmd.Flags().BoolVarP(&opts.yes, "yes", "y", false, "Skip confirmation")
	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runMerge(opts *mergeOptions, id int) error {
	// Get repository
	repo, err := git.ResolveRepo(opts.repo, config.DefaultHost())
	if err != nil {
		return fmt.Errorf("could not determine repository: %w", err)
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

	// Interactive mode: select from open MRs if no ID provided
	if id == 0 {
		mrs, err := client.MergeRequests().List(repo.Owner, repo.Name, &api.MRListOptions{
			State: "open",
		})
		if err != nil {
			return fmt.Errorf("failed to list merge requests: %w", err)
		}

		if len(mrs) == 0 {
			return fmt.Errorf("no open merge requests in %s", repo.FullName())
		}

		fmt.Println("Open merge requests:")
		for i, mr := range mrs {
			if i >= 10 {
				fmt.Printf("  ... and %d more\n", len(mrs)-10)
				break
			}
			fmt.Printf("  #%-4d %s [%s → %s]\n",
				mr.LocalID, truncateTitle(mr.Title, 40),
				truncateTitle(mr.SourceBranch.Title, 15),
				truncateTitle(mr.TargetBranch.Title, 15))
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter MR number to merge: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			return fmt.Errorf("no MR selected")
		}

		id, err = strconv.Atoi(strings.TrimPrefix(input, "#"))
		if err != nil {
			return fmt.Errorf("invalid merge request ID: %s", input)
		}
	}

	// Get merge request first to show info
	mr, err := client.MergeRequests().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("merge request #%d not found", id)
		}
		return fmt.Errorf("failed to get merge request: %w", err)
	}

	if mr.State() != "open" {
		return fmt.Errorf("merge request #%d is %s, cannot merge", id, mr.State())
	}

	if mr.HasConflicts {
		return fmt.Errorf("merge request #%d has conflicts, resolve them first", id)
	}

	// Confirm if not --yes
	if !opts.yes {
		fmt.Printf("Merge request #%d: %s\n", mr.LocalID, mr.Title)
		fmt.Printf("  %s → %s\n\n", mr.SourceBranch.Title, mr.TargetBranch.Title)

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Merge this merge request? [y/N] ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Merge
	err = client.MergeRequests().Merge(repo.Owner, repo.Name, id, &api.MergeMRRequest{
		SquashCommit:       opts.squash,
		RemoveSourceBranch: opts.deleteBranch,
	})
	if err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	fmt.Printf("✓ Merged merge request #%d (%s → %s)\n", mr.LocalID, mr.SourceBranch.Title, mr.TargetBranch.Title)

	if opts.deleteBranch {
		fmt.Printf("✓ Deleted branch %s\n", mr.SourceBranch.Title)
	}

	return nil
}

func truncateTitle(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
