package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPipeline_SHA(t *testing.T) {
	tests := []struct {
		commitID string
		want     string
	}{
		{"abc123def456789", "abc123d"},
		{"short", "short"},
		{"1234567", "1234567"},
		{"12345678", "1234567"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.commitID, func(t *testing.T) {
			p := &Pipeline{CommitID: tt.commitID}
			if got := p.SHA(); got != tt.want {
				t.Errorf("SHA() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPipeline_NormalizedStatus(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"SUCCESS", "success"},
		{"FAILED", "failed"},
		{"RUNNING", "running"},
		{"PENDING", "pending"},
		{"CANCELED", "canceled"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			p := &Pipeline{Status: tt.status}
			if got := p.NormalizedStatus(); got != tt.want {
				t.Errorf("NormalizedStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJob_NormalizedStatus(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"SUCCESS", "success"},
		{"FAILED", "failed"},
		{"SKIPPED", "skipped"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			j := &Job{Status: tt.status}
			if got := j.NormalizedStatus(); got != tt.want {
				t.Errorf("NormalizedStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStatusIcon(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"success", "✓"},
		{"SUCCESS", "✓"},
		{"passed", "✓"},
		{"failed", "✗"},
		{"FAILED", "✗"},
		{"running", "⧖"},
		{"pending", "◯"},
		{"canceled", "⊘"},
		{"skipped", "↷"},
		{"unknown", "?"},
		{"", "?"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := StatusIcon(tt.status); got != tt.want {
				t.Errorf("StatusIcon(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestStatusColor(t *testing.T) {
	tests := []struct {
		status    string
		wantEmpty bool
	}{
		{"success", false},
		{"failed", false},
		{"running", false},
		{"pending", false},
		{"canceled", false},
		{"unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := StatusColor(tt.status)
			if tt.wantEmpty && got != "" {
				t.Errorf("StatusColor(%q) = %q, want empty", tt.status, got)
			}
			if !tt.wantEmpty && got == "" {
				t.Errorf("StatusColor(%q) = empty, want non-empty", tt.status)
			}
		})
	}
}

func TestPipelineService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/project/owner/repo/cicd/pipeline" {
			t.Errorf("path = %q", r.URL.Path)
		}

		response := `{
			"_embedded": {
				"restPipelineModelList": [
					{
						"id": "uuid-1",
						"localId": 100,
						"status": "SUCCESS",
						"ref": "main",
						"commitId": "abc123def",
						"duration": 120,
						"createdAt": "2026-02-05T10:00:00"
					},
					{
						"id": "uuid-2",
						"localId": 99,
						"status": "FAILED",
						"ref": "feature",
						"commitId": "def456",
						"duration": 60,
						"createdAt": "2026-02-05T09:00:00"
					}
				]
			},
			"page": {"size": 10, "totalElements": 2}
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	pipelines, err := client.Pipelines().List("owner", "repo")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pipelines) != 2 {
		t.Fatalf("got %d pipelines, want 2", len(pipelines))
	}

	if pipelines[0].LocalID != 100 {
		t.Errorf("pipelines[0].LocalID = %d, want 100", pipelines[0].LocalID)
	}
	if pipelines[0].Status != "SUCCESS" {
		t.Errorf("pipelines[0].Status = %q", pipelines[0].Status)
	}
	if pipelines[0].SHA() != "abc123d" {
		t.Errorf("pipelines[0].SHA() = %q", pipelines[0].SHA())
	}
}

func TestPipelineService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get uses List internally
		response := `{
			"_embedded": {
				"restPipelineModelList": [
					{"id": "uuid-1", "localId": 100, "status": "SUCCESS"},
					{"id": "uuid-2", "localId": 99, "status": "FAILED"}
				]
			}
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	// Test found
	p, err := client.Pipelines().Get("owner", "repo", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.LocalID != 100 {
		t.Errorf("LocalID = %d, want 100", p.LocalID)
	}

	// Test not found
	_, err = client.Pipelines().Get("owner", "repo", 999)
	if err == nil {
		t.Error("expected error for non-existent pipeline")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestPipelineService_Jobs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/project/owner/repo/cicd/pipeline/100/jobs" {
			t.Errorf("path = %q", r.URL.Path)
		}

		response := `{
			"_embedded": {
				"restPipelineJobModelList": [
					{
						"id": "job-1",
						"localId": 200,
						"name": "build",
						"stageName": "build",
						"status": "SUCCESS"
					},
					{
						"id": "job-2",
						"localId": 201,
						"name": "test",
						"stageName": "test",
						"status": "FAILED"
					}
				]
			}
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	jobs, err := client.Pipelines().Jobs("owner", "repo", 100)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("got %d jobs, want 2", len(jobs))
	}

	if jobs[0].Name != "build" {
		t.Errorf("jobs[0].Name = %q", jobs[0].Name)
	}
	if jobs[0].Stage != "build" {
		t.Errorf("jobs[0].Stage = %q, want build", jobs[0].Stage)
	}
	if jobs[1].Status != "FAILED" {
		t.Errorf("jobs[1].Status = %q", jobs[1].Status)
	}
}

func TestPipeline_JSONParsing(t *testing.T) {
	// Test real API response structure
	response := `{
		"id": "uuid-123",
		"localId": 412,
		"status": "FAILED",
		"duration": 2208,
		"commitId": "3218564dfddd5071c566236d7221ecc0352b7912",
		"ref": "master",
		"source": "PUSH",
		"createdAt": "2026-02-05T17:35:51.933746",
		"finishedAt": "2026-02-05T18:17:51.081636"
	}`

	var p Pipeline
	if err := json.Unmarshal([]byte(response), &p); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if p.LocalID != 412 {
		t.Errorf("LocalID = %d, want 412", p.LocalID)
	}
	if p.Status != "FAILED" {
		t.Errorf("Status = %q", p.Status)
	}
	if p.SHA() != "3218564" {
		t.Errorf("SHA() = %q", p.SHA())
	}
	if p.Duration != 2208 {
		t.Errorf("Duration = %d", p.Duration)
	}
	if p.CreatedAt.Year() != 2026 {
		t.Errorf("CreatedAt.Year() = %d", p.CreatedAt.Year())
	}
	if p.FinishedAt == nil {
		t.Error("FinishedAt is nil")
	}
}
