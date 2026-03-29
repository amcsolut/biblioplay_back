package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasswordReset struct {
	ID        string    `gorm:"type:varchar(36);primary_key" json:"id"`
	UserID    string    `gorm:"type:varchar(36);not null;index" json:"user_id"`
	Token     string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"token"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	Used      bool      `gorm:"default:false;not null" json:"used"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationship
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (PasswordReset) TableName() string {
	return "password_resets"
}

func (pr *PasswordReset) BeforeCreate(tx *gorm.DB) error {
	if pr.ID == "" {
		pr.ID = uuid.New().String()
	}
	return nil
}

