package services

import (
	"errors"
	"time"

	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/dto"
	"github.com/Haidarr-h/backend-go/models"
	"github.com/Haidarr-h/backend-go/pkg/otp"
	"github.com/Haidarr-h/backend-go/pkg/utils"
	"github.com/Haidarr-h/backend-go/repositories"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	authRepo    *repositories.UserRepository
	refreshRepo *repositories.RefreshTokenRepository
	otpRepo     *repositories.OtpRepository
	cfg         *config.Config
}

func NewAuthService(authRepo *repositories.UserRepository, cfg *config.Config, refreshRepo *repositories.RefreshTokenRepository, otpRepo *repositories.OtpRepository) *AuthService {
	return &AuthService{authRepo: authRepo, cfg: cfg, refreshRepo: refreshRepo, otpRepo: otpRepo}
}

// SIGN UP
func (s *AuthService) SignUp(req dto.SignUpRequest) (dto.SignUpResponse, error) {

	// 1. Check if user email already exist
	isEmailExist, emailErr := s.authRepo.ExistByEmail(req.Email)
	if emailErr != nil {
		return dto.SignUpResponse{}, emailErr
	}

	if isEmailExist {
		return dto.SignUpResponse{}, ErrEmailIsExists
	}

	// 2. Check if username already exist
	isUsernameExist, usernameErr := s.authRepo.ExistByUsername(req.Username)
	if usernameErr != nil {
		return dto.SignUpResponse{}, usernameErr
	}

	if isUsernameExist {
		return dto.SignUpResponse{}, ErrUsernameIsExists
	}

	// 3. Hash the passowrd
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return dto.SignUpResponse{}, err
	}

	// 4. Create the user model
	hashedPassword := string(hash)
	user := models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		Password:  &hashedPassword,
	}

	// 5. create the user
	result, err := s.authRepo.CreateUser(user)
	if err != nil {
		return dto.SignUpResponse{}, err
	}

	// 6. generate OTP
	plainOTP, hashedOTP, otpErr := otp.GenerateOtp()

	if otpErr != nil {
		return dto.SignUpResponse{}, otpErr
	}

	// 7. create data at OTP table
	otpData := models.OtpVerification{
		UserID: result.ID,
		OTPHash: hashedOTP,
		ExpiresAt: time.Now().Add(time.Minute * 60),
		Attempts: 0,
		Used: false,
	}

	if _, createOtpErr := s.otpRepo.Create(otpData); createOtpErr != nil {
		return dto.SignUpResponse{}, createOtpErr
	}

	// 8. send OTP to the email
	if err := otp.SendOTP(user.Email, plainOTP, s.cfg); err != nil {
		return dto.SignUpResponse{}, err
	}

	// 9. success response
	response := dto.SignUpResponse{
		ID:         result.ID,
		FirstName:  result.FirstName,
		LastName:   result.LastName,
		Username:   result.Username,
		IsVerified: result.IsVerified,
	}

	return response, nil
}

// SIGN IN
func (s *AuthService) SignIn(req dto.SignInReq) (dto.SignInRes, error) {

	// 1. check sign in by email or username
	isEmail := utils.IsEmail(req.Identifier)

	var user models.User
	var err error

	// 2. check if user exist
	if isEmail {
		user, err = s.authRepo.FindByEmail(req.Identifier)
	} else {
		user, err = s.authRepo.FindByUsername(req.Identifier)
	}

	// 3. Error checks
	if err != nil {

		if errors.Is(err, repositories.ErrUserNotFound) {
			return dto.SignInRes{}, ErrInvalidCredentials
		}

		return dto.SignInRes{}, err
	}

	if user.ID == 0 {
		return dto.SignInRes{}, ErrInvalidCredentials
	}

	if user.Password == nil {
		return dto.SignInRes{}, ErrUserGoogleSignIn
	}

	// 4. Make sure password match
	if passwordErr := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password)); passwordErr != nil {
		return dto.SignInRes{}, ErrInvalidCredentials
	}

	// 5. Create refresh token
	refreshTokenString, refreshErr := s.CreateToken(24*30, user.ID, true)

	if refreshErr != nil {
		return dto.SignInRes{}, ErrFailedCreateToken
	}

	// 6. create refresh token in table
	refreshTokenModel := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
	}

	if _, createErr := s.refreshRepo.Create(&refreshTokenModel); createErr != nil {
		return dto.SignInRes{}, createErr
	}

	// 7. create the access token
	accessTokenString, accessErr := s.CreateToken(12, user.ID, false)

	if accessErr != nil {
		return dto.SignInRes{}, ErrFailedCreateToken
	}

	// 8. send token and success
	return dto.SignInRes{RefreshToken: refreshTokenString, AccessToken: accessTokenString}, nil
}

// REFRESH TOKEN
func (s *AuthService) Refresh(req dto.RefreshTokenReq) (dto.RefreshTokenRes, error) {

	// 1. find the refresh token and check if any
	result, err := s.refreshRepo.FindByToken(req.RefreshToken)

	// 2. error checks
	if err != nil {
		return dto.RefreshTokenRes{}, err
	}

	if result.ExpiresAt.Before(time.Now()) {
		return dto.RefreshTokenRes{}, errors.New("refresh token is expired")
	}

	// 3. create a new refresh token
	refreshTokenString, refreshErr := s.CreateToken(24*30, result.UserID, true)

	if refreshErr != nil {
		return dto.RefreshTokenRes{}, ErrFailedCreateToken
	}

	// 4. update the NEW refresh token in table
	refreshTokenModel := models.RefreshToken{
		UserID:    result.UserID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
	}

	refreshTokenModel.ID = result.ID

	if _, updateErr := s.refreshRepo.Update(&refreshTokenModel); updateErr != nil {
		return dto.RefreshTokenRes{}, updateErr
	}

	// 5. create the access token
	accessTokenString, accessErr := s.CreateToken(12, result.UserID, false)

	if accessErr != nil {
		return dto.RefreshTokenRes{}, ErrFailedCreateToken
	}

	return dto.RefreshTokenRes{AccessToken: accessTokenString, RefreshToken: refreshTokenString}, nil
}

// DELETE
func (s *AuthService) DeleteToken(req dto.RefreshTokenReq) error {

	// 1. Delete it
	if err := s.refreshRepo.DeleteByToken(req.RefreshToken); err != nil {
		return err
	}

	return nil
}

// HELPER FUNCTION
func (s *AuthService) CreateToken(hour int, userID uint, isRefresh bool) (string, error) {

	// 1. convert the hour to time.duration
	duration := time.Duration(hour) * time.Hour

	// 2. Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(duration).Unix(),
	})

	// 3. secret adjustments
	var secret = []byte(s.cfg.JWTSecret)

	if isRefresh {
		secret = []byte(s.cfg.RefreshSecret)
	}

	// 3. signed the token
	tokenString, err := token.SignedString(secret)

	// 4. error checks
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// verify otp
func (s *AuthService) VerifyOTP(req dto.VerifyOTPreq) (bool, error) {
	// 1. find data based on email

	// 2. check the attempts 
	// 
	// 3. read and compare the otp

	// 4. update the verify otp table,

	// 4. 

}