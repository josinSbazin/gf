package cmd

import (
	"fmt"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/josinSbazin/gf/internal/output"
	"github.com/spf13/cobra"
)

type statusOptions struct {
	repo string
}

func newStatusCmd() *cobra.Command {
	opts := &statusOptions{}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of current branch",
		Long: `Show the status of the current branch, including associated
merge requests and pipeline status.`,
		Example: `  # Show status
  gf status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runStatus(opts *statusOptions) error {
	// Get repository
	repo, err := git.ResolveRepo(opts.repo, config.DefaultHost())
	if err != nil {
		return fmt.Errorf("could not determine repository: %w\nUse --repo owner/name to specify", err)
	}

	// Get current branch
	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("could not get current branch: %w", err)
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

	fmt.Printf("\nCurrent branch: %s\n", currentBranch)
	fmt.Println(strings.Repeat("─", 50))

	// Find MR for current branch
	mrs, err := client.MergeRequests().List(repo.Owner, repo.Name, &api.MRListOptions{
		State: "open",
	})
	if err != nil {
		fmt.Printf("  Could not fetch MRs: %v\n", err)
	} else {
		foundMR := false
		for _, mr := range mrs {
			if mr.SourceBranch.Title == currentBranch {
				stateIcon := "●"
				if mr.State() == "merged" {
					stateIcon = "✓"
				} else if mr.State() == "closed" {
					stateIcon = "✗"
				}
				fmt.Printf("\nAssociated MR:\n")
				fmt.Printf("  %s #%d %s\n", stateIcon, mr.LocalID, mr.Title)
				fmt.Printf("    %s → %s\n", mr.SourceBranch.Title, mr.TargetBranch.Title)
				if mr.HasConflicts {
					fmt.Printf("    ⚠ Has conflicts\n")
				}
				foundMR = true
				break
			}
		}
		if !foundMR {
			fmt.Printf("\n  No associated merge request\n")
		}
	}

	// Get latest pipeline for current branch
	pipelines, err := client.Pipelines().List(repo.Owner, repo.Name)
	if err == nil && len(pipelines) > 0 {
		fmt.Printf("\nLatest pipelines:\n")
		count := 0
		for _, p := range pipelines {
			if p.Ref == currentBranch || count < 3 {
				icon := api.StatusIcon(p.Status)
				status := p.NormalizedStatus()
				updated := output.FormatRelativeTime(p.CreatedAt.Time)

				branchInfo := ""
				if p.Ref != currentBranch {
					branchInfo = fmt.Sprintf(" [%s]", truncate(p.Ref, 20))
				}

				fmt.Printf("  %s #%d %s%s (%s)\n", icon, p.LocalID, status, branchInfo, updated)
				count++
				if count >= 5 {
					break
				}
			}
		}
	}

	// Show user's other open MRs
	if len(mrs) > 1 {
		fmt.Printf("\nYour open merge requests:\n")
		count := 0
		for _, mr := range mrs {
			if mr.SourceBranch.Title != currentBranch {
				fmt.Printf("  #%-4d %s [%s]\n", mr.LocalID, truncate(mr.Title, 40), truncate(mr.SourceBranch.Title, 15))
				count++
				if count >= 5 {
					remaining := len(mrs) - count - 1 // -1 for current branch MR
					if remaining > 0 {
						fmt.Printf("  ... and %d more\n", remaining)
					}
					break
				}
			}
		}
	}

	fmt.Println()
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
