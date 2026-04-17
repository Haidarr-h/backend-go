package config

import (
	"fmt"

	"github.com/Haidarr-h/backend-go/models"
)

func SyncDatabase() {
	err := DB.AutoMigrate(&models.User{}, &models.Exercise{}, &models.Routine{}, &models.RoutineExercises{})

	if err != nil {
		fmt.Println("Auto Migration error: ", err)
	} else {
		fmt.Println("Database sync successful")
	}
}
