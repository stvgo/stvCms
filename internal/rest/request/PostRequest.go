package request

type CreatePostRequest struct {
	Title         string         `json:"title" binding:"required"`
	UserID        string         `json:"user_id" binding:"required"`
	ContentBlocks []ContentBlock `json:"content_blocks"`
}

type ContentBlock struct {
	Type     string `json:"type" binding:"required"` // images, code, text, urls
	Order    int    `json:"order" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Language string `json:"language,omitempty"`
}

type UpdatePostRequest struct {
	Id            uint           `json:"id" binding:"required"`
	Title         string         `json:"title"`
	ContentBlocks []ContentBlock `json:"content_blocks"`
}
