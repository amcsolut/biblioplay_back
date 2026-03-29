package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserInvitation struct {
	ID             string     `gorm:"type:varchar(36);primary_key" json:"id"`
	UserID         *string    `gorm:"type:varchar(36);index" json:"user_id"` // Nullable, set when accepted
	Email          string     `gorm:"type:varchar(255);not null;index" json:"email"`
	OrganizationID string     `gorm:"type:varchar(36);not null;index" json:"organization_id"`
	InvitedBy      string     `gorm:"type:varchar(36);not null" json:"invited_by"`
	Status         string     `gorm:"type:varchar(20);default:'pending';not null" json:"status"` // pending, accepted, rejected, expired
	Token          string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	ExpiresAt      time.Time  `gorm:"not null;index" json:"expires_at"`
	AcceptedAt     *time.Time `json:"accepted_at"`
	CreatedAt      time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL" json:"-"`
}

func (UserInvitation) TableName() string {
	return "user_invitations"
}

func (ui *UserInvitation) BeforeCreate(tx *gorm.DB) error {
	if ui.ID == "" {
		ui.ID = uuid.New().String()
	}
	return nil
}

