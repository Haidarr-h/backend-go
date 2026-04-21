package models

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	gorm.Model
	UserID    uint
	User      User `gorm:"foreignKey:UserID"`
	Token     string `gorm:"uniqueIndex"`
	ExpiresAt time.Time
}
