package api

import (
	"fmt"
	"net/url"
	"time"
)

// TagService handles tag API calls
type TagService struct {
	client *Client
}

// Tag represents a git tag in GitFlic
type Tag struct {
	Name         string    `json:"name"`
	FullName     string    `json:"fullName,omitempty"`
	ObjectID     string    `json:"objectId,omitempty"`
	CommitID     string    `json:"commitId,omitempty"`
	ShortMessage string    `json:"shortMessage,omitempty"`
	FullMessage  string    `json:"fullMessage,omitempty"`
	LightWeight  bool      `json:"lightWeight,omitempty"`
	PersonIdent  *Ident    `json:"personIdent,omitempty"`
	CreatedAt    time.Time `json:"createdAt,omitempty"`
}

// TagListResponse represents the paginated response from tag list API
type TagListResponse struct {
	Embedded struct {
		Tags []Tag `json:"tagList"`
	} `json:"_embedded"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages    int `json:"totalPages"`
		Number        int `json:"number"`
	} `json:"page"`
}

// CreateTagRequest specifies parameters for creating a tag
type CreateTagRequest struct {
	TagName    string `json:"tagName"`              // Name of the tag
	BranchName string `json:"branchName,omitempty"` // Branch to create tag on
	CommitID   string `json:"commitId,omitempty"`   // Or commit hash to tag
	Message    string `json:"message"`              // Tag message (required)
}

// List returns all tags for a project
func (s *TagService) List(owner, project string) ([]Tag, error) {
	path := fmt.Sprintf("/project/%s/%s/tag",
		url.PathEscape(owner),
		url.PathEscape(project))

	var resp TagListResponse
	if err := s.client.Get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Embedded.Tags, nil
}

// Get returns a specific tag by name
func (s *TagService) Get(owner, project, tagName string) (*Tag, error) {
	path := fmt.Sprintf("/project/%s/%s/tag/%s",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(tagName))

	var tag Tag
	if err := s.client.Get(path, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

// Create creates a new tag
func (s *TagService) Create(owner, project string, req *CreateTagRequest) (*Tag, error) {
	// GitFlic API: POST /project/{owner}/{project}/tag/create
	path := fmt.Sprintf("/project/%s/%s/tag/create",
		url.PathEscape(owner),
		url.PathEscape(project))

	var tag Tag
	if err := s.client.Post(path, req, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

// Delete deletes a tag by name
func (s *TagService) Delete(owner, project, tagName string) error {
	path := fmt.Sprintf("/project/%s/%s/tag/%s",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(tagName))

	return s.client.Delete(path)
}
