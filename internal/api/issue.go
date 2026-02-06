package api

import (
	"fmt"
	"net/url"
)

// IssueService handles issue API calls
type IssueService struct {
	client *Client
}

// IssueStatus represents an issue status
type IssueStatus struct {
	ID    string `json:"id"` // OPEN, CLOSED, IN_PROGRESS, etc.
	Title string `json:"title"`
	Color string `json:"color"`
}

// Issue represents a GitFlic issue
type Issue struct {
	ID          string      `json:"id"`
	LocalID     int         `json:"localId"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Status      IssueStatus `json:"status"`
	Author      User        `json:"updatedBy"` // GitFlic uses updatedBy for author in responses
	CreatedAt   FlexTime    `json:"createdAt"`
	UpdatedAt   FlexTime    `json:"updatedAt"`
}

// State returns normalized state string (open, closed)
func (i *Issue) State() string {
	switch i.Status.ID {
	case "OPEN", "OPENED", "IN_PROGRESS":
		return "open"
	case "CLOSED", "RESOLVED", "DONE":
		return "closed"
	default:
		return i.Status.ID
	}
}

// IssueListResponse represents the paginated response from issue list API
type IssueListResponse struct {
	Embedded struct {
		Issues []Issue `json:"issueModelList"`
	} `json:"_embedded"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages    int `json:"totalPages"`
		Number        int `json:"number"`
	} `json:"page"`
}

// IssueListOptions specifies options for listing issues
type IssueListOptions struct {
	State   string // open, closed, all
	Page    int
	PerPage int
}

// CreateIssueRequest specifies the parameters for creating an issue
type CreateIssueRequest struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	AssignedUsers []string `json:"assignedUsers"` // Required by GitFlic API (can be empty)
}

// List returns issues for a project
func (s *IssueService) List(owner, project string, opts *IssueListOptions) ([]Issue, error) {
	path := fmt.Sprintf("/project/%s/%s/issue", owner, project)

	params := url.Values{}
	params.Set("page", "0")
	params.Set("size", "100")

	filterState := ""
	if opts != nil {
		filterState = opts.State
		if opts.Page > 0 {
			params.Set("page", fmt.Sprintf("%d", opts.Page))
		}
		if opts.PerPage > 0 {
			params.Set("size", fmt.Sprintf("%d", opts.PerPage))
		}
		// API may support status filter
		switch opts.State {
		case "closed":
			params.Set("status", "CLOSED")
		case "open":
			params.Set("status", "OPEN")
		}
	}

	path += "?" + params.Encode()

	var resp IssueListResponse
	if err := s.client.Get(path, &resp); err != nil {
		return nil, err
	}

	issues := resp.Embedded.Issues

	// Note: Server-side filtering is done via params.Set("status", ...)
	// Client-side fallback only if API doesn't respect the filter
	// This is detected by checking if we got unexpected results
	if filterState != "" && filterState != "all" && len(issues) > 0 {
		// Check if first result matches filter - if not, API didn't filter
		needsClientFilter := false
		if filterState == "open" && issues[0].State() != "open" {
			needsClientFilter = true
		} else if filterState == "closed" && issues[0].State() != "closed" {
			needsClientFilter = true
		}

		if needsClientFilter {
			filtered := make([]Issue, 0, len(issues))
			for _, issue := range issues {
				if issue.State() == filterState {
					filtered = append(filtered, issue)
				}
			}
			issues = filtered
		}
	}

	return issues, nil
}

// Get returns a specific issue
func (s *IssueService) Get(owner, project string, localID int) (*Issue, error) {
	path := fmt.Sprintf("/project/%s/%s/issue/%d", owner, project, localID)

	var issue Issue
	if err := s.client.Get(path, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// Create creates a new issue
func (s *IssueService) Create(owner, project string, req *CreateIssueRequest) (*Issue, error) {
	path := fmt.Sprintf("/project/%s/%s/issue", owner, project)

	// Ensure assignedUsers is set (required by GitFlic API)
	if req.AssignedUsers == nil {
		req.AssignedUsers = []string{}
	}

	var issue Issue
	if err := s.client.Post(path, req, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// Close closes an issue
func (s *IssueService) Close(owner, project string, localID int) error {
	path := fmt.Sprintf("/project/%s/%s/issue/%d/close", owner, project, localID)
	return s.client.Post(path, nil, nil)
}

// Reopen reopens a closed issue
func (s *IssueService) Reopen(owner, project string, localID int) error {
	path := fmt.Sprintf("/project/%s/%s/issue/%d/reopen", owner, project, localID)
	return s.client.Post(path, nil, nil)
}

// UpdateIssueRequest specifies parameters for updating an issue
type UpdateIssueRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// Update updates an issue
func (s *IssueService) Update(owner, project string, localID int, req *UpdateIssueRequest) (*Issue, error) {
	path := fmt.Sprintf("/project/%s/%s/issue/%d", owner, project, localID)

	var issue Issue
	if err := s.client.Put(path, req, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// Delete deletes an issue
func (s *IssueService) Delete(owner, project string, localID int) error {
	path := fmt.Sprintf("/project/%s/%s/issue/%d", owner, project, localID)
	return s.client.Delete(path)
}

// IssueComment represents a comment on an issue
type IssueComment struct {
	ID        string   `json:"id"`
	Note      string   `json:"note"`
	Author    User     `json:"createdBy"`
	CreatedAt FlexTime `json:"createdAt"`
	UpdatedAt FlexTime `json:"updatedAt"`
}

// IssueCommentListResponse represents the response from listing comments
type IssueCommentListResponse struct {
	Embedded struct {
		Comments []IssueComment `json:"issueNoteModelList"`
	} `json:"_embedded"`
}

// ListComments returns all comments for an issue
func (s *IssueService) ListComments(owner, project string, localID int) ([]IssueComment, error) {
	path := fmt.Sprintf("/project/%s/%s/issue-discussion/%d", owner, project, localID)

	var resp IssueCommentListResponse
	if err := s.client.Get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Embedded.Comments, nil
}

// CreateCommentRequest specifies parameters for creating a comment
type CreateCommentRequest struct {
	Note string `json:"note"`
}

// CreateComment creates a new comment on an issue
func (s *IssueService) CreateComment(owner, project string, localID int, note string) (*IssueComment, error) {
	path := fmt.Sprintf("/project/%s/%s/issue-discussion/%d/create", owner, project, localID)

	req := &CreateCommentRequest{Note: note}
	var comment IssueComment
	if err := s.client.Post(path, req, &comment); err != nil {
		return nil, err
	}
	return &comment, nil
}
