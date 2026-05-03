package repositories

import (
	"github.com/Haidarr-h/backend-go/models"
	"github.com/Haidarr-h/backend-go/pkg/logger"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CREATE USER
func (r *UserRepository) CreateUser(user models.User) (models.User, error) {

	// 1. Create directly
	// result := r.db.Create(&user)
	query := "INSERT INTO users (email, first_name, last_name, username, password, is_verified, google_id, picture, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()) RETURNING *"
	result := r.db.Raw(query, user.Email, user.FirstName, user.LastName, user.Username, user.Password, user.IsVerified, user.GoogleID, user.Picture).Scan(&user)

	// 2. check error
	if result.Error != nil {
		return models.User{}, result.Error
	}

	if result.RowsAffected == 0 {
		return models.User{}, ErrUserCreation
	}

	// 3. return success
	return user, nil
}

// UPDATE USER
func (r *UserRepository) UpdateUser(user models.User) (models.User, error) {

	// 1. query
	query := "UPDATE users SET google_id = ?, picture = ?, is_verified = ? WHERE id = ? RETURNING *"
	result := r.db.Raw(query, user.GoogleID, user.Picture, user.IsVerified, user.ID).Scan(&user)

	// 2. error check
	if result.Error != nil {
		logger.Log.Error("UpdateUser query function failed", "error", result.Error)
		return models.User{}, result.Error
	}

	if result.RowsAffected == 0 {
		logger.Log.Error("UpdateUser query gives no effect", "error", result.Error)
		return models.User{}, ErrUserUpdate
	}

	// 3. success
	return user, nil
}

// FIND USER BY EMAIL
func (r *UserRepository) FindByEmail(email string) (models.User, error) {

	var user models.User

	// 1. Search to database
	result := r.db.Raw("SELECT * FROM users WHERE email = ? LIMIT 1", email).Scan(&user)
	// result := r.db.Where("email = ?", email).First(&user)

	// 2. Check error
	if result.Error != nil {
		return models.User{}, result.Error
	}

	if result.RowsAffected == 0 {
		return models.User{}, ErrUserNotFound
	}

	// 3. return if found
	return user, nil
}

// FIND USER BY USERNAME
func (r *UserRepository) FindByUsername(username string) (models.User, error) {

	var user models.User

	// 1. Search to database
	query := "SELECT * FROM users WHERE username = ?"
	result := r.db.Raw(query, username).Scan(&user)

	// 2. Check error
	if result.Error != nil {
		logger.Log.Error("FindByUsername query function failed", "error", result.Error, "username", username)
		return models.User{}, result.Error
	}

	// check empty
	if result.RowsAffected == 0 {
		return models.User{}, ErrUserNotFound
	}

	// 3. return if found
	return user, nil
}

// FIND USER BY GOOGLE ID
func (r *UserRepository) FindByGoogleID(googleID string) (models.User, error) {

	var user models.User

	// 1. Search to database
	query := "SELECT * FROM users WHERE google_id = ? LIMIT 1"
	result := r.db.Raw(query, googleID).Scan(&user)

	// 2. Check error
	if result.Error != nil {
		logger.Log.Error("FindByGoogleID query function failed", "error", result.Error, "id", googleID)
		return models.User{}, result.Error
	}

	// check empty
	if result.RowsAffected == 0 {
		return models.User{}, ErrUserNotFound
	}

	// 3. return if found
	return user, nil
}

// IS USER EXIST BY google id
func (r *UserRepository) ExistByGoogleID(googleID string) (bool, error) {

	var count int64

	// 1. query
	query := "SELECT COUNT(*) FROM users WHERE google_id = ?"
	result := r.db.Raw(query, googleID).Scan(&count)

	// 2. error checks
	if result.Error != nil {
		logger.Log.Error("ExistByGoogleID function query failed", "error", result.Error)
		return false, result.Error
	}

	return count > 0, nil
}

// IS USER EXIST BY EMAIL
func (r *UserRepository) ExistByEmail(email string) (bool, error) {
	var count int64

	// 1. query to db
	query := "SELECT COUNT(*) FROM users where email = ?"
	result := r.db.Raw(query, email).Scan(&count)

	// 2. failed query (db error)
	if result.Error != nil {
		logger.Log.Error("ExistByEmail function query failed", "email", email, "error", result.Error)
		return false, result.Error
	}

	// 3. success
	return count > 0, nil
}

// IS USER EXIST BY USERNAME
func (r *UserRepository) ExistByUsername(username string) (bool, error) {
	var count int64

	// 1. query to db
	query := "SELECT COUNT(*) FROM users where username = ?"
	result := r.db.Raw(query, username).Scan(&count)

	// 2. failed query (db error)
	if result.Error != nil {
		logger.Log.Error("ExistByUsername function query failed", "username", username, "error", result.Error)
		return false, result.Error
	}

	// 3. success
	return count > 0, nil
}

// SET VERIFY
func (r *UserRepository) Verify(id uint) error {

	// 1. query
	query := `
		UPDATE users
		SET is_verified = TRUE
		WHERE id = ?
	`
	result := r.db.Exec(query, id)

	// 2. error
	if result.Error != nil {
		logger.Log.Error("verify user update function failed", "error", result.Error)
		return result.Error
	}

	// 3. success response
	return nil
}
