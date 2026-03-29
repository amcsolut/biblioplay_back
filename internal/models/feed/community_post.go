package feed

import (
	"time"

	"api-backend-infinitrum/internal/models/profile"
	"api-backend-infinitrum/internal/models/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	PostStatusDraft     = "draft"
	PostStatusPublished = "published"
	PostStatusHidden    = "hidden"
	PostStatusDeleted   = "deleted"
)

// CommunityPost é uma postagem no feed de uma comunidade (qualquer usuário autorizado pode postar).
type CommunityPost struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	CommunityID  string `gorm:"type:varchar(36);not null;index:idx_community_posts_feed" json:"community_id"`
	AuthorUserID string `gorm:"type:varchar(36);not null;index" json:"author_user_id"`

	Body *string `gorm:"type:text" json:"body,omitempty"`

	Status string `gorm:"type:varchar(24);not null;default:'published'" json:"status"`
	Pinned bool   `gorm:"not null;default:false" json:"pinned"`

	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP;index:idx_community_posts_feed" json:"created_at"`
	UpdatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	Community profile.Community `gorm:"foreignKey:CommunityID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Author    user.User         `gorm:"foreignKey:AuthorUserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Media     []PostMedia       `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"-"`
	Comments  []PostComment     `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"-"`
}

func (CommunityPost) TableName() string {
	return "community_posts"
}

func (p *CommunityPost) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if p.Status == "" {
		p.Status = PostStatusPublished
	}
	return nil
}
