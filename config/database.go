package config

import (
	"fmt"
	"os"

	"github.com/Haidarr-h/backend-go/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	var err error
	dsn := os.Getenv("DB")

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database")
	}

	fmt.Println("Database connection success")

	if os.Getenv("AUTO_MIGRATE") == "true" {
		if err = db.AutoMigrate(&models.User{}, &models.Exercise{}, &models.Routine{}, &models.RoutineExercises{}, &models.RefreshToken{}); err != nil {
			fmt.Println("Auto migration error:", err)
		} else {
			fmt.Println("Auto migration success:")
		}
	}

	return db
}
