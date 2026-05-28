package auth

import (
	"time"

	"github.com/Haidarr-h/backend-go/internal/user"
	"gorm.io/gorm"
)

type OtpVerification struct {
	gorm.Model
	UserID    uint      `gorm:"not null;index"`
	OTPHash   string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Attempts  int       `gorm:"default:0"`
	Used      bool      `gorm:"default:false"`
	User      user.User      `gorm:"foreignKey:UserID"`
}

type RefreshToken struct {
	gorm.Model
	UserID    uint
	User      user.User `gorm:"foreignKey:UserID"`
	Token     string `gorm:"uniqueIndex"`
	ExpiresAt time.Time
}
