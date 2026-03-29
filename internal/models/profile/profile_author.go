package profile

import (
	"time"

	"api-backend-infinitrum/internal/models/user"

	"gorm.io/gorm"
)

// Política de mensagens para o autor
const (
	AuthorAcceptsMessagesEveryone  = "everyone"
	AuthorAcceptsMessagesFollowers = "followers"
	AuthorAcceptsMessagesNobody    = "nobody"
)

// ProfileAuthor dados públicos e de catálogo do autor (ebooks/audiobooks).
type ProfileAuthor struct {
	UserID string `gorm:"type:varchar(36);primaryKey" json:"user_id"`

	PenName string  `gorm:"type:varchar(200);not null" json:"pen_name"`
	Slug    string  `gorm:"type:varchar(100);not null;uniqueIndex" json:"slug"`
	Tagline *string `gorm:"type:varchar(255)" json:"tagline,omitempty"`
	Bio     *string `gorm:"type:text" json:"bio,omitempty"`

	WebsiteURL  *string `gorm:"type:varchar(500)" json:"website_url,omitempty"`
	SocialLinks string `gorm:"type:text" json:"social_links,omitempty"` // JSON livre (redes, etc.)

	DefaultLanguage string `gorm:"type:varchar(10);not null;default:'pt-BR'" json:"default_language"`
	PrimaryGenres   string `gorm:"type:text" json:"primary_genres,omitempty"` // JSON array de strings

	IsVerified bool `gorm:"not null;default:false" json:"is_verified"`

	AcceptsMessagesFrom string `gorm:"type:varchar(32);not null;default:'followers'" json:"accepts_messages_from"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	User user.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (ProfileAuthor) TableName() string {
	return "profile_authors"
}

func (a *ProfileAuthor) BeforeCreate(tx *gorm.DB) error {
	if a.DefaultLanguage == "" {
		a.DefaultLanguage = "pt-BR"
	}
	if a.AcceptsMessagesFrom == "" {
		a.AcceptsMessagesFrom = AuthorAcceptsMessagesFollowers
	}
	return nil
}
