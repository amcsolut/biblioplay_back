package community

import (
	"errors"
	"time"

	communityDTO "api-backend-infinitrum/api/v1/dto/community"
	"api-backend-infinitrum/internal/models/profile"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ownerUserID string, req *communityDTO.CreateCommunityRequest) (*communityDTO.CommunityResponse, error) {
	var count int64
	if err := s.db.Model(&profile.Community{}).Where("owner_user_id = ?", ownerUserID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("usuário já possui uma comunidade")
	}

	c := &profile.Community{
		OwnerUserID:    ownerUserID,
		Name:           req.Name,
		Slug:           req.Slug,
		Description:    req.Description,
		WelcomeMessage: req.WelcomeMessage,
		CoverImageURL:  req.CoverImageURL,
		Settings:       req.Settings,
	}
	if req.Visibility != "" {
		c.Visibility = req.Visibility
	}
	if req.JoinPolicy != "" {
		c.JoinPolicy = req.JoinPolicy
	}
	if req.PostPolicy != "" {
		c.PostPolicy = req.PostPolicy
	}
	if req.IsActive != nil {
		c.IsActive = *req.IsActive
	}

	if err := s.db.Create(c).Error; err != nil {
		return nil, err
	}
	return toCommunityResponse(c), nil
}

func (s *Service) ListByOwner(ownerUserID string) ([]communityDTO.CommunityResponse, error) {
	var list []profile.Community
	if err := s.db.Where("owner_user_id = ?", ownerUserID).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]communityDTO.CommunityResponse, 0, len(list))
	for i := range list {
		out = append(out, *toCommunityResponse(&list[i]))
	}
	return out, nil
}

func (s *Service) GetByID(id, currentUserID string) (*communityDTO.CommunityResponse, error) {
	var c profile.Community
	if err := s.db.Where("id = ?", id).First(&c).Error; err != nil {
		return nil, err
	}
	if c.OwnerUserID != currentUserID {
		if c.Visibility != profile.CommunityVisibilityPublic || !c.IsActive {
			return nil, gorm.ErrRecordNotFound
		}
	}
	return toCommunityResponse(&c), nil
}

func (s *Service) Update(id, ownerUserID string, req *communityDTO.UpdateCommunityRequest) (*communityDTO.CommunityResponse, error) {
	var c profile.Community
	if err := s.db.Where("id = ? AND owner_user_id = ?", id, ownerUserID).First(&c).Error; err != nil {
		return nil, err
	}
	updates := map[string]interface{}{"updated_at": time.Now()}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Slug != nil {
		updates["slug"] = *req.Slug
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.WelcomeMessage != nil {
		updates["welcome_message"] = *req.WelcomeMessage
	}
	if req.CoverImageURL != nil {
		updates["cover_image_url"] = *req.CoverImageURL
	}
	if req.Visibility != nil {
		updates["visibility"] = *req.Visibility
	}
	if req.JoinPolicy != nil {
		updates["join_policy"] = *req.JoinPolicy
	}
	if req.PostPolicy != nil {
		updates["post_policy"] = *req.PostPolicy
	}
	if req.Settings != nil {
		updates["settings"] = *req.Settings
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if err := s.db.Model(&c).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&c, "id = ?", c.ID).Error; err != nil {
		return nil, err
	}
	return toCommunityResponse(&c), nil
}

func (s *Service) Delete(id, ownerUserID string) error {
	res := s.db.Where("id = ? AND owner_user_id = ?", id, ownerUserID).Delete(&profile.Community{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func toCommunityResponse(c *profile.Community) *communityDTO.CommunityResponse {
	return &communityDTO.CommunityResponse{
		ID:             c.ID,
		OwnerUserID:    c.OwnerUserID,
		Name:           c.Name,
		Slug:           c.Slug,
		Description:    c.Description,
		WelcomeMessage: c.WelcomeMessage,
		CoverImageURL:  c.CoverImageURL,
		Visibility:     c.Visibility,
		JoinPolicy:     c.JoinPolicy,
		PostPolicy:     c.PostPolicy,
		Settings:       c.Settings,
		IsActive:       c.IsActive,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}
