package feed

import (
	"time"

	"api-backend-infinitrum/internal/models/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Alvo polimórfico da reação (IDs em target_id referenciam a tabela correspondente).
const (
	TargetTypePost    = "post"
	TargetTypeComment = "comment"
	TargetTypeReply   = "reply"
)

// Tipos de reação aceitos (validar também na API).
const (
	ReactionLike   = "like"
	ReactionLove   = "love"
	ReactionLaugh  = "laugh"
	ReactionSad    = "sad"
	ReactionAngry  = "angry"
	ReactionWow    = "wow"
)

// Reaction uma reação por usuário por alvo (atualizar `type` para trocar emoji).
type Reaction struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	UserID string `gorm:"type:varchar(36);not null;uniqueIndex:idx_reactions_user_target" json:"user_id"`

	TargetType string `gorm:"type:varchar(24);not null;uniqueIndex:idx_reactions_user_target;index:idx_reactions_lookup" json:"target_type"`
	TargetID   string `gorm:"type:varchar(36);not null;uniqueIndex:idx_reactions_user_target;index:idx_reactions_lookup" json:"target_id"`

	Type string `gorm:"type:varchar(24);not null" json:"type"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	User user.User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Reaction) TableName() string {
	return "reactions"
}

func (r *Reaction) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}
