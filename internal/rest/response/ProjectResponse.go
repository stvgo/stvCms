package response

import "time"

type ProjectResponse struct {
	Id          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	EmbedURL    string    `json:"embed_url"`
	ImageURL    string    `json:"image_url"`
	GitHubURL   string    `json:"github_url"`
	TechStack   string    `json:"tech_stack"`
	UserID      string    `json:"user_id"`
}