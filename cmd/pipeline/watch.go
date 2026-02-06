package pipeline

import (
	"context"
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
	"github.com/josinSbazin/gf/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	minInterval     = 1
	maxInterval     = 300
	apiCallTimeout  = 30 * time.Second
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
	// Validate interval
	if opts.interval < minInterval {
		opts.interval = minInterval
	} else if opts.interval > maxInterval {
		return fmt.Errorf("interval must be between %d and %d seconds", minInterval, maxInterval)
	}

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

	// Check if we're in a terminal (for ANSI escape codes)
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	// Setup signal handler for clean exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan) // Clean up signal handler

	ticker := time.NewTicker(time.Duration(opts.interval) * time.Second)
	defer ticker.Stop()

	// Initial fetch with timeout
	ctx, cancel := context.WithTimeout(context.Background(), apiCallTimeout)
	finalStatus, err := displayPipelineWithContext(ctx, client, repo, id)
	cancel()
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
			// Clear screen only if TTY (avoid garbage in redirected output)
			if isTTY {
				fmt.Print("\033[H\033[2J")
			} else {
				fmt.Println("\n---") // Separator for non-TTY
			}

			ctx, cancel := context.WithTimeout(context.Background(), apiCallTimeout)
			finalStatus, err = displayPipelineWithContext(ctx, client, repo, id)
			cancel()
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

func displayPipelineWithContext(ctx context.Context, client *api.Client, repo *git.Repository, id int) (string, error) {
	// Fetch pipeline with context
	pipeline, err := client.Pipelines().GetWithContext(ctx, repo.Owner, repo.Name, id)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("API request timed out")
		}
		return "", fmt.Errorf("failed to get pipeline: %w", err)
	}

	// Fetch jobs with context
	jobs, err := client.Pipelines().JobsWithContext(ctx, repo.Owner, repo.Name, id)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("API request timed out")
		}
		return "", fmt.Errorf("failed to get jobs: %w", err)
	}

	// Print pipeline info
	statusColor := api.StatusColor(pipeline.Status)
	fmt.Printf("\nPipeline #%d for %s (%s)\n\n",
		pipeline.LocalID,
		pipeline.Ref,
		pipeline.SHA(),
	)

	// Print jobs with status
	for _, job := range jobs {
		icon := api.StatusIcon(job.Status)
		color := api.StatusColor(job.Status)

		status := job.NormalizedStatus()
		if job.NormalizedStatus() == "running" {
			status = "running..."
		}

		fmt.Printf(" %s%s%s %-20s %-15s %s\n",
			color,
			icon,
			api.ColorReset(),
			job.Name,
			status,
			output.FormatDuration(job.Duration),
		)
	}

	// Print overall status
	fmt.Printf("\n%sOverall: %s %s%s\n",
		statusColor,
		api.StatusIcon(pipeline.Status),
		pipeline.NormalizedStatus(),
		api.ColorReset(),
	)

	return pipeline.NormalizedStatus(), nil
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
		return api.NewExitError(0)
	default:
		return api.NewExitError(1)
	}
}
