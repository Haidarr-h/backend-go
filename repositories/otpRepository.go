package repositories

import (
	"github.com/Haidarr-h/backend-go/models"
	"github.com/Haidarr-h/backend-go/pkg/logger"
	"gorm.io/gorm"
)

type OtpRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) *OtpRepository {
	return &OtpRepository{db: db}
}

func (r *OtpRepository) Create(otp models.OtpVerification) (models.OtpVerification, error) {

	// 1. query
	query := "INSERT INTO otp_verifications (user_id, otp_hash, expires_at, attempts, used, created_at, updated_at) VALUES (?,?,?,?,?,NOW(), NOW()) RETURNING *"
	result := r.db.Raw(query, otp.UserID, otp.OTPHash, otp.ExpiresAt, otp.Attempts, otp.Used).Scan(&otp)

	// 2. error checks
	if result.Error != nil {
		logger.Log.Error("failed to create opt data to database", "error", result.Error)
		return models.OtpVerification{}, result.Error
	}

	// 3. return
	return otp, nil
}

func (r *OtpRepository) FindValidOtp(otp models.OtpVerification) (models.OtpVerification, error) {
	// 1. query
	query := "SELECT "
}
