package library

import (
	"errors"

	librarydto "api-backend-infinitrum/api/v1/dto/library"
	catalogmodel "api-backend-infinitrum/internal/models/catalog"
	libmodel "api-backend-infinitrum/internal/models/library"
	usermodel "api-backend-infinitrum/internal/models/user"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func isFreeCatalogBook(b *catalogmodel.CatalogBook) bool {
	return b.Status == catalogmodel.BookStatusPublished &&
		b.Price <= 0 &&
		b.AccessTier == catalogmodel.AccessTierFree
}

func isFreeCatalogCollection(c *catalogmodel.CatalogCollection) bool {
	return c.Status == catalogmodel.BookStatusPublished &&
		c.Price <= 0 &&
		c.AccessTier == catalogmodel.AccessTierFree
}

func isPaidCatalogBook(b *catalogmodel.CatalogBook) bool {
	return b.Status == catalogmodel.BookStatusPublished && !isFreeCatalogBook(b)
}

func isPaidCatalogCollection(c *catalogmodel.CatalogCollection) bool {
	return c.Status == catalogmodel.BookStatusPublished && !isFreeCatalogCollection(c)
}

func (s *Service) getBook(itemID string) (*catalogmodel.CatalogBook, error) {
	var b catalogmodel.CatalogBook
	if err := s.db.Where("id = ?", itemID).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Service) getCollection(itemID string) (*catalogmodel.CatalogCollection, error) {
	var c catalogmodel.CatalogCollection
	if err := s.db.Where("id = ?", itemID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) AddFreeItem(userID string, req *librarydto.AddLibraryItemRequest) (*librarydto.LibraryItemResponse, error) {
	var row *libmodel.UserLibraryItem

	switch req.ItemType {
	case libmodel.ItemTypeBook:
		b, err := s.getBook(req.ItemID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("obra não encontrada")
			}
			return nil, err
		}
		if !isFreeCatalogBook(b) {
			return nil, errors.New("apenas obras gratuitas e publicadas podem ser adicionadas manualmente; conteúdo pago exige compra ou liberação do autor")
		}
		item := &libmodel.UserLibraryItem{
			UserID:   userID,
			ItemType: libmodel.ItemTypeBook,
			ItemID:   b.ID,
			Source:   libmodel.SourceFreeSelf,
		}
		if err := s.db.Create(item).Error; err != nil {
			return nil, err
		}
		row = item

	case libmodel.ItemTypeCollection:
		c, err := s.getCollection(req.ItemID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("coleção não encontrada")
			}
			return nil, err
		}
		if !isFreeCatalogCollection(c) {
			return nil, errors.New("apenas coleções gratuitas e publicadas podem ser adicionadas manualmente; conteúdo pago exige compra ou liberação do autor")
		}
		item := &libmodel.UserLibraryItem{
			UserID:   userID,
			ItemType: libmodel.ItemTypeCollection,
			ItemID:   c.ID,
			Source:   libmodel.SourceFreeSelf,
		}
		if err := s.db.Create(item).Error; err != nil {
			return nil, err
		}
		row = item

	default:
		return nil, errors.New("item_type inválido")
	}

	return s.toLibraryItemResponse(row)
}

func (s *Service) RemoveFreeItem(userID, itemType, itemID string) error {
	res := s.db.Where("user_id = ? AND item_type = ? AND item_id = ? AND source = ?",
		userID, itemType, itemID, libmodel.SourceFreeSelf).Delete(&libmodel.UserLibraryItem{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *Service) ListLibrary(userID string) ([]librarydto.LibraryItemResponse, error) {
	var items []libmodel.UserLibraryItem
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]librarydto.LibraryItemResponse, 0, len(items))
	for i := range items {
		r, err := s.toLibraryItemResponse(&items[i])
		if err != nil {
			return nil, err
		}
		out = append(out, *r)
	}
	return out, nil
}

func (s *Service) toLibraryItemResponse(row *libmodel.UserLibraryItem) (*librarydto.LibraryItemResponse, error) {
	resp := &librarydto.LibraryItemResponse{
		ID:              row.ID,
		ItemType:        row.ItemType,
		ItemID:          row.ItemID,
		Source:          row.Source,
		PurchaseID:      row.PurchaseID,
		GrantedByUserID: row.GrantedByUserID,
		CreatedAt:       row.CreatedAt,
	}
	switch row.ItemType {
	case libmodel.ItemTypeBook:
		b, err := s.getBook(row.ItemID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return resp, nil
			}
			return nil, err
		}
		resp.Book = &librarydto.BookSummary{
			ID:            b.ID,
			Title:         b.Title,
			Slug:          b.Slug,
			CoverImageURL: b.CoverImageURL,
			Price:         b.Price,
			Currency:      b.Currency,
			AccessTier:    b.AccessTier,
			Status:        b.Status,
		}
	case libmodel.ItemTypeCollection:
		c, err := s.getCollection(row.ItemID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return resp, nil
			}
			return nil, err
		}
		resp.Collection = &librarydto.CollectionSummary{
			ID:            c.ID,
			Title:         c.Title,
			Slug:          c.Slug,
			CoverImageURL: c.CoverImageURL,
			Price:         c.Price,
			Currency:      c.Currency,
			AccessTier:    c.AccessTier,
			Status:        c.Status,
		}
	}
	return resp, nil
}

// AddOrSkipFromPurchase adiciona itens à biblioteca após compra (ignora se já existir).
func (s *Service) AddOrSkipFromPurchase(tx *gorm.DB, userID, purchaseID string, lines []struct {
	ItemType string
	ItemID   string
}) error {
	for _, ln := range lines {
		var existing libmodel.UserLibraryItem
		err := tx.Where("user_id = ? AND item_type = ? AND item_id = ?", userID, ln.ItemType, ln.ItemID).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		pid := purchaseID
		row := &libmodel.UserLibraryItem{
			UserID:     userID,
			ItemType:   ln.ItemType,
			ItemID:     ln.ItemID,
			Source:     libmodel.SourcePurchase,
			PurchaseID: &pid,
		}
		if err := tx.Create(row).Error; err != nil {
			return err
		}
	}
	return nil
}

// GrantFromAuthor concede acesso pago a um leitor (apenas dono do catálogo).
func (s *Service) GrantFromAuthor(authorUserID string, req *librarydto.AuthorGrantRequest) (*librarydto.LibraryItemResponse, error) {
	var readerCount int64
	if err := s.db.Model(&usermodel.User{}).Where("id = ?", req.ReaderUserID).Count(&readerCount).Error; err != nil {
		return nil, err
	}
	if readerCount == 0 {
		return nil, errors.New("usuário leitor não encontrado")
	}

	switch req.ItemType {
	case libmodel.ItemTypeBook:
		b, err := s.getBook(req.ItemID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("obra não encontrada")
			}
			return nil, err
		}
		if b.AuthorUserID != authorUserID {
			return nil, errors.New("apenas o autor da obra pode conceder acesso")
		}
		if !isPaidCatalogBook(b) {
			return nil, errors.New("obras gratuitas são adicionadas pelo próprio leitor; use conteúdo pago para liberação manual")
		}

	case libmodel.ItemTypeCollection:
		c, err := s.getCollection(req.ItemID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("coleção não encontrada")
			}
			return nil, err
		}
		if c.AuthorUserID != authorUserID {
			return nil, errors.New("apenas o autor da coleção pode conceder acesso")
		}
		if !isPaidCatalogCollection(c) {
			return nil, errors.New("coleções gratuitas são adicionadas pelo próprio leitor; use conteúdo pago para liberação manual")
		}

	default:
		return nil, errors.New("item_type inválido")
	}

	var existing libmodel.UserLibraryItem
	err := s.db.Where("user_id = ? AND item_type = ? AND item_id = ?", req.ReaderUserID, req.ItemType, req.ItemID).First(&existing).Error
	if err == nil {
		return s.toLibraryItemResponse(&existing)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	row := &libmodel.UserLibraryItem{
		UserID:          req.ReaderUserID,
		ItemType:        req.ItemType,
		ItemID:          req.ItemID,
		Source:          libmodel.SourceAuthorGrant,
		GrantedByUserID: &authorUserID,
	}
	if err := s.db.Create(row).Error; err != nil {
		return nil, err
	}
	return s.toLibraryItemResponse(row)
}

// UserHasLibraryAccess verifica se o usuário possui o livro ou coleção na biblioteca.
func (s *Service) UserHasLibraryAccess(userID, itemType, itemID string) (bool, error) {
	var n int64
	err := s.db.Model(&libmodel.UserLibraryItem{}).
		Where("user_id = ? AND item_type = ? AND item_id = ?", userID, itemType, itemID).
		Count(&n).Error
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// --- uso interno: catálogo pago para checkout ---

func (s *Service) ResolvePaidLine(itemType, itemID string) (unitPrice float64, currency string, title string, err error) {
	switch itemType {
	case libmodel.ItemTypeBook:
		b, e := s.getBook(itemID)
		if e != nil {
			if errors.Is(e, gorm.ErrRecordNotFound) {
				return 0, "", "", errors.New("obra não encontrada")
			}
			return 0, "", "", e
		}
		if b.Status != catalogmodel.BookStatusPublished {
			return 0, "", "", errors.New("obra não está publicada")
		}
		if !isPaidCatalogBook(b) {
			return 0, "", "", errors.New("obra não é paga; adicione gratuitamente à biblioteca")
		}
		return b.Price, b.Currency, b.Title, nil

	case libmodel.ItemTypeCollection:
		c, e := s.getCollection(itemID)
		if e != nil {
			if errors.Is(e, gorm.ErrRecordNotFound) {
				return 0, "", "", errors.New("coleção não encontrada")
			}
			return 0, "", "", e
		}
		if c.Status != catalogmodel.BookStatusPublished {
			return 0, "", "", errors.New("coleção não está publicada")
		}
		if !isPaidCatalogCollection(c) {
			return 0, "", "", errors.New("coleção não é paga; adicione gratuitamente à biblioteca")
		}
		return c.Price, c.Currency, c.Title, nil
	default:
		return 0, "", "", errors.New("item_type inválido")
	}
}
