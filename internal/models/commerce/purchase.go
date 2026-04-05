package commerce

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	PurchaseStatusPending  = "pending"
	PurchaseStatusPaid     = "paid"
	PurchaseStatusFailed   = "failed"
	PurchaseStatusRefunded = "refunded"
)

type Purchase struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	UserID string `gorm:"type:varchar(36);not null;index" json:"user_id"`

	Status string `gorm:"type:varchar(24);not null;default:'pending'" json:"status"`

	Currency    string  `gorm:"type:varchar(3);not null;default:'BRL'" json:"currency"`
	TotalAmount float64 `gorm:"type:numeric(12,2);not null;default:0" json:"total_amount"`

	Provider          *string `gorm:"type:varchar(64)" json:"provider,omitempty"`
	ExternalPaymentID *string `gorm:"type:varchar(255);index" json:"external_payment_id,omitempty"`

	PaidAt *time.Time `json:"paid_at,omitempty"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	Lines []PurchaseLine `gorm:"foreignKey:PurchaseID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Purchase) TableName() string {
	return "purchases"
}

func (p *Purchase) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if p.Currency == "" {
		p.Currency = "BRL"
	}
	return nil
}

type PurchaseLine struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	PurchaseID string `gorm:"type:varchar(36);not null;index" json:"purchase_id"`

	ItemType string `gorm:"type:varchar(32);not null" json:"item_type"`
	ItemID   string `gorm:"type:varchar(36);not null" json:"item_id"`

	UnitPrice     float64 `gorm:"type:numeric(12,2);not null" json:"unit_price"`
	TitleSnapshot *string `gorm:"type:varchar(400)" json:"title_snapshot,omitempty"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	Purchase Purchase `gorm:"foreignKey:PurchaseID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (PurchaseLine) TableName() string {
	return "purchase_lines"
}

func (l *PurchaseLine) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}
