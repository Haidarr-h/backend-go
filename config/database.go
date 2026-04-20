package config

import (
	"fmt"
	"os"

	"github.com/Haidarr-h/backend-go/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := os.Getenv("DB")

	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database")
	}

	fmt.Println("Database connection success")

	if err = DB.AutoMigrate(&models.User{}, &models.Exercise{}, &models.Routine{}, &models.RoutineExercises{}); err != nil {
		fmt.Println("Auto migration error:", err)
	} else {
		fmt.Println("Auto migration success:")
	}
}