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

// CREATE USER
func (r *UserRepository) CreateUser(user models.User) (models.User, error) {

	// 1. Create directly
	// result := r.db.Create(&user)
	query := "INSERT INTO users (email, first_name, last_name, username, password, created_at, updated_at) VALUES (?, ?, ?, ?, ?, NOW(), NOW()) RETURNING *"
	result := r.db.Raw(query, user.Email, user.FirstName, user.LastName, user.Username, user.Password).Scan(&user)

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
