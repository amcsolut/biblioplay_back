package library

import "time"

type AddLibraryItemRequest struct {
	ItemType string `json:"item_type" binding:"required,oneof=book collection"`
	ItemID   string `json:"item_id" binding:"required,min=1"`
}

type BookSummary struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
	Price         float64 `json:"price"`
	Currency      string  `json:"currency"`
	AccessTier    string  `json:"access_tier"`
	Status        string  `json:"status"`
}

type CollectionSummary struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
	Price         float64 `json:"price"`
	Currency      string  `json:"currency"`
	AccessTier    string  `json:"access_tier"`
	Status        string  `json:"status"`
}

type LibraryItemResponse struct {
	ID              string     `json:"id"`
	ItemType        string     `json:"item_type"`
	ItemID          string     `json:"item_id"`
	Source          string     `json:"source"`
	PurchaseID      *string    `json:"purchase_id,omitempty"`
	GrantedByUserID *string    `json:"granted_by_user_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	Book            *BookSummary       `json:"book,omitempty"`
	Collection      *CollectionSummary `json:"collection,omitempty"`
}

type AuthorGrantRequest struct {
	ReaderUserID string `json:"reader_user_id" binding:"required,min=1"`
	ItemType     string `json:"item_type" binding:"required,oneof=book collection"`
	ItemID       string `json:"item_id" binding:"required,min=1"`
}
