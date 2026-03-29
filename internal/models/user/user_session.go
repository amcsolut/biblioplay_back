package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserSession struct {
	ID           string    `gorm:"type:varchar(36);primary_key" json:"id"`
	UserID       string    `gorm:"type:varchar(36);not null;index" json:"user_id"`
	RefreshToken string    `gorm:"type:text;not null" json:"-"`
	ExpiresAt    time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationship
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}

func (us *UserSession) BeforeCreate(tx *gorm.DB) error {
	if us.ID == "" {
		us.ID = uuid.New().String()
	}
	return nil
}

