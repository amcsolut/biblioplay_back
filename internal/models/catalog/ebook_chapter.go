package catalog

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EbookChapter capítulo em texto de uma obra.
type EbookChapter struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	BookID string `gorm:"type:varchar(36);not null;uniqueIndex:idx_catalog_ebook_book_position" json:"book_id"`

	Position int    `gorm:"not null;uniqueIndex:idx_catalog_ebook_book_position" json:"position"`
	Title    string `gorm:"type:varchar(500);not null" json:"title"`

	BodyText *string `gorm:"type:text" json:"body_text,omitempty"`
	// Conteúdo em storage (ex.: S3) em vez de body_text inline
	ContentStorageKey *string `gorm:"type:varchar(500)" json:"content_storage_key,omitempty"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	Book CatalogBook `gorm:"foreignKey:BookID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (EbookChapter) TableName() string {
	return "catalog_ebook_chapters"
}

func (c *EbookChapter) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
