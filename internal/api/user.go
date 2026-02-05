package api

import "time"

// UserService handles user-related API calls
type UserService struct {
	client *Client
}

// User represents a GitFlic user
type User struct {
	UUID      string    `json:"uuid"`
	Alias     string    `json:"alias"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatarUrl"`
	CreatedAt time.Time `json:"createdAt"`
}

// Me returns the authenticated user
func (s *UserService) Me() (*User, error) {
	var user User
	if err := s.client.Get("/user/me", &user); err != nil {
		return nil, err
	}
	return &user, nil
}
