package feed

import (
	"time"

	"api-backend-infinitrum/internal/models/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommentReply resposta a um comentário (thread: pode responder a outra reply via ParentReplyID).
type CommentReply struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	CommentID string `gorm:"type:varchar(36);not null;index" json:"comment_id"`
	// NULL = resposta direta ao comentário raiz; senão FK para outra reply
	ParentReplyID *string `gorm:"type:varchar(36);index" json:"parent_reply_id,omitempty"`

	AuthorUserID string `gorm:"type:varchar(36);not null;index" json:"author_user_id"`

	Body string `gorm:"type:text;not null" json:"body"`

	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	Comment     PostComment    `gorm:"foreignKey:CommentID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	ParentReply *CommentReply  `gorm:"foreignKey:ParentReplyID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Author      user.User      `gorm:"foreignKey:AuthorUserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (CommentReply) TableName() string {
	return "comment_replies"
}

func (r *CommentReply) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}
