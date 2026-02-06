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

func newRetryCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:     "retry <id>",
		Aliases: []string{"restart"},
		Short:   "Retry a failed pipeline",
		Long:    `Retry (restart) a failed pipeline.`,
		Example: `  # Retry pipeline
  gf pipeline retry 42`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid pipeline ID: %s", args[0])
			}
			return runRetry(repo, id)
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runRetry(repoFlag string, id int) error {
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

	// Retry pipeline
	pipeline, err := client.Pipelines().Restart(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("pipeline #%d not found in %s", id, repo.FullName())
		}
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to restart pipelines in %s", repo.FullName())
		}
		return fmt.Errorf("failed to retry pipeline: %w", err)
	}

	fmt.Printf("✓ Retried pipeline #%d\n", pipeline.LocalID)
	return nil
}
