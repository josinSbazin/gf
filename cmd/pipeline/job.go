package pipeline

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/josinSbazin/gf/internal/output"
	"github.com/spf13/cobra"
)

func newJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Work with pipeline jobs",
		Long:  `View and manage individual jobs within a pipeline.`,
	}

	cmd.AddCommand(newJobViewCmd())
	cmd.AddCommand(newJobLogCmd())
	cmd.AddCommand(newJobRetryCmd())
	cmd.AddCommand(newJobCancelCmd())

	return cmd
}

// parseJobArgs parses "pipeline_id job_id" or "pipeline_id:job_id" format
func parseJobArgs(args []string) (pipelineID, jobID int, err error) {
	if len(args) == 2 {
		pipelineID, err = strconv.Atoi(strings.TrimPrefix(args[0], "#"))
		if err != nil {
			return 0, 0, fmt.Errorf("invalid pipeline ID: %s", args[0])
		}
		jobID, err = strconv.Atoi(strings.TrimPrefix(args[1], "#"))
		if err != nil {
			return 0, 0, fmt.Errorf("invalid job ID: %s", args[1])
		}
		return pipelineID, jobID, nil
	}

	if len(args) == 1 && strings.Contains(args[0], ":") {
		parts := strings.SplitN(args[0], ":", 2)
		pipelineID, err = strconv.Atoi(strings.TrimPrefix(parts[0], "#"))
		if err != nil {
			return 0, 0, fmt.Errorf("invalid pipeline ID: %s", parts[0])
		}
		jobID, err = strconv.Atoi(strings.TrimPrefix(parts[1], "#"))
		if err != nil {
			return 0, 0, fmt.Errorf("invalid job ID: %s", parts[1])
		}
		return pipelineID, jobID, nil
	}

	return 0, 0, fmt.Errorf("expected format: <pipeline-id> <job-id> or <pipeline-id>:<job-id>")
}

func newJobViewCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "view <pipeline-id> <job-id>",
		Short: "View job details",
		Long:  `View details of a specific job within a pipeline.`,
		Example: `  # View job
  gf pipeline job view 42 1

  # Alternative format
  gf pipeline job view 42:1`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pipelineID, jobID, err := parseJobArgs(args)
			if err != nil {
				return err
			}
			return runJobView(repo, pipelineID, jobID)
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runJobView(repoFlag string, pipelineID, jobID int) error {
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

	// Get jobs for pipeline
	jobs, err := client.Pipelines().Jobs(repo.Owner, repo.Name, pipelineID)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("pipeline #%d not found in %s", pipelineID, repo.FullName())
		}
		return fmt.Errorf("failed to get jobs: %w", err)
	}

	// Find the job
	var job *api.Job
	for i := range jobs {
		if jobs[i].LocalID == jobID {
			job = &jobs[i]
			break
		}
	}

	if job == nil {
		return fmt.Errorf("job #%d not found in pipeline #%d", jobID, pipelineID)
	}

	// Print job details
	color := api.StatusColor(job.Status)
	reset := api.ColorReset()

	fmt.Printf("\nJob #%d: %s\n", job.LocalID, job.Name)
	fmt.Printf("Stage:    %s\n", job.Stage)
	fmt.Printf("Status:   %s%s %s%s\n", color, api.StatusIcon(job.Status), job.NormalizedStatus(), reset)
	if job.Duration > 0 {
		fmt.Printf("Duration: %s\n", output.FormatDuration(job.Duration))
	}
	if job.Runner != "" {
		fmt.Printf("Runner:   %s\n", job.Runner)
	}

	return nil
}

func newJobLogCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "log <pipeline-id> <job-id>",
		Short: "View job log",
		Long:  `View the log output of a specific job.`,
		Example: `  # View job log
  gf pipeline job log 42 1`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pipelineID, jobID, err := parseJobArgs(args)
			if err != nil {
				return err
			}
			return runJobLog(repo, pipelineID, jobID)
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runJobLog(repoFlag string, pipelineID, jobID int) error {
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

	// Get job log
	log, err := client.Pipelines().GetJobLog(repo.Owner, repo.Name, pipelineID, jobID)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("job log not found for pipeline #%d job #%d", pipelineID, jobID)
		}
		return fmt.Errorf("failed to get job log: %w", err)
	}

	if log == "" {
		fmt.Println("(no log output)")
		return nil
	}

	fmt.Println(log)
	return nil
}

func newJobRetryCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "retry <pipeline-id> <job-id>",
		Short: "Retry a failed job",
		Long:  `Retry (restart) a failed job within a pipeline.`,
		Example: `  # Retry job
  gf pipeline job retry 42 1`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pipelineID, jobID, err := parseJobArgs(args)
			if err != nil {
				return err
			}
			return runJobRetry(repo, pipelineID, jobID)
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runJobRetry(repoFlag string, pipelineID, jobID int) error {
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

	// Retry job
	job, err := client.Pipelines().RestartJob(repo.Owner, repo.Name, pipelineID, jobID)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("job #%d not found in pipeline #%d", jobID, pipelineID)
		}
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to restart jobs in %s", repo.FullName())
		}
		return fmt.Errorf("failed to retry job: %w", err)
	}

	fmt.Printf("✓ Retried job #%d (%s)\n", job.LocalID, job.Name)
	return nil
}

func newJobCancelCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "cancel <pipeline-id> <job-id>",
		Short: "Cancel a running job",
		Long:  `Cancel a running job within a pipeline.`,
		Example: `  # Cancel job
  gf pipeline job cancel 42 1`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pipelineID, jobID, err := parseJobArgs(args)
			if err != nil {
				return err
			}
			return runJobCancel(repo, pipelineID, jobID)
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runJobCancel(repoFlag string, pipelineID, jobID int) error {
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

	// Cancel job
	err = client.Pipelines().CancelJob(repo.Owner, repo.Name, pipelineID, jobID)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("job #%d not found in pipeline #%d", jobID, pipelineID)
		}
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to cancel jobs in %s", repo.FullName())
		}
		return fmt.Errorf("failed to cancel job: %w", err)
	}

	fmt.Printf("✓ Canceled job #%d\n", jobID)
	return nil
}
