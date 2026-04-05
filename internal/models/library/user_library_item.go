package library

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ItemTypeBook       = "book"
	ItemTypeCollection = "collection"
)

const (
	SourceFreeSelf    = "free_self"
	SourcePurchase    = "purchase"
	SourceAuthorGrant = "author_grant"
)

// UserLibraryItem liga leitor a livro ou coleção (polimórfico).
type UserLibraryItem struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	UserID   string `gorm:"type:varchar(36);not null;index;uniqueIndex:idx_user_library_item" json:"user_id"`
	ItemType string `gorm:"type:varchar(32);not null;uniqueIndex:idx_user_library_item" json:"item_type"`
	ItemID   string `gorm:"type:varchar(36);not null;uniqueIndex:idx_user_library_item" json:"item_id"`

	Source string `gorm:"type:varchar(32);not null" json:"source"`

	PurchaseID      *string `gorm:"type:varchar(36);index" json:"purchase_id,omitempty"`
	GrantedByUserID *string `gorm:"type:varchar(36);index" json:"granted_by_user_id,omitempty"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (UserLibraryItem) TableName() string {
	return "user_library_items"
}

func (u *UserLibraryItem) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}
