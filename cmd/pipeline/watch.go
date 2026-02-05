package pipeline

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type watchOptions struct {
	interval   int
	exitStatus bool
	repo       string
}

func newWatchCmd() *cobra.Command {
	opts := &watchOptions{}

	cmd := &cobra.Command{
		Use:   "watch <id>",
		Short: "Watch a pipeline in real-time",
		Long:  `Watch a pipeline and its jobs update in real-time.`,
		Example: `  # Watch pipeline #45
  gf pipeline watch 45

  # Watch with custom interval
  gf pipeline watch 45 --interval 5

  # Exit with pipeline's exit status
  gf pipeline watch 45 --exit-status`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid pipeline ID: %s", args[0])
			}
			return runWatch(opts, id)
		},
	}

	cmd.Flags().IntVarP(&opts.interval, "interval", "i", 3, "Refresh interval in seconds")
	cmd.Flags().BoolVar(&opts.exitStatus, "exit-status", false, "Exit with pipeline status (0=success, 1=failed)")
	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runWatch(opts *watchOptions, id int) error {
	// Get repository
	var repo *git.Repository
	var err error

	if opts.repo != "" {
		parts := strings.Split(opts.repo, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format, expected owner/name")
		}
		repo = &git.Repository{
			Host:  config.DefaultHost(),
			Owner: parts[0],
			Name:  parts[1],
		}
	} else {
		repo, err = git.DetectRepo()
		if err != nil {
			return fmt.Errorf("could not determine repository: %w", err)
		}
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

	// Setup signal handler for clean exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(opts.interval) * time.Second)
	defer ticker.Stop()

	// Initial fetch
	finalStatus, err := displayPipeline(client, repo, id)
	if err != nil {
		return err
	}

	// If already finished, exit
	if isFinished(finalStatus) {
		return exitWithStatus(finalStatus, opts.exitStatus)
	}

	fmt.Println("\n[Ctrl+C to stop watching]")

	for {
		select {
		case <-sigChan:
			fmt.Println("\nStopped watching.")
			return nil
		case <-ticker.C:
			// Clear screen and redraw
			fmt.Print("\033[H\033[2J")

			finalStatus, err = displayPipeline(client, repo, id)
			if err != nil {
				return err
			}

			if isFinished(finalStatus) {
				return exitWithStatus(finalStatus, opts.exitStatus)
			}

			fmt.Println("\n[Ctrl+C to stop watching]")
		}
	}
}

func displayPipeline(client *api.Client, repo *git.Repository, id int) (string, error) {
	// Fetch pipeline
	pipeline, err := client.Pipelines().Get(repo.Owner, repo.Name, id)
	if err != nil {
		return "", fmt.Errorf("failed to get pipeline: %w", err)
	}

	// Fetch jobs
	jobs, err := client.Pipelines().Jobs(repo.Owner, repo.Name, id)
	if err != nil {
		return "", fmt.Errorf("failed to get jobs: %w", err)
	}

	// Print pipeline info
	sha := pipeline.SHA
	if len(sha) > 7 {
		sha = sha[:7]
	}

	statusColor := api.StatusColor(pipeline.Status)
	fmt.Printf("\nPipeline #%d for %s (%s)\n\n",
		pipeline.LocalID,
		pipeline.Ref,
		sha,
	)

	// Print jobs with status
	for _, job := range jobs {
		icon := api.StatusIcon(job.Status)
		color := api.StatusColor(job.Status)

		status := job.Status
		if job.Status == "running" {
			status = "running..."
		}

		fmt.Printf(" %s%s%s %-20s %-15s %s\n",
			color,
			icon,
			api.ColorReset,
			job.Name,
			status,
			formatDuration(job.Duration),
		)
	}

	// Print overall status
	fmt.Printf("\n%sOverall: %s %s%s\n",
		statusColor,
		api.StatusIcon(pipeline.Status),
		pipeline.Status,
		api.ColorReset,
	)

	return pipeline.Status, nil
}

func isFinished(status string) bool {
	switch status {
	case "success", "passed", "failed", "canceled":
		return true
	default:
		return false
	}
}

func exitWithStatus(status string, useExitStatus bool) error {
	if !useExitStatus {
		return nil
	}

	switch status {
	case "success", "passed":
		os.Exit(0)
	default:
		os.Exit(1)
	}
	return nil
}
