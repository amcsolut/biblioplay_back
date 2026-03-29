package user

import (
	"time"

	"api-backend-infinitrum/internal/models/user"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(user *user.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetByID(id string) (*user.User, error) {
	var user user.User
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetByEmail(email string) (*user.User, error) {
	var user user.User
	err := r.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetAll(limit, offset int) ([]user.User, error) {
	var users []user.User
	err := r.db.Where("deleted_at IS NULL").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

func (r *Repository) Update(user *user.User) error {
	user.UpdatedAt = time.Now()
	return r.db.Save(user).Error
}

func (r *Repository) SoftDelete(id string) error {
	now := time.Now()
	return r.db.Model(&user.User{}).
		Where("id = ?", id).
		Update("deleted_at", now).Error
}

func (r *Repository) UpdatePassword(userID, hashedPassword string) error {
	return r.db.Model(&user.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"password_hash": hashedPassword,
			"updated_at":    time.Now(),
		}).Error
}

func (r *Repository) SaveRefreshToken(userID, refreshToken string, expiresAt time.Time) error {
	// First, delete any existing refresh tokens for this user
	r.db.Where("user_id = ?", userID).Delete(&user.UserSession{})

	// Create new session with refresh token
	session := &user.UserSession{
		UserID:       userID,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}

	return r.db.Create(session).Error
}

func (r *Repository) GetRefreshToken(userID string) (*user.UserSession, error) {
	var session user.UserSession
	err := r.db.Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *Repository) GetRefreshTokenByToken(userID, refreshToken string) (*user.UserSession, error) {
	var session user.UserSession
	err := r.db.Where("user_id = ? AND refresh_token = ? AND expires_at > ?", userID, refreshToken, time.Now()).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *Repository) DeleteRefreshToken(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&user.UserSession{}).Error
}

func (r *Repository) GetUserCount() (int64, error) {
	var count int64
	err := r.db.Model(&user.User{}).Where("deleted_at IS NULL").Count(&count).Error
	return count, err
}

func (r *Repository) SearchUsers(query string, limit, offset int) ([]user.User, error) {
	var users []user.User
	searchQuery := "%" + query + "%"
	
	err := r.db.Where("deleted_at IS NULL").
		Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?", 
			searchQuery, searchQuery, searchQuery).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error
	
	return users, err
}

func (r *Repository) UpdateLastLogin(userID string) error {
	return r.db.Model(&user.User{}).
		Where("id = ?", userID).
		Update("updated_at", time.Now()).Error
}

func (r *Repository) GetByProviderID(provider string, providerID string) (*user.User, error) {
	var user user.User
	err := r.db.Where("provider = ? AND provider_id = ? AND deleted_at IS NULL", provider, providerID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Password Reset methods
func (r *Repository) CreatePasswordResetToken(resetToken *user.PasswordReset) error {
	// Delete any existing unused tokens for this user
	r.db.Where("user_id = ? AND used = ?", resetToken.UserID, false).Delete(&user.PasswordReset{})
	return r.db.Create(resetToken).Error
}

func (r *Repository) GetPasswordResetToken(token string) (*user.PasswordReset, error) {
	var resetToken user.PasswordReset
	err := r.db.Where("token = ? AND used = ? AND expires_at > ?", token, false, time.Now()).
		First(&resetToken).Error
	if err != nil {
		return nil, err
	}
	return &resetToken, nil
}

func (r *Repository) MarkPasswordResetTokenAsUsed(token string) error {
	return r.db.Model(&user.PasswordReset{}).
		Where("token = ?", token).
		Update("used", true).Error
}

