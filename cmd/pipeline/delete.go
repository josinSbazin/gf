package pipeline

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

type deleteOptions struct {
	repo  string
	force bool
}

func newDeleteCmd() *cobra.Command {
	opts := &deleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a pipeline",
		Long: `Delete a pipeline from the repository.

By default, asks for confirmation before deleting.
Use --force to skip confirmation.`,
		Example: `  # Delete pipeline (with confirmation)
  gf pipeline delete 42

  # Delete pipeline without confirmation
  gf pipeline delete 42 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid pipeline ID: %s", args[0])
			}
			return runDelete(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runDelete(opts *deleteOptions, id int) error {
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

	// Check if pipeline exists
	pipeline, err := client.Pipelines().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("pipeline #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to get pipeline: %w", err)
	}

	// Confirm deletion
	if !opts.force {
		fmt.Printf("Are you sure you want to delete pipeline #%d (%s on %s)? [y/N]: ",
			pipeline.LocalID, pipeline.NormalizedStatus(), pipeline.Ref)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Delete pipeline
	err = client.Pipelines().Delete(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to delete pipelines in %s", repo.FullName())
		}
		return fmt.Errorf("failed to delete pipeline: %w", err)
	}

	fmt.Printf("✓ Deleted pipeline #%d\n", id)
	return nil
}
