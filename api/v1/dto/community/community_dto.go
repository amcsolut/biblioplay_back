package community

import "time"

type CreateCommunityRequest struct {
	Name            string  `json:"name" binding:"required,min=2,max=200"`
	Slug            string  `json:"slug" binding:"required,min=2,max=120"`
	Description     *string `json:"description"`
	WelcomeMessage  *string `json:"welcome_message"`
	CoverImageURL   *string `json:"cover_image_url"`
	Visibility      string  `json:"visibility"`
	JoinPolicy      string  `json:"join_policy"`
	PostPolicy      string  `json:"post_policy"`
	Settings        string  `json:"settings"`
	IsActive        *bool   `json:"is_active"`
}

type UpdateCommunityRequest struct {
	Name            *string `json:"name" binding:"omitempty,min=2,max=200"`
	Slug            *string `json:"slug" binding:"omitempty,min=2,max=120"`
	Description     *string `json:"description"`
	WelcomeMessage  *string `json:"welcome_message"`
	CoverImageURL   *string `json:"cover_image_url"`
	Visibility      *string `json:"visibility"`
	JoinPolicy      *string `json:"join_policy"`
	PostPolicy      *string `json:"post_policy"`
	Settings        *string `json:"settings"`
	IsActive        *bool   `json:"is_active"`
}

type CommunityResponse struct {
	ID             string    `json:"id"`
	OwnerUserID    string    `json:"owner_user_id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    *string   `json:"description,omitempty"`
	WelcomeMessage *string   `json:"welcome_message,omitempty"`
	CoverImageURL  *string   `json:"cover_image_url,omitempty"`
	Visibility     string    `json:"visibility"`
	JoinPolicy     string    `json:"join_policy"`
	PostPolicy     string    `json:"post_policy"`
	Settings       string    `json:"settings,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
