package auth

import (
	"github.com/Haidarr-h/backend-go/pkg/logger"
	"gorm.io/gorm"
)

type OtpRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) *OtpRepository {
	return &OtpRepository{db: db}
}

func (r *OtpRepository) Create(otp OtpVerification) (OtpVerification, error) {

	// 1. query
	query := "INSERT INTO otp_verifications (user_id, otp_hash, expires_at, attempts, used, created_at, updated_at) VALUES (?,?,?,?,?,NOW(), NOW()) RETURNING *"
	result := r.db.Raw(query, otp.UserID, otp.OTPHash, otp.ExpiresAt, otp.Attempts, otp.Used).Scan(&otp)

	// 2. error checks
	if result.Error != nil {
		logger.Log.Error("failed to create otp data to database", "error", result.Error)
		return OtpVerification{}, result.Error
	}

	// 3. return
	return otp, nil
}

// FIND OTP
func (r *OtpRepository) FindByEmail(email string) (OtpVerification, error) {
	var otp OtpVerification

	// 1. query
	query := `
		SELECT otp_verifications.* 
		FROM otp_verifications
		INNER JOIN users
		ON otp_verifications.user_id = users.id
		WHERE users.email = ?
		ORDER BY otp_verifications.created_at DESC
		LIMIT 1
		`
	result := r.db.Raw(query, email).Scan(&otp)

	// 2. error check
	if result.Error != nil {
		logger.Log.Error("findValidOtp query function failed", "error", result.Error)
		return OtpVerification{}, result.Error
	}

	if result.RowsAffected == 0 {
		return OtpVerification{}, ErrEmailNotFound
	}

	// 3. success return
	return otp, nil
}

// UPDATE OTP
func (r *OtpRepository) Update(otp OtpVerification) (OtpVerification, error) {

	// 1. query
	query := `
		UPDATE otp_verifications
		SET attempts = ?, used = ?
		WHERE id = ?
		RETURNING *
	`
	result := r.db.Raw(query, otp.Attempts, otp.Used, otp.ID).Scan(&otp)

	// 2. error check
	if result.Error != nil {
		logger.Log.Error("update otp query function failed", "error", result.Error)
		return OtpVerification{}, result.Error
	}

	if result.RowsAffected == 0 {
		logger.Log.Error("update otp query function failed. no rows updated")
		return OtpVerification{}, ErrOTPUpdate
	}

	// 3. success response
	return otp, nil
}
