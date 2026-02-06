package commit

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/josinSbazin/gf/internal/output"
	"github.com/spf13/cobra"
)

type viewOptions struct {
	repo string
	json bool
}

func newViewCmd() *cobra.Command {
	opts := &viewOptions{}

	cmd := &cobra.Command{
		Use:   "view <hash>",
		Short: "View a commit",
		Long:  `View details of a specific commit.`,
		Example: `  # View commit
  gf commit view abc1234

  # Output as JSON
  gf commit view abc1234 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output as JSON")

	return cmd
}

func runView(opts *viewOptions, hash string) error {
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

	// Get commit
	commit, err := client.Commits().Get(repo.Owner, repo.Name, hash)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("commit %s not found in %s", hash, repo.FullName())
		}
		return fmt.Errorf("failed to get commit: %w", err)
	}

	// JSON output
	if opts.json {
		data, err := json.MarshalIndent(commit, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Print commit details
	fmt.Printf("\ncommit %s\n", commit.Hash)
	if len(commit.ParentHashes) > 0 {
		fmt.Printf("Parent: %s\n", strings.Join(commit.ParentHashes, " "))
	}
	fmt.Printf("Author: %s <%s>\n", commit.AuthorName, commit.AuthorEmail)
	if commit.CommitterName != commit.AuthorName {
		fmt.Printf("Committer: %s <%s>\n", commit.CommitterName, commit.CommitterEmail)
	}
	fmt.Printf("Date:   %s\n", output.FormatRelativeTime(commit.CreatedAt))
	fmt.Printf("\n    %s\n", strings.ReplaceAll(commit.Message, "\n", "\n    "))

	return nil
}
