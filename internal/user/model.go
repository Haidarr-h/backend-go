package user

import (
	"github.com/Haidarr-h/backend-go/internal/routine"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email      string  `gorm:"uniqueIndex;not null"`
	FirstName  string  `gorm:"not null"`
	LastName   string  `gorm:"not null"`
	Username   string  `gorm:"uniqueIndex;not null"`
	GoogleID   *string `gorm:"default:null"`
	Picture    *string
	Password   *string
	Routines   []routine.Routine
	IsVerified bool `gorm:"default:false"`
}
