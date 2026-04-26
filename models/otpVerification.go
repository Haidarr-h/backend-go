package models

import (
	"time"

	"gorm.io/gorm"
)

type OtpVerification struct {
	gorm.Model
	UserID    uint      `gorm:"not null;index"`
	OTPHash   string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Attempts  int       `gorm:"default:0"`
	Used      bool      `gorm:"default:false"`
	User      User      `gorm:"foreignKey:UserID"`
}
