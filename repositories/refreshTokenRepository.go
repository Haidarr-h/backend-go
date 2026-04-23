package repositories

import (
	"errors"

	"github.com/Haidarr-h/backend-go/models"
	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// CREATE
func (r *RefreshTokenRepository) Create(refreshToken *models.RefreshToken) (*models.RefreshToken, error) {

	// 1. Create directly
	query := "INSERT INTO refresh_tokens (user_id, token, expires_at, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW()) RETURNING *"

	result := r.db.Raw(query, refreshToken.UserID, refreshToken.Token, refreshToken.ExpiresAt).Scan(&refreshToken)

	// 2. Error checking
	if result.Error != nil {
		return nil, result.Error
	}

	if refreshToken.ID == 0 {
		return nil, ErrGenerateToken
	}

	// 3. Result
	return refreshToken, nil
}

// FINDBYTOKEN
func (r *RefreshTokenRepository) FindByToken(token string) (*models.RefreshToken, error) {

	var refreshToken models.RefreshToken

	// 1. Find
	query := "SELECT * FROM refresh_tokens WHERE token = ? LIMIT 1"
	result := r.db.Raw(query, token).Scan(&refreshToken)

	// 2. error checks
	if result.Error != nil {
		return nil, result.Error
	}

	if refreshToken.ID == 0 {
		return nil, errors.New("failed to find refresh token")
	}

	// 3. return
	return &refreshToken, nil
}

// UPDATE
func (r *RefreshTokenRepository) Update(refreshToken *models.RefreshToken) (*models.RefreshToken, error) {

	// 1. update direcly
	query := "UPDATE refresh_tokens SET token = ?, expires_at = ?, updated_at = NOW() WHERE id = ? RETURNING *"
	result := r.db.Raw(query, refreshToken.Token, refreshToken.ExpiresAt, refreshToken.ID).Scan(&refreshToken)

	// 2. Error checks
	if result.Error != nil {
		return nil, result.Error
	}

	if refreshToken.ID == 0 {
		return nil, errors.New("failed to update refresh token")
	}

	// 3. return
	return refreshToken, nil
}

// DELETE BY Token
func (r *RefreshTokenRepository) DeleteByToken(token string) error {

	// 1. Delete
	query := "DELETE from refresh_tokens WHERE token = ?"
	result := r.db.Exec(query, token)

	// 2. error checks
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("failed to delete refresh token")
	}

	return nil
}

// DELETE BY user ID
func (r *RefreshTokenRepository) DeleteByUserID(userID uint) error {

	// 1. Delete
	query := "DELETE from refresh_tokens WHERE user_id = ?"
	result := r.db.Exec(query, userID)

	// 2. error checks
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("failed to delete refresh token")
	}

	return nil
}
