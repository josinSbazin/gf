package api

import (
	"fmt"
	"time"
)

// MergeRequestService handles merge request API calls
type MergeRequestService struct {
	client *Client
}

// MergeRequest represents a GitFlic merge request
type MergeRequest struct {
	UUID           string    `json:"uuid"`
	LocalID        int       `json:"localId"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	State          string    `json:"state"` // open, merged, closed
	SourceBranch   string    `json:"sourceBranch"`
	TargetBranch   string    `json:"targetBranch"`
	Author         User      `json:"author"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	MergedAt       *time.Time `json:"mergedAt,omitempty"`
	ClosedAt       *time.Time `json:"closedAt,omitempty"`
	IsDraft        bool      `json:"isDraft"`
	CanMerge       bool      `json:"canMerge"`
	HasConflicts   bool      `json:"hasConflicts"`
	CommentsCount  int       `json:"commentsCount"`
	ApprovalsCount int       `json:"approvalsCount"`
}

// MRListOptions specifies options for listing merge requests
type MRListOptions struct {
	State        string // open, merged, closed, all
	SourceBranch string
	TargetBranch string
	AuthorAlias  string
	Page         int
	PerPage      int
}

// CreateMRRequest specifies the parameters for creating a merge request
type CreateMRRequest struct {
	Title              string `json:"title"`
	Description        string `json:"description,omitempty"`
	SourceBranch       string `json:"sourceBranch"`
	TargetBranch       string `json:"targetBranch"`
	RemoveSourceBranch bool   `json:"removeSourceBranch,omitempty"`
	IsDraft            bool   `json:"workInProgress,omitempty"`
	SquashCommit       bool   `json:"squashCommit,omitempty"`
}

// MergeMRRequest specifies the parameters for merging a merge request
type MergeMRRequest struct {
	SquashCommit       bool   `json:"squashCommit,omitempty"`
	RemoveSourceBranch bool   `json:"removeSourceBranch,omitempty"`
	MergeCommitMessage string `json:"mergeCommitMessage,omitempty"`
}

// List returns merge requests for a project
func (s *MergeRequestService) List(owner, project string, opts *MRListOptions) ([]MergeRequest, error) {
	path := fmt.Sprintf("/project/%s/%s/merge-request/list", owner, project)

	// TODO: Add query params for filtering

	var mrs []MergeRequest
	if err := s.client.Get(path, &mrs); err != nil {
		return nil, err
	}
	return mrs, nil
}

// Get returns a specific merge request
func (s *MergeRequestService) Get(owner, project string, localID int) (*MergeRequest, error) {
	path := fmt.Sprintf("/project/%s/%s/merge-request/%d", owner, project, localID)

	var mr MergeRequest
	if err := s.client.Get(path, &mr); err != nil {
		return nil, err
	}
	return &mr, nil
}

// Create creates a new merge request
func (s *MergeRequestService) Create(owner, project string, req *CreateMRRequest) (*MergeRequest, error) {
	path := fmt.Sprintf("/project/%s/%s/merge-request", owner, project)

	var mr MergeRequest
	if err := s.client.Post(path, req, &mr); err != nil {
		return nil, err
	}
	return &mr, nil
}

// Merge merges a merge request
func (s *MergeRequestService) Merge(owner, project string, localID int, req *MergeMRRequest) error {
	path := fmt.Sprintf("/project/%s/%s/merge-request/%d/merge", owner, project, localID)
	return s.client.Post(path, req, nil)
}

// Approve approves a merge request
func (s *MergeRequestService) Approve(owner, project string, localID int) error {
	path := fmt.Sprintf("/project/%s/%s/merge-request/%d/approve", owner, project, localID)
	return s.client.Post(path, nil, nil)
}

// Close closes a merge request without merging
func (s *MergeRequestService) Close(owner, project string, localID int) error {
	path := fmt.Sprintf("/project/%s/%s/merge-request/%d/close", owner, project, localID)
	return s.client.Post(path, nil, nil)
}
