package user

import "github.com/Haidarr-h/backend-go/models"

type Repository interface {
    CreateUser(user models.User) (models.User, error)
    UpdateUser(user models.User) (models.User, error)
    FindByEmail(email string) (models.User, error)
    FindByUsername(username string) (models.User, error)
    FindByGoogleID(googleID string) (models.User, error)
    ExistByEmail(email string) (bool, error)
    ExistByUsername(username string) (bool, error)
    Verify(userID uint) error
}