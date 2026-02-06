package api

import (
	"fmt"
	"net/url"
	"time"
)

// MergeRequestService handles merge request API calls
type MergeRequestService struct {
	client *Client
}

// Branch represents a git branch in GitFlic
type Branch struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Hash      string `json:"hash"`
	IsDeleted bool   `json:"isDeleted"`
}

// Status represents a merge request status
type Status struct {
	ID       string `json:"id"` // OPEN, MERGED, CANCELED, CLOSED
	Title    string `json:"title"`
	Color    string `json:"color"`
	HexColor string `json:"hexColor"`
}

// MergeRequest represents a GitFlic merge request
type MergeRequest struct {
	ID           string    `json:"id"`
	LocalID      int       `json:"localId"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	SourceBranch Branch    `json:"sourceBranch"`
	TargetBranch Branch    `json:"targetBranch"`
	Status       Status    `json:"status"`
	Author       User      `json:"createdBy"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	CanMerge     bool      `json:"canMerge"`
	HasConflicts bool      `json:"hasConflicts"`
}

// State returns normalized state string (open, merged, closed)
func (mr *MergeRequest) State() string {
	switch mr.Status.ID {
	case "OPEN", "OPENED":
		return "open"
	case "MERGED":
		return "merged"
	case "CANCELED", "CLOSED":
		return "closed"
	default:
		return mr.Status.ID
	}
}

// MRListResponse represents the paginated response from MR list API
type MRListResponse struct {
	Embedded struct {
		MergeRequests []MergeRequest `json:"mergeRequestModelList"`
	} `json:"_embedded"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages    int `json:"totalPages"`
		Number        int `json:"number"`
	} `json:"page"`
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

// BranchRef is a reference to a branch for API requests
type BranchRef struct {
	ID string `json:"id"`
}

// ProjectRef is a reference to a project for API requests
type ProjectRef struct {
	ID string `json:"id"`
}

// CreateMRRequest specifies the parameters for creating a merge request
type CreateMRRequest struct {
	Title              string     `json:"title"`
	Description        string     `json:"description"` // Required by GitFlic API
	SourceBranch       BranchRef  `json:"sourceBranch"`
	TargetBranch       BranchRef  `json:"targetBranch"`
	SourceProject      ProjectRef `json:"sourceProject"`
	TargetProject      ProjectRef `json:"targetProject"`
	RemoveSourceBranch bool       `json:"removeSourceBranch,omitempty"`
	IsDraft            bool       `json:"workInProgress,omitempty"`
	SquashCommit       bool       `json:"squashCommit,omitempty"`
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

	filterState := ""
	if opts != nil {
		filterState = opts.State

		// API supports: MERGED, CANCELED (not OPEN)
		// For "open" we fetch all and filter client-side
		params := url.Values{}
		switch opts.State {
		case "merged":
			params.Set("status", "MERGED")
		case "closed":
			params.Set("status", "CANCELED")
		}
		if q := params.Encode(); q != "" {
			path += "?" + q
		}
	}

	var resp MRListResponse
	if err := s.client.Get(path, &resp); err != nil {
		return nil, err
	}

	mrs := resp.Embedded.MergeRequests

	// Client-side filter for "open" (API doesn't support this filter)
	if filterState == "open" {
		filtered := make([]MergeRequest, 0)
		for _, mr := range mrs {
			if mr.Status.ID != "MERGED" && mr.Status.ID != "CANCELED" && mr.Status.ID != "CLOSED" {
				filtered = append(filtered, mr)
			}
		}
		mrs = filtered
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

// UpdateMRRequest specifies the parameters for updating a merge request
type UpdateMRRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	IsDraft     *bool  `json:"workInProgress,omitempty"`
}

// Update updates a merge request
func (s *MergeRequestService) Update(owner, project string, localID int, req *UpdateMRRequest) (*MergeRequest, error) {
	path := fmt.Sprintf("/project/%s/%s/merge-request/%d", owner, project, localID)

	var mr MergeRequest
	if err := s.client.Put(path, req, &mr); err != nil {
		return nil, err
	}
	return &mr, nil
}

// Reopen reopens a closed merge request
func (s *MergeRequestService) Reopen(owner, project string, localID int) error {
	path := fmt.Sprintf("/project/%s/%s/merge-request/%d/reopen", owner, project, localID)
	return s.client.Post(path, nil, nil)
}

// MRDiscussion represents a discussion thread on a merge request
type MRDiscussion struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Author    User      `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	Resolved  bool      `json:"resolved"`
	// Line comment fields (optional)
	NewLine *int    `json:"newLine,omitempty"`
	OldLine *int    `json:"oldLine,omitempty"`
	NewPath *string `json:"newPath,omitempty"`
	OldPath *string `json:"oldPath,omitempty"`
}

// MRDiscussionListResponse represents the response from listing discussions
type MRDiscussionListResponse struct {
	Embedded struct {
		Discussions []MRDiscussion `json:"mergeRequestDiscussionModelList"`
	} `json:"_embedded"`
}

// ListDiscussions returns all discussions for a merge request
func (s *MergeRequestService) ListDiscussions(owner, project string, localID int) ([]MRDiscussion, error) {
	path := fmt.Sprintf("/project/%s/%s/merge-request/%d/discussions", owner, project, localID)

	var resp MRDiscussionListResponse
	if err := s.client.Get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Embedded.Discussions, nil
}

// CreateDiscussionRequest specifies parameters for creating a discussion
type CreateDiscussionRequest struct {
	Message string `json:"message"`
	// Optional: for line-specific comments
	NewLine *int    `json:"newLine,omitempty"`
	OldLine *int    `json:"oldLine,omitempty"`
	NewPath *string `json:"newPath,omitempty"`
	OldPath *string `json:"oldPath,omitempty"`
}

// CreateDiscussion creates a new discussion on a merge request
func (s *MergeRequestService) CreateDiscussion(owner, project string, localID int, req *CreateDiscussionRequest) (*MRDiscussion, error) {
	path := fmt.Sprintf("/project/%s/%s/merge-request/%d/discussions/create", owner, project, localID)

	var discussion MRDiscussion
	if err := s.client.Post(path, req, &discussion); err != nil {
		return nil, err
	}
	return &discussion, nil
}
