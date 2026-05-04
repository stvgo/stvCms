package models

import "time"

// Notification represents an admin notification about a pending post.
type Notification struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "post_pending"
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	PostID    uint      `json:"post_id"`
	AuthorID  string    `json:"author_id"`
	AuthorName string   `json:"author_name"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}