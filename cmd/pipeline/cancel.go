package pipeline

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

func newCancelCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "cancel <id>",
		Short: "Cancel a running pipeline",
		Long:  `Cancel a running pipeline.`,
		Example: `  # Cancel pipeline
  gf pipeline cancel 42`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid pipeline ID: %s", args[0])
			}
			return runCancel(repo, id)
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runCancel(repoFlag string, id int) error {
	// Get repository
	repo, err := git.ResolveRepo(repoFlag, config.DefaultHost())
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

	// Check if pipeline exists
	pipeline, err := client.Pipelines().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("pipeline #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to get pipeline: %w", err)
	}

	if pipeline.NormalizedStatus() != "running" && pipeline.NormalizedStatus() != "pending" {
		return fmt.Errorf("pipeline #%d is not running (status: %s)", id, pipeline.NormalizedStatus())
	}

	// Cancel pipeline
	err = client.Pipelines().Cancel(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to cancel pipelines in %s", repo.FullName())
		}
		return fmt.Errorf("failed to cancel pipeline: %w", err)
	}

	fmt.Printf("✓ Canceled pipeline #%d\n", id)
	return nil
}
