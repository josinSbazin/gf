package api

import (
	"fmt"
	"time"
)

// PipelineService handles pipeline API calls
type PipelineService struct {
	client *Client
}

// Pipeline represents a GitFlic CI/CD pipeline
type Pipeline struct {
	UUID       string     `json:"uuid"`
	LocalID    int        `json:"localId"`
	Status     string     `json:"status"` // pending, running, success, failed, canceled
	Ref        string     `json:"ref"`    // branch or tag
	SHA        string     `json:"sha"`
	CreatedAt  time.Time  `json:"createdAt"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	Duration   int        `json:"duration"` // seconds
	Author     User       `json:"author"`
}

// Job represents a job within a pipeline
type Job struct {
	UUID       string     `json:"uuid"`
	LocalID    int        `json:"localId"`
	Name       string     `json:"name"`
	Stage      string     `json:"stage"`
	Status     string     `json:"status"` // pending, running, success, failed, canceled, skipped
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	Duration   int        `json:"duration"`
	Runner     string     `json:"runner,omitempty"`
}

// List returns pipelines for a project
func (s *PipelineService) List(owner, project string) ([]Pipeline, error) {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline", owner, project)

	var pipelines []Pipeline
	if err := s.client.Get(path, &pipelines); err != nil {
		return nil, err
	}
	return pipelines, nil
}

// Get returns a specific pipeline
func (s *PipelineService) Get(owner, project string, localID int) (*Pipeline, error) {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline/%d", owner, project, localID)

	var p Pipeline
	if err := s.client.Get(path, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Jobs returns jobs for a pipeline
func (s *PipelineService) Jobs(owner, project string, localID int) ([]Job, error) {
	path := fmt.Sprintf("/project/%s/%s/cicd/pipeline/%d/jobs", owner, project, localID)

	var jobs []Job
	if err := s.client.Get(path, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
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
	switch status {
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

// StatusColor returns color for terminal output
func StatusColor(status string) string {
	switch status {
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

const ColorReset = "\033[0m"
