package commit

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type diffOptions struct {
	repo    string
	json    bool
	stat    bool
}

func newDiffCmd() *cobra.Command {
	opts := &diffOptions{}

	cmd := &cobra.Command{
		Use:   "diff <hash>",
		Short: "View commit diff",
		Long:  `View the diff (changes) introduced by a commit.`,
		Example: `  # View commit diff
  gf commit diff abc1234

  # Show only stats
  gf commit diff abc1234 --stat

  # Output as JSON
  gf commit diff abc1234 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiff(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&opts.stat, "stat", false, "Show diffstat only")

	return cmd
}

func runDiff(opts *diffOptions, hash string) error {
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

	// Get diff
	diffs, err := client.Commits().Diff(repo.Owner, repo.Name, hash)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("commit %s not found in %s", hash, repo.FullName())
		}
		return fmt.Errorf("failed to get diff: %w", err)
	}

	if len(diffs) == 0 {
		fmt.Println("No changes in this commit")
		return nil
	}

	// JSON output
	if opts.json {
		data, err := json.MarshalIndent(diffs, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Stats only
	if opts.stat {
		totalAdditions := 0
		totalDeletions := 0

		for _, d := range diffs {
			path := d.FilePath
			if d.OldPath != "" && d.OldPath != d.FilePath {
				path = fmt.Sprintf("%s → %s", d.OldPath, d.FilePath)
			}

			change := changeTypeSymbol(d.ChangeType)
			fmt.Printf("%s %s | +%d -%d\n", change, path, d.Additions, d.Deletions)
			totalAdditions += d.Additions
			totalDeletions += d.Deletions
		}

		fmt.Printf("\n%d files changed, %d insertions(+), %d deletions(-)\n",
			len(diffs), totalAdditions, totalDeletions)
		return nil
	}

	// Full diff output
	for _, d := range diffs {
		fmt.Printf("\n%s %s\n", changeTypeSymbol(d.ChangeType), d.FilePath)
		fmt.Println(strings.Repeat("-", 60))

		if d.DiffContent != "" {
			// Colorize diff output
			lines := strings.Split(d.DiffContent, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
					fmt.Printf("%s%s%s\n", api.StatusColor("success"), line, api.ColorReset())
				} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
					fmt.Printf("%s%s%s\n", api.StatusColor("failed"), line, api.ColorReset())
				} else if strings.HasPrefix(line, "@@") {
					fmt.Printf("%s%s%s\n", api.StatusColor("running"), line, api.ColorReset())
				} else {
					fmt.Println(line)
				}
			}
		} else {
			fmt.Printf("+%d -%d\n", d.Additions, d.Deletions)
		}
	}

	return nil
}

func changeTypeSymbol(changeType string) string {
	switch strings.ToUpper(changeType) {
	case "ADD", "ADDED":
		return fmt.Sprintf("%sA%s", api.StatusColor("success"), api.ColorReset())
	case "MODIFY", "MODIFIED":
		return fmt.Sprintf("%sM%s", api.StatusColor("running"), api.ColorReset())
	case "DELETE", "DELETED":
		return fmt.Sprintf("%sD%s", api.StatusColor("failed"), api.ColorReset())
	case "RENAME", "RENAMED":
		return fmt.Sprintf("%sR%s", api.StatusColor("pending"), api.ColorReset())
	default:
		return "?"
	}
}
