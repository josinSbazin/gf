package api

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// Pagination constants for pipeline fallback search
const (
	pipelineSearchMaxPages = 5
	pipelineSearchPageSize = 50
)

// FlexTime handles time parsing with or without timezone
type FlexTime struct {
	time.Time
}

func (ft *FlexTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		return nil
	}

	// Try RFC3339 first (with timezone)
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		ft.Time = t
		return nil
	}

	// Try without timezone
	t, err = time.Parse("2006-01-02T15:04:05.999999", s)
	if err == nil {
		ft.Time = t
		return nil
	}

	// Try ISO format
	t, err = time.Parse("2006-01-02T15:04:05", s)
	if err == nil {
		ft.Time = t
		return nil
	}

	// GitFlic sometimes returns relative time strings like "2 time.minute.multi time.ago"
	// In this case, use current time as fallback
	if strings.Contains(s, "time.") {
		ft.Time = time.Now()
		return nil
	}

	return fmt.Errorf("cannot parse time: %s", s)
}

// PipelineService handles pipeline API calls
type PipelineService struct {
	client *Client
}

// Pipeline represents a GitFlic CI/CD pipeline
type Pipeline struct {
	ID         string    `json:"id"`
	LocalID    int       `json:"localId"`
	Status     string    `json:"status"` // PENDING, RUNNING, SUCCESS, FAILED, CANCELED
	Ref        string    `json:"ref"`    // branch or tag
	CommitID   string    `json:"commitId"`
	Source     string    `json:"source"` // PUSH, MERGE_REQUEST, etc.
	CreatedAt  FlexTime  `json:"createdAt"`
	StartedAt  *FlexTime `json:"startedAt,omitempty"`
	FinishedAt *FlexTime `json:"finishedAt,omitempty"`
	Duration   int       `json:"duration"` // seconds
}

// SHA returns short commit hash
func (p *Pipeline) SHA() string {
	if len(p.CommitID) > 7 {
		return p.CommitID[:7]
	}
	return p.CommitID
}

// NormalizedStatus returns lowercase status
func (p *Pipeline) NormalizedStatus() string {
	return strings.ToLower(p.Status)
}

// PipelineListResponse represents the paginated response from pipeline list API
type PipelineListResponse struct {
	Embedded struct {
		Pipelines []Pipeline `json:"restPipelineModelList"`
	} `json:"_embedded"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages    int `json:"totalPages"`
		Number        int `json:"number"`
	} `json:"page"`
}

// Job represents a job within a pipeline
type Job struct {
	ID         string     `json:"id"`
	LocalID    int        `json:"localId"`
	Name       string     `json:"name"`
	Stage      string     `json:"stageName"` // API returns stageName
	Status     string     `json:"status"`    // PENDING, RUNNING, SUCCESS, FAILED, CANCELED, SKIPPED
	StartedAt  *FlexTime  `json:"startedAt,omitempty"`
	FinishedAt *FlexTime  `json:"finishedAt,omitempty"`
	Duration   int        `json:"duration"`
	Runner     string     `json:"runner,omitempty"`
}

// NormalizedStatus returns lowercase status
func (j *Job) NormalizedStatus() string {
	return strings.ToLower(j.Status)
}

// JobListResponse represents the paginated response from job list API
type JobListResponse struct {
	Embedded struct {
		Jobs []Job `json:"restPipelineJobModelList"`
	} `json:"_embedded"`
}

// PipelineListOptions specifies options for listing pipelines
type PipelineListOptions struct {
	Page int // 0-indexed page number
	Size int // items per page (default: 20)
}

// List returns pipelines for a project
func (s *PipelineService) List(owner, project string) ([]Pipeline, error) {
	return s.ListWithOptions(owner, project, nil)
}

// ListWithOptions returns pipelines with pagination
func (s *PipelineService) ListWithOptions(owner, project string, opts *PipelineListOptions) ([]Pipeline, error) {
	return s.listWithContext(context.Background(), owner, project, opts)
}

// listWithContext is the internal implementation with context support
func (s *PipelineService) listWithContext(ctx context.Context, owner, project string, opts *PipelineListOptions) ([]Pipeline, error) {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline", owner, project)

	if opts != nil && (opts.Page > 0 || opts.Size > 0) {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", fmt.Sprintf("%d", opts.Page))
		}
		if opts.Size > 0 {
			params.Set("size", fmt.Sprintf("%d", opts.Size))
		}
		path += "?" + params.Encode()
	}

	var resp PipelineListResponse
	if err := s.client.GetWithContext(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Embedded.Pipelines, nil
}

// Get returns a specific pipeline by localID
// Note: GitFlic API may not have a direct endpoint for single pipeline,
// so we try direct access first, then fall back to list search
func (s *PipelineService) Get(owner, project string, localID int) (*Pipeline, error) {
	return s.GetWithContext(context.Background(), owner, project, localID)
}

// Jobs returns jobs for a pipeline
func (s *PipelineService) Jobs(owner, project string, localID int) ([]Job, error) {
	return s.JobsWithContext(context.Background(), owner, project, localID)
}

// JobsWithContext returns jobs for a pipeline with context support
func (s *PipelineService) JobsWithContext(ctx context.Context, owner, project string, localID int) ([]Job, error) {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline/%d/jobs", owner, project, localID)

	var resp JobListResponse
	if err := s.client.GetWithContext(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Embedded.Jobs, nil
}

// GetWithContext returns a specific pipeline by localID with context support
func (s *PipelineService) GetWithContext(ctx context.Context, owner, project string, localID int) (*Pipeline, error) {
	// Try direct endpoint first (may not exist in all GitFlic versions)
	directPath := fmt.Sprintf("/project/%s/%s/cicd/pipeline/%d", owner, project, localID)
	var pipeline Pipeline
	if err := s.client.GetWithContext(ctx, directPath, &pipeline); err == nil && pipeline.LocalID == localID {
		return &pipeline, nil
	}

	// Fallback: search through paginated list
	return s.findPipelineByLocalID(ctx, owner, project, localID)
}

// findPipelineByLocalID searches for a pipeline by localID through paginated results
func (s *PipelineService) findPipelineByLocalID(ctx context.Context, owner, project string, localID int) (*Pipeline, error) {
	for page := 0; page < pipelineSearchMaxPages; page++ {
		pipelines, err := s.listWithContext(ctx, owner, project, &PipelineListOptions{
			Page: page,
			Size: pipelineSearchPageSize,
		})
		if err != nil {
			return nil, err
		}

		for i := range pipelines {
			if pipelines[i].LocalID == localID {
				return &pipelines[i], nil
			}
		}

		// Reached the end of results
		if len(pipelines) < pipelineSearchPageSize {
			break
		}
	}

	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("pipeline #%d not found", localID)}
}

// Start starts a new pipeline
func (s *PipelineService) Start(owner, project string, ref string) (*Pipeline, error) {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline/start", owner, project)

	body := map[string]string{"ref": ref}
	var p Pipeline
	if err := s.client.Post(path, body, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Restart restarts a pipeline
func (s *PipelineService) Restart(owner, project string, localID int) (*Pipeline, error) {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline/%d/restart", owner, project, localID)

	// API returns empty body on restart, so don't try to decode
	if err := s.client.Post(path, nil, nil); err != nil {
		return nil, err
	}

	// Fetch the pipeline info after restart
	return s.Get(owner, project, localID)
}

// Cancel cancels a running pipeline
func (s *PipelineService) Cancel(owner, project string, localID int) error {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline/%d/cancel", owner, project, localID)
	return s.client.Post(path, nil, nil)
}

// Delete deletes a pipeline
func (s *PipelineService) Delete(owner, project string, localID int) error {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline/%d", owner, project, localID)
	return s.client.Delete(path)
}

// GetJob returns a specific job by localID
func (s *PipelineService) GetJob(owner, project string, pipelineID, jobID int) (*Job, error) {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline/%d/job/%d", owner, project, pipelineID, jobID)

	var job Job
	if err := s.client.Get(path, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// RestartJob restarts a job
func (s *PipelineService) RestartJob(owner, project string, pipelineID, jobID int) (*Job, error) {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline/%d/job/%d/restart", owner, project, pipelineID, jobID)

	var job Job
	if err := s.client.Post(path, nil, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// CancelJob cancels a running job
func (s *PipelineService) CancelJob(owner, project string, pipelineID, jobID int) error {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline/%d/job/%d/cancel", owner, project, pipelineID, jobID)
	return s.client.Post(path, nil, nil)
}

// GetJobLog returns the log output for a job
func (s *PipelineService) GetJobLog(owner, project string, pipelineID, jobID int) (string, error) {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline/%d/job/%d/log", owner, project, pipelineID, jobID)

	var log struct {
		Content string `json:"content"`
	}
	if err := s.client.Get(path, &log); err != nil {
		return "", err
	}
	return log.Content, nil
}

// StatusIcon returns an icon for the pipeline status
func StatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "success", "passed":
		return "✓"
	case "failed":
		return "✗"
	case "running":
		return "⧖"
	case "pending":
		return "◯"
	case "canceled":
		return "⊘"
	case "skipped":
		return "↷"
	default:
		return "?"
	}
}

var (
	noColorOnce   sync.Once
	noColorCached bool
)

// NoColor returns true if color output should be disabled
// Result is cached on first call for performance
func NoColor() bool {
	noColorOnce.Do(func() {
		noColorCached = os.Getenv("NO_COLOR") != ""
	})
	return noColorCached
}

// StatusColor returns color for terminal output
// Returns empty string if NO_COLOR env is set
func StatusColor(status string) string {
	if NoColor() {
		return ""
	}

	switch strings.ToLower(status) {
	case "success", "passed":
		return "\033[32m" // green
	case "failed":
		return "\033[31m" // red
	case "running":
		return "\033[33m" // yellow
	case "pending":
		return "\033[90m" // gray
	case "canceled", "skipped":
		return "\033[90m" // gray
	default:
		return ""
	}
}

// ColorReset resets terminal color
// Returns empty string if NO_COLOR env is set
func ColorReset() string {
	if NoColor() {
		return ""
	}
	return "\033[0m"
}

// MRStateColor returns color for MR state
func MRStateColor(state string) string {
	if NoColor() {
		return ""
	}

	switch strings.ToLower(state) {
	case "open":
		return "\033[32m" // green
	case "merged":
		return "\033[35m" // magenta
	case "closed":
		return "\033[31m" // red
	default:
		return ""
	}
}

// IssueStateColor returns color for issue state
func IssueStateColor(state string) string {
	if NoColor() {
		return ""
	}

	switch strings.ToLower(state) {
	case "open":
		return "\033[32m" // green
	case "closed":
		return "\033[31m" // red
	default:
		return ""
	}
}
