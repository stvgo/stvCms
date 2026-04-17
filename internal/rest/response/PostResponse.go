package response

import "time"

// TODO: cambiar a snake case la firma del response, también en front
type PostResponse struct {
	Id            uint                   `json:"id"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
	Title         string                 `json:"title"`
	UserID        string                 `json:"userId"`
	ContentBlocks []ContentBlockResponse `json:"contentBlocks"`
	IsVisible     bool                   `json:"isVisible"`
}

type ContentBlockResponse struct {
	Id       uint   `json:"id"`
	Type     string `json:"type"`
	Order    int    `json:"order"`
	Content  string `json:"content"`
	Language string `json:"language,omitempty"`
}
