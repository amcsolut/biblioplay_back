package profile

import (
	"crypto/rand"
	"errors"
	"strings"

	profileModel "api-backend-infinitrum/internal/models/profile"

	"gorm.io/gorm"
)

func randomAlnum(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateMember(tx *gorm.DB, m *profileModel.ProfileMember) error {
	return tx.Create(m).Error
}

func (r *Repository) CreateAuthor(tx *gorm.DB, a *profileModel.ProfileAuthor) error {
	return tx.Create(a).Error
}

func (r *Repository) UsernameExists(tx *gorm.DB, username string) (bool, error) {
	db := tx
	if db == nil {
		db = r.db
	}
	var count int64
	err := db.Model(&profileModel.ProfileMember{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) SlugExists(tx *gorm.DB, slug string) (bool, error) {
	db := tx
	if db == nil {
		db = r.db
	}
	var count int64
	err := db.Model(&profileModel.ProfileAuthor{}).Where("slug = ?", slug).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AuthorOrCommunitySlugTaken indica se o slug já está em uso em profile_authors ou communities.
func (r *Repository) AuthorOrCommunitySlugTaken(tx *gorm.DB, slug string) (bool, error) {
	db := tx
	if db == nil {
		db = r.db
	}
	var n int64
	if err := db.Model(&profileModel.ProfileAuthor{}).Where("slug = ?", slug).Count(&n).Error; err != nil {
		return false, err
	}
	if n > 0 {
		return true, nil
	}
	if err := db.Model(&profileModel.Community{}).Where("slug = ?", slug).Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *Repository) CreateCommunity(tx *gorm.DB, c *profileModel.Community) error {
	return tx.Create(c).Error
}

// AllocateUniqueAuthorCommunitySlug gera slug único em profile_authors e communities a partir do pen name.
func (r *Repository) AllocateUniqueAuthorCommunitySlug(tx *gorm.DB, penName string) (string, error) {
	base := profileModel.SlugFromPenName(penName)
	for i := 0; i < 50; i++ {
		candidate := base
		if i > 0 {
			suf := randomAlnum(6)
			maxBase := profileModel.MaxAuthorCommunitySlugLen - 1 - len(suf)
			if maxBase < 3 {
				maxBase = 3
			}
			b := base
			if len(b) > maxBase {
				b = strings.TrimRight(b[:maxBase], "-")
			}
			candidate = b + "-" + suf
		}
		if len(candidate) > profileModel.MaxAuthorCommunitySlugLen {
			candidate = strings.TrimRight(candidate[:profileModel.MaxAuthorCommunitySlugLen], "-")
		}
		taken, err := r.AuthorOrCommunitySlugTaken(tx, candidate)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
	}
	return "", errors.New("could not allocate slug")
}
