package initializers

import (
	"fmt"

	"github.com/Haidarr-h/backend-go/models"
)

func SyncDatabase() {
	err := DB.AutoMigrate(&models.User{})

	if err != nil {
		fmt.Println("Auto Migration error: ", err)
	} else {
		fmt.Println("Database sync successful")
	}
}