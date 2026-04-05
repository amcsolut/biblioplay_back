package commerce

import (
	"errors"
	"fmt"
	"time"

	commercedto "api-backend-infinitrum/api/v1/dto/commerce"
	libsvc "api-backend-infinitrum/api/v1/services/library"
	commercemodel "api-backend-infinitrum/internal/models/commerce"

	"gorm.io/gorm"
)

type Service struct {
	db      *gorm.DB
	library *libsvc.Service
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db, library: libsvc.NewService(db)}
}

// CreatePaidPurchase cria pedido como pago e adiciona itens à biblioteca (integração com gateway pode alterar status depois).
func (s *Service) CreatePaidPurchase(buyerUserID string, req *commercedto.CreatePurchaseRequest) (*commercedto.PurchaseResponse, error) {
	if len(req.Lines) == 0 {
		return nil, errors.New("informe ao menos uma linha")
	}

	seen := make(map[string]struct{}, len(req.Lines))
	for _, ln := range req.Lines {
		key := fmt.Sprintf("%s:%s", ln.ItemType, ln.ItemID)
		if _, ok := seen[key]; ok {
			return nil, errors.New("linhas duplicadas no mesmo pedido")
		}
		seen[key] = struct{}{}
	}

	type lineResolved struct {
		ItemType  string
		ItemID    string
		UnitPrice float64
		Currency  string
		Title     string
	}
	resolved := make([]lineResolved, 0, len(req.Lines))
	currency := ""

	for _, ln := range req.Lines {
		price, cur, title, err := s.library.ResolvePaidLine(ln.ItemType, ln.ItemID)
		if err != nil {
			return nil, err
		}
		if currency == "" {
			currency = cur
		}
		if cur != currency {
			return nil, errors.New("todas as linhas devem usar a mesma moeda nesta versão da API")
		}
		resolved = append(resolved, lineResolved{
			ItemType:  ln.ItemType,
			ItemID:    ln.ItemID,
			UnitPrice: price,
			Currency:  cur,
			Title:     title,
		})
	}

	var total float64
	for _, r := range resolved {
		total += r.UnitPrice
	}

	now := time.Now()
	p := &commercemodel.Purchase{
		UserID:        buyerUserID,
		Status:        commercemodel.PurchaseStatusPaid,
		Currency:      currency,
		TotalAmount:   total,
		PaidAt:        &now,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		for _, r := range resolved {
			ts := r.Title
			line := &commercemodel.PurchaseLine{
				PurchaseID:    p.ID,
				ItemType:      r.ItemType,
				ItemID:        r.ItemID,
				UnitPrice:     r.UnitPrice,
				TitleSnapshot: &ts,
			}
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}
		libLines := make([]struct {
			ItemType string
			ItemID   string
		}, len(resolved))
		for i := range resolved {
			libLines[i].ItemType = resolved[i].ItemType
			libLines[i].ItemID = resolved[i].ItemID
		}
		return s.library.AddOrSkipFromPurchase(tx, buyerUserID, p.ID, libLines)
	})
	if err != nil {
		return nil, err
	}

	return s.toPurchaseResponse(p.ID)
}

func (s *Service) toPurchaseResponse(purchaseID string) (*commercedto.PurchaseResponse, error) {
	var p commercemodel.Purchase
	if err := s.db.Preload("Lines").Where("id = ?", purchaseID).First(&p).Error; err != nil {
		return nil, err
	}
	lines := make([]commercedto.PurchaseLineResponse, 0, len(p.Lines))
	for i := range p.Lines {
		lines = append(lines, commercedto.PurchaseLineResponse{
			ID:            p.Lines[i].ID,
			ItemType:      p.Lines[i].ItemType,
			ItemID:        p.Lines[i].ItemID,
			UnitPrice:     p.Lines[i].UnitPrice,
			TitleSnapshot: p.Lines[i].TitleSnapshot,
		})
	}
	return &commercedto.PurchaseResponse{
		ID:          p.ID,
		UserID:      p.UserID,
		Status:      p.Status,
		Currency:    p.Currency,
		TotalAmount: p.TotalAmount,
		PaidAt:      p.PaidAt,
		CreatedAt:   p.CreatedAt,
		Lines:       lines,
	}, nil
}
