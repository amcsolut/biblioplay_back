package feed

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	MediaTypeImage = "image"
	MediaTypeVideo = "video"
)

// PostMedia mídia anexada a um post (imagens ou vídeos).
type PostMedia struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	PostID string `gorm:"type:varchar(36);not null;uniqueIndex:idx_post_media_post_position" json:"post_id"`

	MediaType string `gorm:"type:varchar(24);not null" json:"media_type"`

	StorageKey *string `gorm:"type:varchar(500)" json:"storage_key,omitempty"`
	MediaURL   *string `gorm:"type:varchar(500)" json:"media_url,omitempty"`

	ThumbnailURL   *string `gorm:"type:varchar(500)" json:"thumbnail_url,omitempty"`
	ThumbnailKey   *string `gorm:"type:varchar(500)" json:"thumbnail_key,omitempty"`

	Position int `gorm:"not null;default:0;uniqueIndex:idx_post_media_post_position" json:"position"`

	Width           *int `json:"width,omitempty"`
	Height          *int `json:"height,omitempty"`
	DurationSeconds *int `json:"duration_seconds,omitempty"`
	MimeType        *string `gorm:"type:varchar(100)" json:"mime_type,omitempty"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	Post CommunityPost `gorm:"foreignKey:PostID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (PostMedia) TableName() string {
	return "post_media"
}

func (m *PostMedia) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
