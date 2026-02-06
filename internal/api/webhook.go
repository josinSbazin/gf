package api

import (
	"fmt"
	"net/url"
	"time"
)

// WebhookService handles webhook API calls
type WebhookService struct {
	client *Client
}

// Webhook represents a GitFlic webhook
type Webhook struct {
	ID        string         `json:"id"`
	URL       string         `json:"url"`
	Secret    string         `json:"secret,omitempty"`
	Events    *WebhookEvents `json:"events"`
	ProjectID string         `json:"projectId,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// WebhookListResponse represents the paginated response from webhook list API
type WebhookListResponse struct {
	Embedded struct {
		Webhooks []Webhook `json:"webhookList"`
	} `json:"_embedded"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages    int `json:"totalPages"`
		Number        int `json:"number"`
	} `json:"page"`
}

// WebhookEvents represents event flags for a webhook
// GitFlic uses boolean flags for each event type
type WebhookEvents struct {
	CollaboratorAdd     bool `json:"COLLABORATOR_ADD,omitempty"`
	CollaboratorDelete  bool `json:"COLLABORATOR_DELETE,omitempty"`
	IssueCreate         bool `json:"ISSUE_CREATE,omitempty"`
	IssueUpdate         bool `json:"ISSUE_UPDATE,omitempty"`
	NewIssueNote        bool `json:"NEW_ISSUE_NOTE,omitempty"`
	MergeRequestCreate  bool `json:"MERGE_REQUEST_CREATE,omitempty"`
	MergeRequestUpdate  bool `json:"MERGE_REQUEST_UPDATE,omitempty"`
	Merge               bool `json:"MERGE,omitempty"`
	PipelineNew         bool `json:"PIPELINE_NEW,omitempty"`
	PipelineSuccess     bool `json:"PIPELINE_SUCCESS,omitempty"`
	PipelineFail        bool `json:"PIPELINE_FAIL,omitempty"`
	TagCreate           bool `json:"TAG_CREATE,omitempty"`
	TagDelete           bool `json:"TAG_DELETE,omitempty"`
	BranchCreate        bool `json:"BRANCH_CREATE,omitempty"`
	BranchUpdate        bool `json:"BRANCH_UPDATE,omitempty"`
	BranchDelete        bool `json:"BRANCH_DELETE,omitempty"`
	DiscussionCreate    bool `json:"DISCUSSION_CREATE,omitempty"`
	Push                bool `json:"PUSH,omitempty"`
	ReleaseCreate       bool `json:"RELEASE_CREATE,omitempty"`
	ReleaseUpdate       bool `json:"RELEASE_UPDATE,omitempty"`
	ReleaseDelete       bool `json:"RELEASE_DELETE,omitempty"`
}

// CreateWebhookRequest specifies parameters for creating a webhook
type CreateWebhookRequest struct {
	URL    string         `json:"url"`
	Secret string         `json:"secret"`
	Events *WebhookEvents `json:"events"`
}

// UpdateWebhookRequest specifies parameters for updating a webhook
type UpdateWebhookRequest struct {
	URL    string   `json:"url,omitempty"`
	Secret string   `json:"secret,omitempty"`
	Events []string `json:"events,omitempty"`
	Active *bool    `json:"active,omitempty"`
}

// List returns all webhooks for a project
func (s *WebhookService) List(owner, project string) ([]Webhook, error) {
	// GitFlic API: GET /project/{owner}/{project}/setting/webhook
	path := fmt.Sprintf("/project/%s/%s/setting/webhook",
		url.PathEscape(owner),
		url.PathEscape(project))

	var resp WebhookListResponse
	if err := s.client.Get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Embedded.Webhooks, nil
}

// Get returns a specific webhook by ID
func (s *WebhookService) Get(owner, project, webhookID string) (*Webhook, error) {
	// GitFlic API: GET /project/{owner}/{project}/setting/webhook/{id}
	path := fmt.Sprintf("/project/%s/%s/setting/webhook/%s",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(webhookID))

	var webhook Webhook
	if err := s.client.Get(path, &webhook); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// Create creates a new webhook
func (s *WebhookService) Create(owner, project string, req *CreateWebhookRequest) (*Webhook, error) {
	// GitFlic API: POST /project/{owner}/{project}/setting/webhook
	path := fmt.Sprintf("/project/%s/%s/setting/webhook",
		url.PathEscape(owner),
		url.PathEscape(project))

	var webhook Webhook
	if err := s.client.Post(path, req, &webhook); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// Update updates a webhook
func (s *WebhookService) Update(owner, project, webhookID string, req *UpdateWebhookRequest) (*Webhook, error) {
	// GitFlic API: POST /project/{owner}/{project}/setting/webhook/{id}
	path := fmt.Sprintf("/project/%s/%s/setting/webhook/%s",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(webhookID))

	var webhook Webhook
	// GitFlic uses POST for updates, not PUT
	if err := s.client.Post(path, req, &webhook); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// Delete deletes a webhook
func (s *WebhookService) Delete(owner, project, webhookID string) error {
	// GitFlic API: POST /project/{owner}/{project}/setting/webhook/{id}/delete
	path := fmt.Sprintf("/project/%s/%s/setting/webhook/%s/delete",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(webhookID))

	// GitFlic uses POST to /delete endpoint, not DELETE method
	return s.client.Post(path, nil, nil)
}

// Test triggers a test webhook
func (s *WebhookService) Test(owner, project, webhookID string) error {
	// Note: Test endpoint not documented - this may not work
	path := fmt.Sprintf("/project/%s/%s/setting/webhook/%s/test",
		url.PathEscape(owner),
		url.PathEscape(project),
		url.PathEscape(webhookID))

	return s.client.Post(path, nil, nil)
}
