package catalog

import (
	"time"

	"api-backend-infinitrum/internal/models/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Acesso à obra (FREE vs assinatura PRO)
const (
	AccessTierFree = "free"
	AccessTierPro  = "pro"
)

// Status de publicação
const (
	BookStatusDraft     = "draft"
	BookStatusPublished = "published"
	BookStatusArchived  = "archived"
)

// CatalogBook é uma obra no catálogo do autor (ebook e/ou audiobook via capítulos nas tabelas específicas).
type CatalogBook struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	AuthorUserID string `gorm:"type:varchar(36);not null;uniqueIndex:idx_catalog_book_author_slug" json:"author_user_id"`

	Title string  `gorm:"type:varchar(300);not null" json:"title"`
	Slug  string  `gorm:"type:varchar(160);not null;uniqueIndex:idx_catalog_book_author_slug" json:"slug"`
	Subtitle *string `gorm:"type:varchar(300)" json:"subtitle,omitempty"`
	Synopsis *string `gorm:"type:text" json:"synopsis,omitempty"`

	CoverImageURL *string `gorm:"type:varchar(500)" json:"cover_image_url,omitempty"`

	Price    float64 `gorm:"type:numeric(10,2);not null;default:0" json:"price"`
	Currency string  `gorm:"type:varchar(3);not null;default:'BRL'" json:"currency"`

	AccessTier string `gorm:"type:varchar(16);not null;default:'free'" json:"access_tier"`

	Status string `gorm:"type:varchar(24);not null;default:'draft'" json:"status"`

	Language    string     `gorm:"type:varchar(10);not null;default:'pt-BR'" json:"language"`
	PublishedAt *time.Time `json:"published_at,omitempty"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	Author          user.User           `gorm:"foreignKey:AuthorUserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	EbookChapters   []EbookChapter      `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE" json:"-"`
	AudiobookChapters []AudiobookChapter `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE" json:"-"`
}

func (CatalogBook) TableName() string {
	return "catalog_books"
}

func (b *CatalogBook) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	if b.Currency == "" {
		b.Currency = "BRL"
	}
	if b.AccessTier == "" {
		b.AccessTier = AccessTierFree
	}
	if b.Status == "" {
		b.Status = BookStatusDraft
	}
	if b.Language == "" {
		b.Language = "pt-BR"
	}
	return nil
}
