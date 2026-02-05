package api

import (
	"fmt"
	"time"
)

// ProjectService handles project-related API calls
type ProjectService struct {
	client *Client
}

// Project represents a GitFlic project
type Project struct {
	UUID        string    `json:"uuid"`
	Alias       string    `json:"alias"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IsPrivate   bool      `json:"isPrivate"`
	OwnerAlias  string    `json:"ownerAlias"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	StarsCount  int       `json:"starsCount"`
	ForksCount  int       `json:"forksCount"`
	Language    string    `json:"language"`
}

// Get returns a project by owner and name
func (s *ProjectService) Get(owner, project string) (*Project, error) {
	var p Project
	path := fmt.Sprintf("/project/%s/%s", owner, project)
	if err := s.client.Get(path, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// MyProjects returns projects belonging to the authenticated user
func (s *ProjectService) MyProjects() ([]Project, error) {
	var projects []Project
	if err := s.client.Get("/project/my", &projects); err != nil {
		return nil, err
	}
	return projects, nil
}
