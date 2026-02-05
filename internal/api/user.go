package api

// UserService handles user-related API calls
type UserService struct {
	client *Client
}

// User represents a GitFlic user
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Surname   string `json:"surname"`
	FullName  string `json:"fullName"`
	AvatarURL string `json:"avatar"`
}

// Alias returns username for compatibility
func (u *User) Alias() string {
	return u.Username
}

// Me returns the authenticated user
func (s *UserService) Me() (*User, error) {
	var user User
	if err := s.client.Get("/user/me", &user); err != nil {
		return nil, err
	}
	return &user, nil
}
