package profile

import (
	"time"

	"api-backend-infinitrum/internal/models/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Comunidade: visibilidade geral
const (
	CommunityVisibilityPublic            = "public"
	CommunityVisibilityPrivate           = "private"
	CommunityVisibilityUnlisted          = "unlisted"
	CommunityVisibilitySubscribersOnly   = "subscribers_only"
)

// Como novos membros entram
const (
	JoinPolicyOpen   = "open"
	JoinPolicyApproval = "approval"
	JoinPolicyInviteOnly = "invite_only"
)

// Quem pode publicar no feed da comunidade
const (
	PostPolicyAuthorOnly   = "author_only"
	PostPolicyMembers      = "members"
	PostPolicyModeratorsAndAuthor = "moderators_and_author"
)

// Community representa a comunidade interna de um autor (1:1 com owner_user_id).
type Community struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	OwnerUserID string `gorm:"type:varchar(36);not null;uniqueIndex" json:"owner_user_id"`

	Name        string  `gorm:"type:varchar(200);not null" json:"name"`
	Slug        string  `gorm:"type:varchar(120);not null;uniqueIndex" json:"slug"`
	Description *string `gorm:"type:text" json:"description,omitempty"`
	WelcomeMessage *string `gorm:"type:text" json:"welcome_message,omitempty"`

	CoverImageURL *string `gorm:"type:varchar(500)" json:"cover_image_url,omitempty"`

	Visibility string `gorm:"type:varchar(32);not null;default:'public'" json:"visibility"`
	JoinPolicy string `gorm:"type:varchar(32);not null;default:'open'" json:"join_policy"`
	PostPolicy string `gorm:"type:varchar(32);not null;default:'members'" json:"post_policy"`

	Settings string `gorm:"type:text" json:"settings,omitempty"`

	IsActive bool `gorm:"not null;default:true" json:"is_active"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	Owner user.User `gorm:"foreignKey:OwnerUserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Community) TableName() string {
	return "communities"
}

func (c *Community) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.Visibility == "" {
		c.Visibility = CommunityVisibilityPublic
	}
	if c.JoinPolicy == "" {
		c.JoinPolicy = JoinPolicyOpen
	}
	if c.PostPolicy == "" {
		c.PostPolicy = PostPolicyMembers
	}
	return nil
}
