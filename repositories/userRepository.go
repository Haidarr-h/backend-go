package repositories

import (
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
	result := r.db.Where("email = ?", email).First(&user)

	// 2. Check error
	if result.Error != nil {
		return models.User{}, result.Error
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
		return models.User{}, result.Error
	}

	// 3. return if found
	return user, nil
}

// CREATE USER
func (r *UserRepository) CreateUser(user models.User) (models.User, error) {

	// 1. Create directly
	result := r.db.Create(&user)

	// 2. check error
	if result.Error != nil {
		return  models.User{}, result.Error
	}

	// 3. return success
	return user, nil
}