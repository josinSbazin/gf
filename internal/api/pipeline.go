package api

import (
	"fmt"
	"os"
	"strings"
	"time"
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
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline", owner, project)

	if opts != nil && (opts.Page > 0 || opts.Size > 0) {
		path += "?"
		if opts.Page > 0 {
			path += fmt.Sprintf("page=%d&", opts.Page)
		}
		if opts.Size > 0 {
			path += fmt.Sprintf("size=%d&", opts.Size)
		}
		path = strings.TrimSuffix(path, "&")
	}

	var resp PipelineListResponse
	if err := s.client.Get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Embedded.Pipelines, nil
}

// Get returns a specific pipeline by localID
// Note: GitFlic API may not have a direct endpoint for single pipeline,
// so we try direct access first, then fall back to list search
func (s *PipelineService) Get(owner, project string, localID int) (*Pipeline, error) {
	// Try direct endpoint first (may not exist in all GitFlic versions)
	directPath := fmt.Sprintf("/project/%s/%s/cicd/pipeline/%d", owner, project, localID)
	var pipeline Pipeline
	if err := s.client.Get(directPath, &pipeline); err == nil && pipeline.LocalID == localID {
		return &pipeline, nil
	}

	// Fallback: fetch recent pipelines and search
	// Use pagination to limit data transfer - most lookups are for recent pipelines
	const maxPages = 5
	const pageSize = 50

	for page := 0; page < maxPages; page++ {
		pipelines, err := s.ListWithOptions(owner, project, &PipelineListOptions{
			Page: page,
			Size: pageSize,
		})
		if err != nil {
			return nil, err
		}

		for i := range pipelines {
			if pipelines[i].LocalID == localID {
				return &pipelines[i], nil
			}
		}

		// If less than pageSize returned, we've reached the end
		if len(pipelines) < pageSize {
			break
		}
	}

	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("pipeline #%d not found", localID)}
}

// Jobs returns jobs for a pipeline
func (s *PipelineService) Jobs(owner, project string, localID int) ([]Job, error) {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline/%d/jobs", owner, project, localID)

	var resp JobListResponse
	if err := s.client.Get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Embedded.Jobs, nil
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

	var p Pipeline
	if err := s.client.Post(path, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Cancel cancels a running pipeline
func (s *PipelineService) Cancel(owner, project string, localID int) error {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline/%d/cancel", owner, project, localID)
	return s.client.Post(path, nil, nil)
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

// NoColor returns true if color output should be disabled
func NoColor() bool {
	return os.Getenv("NO_COLOR") != ""
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
