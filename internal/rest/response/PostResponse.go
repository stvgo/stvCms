package response

import "time"

type PostResponse struct {
	Id            uint                   `json:"id"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Title         string                 `json:"title"`
	UserID        string                 `json:"user_id"`
	ContentBlocks []ContentBlockResponse `json:"content_blocks"`
	IsVisible     bool                   `json:"is_visible"`
}

type ContentBlockResponse struct {
	Id       uint   `json:"id"`
	Type     string `json:"type"`
	Order    int    `json:"order"`
	Content  string `json:"content"`
	Language string `json:"language,omitempty"`
}
