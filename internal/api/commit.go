package api

import (
	"fmt"
	"net/url"
	"time"
)

// CommitService handles commit API calls
type CommitService struct {
	client *Client
}

// CommitDetail represents a detailed commit from GitFlic API
type CommitDetail struct {
	Hash           string    `json:"hash"`
	ShortHash      string    `json:"shortHash"`
	Message        string    `json:"message"`
	Author         *User     `json:"author,omitempty"`
	AuthorName     string    `json:"authorName"`
	AuthorEmail    string    `json:"authorEmail"`
	Committer      *User     `json:"committer,omitempty"`
	CommitterName  string    `json:"committerName"`
	CommitterEmail string    `json:"committerEmail"`
	CreatedAt      time.Time `json:"createdAt"`
	ParentHashes   []string  `json:"parentHashes,omitempty"`
}

// CommitListResponse represents the paginated response from commit list API
type CommitListResponse struct {
	Embedded struct {
		Commits []CommitDetail `json:"commitList"`
	} `json:"_embedded"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages    int `json:"totalPages"`
		Number        int `json:"number"`
	} `json:"page"`
}

// CommitListOptions specifies options for listing commits
type CommitListOptions struct {
	Ref     string // Branch or tag name
	Page    int
	PerPage int
}

// List returns commits for a project
func (s *CommitService) List(owner, project string, opts *CommitListOptions) ([]CommitDetail, error) {
	// GitFlic API uses /commits (plural) for listing
	path := fmt.Sprintf("/project/%s/%s/commits",
		url.PathEscape(owner),
		url.PathEscape(project))

	params := url.Values{}
	if opts != nil {
		if opts.Ref != "" {
			// GitFlic uses "branch" param, not "ref"
			params.Set("branch", opts.Ref)
		}
		if opts.Page > 0 {
			params.Set("page", fmt.Sprintf("%d", opts.Page))
		}
		if opts.PerPage > 0 {
			params.Set("size", fmt.Sprintf("%d", opts.PerPage))
		}
	}

	if q := params.Encode(); q != "" {
		path += "?" + q
	}

	var resp CommitListResponse
	if err := s.client.Get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Embedded.Commits, nil
}

// Get returns a specific commit by hash
func (s *CommitService) Get(owner, project, hash string) (*CommitDetail, error) {
	path := fmt.Sprintf("/project/%s/%s/commit/%s",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(hash))

	var commit CommitDetail
	if err := s.client.Get(path, &commit); err != nil {
		return nil, err
	}
	return &commit, nil
}

// CommitDiff represents diff information for a commit
type CommitDiff struct {
	FilePath     string `json:"filePath"`
	OldPath      string `json:"oldPath,omitempty"`
	ChangeType   string `json:"changeType"` // ADD, MODIFY, DELETE, RENAME
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	DiffContent  string `json:"diffContent,omitempty"`
}

// CommitDiffResponse represents the response from commit diff API
type CommitDiffResponse struct {
	Diffs []CommitDiff `json:"diffs"`
}

// Diff returns the diff for a commit
func (s *CommitService) Diff(owner, project, hash string) ([]CommitDiff, error) {
	path := fmt.Sprintf("/project/%s/%s/commit/%s/diff",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(hash))

	var resp CommitDiffResponse
	if err := s.client.Get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Diffs, nil
}
