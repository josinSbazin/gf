package pipeline

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type viewOptions struct {
	repo string
	web  bool
}

func newViewCmd() *cobra.Command {
	opts := &viewOptions{}

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a pipeline",
		Long:  `View details of a pipeline and its jobs.`,
		Example: `  # View pipeline #45
  gf pipeline view 45

  # Open in browser
  gf pipeline view 45 --web`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid pipeline ID: %s", args[0])
			}
			return runView(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVarP(&opts.web, "web", "w", false, "Open in browser")

	return cmd
}

func runView(opts *viewOptions, id int) error {
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

	// Fetch pipeline
	pipeline, err := client.Pipelines().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("pipeline #%d not found", id)
		}
		return fmt.Errorf("failed to get pipeline: %w", err)
	}

	// Fetch jobs
	jobs, err := client.Pipelines().Jobs(repo.Owner, repo.Name, id)
	if err != nil {
		return fmt.Errorf("failed to get jobs: %w", err)
	}

	// Print pipeline info
	fmt.Printf("\nPipeline #%d for %s (%s) - %s %s\n\n",
		pipeline.LocalID,
		pipeline.Ref,
		pipeline.SHA(),
		api.StatusIcon(pipeline.Status),
		pipeline.NormalizedStatus(),
	)

	fmt.Printf("Duration: %s\n", formatDuration(pipeline.Duration))
	fmt.Printf("Started:  %s\n", formatRelativeTime(pipeline.CreatedAt.Time))
	if pipeline.FinishedAt != nil {
		fmt.Printf("Finished: %s\n", formatRelativeTime(pipeline.FinishedAt.Time))
	}

	// Print jobs
	if len(jobs) > 0 {
		fmt.Printf("\nJOBS\n")
		fmt.Printf("%-12s %-25s %-12s %s\n", "STAGE", "NAME", "STATUS", "DURATION")
		fmt.Println(strings.Repeat("-", 60))

		for _, job := range jobs {
			status := fmt.Sprintf("%s %s", api.StatusIcon(job.Status), job.NormalizedStatus())

			name := job.Name
			if len(name) > 22 {
				name = name[:22] + "..."
			}

			fmt.Printf("%-12s %-25s %-12s %s\n",
				job.Stage,
				name,
				status,
				formatDuration(job.Duration),
			)
		}
	}

	fmt.Println()

	// URL
	url := fmt.Sprintf("https://%s/project/%s/%s/cicd/pipeline/%d",
		repo.Host, repo.Owner, repo.Name, pipeline.LocalID)
	fmt.Printf("View in browser: %s\n", url)

	return nil
}

func formatRelativeTime(t time.Time) string {
	diff := time.Since(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}
