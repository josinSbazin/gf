package api

import (
	"fmt"
	"net/url"
	"time"
)

// BranchService handles branch API calls
type BranchService struct {
	client *Client
}

// BranchDetail represents detailed branch information from GitFlic API
type BranchDetail struct {
	Name       string  `json:"name"`
	FullName   string  `json:"fullName,omitempty"`
	Hash       string  `json:"hash,omitempty"`
	IsDefault  bool    `json:"default"`
	Protected  bool    `json:"protected,omitempty"`
	Merged     bool    `json:"merged,omitempty"`
	Work       bool    `json:"work,omitempty"`
	LastCommit *Commit `json:"lastCommit,omitempty"`
}

// Commit represents a git commit (basic info for branch listing)
type Commit struct {
	Hash         string    `json:"hash"`
	Message      string    `json:"message"`
	ShortMessage string    `json:"shortMessage,omitempty"`
	Author       *User     `json:"author,omitempty"`
	AuthorIdent  *Ident    `json:"authorIdent,omitempty"`
	CreatedAt    time.Time `json:"createdAt,omitempty"`
}

// Ident represents author/committer identity
type Ident struct {
	Name         string    `json:"name"`
	EmailAddress string    `json:"emailAddress"`
	When         time.Time `json:"when,omitempty"`
}

// BranchListResponse represents the paginated response from branch list API
type BranchListResponse struct {
	Embedded struct {
		Branches []BranchDetail `json:"branchList"`
	} `json:"_embedded"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages    int `json:"totalPages"`
		Number        int `json:"number"`
	} `json:"page"`
}

// CreateBranchRequest specifies parameters for creating a branch
type CreateBranchRequest struct {
	NewBranch    string `json:"newBranch"`    // Name of the new branch
	OriginBranch string `json:"originBranch"` // Source branch to create from
}

// List returns all branches for a project
func (s *BranchService) List(owner, project string) ([]BranchDetail, error) {
	path := fmt.Sprintf("/project/%s/%s/branch", url.PathEscape(owner), url.PathEscape(project))

	var resp BranchListResponse
	if err := s.client.Get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Embedded.Branches, nil
}

// Get returns a specific branch by name
func (s *BranchService) Get(owner, project, branchName string) (*BranchDetail, error) {
	// GitFlic API: GET /project/{owner}/{project}/branch?branchName={name}
	path := fmt.Sprintf("/project/%s/%s/branch?branchName=%s",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.QueryEscape(branchName))

	var branch BranchDetail
	if err := s.client.Get(path, &branch); err != nil {
		return nil, err
	}
	return &branch, nil
}

// GetDefault returns the default branch for a project
func (s *BranchService) GetDefault(owner, project string) (*BranchDetail, error) {
	path := fmt.Sprintf("/project/%s/%s/branch/default",
		url.PathEscape(owner),
		url.PathEscape(project))

	var branch BranchDetail
	if err := s.client.Get(path, &branch); err != nil {
		return nil, err
	}
	return &branch, nil
}

// Create creates a new branch
func (s *BranchService) Create(owner, project string, req *CreateBranchRequest) (*BranchDetail, error) {
	path := fmt.Sprintf("/project/%s/%s/branch",
		url.PathEscape(owner),
		url.PathEscape(project))

	var branch BranchDetail
	if err := s.client.Post(path, req, &branch); err != nil {
		return nil, err
	}
	return &branch, nil
}

// Delete deletes a branch by name
func (s *BranchService) Delete(owner, project, branchName string) error {
	path := fmt.Sprintf("/project/%s/%s/branch?branchName=%s",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.QueryEscape(branchName))

	return s.client.Delete(path)
}
