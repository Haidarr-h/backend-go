package initializers

import "github.com/Haidarr-h/backend-go/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})
}