package feed

import (
	"time"

	"api-backend-infinitrum/internal/models/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PostComment comentário de primeiro nível em um post.
type PostComment struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	PostID       string `gorm:"type:varchar(36);not null;index:idx_post_comments_post" json:"post_id"`
	AuthorUserID string `gorm:"type:varchar(36);not null;index" json:"author_user_id"`

	Body string `gorm:"type:text;not null" json:"body"`

	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP;index:idx_post_comments_post" json:"created_at"`
	UpdatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	Post    CommunityPost   `gorm:"foreignKey:PostID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Author  user.User       `gorm:"foreignKey:AuthorUserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Replies []CommentReply  `gorm:"foreignKey:CommentID;constraint:OnDelete:CASCADE" json:"-"`
}

func (PostComment) TableName() string {
	return "post_comments"
}

func (c *PostComment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
