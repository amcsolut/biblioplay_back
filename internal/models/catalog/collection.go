package catalog

import (
	"time"

	"api-backend-infinitrum/internal/models/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CatalogCollection agrupa obras do catálogo do mesmo autor (saga, trilogia, bundle).
type CatalogCollection struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	AuthorUserID string `gorm:"type:varchar(36);not null;uniqueIndex:idx_catalog_collection_author_slug" json:"author_user_id"`

	Title string `gorm:"type:varchar(300);not null" json:"title"`
	Slug  string `gorm:"type:varchar(160);not null;uniqueIndex:idx_catalog_collection_author_slug" json:"slug"`

	Subtitle    *string `gorm:"type:varchar(300)" json:"subtitle,omitempty"`
	Description *string `gorm:"type:text" json:"description,omitempty"`

	CoverImageURL *string `gorm:"type:varchar(500)" json:"cover_image_url,omitempty"`

	Price    float64 `gorm:"type:numeric(10,2);not null;default:0" json:"price"`
	Currency string  `gorm:"type:varchar(3);not null;default:'BRL'" json:"currency"`

	AccessTier string `gorm:"type:varchar(16);not null;default:'free'" json:"access_tier"`

	// Kind rótulo opcional: saga, trilogia, bundle, etc.
	Kind *string `gorm:"type:varchar(64)" json:"kind,omitempty"`

	Status string `gorm:"type:varchar(24);not null;default:'draft'" json:"status"`

	Language    string     `gorm:"type:varchar(10);not null;default:'pt-BR'" json:"language"`
	PublishedAt *time.Time `json:"published_at,omitempty"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	Author user.User               `gorm:"foreignKey:AuthorUserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Items  []CatalogCollectionBook `gorm:"foreignKey:CollectionID;constraint:OnDelete:CASCADE" json:"-"`
}

func (CatalogCollection) TableName() string {
	return "catalog_collections"
}

func (c *CatalogCollection) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.Currency == "" {
		c.Currency = "BRL"
	}
	if c.AccessTier == "" {
		c.AccessTier = AccessTierFree
	}
	if c.Status == "" {
		c.Status = BookStatusDraft
	}
	if c.Language == "" {
		c.Language = "pt-BR"
	}
	return nil
}

// CatalogCollectionBook liga uma coleção a obras do catálogo com ordem fixa.
type CatalogCollectionBook struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	CollectionID string `gorm:"type:varchar(36);not null;uniqueIndex:idx_ccb_collection_book;uniqueIndex:idx_ccb_collection_position" json:"collection_id"`
	BookID       string `gorm:"type:varchar(36);not null;uniqueIndex:idx_ccb_collection_book" json:"book_id"`

	Position int `gorm:"not null;uniqueIndex:idx_ccb_collection_position" json:"position"`

	VolumeLabel *string `gorm:"type:varchar(120)" json:"volume_label,omitempty"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	Collection CatalogCollection `gorm:"foreignKey:CollectionID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Book       CatalogBook       `gorm:"foreignKey:BookID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (CatalogCollectionBook) TableName() string {
	return "catalog_collection_books"
}

func (cb *CatalogCollectionBook) BeforeCreate(tx *gorm.DB) error {
	if cb.ID == "" {
		cb.ID = uuid.New().String()
	}
	return nil
}
