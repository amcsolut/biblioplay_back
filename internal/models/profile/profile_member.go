package profile

import (
	"time"

	"api-backend-infinitrum/internal/models/user"

	"gorm.io/gorm"
)

// Visibilidade do perfil de membro
const (
	ProfileVisibilityPublic          = "public"
	ProfileVisibilityFollowersOnly   = "followers_only"
	ProfileVisibilityPrivate         = "private"
)

// Quem pode enviar mensagens ao membro
const (
	AllowMessagesEveryone  = "everyone"
	AllowMessagesFollowers = "followers"
	AllowMessagesNobody    = "nobody"
)

// ProfileMember estende a conta de usuário com dados de leitor/ouvinte na plataforma.
// Avatar e nome legal ficam em users.
type ProfileMember struct {
	UserID string `gorm:"type:varchar(36);primaryKey" json:"user_id"`

	Username    string  `gorm:"type:varchar(50);not null;uniqueIndex" json:"username"`
	DisplayName *string `gorm:"type:varchar(150)" json:"display_name,omitempty"`
	Bio         *string `gorm:"type:text" json:"bio,omitempty"`

	Locale      string  `gorm:"type:varchar(10);not null;default:'pt-BR'" json:"locale"`
	Timezone    *string `gorm:"type:varchar(50)" json:"timezone,omitempty"`
	CountryCode *string `gorm:"type:char(2)" json:"country_code,omitempty"`

	FavoriteGenres string `gorm:"type:text" json:"favorite_genres,omitempty"` // JSON array de strings

	ContentEbook     bool `gorm:"not null;default:true" json:"content_ebook"`
	ContentAudiobook bool `gorm:"not null;default:true" json:"content_audiobook"`

	ProfileVisibility   string `gorm:"type:varchar(32);not null;default:'public'" json:"profile_visibility"`
	ShowReadingActivity bool   `gorm:"not null;default:true" json:"show_reading_activity"`
	AllowMessagesFrom   string `gorm:"type:varchar(32);not null;default:'followers'" json:"allow_messages_from"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	User user.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (ProfileMember) TableName() string {
	return "profile_members"
}

func (m *ProfileMember) BeforeCreate(tx *gorm.DB) error {
	if m.ProfileVisibility == "" {
		m.ProfileVisibility = ProfileVisibilityPublic
	}
	if m.AllowMessagesFrom == "" {
		m.AllowMessagesFrom = AllowMessagesFollowers
	}
	if m.Locale == "" {
		m.Locale = "pt-BR"
	}
	return nil
}
