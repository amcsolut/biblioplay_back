package catalog

import (
	"errors"
	"time"

	catalogDTO "api-backend-infinitrum/api/v1/dto/catalog"
	catalogmodel "api-backend-infinitrum/internal/models/catalog"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// --- Books ---

func (s *Service) CreateBook(authorUserID string, req *catalogDTO.CreateBookRequest) (*catalogDTO.BookResponse, error) {
	b := &catalogmodel.CatalogBook{
		AuthorUserID:  authorUserID,
		Title:         req.Title,
		Slug:          req.Slug,
		Subtitle:      req.Subtitle,
		Synopsis:      req.Synopsis,
		CoverImageURL: req.CoverImageURL,
	}
	if req.Price != nil {
		b.Price = *req.Price
	}
	if req.Currency != "" {
		b.Currency = req.Currency
	}
	if req.AccessTier != "" {
		b.AccessTier = req.AccessTier
	}
	if req.Status != "" {
		b.Status = req.Status
	}
	if req.Language != "" {
		b.Language = req.Language
	}
	if req.PublishedAt != nil && *req.PublishedAt != "" {
		t, err := time.Parse(time.RFC3339, *req.PublishedAt)
		if err != nil {
			return nil, errors.New("published_at inválido: use RFC3339")
		}
		b.PublishedAt = &t
	}
	if err := s.db.Create(b).Error; err != nil {
		return nil, err
	}
	return toBookResponse(b), nil
}

func (s *Service) ListBooks(authorUserID string) ([]catalogDTO.BookResponse, error) {
	var list []catalogmodel.CatalogBook
	if err := s.db.Where("author_user_id = ?", authorUserID).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]catalogDTO.BookResponse, 0, len(list))
	for i := range list {
		out = append(out, *toBookResponse(&list[i]))
	}
	return out, nil
}

func (s *Service) GetBook(id, authorUserID string) (*catalogDTO.BookResponse, error) {
	var b catalogmodel.CatalogBook
	if err := s.db.Where("id = ? AND author_user_id = ?", id, authorUserID).First(&b).Error; err != nil {
		return nil, err
	}
	return toBookResponse(&b), nil
}

func (s *Service) UpdateBook(id, authorUserID string, req *catalogDTO.UpdateBookRequest) (*catalogDTO.BookResponse, error) {
	var b catalogmodel.CatalogBook
	if err := s.db.Where("id = ? AND author_user_id = ?", id, authorUserID).First(&b).Error; err != nil {
		return nil, err
	}
	updates := map[string]interface{}{"updated_at": time.Now()}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Slug != nil {
		updates["slug"] = *req.Slug
	}
	if req.Subtitle != nil {
		updates["subtitle"] = *req.Subtitle
	}
	if req.Synopsis != nil {
		updates["synopsis"] = *req.Synopsis
	}
	if req.CoverImageURL != nil {
		updates["cover_image_url"] = *req.CoverImageURL
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.Currency != nil {
		updates["currency"] = *req.Currency
	}
	if req.AccessTier != nil {
		updates["access_tier"] = *req.AccessTier
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Language != nil {
		updates["language"] = *req.Language
	}
	if req.PublishedAt != nil {
		if *req.PublishedAt == "" {
			updates["published_at"] = nil
		} else {
			t, err := time.Parse(time.RFC3339, *req.PublishedAt)
			if err != nil {
				return nil, errors.New("published_at inválido: use RFC3339")
			}
			updates["published_at"] = t
		}
	}
	if err := s.db.Model(&b).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&b, "id = ?", b.ID).Error; err != nil {
		return nil, err
	}
	return toBookResponse(&b), nil
}

func (s *Service) DeleteBook(id, authorUserID string) error {
	res := s.db.Where("id = ? AND author_user_id = ?", id, authorUserID).Delete(&catalogmodel.CatalogBook{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func toBookResponse(b *catalogmodel.CatalogBook) *catalogDTO.BookResponse {
	return &catalogDTO.BookResponse{
		ID:            b.ID,
		AuthorUserID:  b.AuthorUserID,
		Title:         b.Title,
		Slug:          b.Slug,
		Subtitle:      b.Subtitle,
		Synopsis:      b.Synopsis,
		CoverImageURL: b.CoverImageURL,
		Price:         b.Price,
		Currency:      b.Currency,
		AccessTier:    b.AccessTier,
		Status:        b.Status,
		Language:      b.Language,
		PublishedAt:   b.PublishedAt,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}

func (s *Service) bookOwnedBy(bookID, authorUserID string) (*catalogmodel.CatalogBook, error) {
	var b catalogmodel.CatalogBook
	if err := s.db.Where("id = ? AND author_user_id = ?", bookID, authorUserID).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

// --- Ebook chapters ---

func (s *Service) ListEbookChapters(bookID, authorUserID string) ([]catalogDTO.EbookChapterResponse, error) {
	if _, err := s.bookOwnedBy(bookID, authorUserID); err != nil {
		return nil, err
	}
	var list []catalogmodel.EbookChapter
	if err := s.db.Where("book_id = ?", bookID).Order("position ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]catalogDTO.EbookChapterResponse, 0, len(list))
	for i := range list {
		out = append(out, *toEbookChapterResponse(&list[i]))
	}
	return out, nil
}

func (s *Service) GetEbookChapter(bookID, chapterID, authorUserID string) (*catalogDTO.EbookChapterResponse, error) {
	if _, err := s.bookOwnedBy(bookID, authorUserID); err != nil {
		return nil, err
	}
	var ch catalogmodel.EbookChapter
	if err := s.db.Where("id = ? AND book_id = ?", chapterID, bookID).First(&ch).Error; err != nil {
		return nil, err
	}
	return toEbookChapterResponse(&ch), nil
}

func (s *Service) CreateEbookChapter(bookID, authorUserID string, req *catalogDTO.CreateEbookChapterRequest) (*catalogDTO.EbookChapterResponse, error) {
	if _, err := s.bookOwnedBy(bookID, authorUserID); err != nil {
		return nil, err
	}
	ch := &catalogmodel.EbookChapter{
		BookID:            bookID,
		Position:          req.Position,
		Title:             req.Title,
		BodyText:          req.BodyText,
		ContentStorageKey: req.ContentStorageKey,
	}
	if err := s.db.Create(ch).Error; err != nil {
		return nil, err
	}
	return toEbookChapterResponse(ch), nil
}

func (s *Service) UpdateEbookChapter(bookID, chapterID, authorUserID string, req *catalogDTO.UpdateEbookChapterRequest) (*catalogDTO.EbookChapterResponse, error) {
	if _, err := s.bookOwnedBy(bookID, authorUserID); err != nil {
		return nil, err
	}
	var ch catalogmodel.EbookChapter
	if err := s.db.Where("id = ? AND book_id = ?", chapterID, bookID).First(&ch).Error; err != nil {
		return nil, err
	}
	updates := map[string]interface{}{"updated_at": time.Now()}
	if req.Position != nil {
		updates["position"] = *req.Position
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.BodyText != nil {
		updates["body_text"] = *req.BodyText
	}
	if req.ContentStorageKey != nil {
		updates["content_storage_key"] = *req.ContentStorageKey
	}
	if err := s.db.Model(&ch).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&ch, "id = ?", ch.ID).Error; err != nil {
		return nil, err
	}
	return toEbookChapterResponse(&ch), nil
}

func (s *Service) DeleteEbookChapter(bookID, chapterID, authorUserID string) error {
	if _, err := s.bookOwnedBy(bookID, authorUserID); err != nil {
		return err
	}
	res := s.db.Where("id = ? AND book_id = ?", chapterID, bookID).Delete(&catalogmodel.EbookChapter{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func toEbookChapterResponse(ch *catalogmodel.EbookChapter) *catalogDTO.EbookChapterResponse {
	return &catalogDTO.EbookChapterResponse{
		ID:                ch.ID,
		BookID:            ch.BookID,
		Position:          ch.Position,
		Title:             ch.Title,
		BodyText:          ch.BodyText,
		ContentStorageKey: ch.ContentStorageKey,
		CreatedAt:         ch.CreatedAt,
		UpdatedAt:         ch.UpdatedAt,
	}
}

// --- Audiobook chapters ---

func (s *Service) ListAudiobookChapters(bookID, authorUserID string) ([]catalogDTO.AudiobookChapterResponse, error) {
	if _, err := s.bookOwnedBy(bookID, authorUserID); err != nil {
		return nil, err
	}
	var list []catalogmodel.AudiobookChapter
	if err := s.db.Where("book_id = ?", bookID).Order("position ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]catalogDTO.AudiobookChapterResponse, 0, len(list))
	for i := range list {
		out = append(out, *toAudiobookChapterResponse(&list[i]))
	}
	return out, nil
}

func (s *Service) GetAudiobookChapter(bookID, chapterID, authorUserID string) (*catalogDTO.AudiobookChapterResponse, error) {
	if _, err := s.bookOwnedBy(bookID, authorUserID); err != nil {
		return nil, err
	}
	var ch catalogmodel.AudiobookChapter
	if err := s.db.Where("id = ? AND book_id = ?", chapterID, bookID).First(&ch).Error; err != nil {
		return nil, err
	}
	return toAudiobookChapterResponse(&ch), nil
}

func (s *Service) CreateAudiobookChapter(bookID, authorUserID string, req *catalogDTO.CreateAudiobookChapterRequest) (*catalogDTO.AudiobookChapterResponse, error) {
	if _, err := s.bookOwnedBy(bookID, authorUserID); err != nil {
		return nil, err
	}
	ch := &catalogmodel.AudiobookChapter{
		BookID:          bookID,
		Position:        req.Position,
		Title:           req.Title,
		AudioStorageKey: req.AudioStorageKey,
		AudioURL:        req.AudioURL,
		DurationSeconds: req.DurationSeconds,
	}
	if err := s.db.Create(ch).Error; err != nil {
		return nil, err
	}
	return toAudiobookChapterResponse(ch), nil
}

func (s *Service) UpdateAudiobookChapter(bookID, chapterID, authorUserID string, req *catalogDTO.UpdateAudiobookChapterRequest) (*catalogDTO.AudiobookChapterResponse, error) {
	if _, err := s.bookOwnedBy(bookID, authorUserID); err != nil {
		return nil, err
	}
	var ch catalogmodel.AudiobookChapter
	if err := s.db.Where("id = ? AND book_id = ?", chapterID, bookID).First(&ch).Error; err != nil {
		return nil, err
	}
	updates := map[string]interface{}{"updated_at": time.Now()}
	if req.Position != nil {
		updates["position"] = *req.Position
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.AudioStorageKey != nil {
		updates["audio_storage_key"] = *req.AudioStorageKey
	}
	if req.AudioURL != nil {
		updates["audio_url"] = *req.AudioURL
	}
	if req.DurationSeconds != nil {
		updates["duration_seconds"] = *req.DurationSeconds
	}
	if err := s.db.Model(&ch).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&ch, "id = ?", ch.ID).Error; err != nil {
		return nil, err
	}
	return toAudiobookChapterResponse(&ch), nil
}

func (s *Service) DeleteAudiobookChapter(bookID, chapterID, authorUserID string) error {
	if _, err := s.bookOwnedBy(bookID, authorUserID); err != nil {
		return err
	}
	res := s.db.Where("id = ? AND book_id = ?", chapterID, bookID).Delete(&catalogmodel.AudiobookChapter{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func toAudiobookChapterResponse(ch *catalogmodel.AudiobookChapter) *catalogDTO.AudiobookChapterResponse {
	return &catalogDTO.AudiobookChapterResponse{
		ID:              ch.ID,
		BookID:          ch.BookID,
		Position:        ch.Position,
		Title:           ch.Title,
		AudioStorageKey: ch.AudioStorageKey,
		AudioURL:        ch.AudioURL,
		DurationSeconds: ch.DurationSeconds,
		CreatedAt:       ch.CreatedAt,
		UpdatedAt:       ch.UpdatedAt,
	}
}
