package catalog

import "time"

type CreateBookRequest struct {
	Title         string   `json:"title" binding:"required,min=1,max=300"`
	Slug          string   `json:"slug" binding:"required,min=1,max=160"`
	Subtitle      *string  `json:"subtitle"`
	Synopsis      *string  `json:"synopsis"`
	CoverImageURL *string  `json:"cover_image_url"`
	Price         *float64 `json:"price"`
	Currency      string   `json:"currency"`
	AccessTier    string   `json:"access_tier"`
	Status        string   `json:"status"`
	Language      string   `json:"language"`
	PublishedAt   *string  `json:"published_at"` // RFC3339
}

type UpdateBookRequest struct {
	Title         *string  `json:"title" binding:"omitempty,min=1,max=300"`
	Slug          *string  `json:"slug" binding:"omitempty,min=1,max=160"`
	Subtitle      *string  `json:"subtitle"`
	Synopsis      *string  `json:"synopsis"`
	CoverImageURL *string  `json:"cover_image_url"`
	Price         *float64 `json:"price"`
	Currency      *string  `json:"currency"`
	AccessTier    *string  `json:"access_tier"`
	Status        *string  `json:"status"`
	Language      *string  `json:"language"`
	PublishedAt   *string  `json:"published_at"`
}

type BookResponse struct {
	ID            string     `json:"id"`
	AuthorUserID  string     `json:"author_user_id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	Subtitle      *string    `json:"subtitle,omitempty"`
	Synopsis      *string    `json:"synopsis,omitempty"`
	CoverImageURL *string    `json:"cover_image_url,omitempty"`
	Price         float64    `json:"price"`
	Currency      string     `json:"currency"`
	AccessTier    string     `json:"access_tier"`
	Status        string     `json:"status"`
	Language      string     `json:"language"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CreateEbookChapterRequest struct {
	Position          int     `json:"position" binding:"required,min=1"`
	Title             string  `json:"title" binding:"required,min=1,max=500"`
	BodyText          *string `json:"body_text"`
	ContentStorageKey *string `json:"content_storage_key"`
}

type UpdateEbookChapterRequest struct {
	Position          *int    `json:"position" binding:"omitempty,min=1"`
	Title             *string `json:"title" binding:"omitempty,min=1,max=500"`
	BodyText          *string `json:"body_text"`
	ContentStorageKey *string `json:"content_storage_key"`
}

type EbookChapterResponse struct {
	ID                string    `json:"id"`
	BookID            string    `json:"book_id"`
	Position          int       `json:"position"`
	Title             string    `json:"title"`
	BodyText          *string   `json:"body_text,omitempty"`
	ContentStorageKey *string   `json:"content_storage_key,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CreateAudiobookChapterRequest struct {
	Position          int     `json:"position" binding:"required,min=1"`
	Title             string  `json:"title" binding:"required,min=1,max=500"`
	AudioStorageKey   *string `json:"audio_storage_key"`
	AudioURL          *string `json:"audio_url"`
	DurationSeconds   *int    `json:"duration_seconds" binding:"omitempty,min=0"`
}

type UpdateAudiobookChapterRequest struct {
	Position          *int    `json:"position" binding:"omitempty,min=1"`
	Title             *string `json:"title" binding:"omitempty,min=1,max=500"`
	AudioStorageKey   *string `json:"audio_storage_key"`
	AudioURL          *string `json:"audio_url"`
	DurationSeconds   *int    `json:"duration_seconds" binding:"omitempty,min=0"`
}

type AudiobookChapterResponse struct {
	ID                string    `json:"id"`
	BookID            string    `json:"book_id"`
	Position          int       `json:"position"`
	Title             string    `json:"title"`
	AudioStorageKey   *string   `json:"audio_storage_key,omitempty"`
	AudioURL          *string   `json:"audio_url,omitempty"`
	DurationSeconds   *int      `json:"duration_seconds,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
