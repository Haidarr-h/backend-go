package repositories

import (
	"errors"

	"github.com/Haidarr-h/backend-go/models"
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

// FIND USER BY USERNAME
func (r *UserRepository) FindByUsername(username string) (models.User, error) {

	var user models.User

	// 1. Search to database
	result := r.db.Where("username = ?", username).First(&user)

	// 2. Check error
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, result.Error
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
		return models.User{}, errors.New("failed to create user")
	}

	// 3. return success
	return user, nil
}
