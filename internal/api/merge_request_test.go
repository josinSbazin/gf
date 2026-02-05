package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMergeRequest_State(t *testing.T) {
	tests := []struct {
		statusID string
		want     string
	}{
		{"OPEN", "open"},
		{"MERGED", "merged"},
		{"CANCELED", "closed"},
		{"CLOSED", "closed"},
		{"UNKNOWN", "UNKNOWN"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.statusID, func(t *testing.T) {
			mr := &MergeRequest{
				Status: Status{ID: tt.statusID},
			}
			if got := mr.State(); got != tt.want {
				t.Errorf("State() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMergeRequestService_List(t *testing.T) {
	tests := []struct {
		name       string
		opts       *MRListOptions
		wantPath   string
		response   string
		wantCount  int
		wantErr    bool
	}{
		{
			name:     "no filter",
			opts:     nil,
			wantPath: "/project/owner/repo/merge-request/list",
			response: `{"_embedded":{"mergeRequestModelList":[
				{"localId":1,"title":"MR 1","status":{"id":"OPEN"}},
				{"localId":2,"title":"MR 2","status":{"id":"MERGED"}}
			]}}`,
			wantCount: 2,
		},
		{
			name:     "filter merged",
			opts:     &MRListOptions{State: "merged"},
			wantPath: "/project/owner/repo/merge-request/list?status=MERGED",
			response: `{"_embedded":{"mergeRequestModelList":[
				{"localId":2,"title":"MR 2","status":{"id":"MERGED"}}
			]}}`,
			wantCount: 1,
		},
		{
			name:     "filter closed",
			opts:     &MRListOptions{State: "closed"},
			wantPath: "/project/owner/repo/merge-request/list?status=CANCELED",
			response: `{"_embedded":{"mergeRequestModelList":[]}}`,
			wantCount: 0,
		},
		{
			name:     "filter open - client side",
			opts:     &MRListOptions{State: "open"},
			wantPath: "/project/owner/repo/merge-request/list",
			response: `{"_embedded":{"mergeRequestModelList":[
				{"localId":1,"title":"MR 1","status":{"id":"OPEN"}},
				{"localId":2,"title":"MR 2","status":{"id":"MERGED"}}
			]}}`,
			wantCount: 1, // Only OPEN should remain after client-side filter
		},
		{
			name:     "filter all",
			opts:     &MRListOptions{State: "all"},
			wantPath: "/project/owner/repo/merge-request/list",
			response: `{"_embedded":{"mergeRequestModelList":[
				{"localId":1,"title":"MR 1","status":{"id":"OPEN"}},
				{"localId":2,"title":"MR 2","status":{"id":"MERGED"}}
			]}}`,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify path
				if r.URL.String() != tt.wantPath {
					t.Errorf("path = %q, want %q", r.URL.String(), tt.wantPath)
				}

				// Verify auth header
				if auth := r.Header.Get("Authorization"); auth != "token test-token" {
					t.Errorf("Authorization = %q, want %q", auth, "token test-token")
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token")
			mrs, err := client.MergeRequests().List("owner", "repo", tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(mrs) != tt.wantCount {
				t.Errorf("got %d MRs, want %d", len(mrs), tt.wantCount)
			}
		})
	}
}

func TestMergeRequestService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/project/owner/repo/merge-request/123" {
			t.Errorf("path = %q, want /project/owner/repo/merge-request/123", r.URL.Path)
		}

		response := `{
			"id": "uuid-123",
			"localId": 123,
			"title": "Test MR",
			"description": "Test description",
			"status": {"id": "OPEN"},
			"sourceBranch": {"title": "feature"},
			"targetBranch": {"title": "main"},
			"createdBy": {"username": "user1"},
			"hasConflicts": false,
			"canMerge": true
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	mr, err := client.MergeRequests().Get("owner", "repo", 123)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mr.LocalID != 123 {
		t.Errorf("LocalID = %d, want 123", mr.LocalID)
	}
	if mr.Title != "Test MR" {
		t.Errorf("Title = %q, want %q", mr.Title, "Test MR")
	}
	if mr.State() != "open" {
		t.Errorf("State() = %q, want %q", mr.State(), "open")
	}
	if mr.SourceBranch.Title != "feature" {
		t.Errorf("SourceBranch.Title = %q, want %q", mr.SourceBranch.Title, "feature")
	}
}

func TestMergeRequest_JSONParsing(t *testing.T) {
	// Test real API response structure
	response := `{
		"id": "uuid-123",
		"localId": 42,
		"title": "Feature: Add login",
		"description": "Adds login functionality",
		"sourceBranch": {
			"id": "feature-login",
			"title": "feature/login",
			"hash": "abc123",
			"isDeleted": false
		},
		"targetBranch": {
			"id": "main",
			"title": "main",
			"hash": "def456",
			"isDeleted": false
		},
		"status": {
			"id": "MERGED",
			"title": "Merged",
			"color": "success"
		},
		"createdBy": {
			"id": "user-uuid",
			"username": "developer",
			"name": "John",
			"surname": "Doe"
		},
		"createdAt": "2026-02-05T10:00:00Z",
		"updatedAt": "2026-02-05T12:00:00Z",
		"canMerge": false,
		"hasConflicts": false
	}`

	var mr MergeRequest
	if err := json.Unmarshal([]byte(response), &mr); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if mr.LocalID != 42 {
		t.Errorf("LocalID = %d, want 42", mr.LocalID)
	}
	if mr.SourceBranch.Title != "feature/login" {
		t.Errorf("SourceBranch.Title = %q", mr.SourceBranch.Title)
	}
	if mr.Author.Username != "developer" {
		t.Errorf("Author.Username = %q", mr.Author.Username)
	}
	if mr.State() != "merged" {
		t.Errorf("State() = %q, want merged", mr.State())
	}
}
