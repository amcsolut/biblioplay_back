package commerce

import "time"

type PurchaseLineRequest struct {
	ItemType string `json:"item_type" binding:"required,oneof=book collection"`
	ItemID   string `json:"item_id" binding:"required,min=1"`
}

type CreatePurchaseRequest struct {
	Lines []PurchaseLineRequest `json:"lines" binding:"required,min=1,dive"`
}

type PurchaseLineResponse struct {
	ID            string   `json:"id"`
	ItemType      string   `json:"item_type"`
	ItemID        string   `json:"item_id"`
	UnitPrice     float64  `json:"unit_price"`
	TitleSnapshot *string  `json:"title_snapshot,omitempty"`
}

type PurchaseResponse struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	Status        string                 `json:"status"`
	Currency      string                 `json:"currency"`
	TotalAmount   float64                `json:"total_amount"`
	PaidAt        *time.Time             `json:"paid_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	Lines         []PurchaseLineResponse `json:"lines"`
}
