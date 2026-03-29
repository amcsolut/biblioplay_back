package catalog

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AudiobookChapter capítulo em áudio de uma obra.
type AudiobookChapter struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	BookID string `gorm:"type:varchar(36);not null;uniqueIndex:idx_catalog_audio_book_position" json:"book_id"`

	Position int    `gorm:"not null;uniqueIndex:idx_catalog_audio_book_position" json:"position"`
	Title    string `gorm:"type:varchar(500);not null" json:"title"`

	AudioStorageKey *string `gorm:"type:varchar(500)" json:"audio_storage_key,omitempty"`
	AudioURL        *string `gorm:"type:varchar(500)" json:"audio_url,omitempty"`
	DurationSeconds *int    `gorm:"type:integer" json:"duration_seconds,omitempty"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	Book CatalogBook `gorm:"foreignKey:BookID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (AudiobookChapter) TableName() string {
	return "catalog_audiobook_chapters"
}

func (c *AudiobookChapter) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
