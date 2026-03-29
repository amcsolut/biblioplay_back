package feed

import "time"

// --- Posts ---

type CreatePostRequest struct {
	Body   *string `json:"body"`
	Status string  `json:"status"`
	Pinned *bool   `json:"pinned"`
}

type UpdatePostRequest struct {
	Body   *string `json:"body"`
	Status *string `json:"status"`
	Pinned *bool   `json:"pinned"`
}

type PostResponse struct {
	ID           string     `json:"id"`
	CommunityID  string     `json:"community_id"`
	AuthorUserID string     `json:"author_user_id"`
	Body         *string    `json:"body,omitempty"`
	Status       string     `json:"status"`
	Pinned       bool       `json:"pinned"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// --- Media ---

type CreatePostMediaRequest struct {
	MediaType       string  `json:"media_type" binding:"required"`
	StorageKey      *string `json:"storage_key"`
	MediaURL        *string `json:"media_url"`
	ThumbnailURL    *string `json:"thumbnail_url"`
	ThumbnailKey    *string `json:"thumbnail_key"`
	Position        int     `json:"position"`
	Width           *int    `json:"width"`
	Height          *int    `json:"height"`
	DurationSeconds *int    `json:"duration_seconds"`
	MimeType        *string `json:"mime_type"`
}

type UpdatePostMediaRequest struct {
	MediaType       *string `json:"media_type"`
	StorageKey      *string `json:"storage_key"`
	MediaURL        *string `json:"media_url"`
	ThumbnailURL    *string `json:"thumbnail_url"`
	ThumbnailKey    *string `json:"thumbnail_key"`
	Position        *int    `json:"position"`
	Width           *int    `json:"width"`
	Height          *int    `json:"height"`
	DurationSeconds *int    `json:"duration_seconds"`
	MimeType        *string `json:"mime_type"`
}

type PostMediaResponse struct {
	ID              string    `json:"id"`
	PostID          string    `json:"post_id"`
	MediaType       string    `json:"media_type"`
	StorageKey      *string   `json:"storage_key,omitempty"`
	MediaURL        *string   `json:"media_url,omitempty"`
	ThumbnailURL    *string   `json:"thumbnail_url,omitempty"`
	ThumbnailKey    *string   `json:"thumbnail_key,omitempty"`
	Position        int       `json:"position"`
	Width           *int      `json:"width,omitempty"`
	Height          *int      `json:"height,omitempty"`
	DurationSeconds *int      `json:"duration_seconds,omitempty"`
	MimeType        *string   `json:"mime_type,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// --- Comments ---

type CreateCommentRequest struct {
	Body string `json:"body" binding:"required,min=1"`
}

type UpdateCommentRequest struct {
	Body string `json:"body" binding:"required,min=1"`
}

type CommentResponse struct {
	ID           string     `json:"id"`
	PostID       string     `json:"post_id"`
	AuthorUserID string     `json:"author_user_id"`
	Body         string     `json:"body"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// --- Replies ---

type CreateReplyRequest struct {
	Body            string  `json:"body" binding:"required,min=1"`
	ParentReplyID   *string `json:"parent_reply_id"`
}

type UpdateReplyRequest struct {
	Body string `json:"body" binding:"required,min=1"`
}

type ReplyResponse struct {
	ID             string     `json:"id"`
	CommentID      string     `json:"comment_id"`
	ParentReplyID  *string    `json:"parent_reply_id,omitempty"`
	AuthorUserID   string     `json:"author_user_id"`
	Body           string     `json:"body"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// --- Reactions ---

type UpsertReactionRequest struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   string `json:"target_id" binding:"required"`
	Type       string `json:"type" binding:"required"`
}

type ReactionResponse struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	Type       string    `json:"type"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
