package api

import (
	"fmt"
	"io"
	"net/url"
	"time"
)

// ReleaseService handles release API calls
type ReleaseService struct {
	client *Client
}

// Release represents a GitFlic release
type Release struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	TagName      string    `json:"tagName"`
	CommitID     string    `json:"commitId,omitempty"`
	IsDraft      bool      `json:"isDraft"`
	IsPrerelease bool      `json:"preRelease"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt,omitempty"`
	PublishedAt  time.Time `json:"publishedAt,omitempty"`
	Author       User      `json:"createdBy,omitempty"`
}

// ReleaseListResponse represents the paginated response from release list API
type ReleaseListResponse struct {
	Embedded struct {
		Releases []Release `json:"releaseTagModelList"`
	} `json:"_embedded"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages    int `json:"totalPages"`
		Number        int `json:"number"`
	} `json:"page"`
}

// ReleaseListOptions specifies options for listing releases
type ReleaseListOptions struct {
	Page    int
	PerPage int
}

// CreateReleaseRequest specifies the parameters for creating a release
type CreateReleaseRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	TagName     string `json:"tagName"`
	IsDraft     bool   `json:"isDraft,omitempty"`
	IsPrerelease bool  `json:"isPrerelease,omitempty"`
}

// List returns releases for a project
func (s *ReleaseService) List(owner, project string, opts *ReleaseListOptions) ([]Release, int, error) {
	path := fmt.Sprintf("/project/%s/%s/release", owner, project)

	// Add pagination params if provided
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", fmt.Sprintf("%d", opts.Page))
		}
		if opts.PerPage > 0 {
			params.Set("size", fmt.Sprintf("%d", opts.PerPage))
		}
		if q := params.Encode(); q != "" {
			path += "?" + q
		}
	}

	var resp ReleaseListResponse
	if err := s.client.Get(path, &resp); err != nil {
		return nil, 0, err
	}

	return resp.Embedded.Releases, resp.Page.TotalElements, nil
}

// Get returns a specific release by tag name
func (s *ReleaseService) Get(owner, project, tagName string) (*Release, error) {
	// GitFlic API requires filtering by tagName query parameter
	path := fmt.Sprintf("/project/%s/%s/release?tagName=%s",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.QueryEscape(tagName))

	var resp ReleaseListResponse
	if err := s.client.Get(path, &resp); err != nil {
		return nil, err
	}

	if len(resp.Embedded.Releases) == 0 {
		return nil, ErrNotFound
	}

	return &resp.Embedded.Releases[0], nil
}

// Create creates a new release
func (s *ReleaseService) Create(owner, project string, req *CreateReleaseRequest) (*Release, error) {
	path := fmt.Sprintf("/project/%s/%s/release", owner, project)

	var release Release
	if err := s.client.Post(path, req, &release); err != nil {
		return nil, err
	}
	return &release, nil
}

// Delete deletes a release by tag name
func (s *ReleaseService) Delete(owner, project, tagName string) error {
	// First get the release to obtain its ID
	release, err := s.Get(owner, project, tagName)
	if err != nil {
		return err
	}

	// Delete using the release ID
	path := fmt.Sprintf("/project/%s/%s/release/%s",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(release.ID))
	return s.client.Delete(path)
}

// UpdateReleaseRequest specifies the parameters for updating a release
type UpdateReleaseRequest struct {
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
	TagName      string `json:"tagName,omitempty"`
	IsDraft      *bool  `json:"isDraft,omitempty"`
	IsPrerelease *bool  `json:"preRelease,omitempty"` // API uses "preRelease"
}

// Update updates a release by tag name
func (s *ReleaseService) Update(owner, project, tagName string, req *UpdateReleaseRequest) (*Release, error) {
	// First get the release to obtain its ID and current values
	existing, err := s.Get(owner, project, tagName)
	if err != nil {
		return nil, err
	}

	// Build complete update payload - API requires all fields
	payload := map[string]interface{}{
		"title":       existing.Title,
		"description": existing.Description,
		"tagName":     tagName,
		"preRelease":  existing.IsPrerelease,
	}

	// Override with provided values
	if req.Title != "" {
		payload["title"] = req.Title
	}
	if req.Description != "" {
		payload["description"] = req.Description
	}
	if req.IsPrerelease != nil {
		payload["preRelease"] = *req.IsPrerelease
	}

	// Update using the release ID
	path := fmt.Sprintf("/project/%s/%s/release/%s",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(existing.ID))

	var release Release
	if err := s.client.Put(path, payload, &release); err != nil {
		return nil, err
	}
	return &release, nil
}

// ReleaseAsset represents a file attached to a release
type ReleaseAsset struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ContentType string    `json:"contentType"`
	DownloadURL string    `json:"downloadUrl"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ReleaseAssetListResponse represents the response from listing assets
type ReleaseAssetListResponse struct {
	Embedded struct {
		Assets []ReleaseAsset `json:"releaseAssetModelList"`
	} `json:"_embedded"`
}

// ListAssets returns all assets for a release
func (s *ReleaseService) ListAssets(owner, project, tagName string) ([]ReleaseAsset, error) {
	path := fmt.Sprintf("/project/%s/%s/release/%s/asset",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(tagName))

	var resp ReleaseAssetListResponse
	if err := s.client.Get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Embedded.Assets, nil
}

// GetAssetDownloadURL returns the download URL for an asset
func (s *ReleaseService) GetAssetDownloadURL(owner, project, tagName, assetName string) string {
	return fmt.Sprintf("%s/project/%s/%s/release/%s/asset/%s/download",
		s.client.BaseURL,
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(tagName),
		url.PathEscape(assetName))
}

// UploadAsset uploads a file as a release asset
func (s *ReleaseService) UploadAsset(owner, project, tagName, fileName string, fileData io.Reader) (*ReleaseAsset, error) {
	// First get the release to obtain its UUID
	release, err := s.Get(owner, project, tagName)
	if err != nil {
		return nil, err
	}

	// GitFlic API: POST /project/{owner}/{project}/release/{releaseUuid}/file
	// Form field name is "files", not "file"
	path := fmt.Sprintf("/project/%s/%s/release/%s/file",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(release.ID))

	var asset ReleaseAsset
	if err := s.client.UploadFile(path, "files", fileName, fileData, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

// DeleteAsset deletes a release asset
func (s *ReleaseService) DeleteAsset(owner, project, tagName, assetName string) error {
	path := fmt.Sprintf("/project/%s/%s/release/%s/asset/%s",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(tagName),
		url.PathEscape(assetName))

	return s.client.Delete(path)
}

// DownloadAsset downloads a release asset
func (s *ReleaseService) DownloadAsset(owner, project, tagName, assetName string) (io.ReadCloser, string, error) {
	path := fmt.Sprintf("/project/%s/%s/release/%s/asset/%s/download",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(tagName),
		url.PathEscape(assetName))

	return s.client.DownloadFile(path)
}
