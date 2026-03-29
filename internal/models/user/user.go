package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID            string     `gorm:"type:varchar(36);primary_key" json:"id"`
	Email         string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash  *string    `gorm:"type:varchar(255)" json:"-"` // Nullable for social auth users
	FirstName     string     `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName      string     `gorm:"type:varchar(100);not null" json:"last_name"`
	Phone         *string    `gorm:"type:varchar(20)" json:"phone"`
	AvatarURL     *string    `gorm:"type:varchar(500)" json:"avatar_url"`
	IsActive      bool       `gorm:"default:true;not null" json:"is_active"`
	EmailVerified bool       `gorm:"default:false;not null" json:"email_verified"`
	RoleLevel     int        `gorm:"type:integer;default:1;not null" json:"role_level"`
	// Social authentication fields
	Provider   *string `gorm:"type:varchar(50);index" json:"provider"` // google, facebook, or null for email/password
	ProviderID *string `gorm:"type:varchar(255);index" json:"provider_id"` // Social provider user ID
	DeletedAt  *time.Time `gorm:"index" json:"deleted_at"`
	CreatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	UserSessions    []UserSession    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	UserInvitations []UserInvitation `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	// Add other relationships as needed
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}
