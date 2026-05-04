package request

type CreateProjectRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required"`
	URL         string `json:"url"`
	EmbedURL    string `json:"embed_url"`
	ImageURL    string `json:"image_url"`
	GitHubURL   string `json:"github_url"`
	TechStack   string `json:"tech_stack"`
	UserID      string `json:"user_id"`
}

type UpdateProjectRequest struct {
	Id          uint   `json:"id" binding:"required"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	EmbedURL    string `json:"embed_url"`
	ImageURL    string `json:"image_url"`
	GitHubURL   string `json:"github_url"`
	TechStack   string `json:"tech_stack"`
}