package auth

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/internal/user"
	"github.com/Haidarr-h/backend-go/models"
	jwtlocal "github.com/Haidarr-h/backend-go/pkg/jwtLocal"
	"github.com/Haidarr-h/backend-go/pkg/logger"
	oauth "github.com/Haidarr-h/backend-go/pkg/oAuth"
	"github.com/Haidarr-h/backend-go/pkg/otp"
	"github.com/Haidarr-h/backend-go/pkg/utils"
	"github.com/Haidarr-h/backend-go/repositories"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    *user.UserRepository
	refreshRepo *RefreshTokenRepository
	otpRepo     *OtpRepository
	cfg         *config.Config
}

func NewAuthService(authRepo *user.UserRepository, cfg *config.Config, refreshRepo *RefreshTokenRepository, otpRepo *OtpRepository) *AuthService {
	return &AuthService{userRepo: authRepo, cfg: cfg, refreshRepo: refreshRepo, otpRepo: otpRepo}
}

// SIGN UP
func (s *AuthService) SignUp(req SignUpRequest) (SignUpResponse, error) {

	// 1. Check if user email already exist
	isEmailExist, emailErr := s.userRepo.ExistByEmail(req.Email)
	if emailErr != nil {
		return SignUpResponse{}, emailErr
	}

	if isEmailExist {
		return SignUpResponse{}, ErrEmailIsExists
	}

	// 2. Check if username already exist
	isUsernameExist, usernameErr := s.userRepo.ExistByUsername(req.Username)
	if usernameErr != nil {
		return SignUpResponse{}, usernameErr
	}

	if isUsernameExist {
		return SignUpResponse{}, ErrUsernameIsExists
	}

	// 3. Hash the passowrd
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return SignUpResponse{}, err
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
	result, err := s.userRepo.CreateUser(user)
	if err != nil {
		return SignUpResponse{}, err
	}

	// 6. generate OTP
	plainOTP, hashedOTP, otpErr := otp.GenerateOtp()

	if otpErr != nil {
		return SignUpResponse{}, otpErr
	}

	// 7. create data at OTP table
	otpData := models.OtpVerification{
		UserID:    result.ID,
		OTPHash:   hashedOTP,
		ExpiresAt: time.Now().Add(time.Minute * 60),
		Attempts:  0,
		Used:      false,
	}

	if _, createOtpErr := s.otpRepo.Create(otpData); createOtpErr != nil {
		return SignUpResponse{}, createOtpErr
	}

	// 8. send OTP to the email
	if err := otp.SendOTP(user.Email, plainOTP, s.cfg); err != nil {
		return SignUpResponse{}, err
	}

	// 9. success response
	response := SignUpResponse{
		ID:         result.ID,
		FirstName:  result.FirstName,
		LastName:   result.LastName,
		Username:   result.Username,
		IsVerified: result.IsVerified,
	}

	return response, nil
}

// SIGN IN
func (s *AuthService) SignIn(req SignInReq) (SignInRes, error) {

	// 1. check sign in by email or username
	isEmail := utils.IsEmail(req.Identifier)

	var user models.User
	var err error

	// 2. check if user exist
	if isEmail {
		user, err = s.userRepo.FindByEmail(req.Identifier)
	} else {
		user, err = s.userRepo.FindByUsername(req.Identifier)
	}

	// 3. Error checks
	if err != nil {

		if errors.Is(err, repositories.ErrUserNotFound) {
			return SignInRes{}, ErrInvalidCredentials
		}

		return SignInRes{}, err
	}

	if user.ID == 0 {
		return SignInRes{}, ErrInvalidCredentials
	}

	if user.Password == nil {
		return SignInRes{}, ErrUserGoogleSignIn
	}

	// 4. Make sure password match
	if passwordErr := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password)); passwordErr != nil {
		return SignInRes{}, ErrInvalidCredentials
	}

	// 5. Create refresh token
	refreshTokenString, refreshErr := s.CreateToken(24*30, user.ID, true)

	if refreshErr != nil {
		return SignInRes{}, ErrFailedCreateToken
	}

	// 6. create refresh token in table
	refreshTokenModel := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
	}

	if _, createErr := s.refreshRepo.Create(&refreshTokenModel); createErr != nil {
		return SignInRes{}, createErr
	}

	// 7. create the access token
	accessTokenString, accessErr := s.CreateToken(12, user.ID, false)

	if accessErr != nil {
		return SignInRes{}, ErrFailedCreateToken
	}

	// 8. send token and success
	return SignInRes{RefreshToken: refreshTokenString, AccessToken: accessTokenString}, nil
}

// REFRESH TOKEN
func (s *AuthService) Refresh(req RefreshTokenReq) (RefreshTokenRes, error) {

	// 1. find the refresh token and check if any
	result, err := s.refreshRepo.FindByToken(req.RefreshToken)

	// 2. error checks
	if err != nil {
		return RefreshTokenRes{}, err
	}

	if result.ExpiresAt.Before(time.Now()) {
		return RefreshTokenRes{}, errors.New("refresh token is expired")
	}

	// 3. create a new refresh token
	refreshTokenString, refreshErr := s.CreateToken(24*30, result.UserID, true)

	if refreshErr != nil {
		return RefreshTokenRes{}, ErrFailedCreateToken
	}

	// 4. update the NEW refresh token in table
	refreshTokenModel := models.RefreshToken{
		UserID:    result.UserID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
	}

	refreshTokenModel.ID = result.ID

	if _, updateErr := s.refreshRepo.Update(&refreshTokenModel); updateErr != nil {
		return RefreshTokenRes{}, updateErr
	}

	// 5. create the access token
	accessTokenString, accessErr := s.CreateToken(12, result.UserID, false)

	if accessErr != nil {
		return RefreshTokenRes{}, ErrFailedCreateToken
	}

	return RefreshTokenRes{AccessToken: accessTokenString, RefreshToken: refreshTokenString}, nil
}

// DELETE
func (s *AuthService) DeleteToken(req RefreshTokenReq) error {

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
func (s *AuthService) VerifyOTP(req VerifyOTPreq) (bool, error) {

	// 1. find data based on email
	otpData, otpDataErr := s.otpRepo.FindByEmail(req.Email)

	if otpDataErr != nil {
		return false, otpDataErr
	}

	// 2. check the attempts, used, and expiry
	if otpData.Attempts >= 5 {
		return false, ErrInvalidOTPAttempts
	}

	if otpData.ExpiresAt.Before(time.Now()) {
		return false, ErrOTPExpired
	}

	if otpData.Used {
		return false, ErrInvalidOTPUsed
	}

	// 3. read and compare the otp
	hashOTPreq := fmt.Sprintf("%x", sha256.Sum256([]byte(req.OtpCode)))

	if hashOTPreq != otpData.OTPHash {
		logger.Log.Error("failed to compare otp code")

		otpData.Attempts += 1
		if _, updateErr := s.otpRepo.Update(otpData); updateErr != nil {
			return false, updateErr
		}

		return false, ErrInvalidOTP
	}

	// 4. update the verify otp table,
	otpData.Attempts += 1
	otpData.Used = true
	if _, updateErr := s.otpRepo.Update(otpData); updateErr != nil {
		return false, updateErr
	}

	// 5. update the user is verified
	if verifyErr := s.userRepo.Verify(otpData.UserID); verifyErr != nil {
		return false, verifyErr
	}

	// 5. response
	return true, nil
}

func (s *AuthService) ResendOTP(req ResendOTPreq) error {
	// 1. find otp data by email
	otpData, otpDataErr := s.otpRepo.FindByEmail(req.Email)
	if otpDataErr != nil {
		return otpDataErr
	}

	// 2. generate new code, generate new otp, set expiry, used, and attempt
	plainOTP, hashedOTP, otpErr := otp.GenerateOtp()
	if otpErr != nil {
		return otpErr
	}

	// 3. create new otp verification data in table
	otpData.OTPHash = hashedOTP
	otpData.ExpiresAt = time.Now().Add(time.Minute * 60)
	otpData.Attempts = 0
	otpData.Used = false
	otpData.UpdatedAt = time.Now()

	_, createOTPErr := s.otpRepo.Create(otpData)
	if createOTPErr != nil {
		return createOTPErr
	}

	// 4. send the new otp to user gmail
	if err := otp.SendOTP(req.Email, plainOTP, s.cfg); err != nil {
		return err
	}

	// 3. return nil if success
	return nil
}

// CreateAccessRefreshToken generates access and refresh tokens
// used for: sign in manual, sign in by google, and refresh
func (s *AuthService) CreateAccessRefreshToken(userID uint, cfg *config.Config) (Tokens, error) {

	// 1. Generate the refresh token
	refreshTokenString, refreshErr := jwtlocal.CreateToken(24*30, userID, true, cfg)

	if refreshErr != nil {
		return Tokens{}, refreshErr
	}

	// 2. create new row in the refresh token table
	refreshTokenModel := models.RefreshToken{
		UserID:    userID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
	}

	if _, createErr := s.refreshRepo.Create(&refreshTokenModel); createErr != nil {
		return Tokens{}, createErr
	}

	// 3. create the access token
	accessTokenString, accessErr := jwtlocal.CreateToken(12, userID, false, cfg)

	if accessErr != nil {
		return Tokens{}, accessErr
	}

	return Tokens{AccessToken: accessTokenString, RefreshToken: refreshTokenString}, nil
}

// Oauth
func (s *AuthService) GoogleSignIn(req GoogleSignInReq) (GoogleSignInRes, error) {

	// 1. verify the token with google api
	googleUserInfo, verifyErr := oauth.VerifyGoogleToken(req.IDToken, s.cfg)

	if verifyErr != nil {
		return GoogleSignInRes{}, ErrInvalidGoogleIDToken
	}

	// 2. find userData with google id
	userData, err := s.userRepo.FindByGoogleID(googleUserInfo.Sub)

	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return GoogleSignInRes{}, err
	}

	// 3. no user found by google id -> create a new one or updates
	if errors.Is(err, repositories.ErrUserNotFound) {

		// 3.1 check user by email
		userData, err = s.userRepo.FindByEmail(googleUserInfo.Email)

		if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
			return GoogleSignInRes{}, err
		}

		// 3.2 no user found by email == CREATE A NEW ONE
		if errors.Is(err, repositories.ErrUserNotFound) {
			
			// user models build
			baseUsername := strings.Split(googleUserInfo.Email, "@")[0]

			userData.FirstName = googleUserInfo.GivenName
			userData.Username = baseUsername
			userData.Email = googleUserInfo.Email
			userData.GoogleID = &googleUserInfo.Sub
			userData.Picture = &googleUserInfo.Picture
			userData.IsVerified = true

			if googleUserInfo.FamilyName == "" {
				userData.LastName = googleUserInfo.GivenName
			} else {
				userData.LastName = googleUserInfo.FamilyName
			}

			// create
			if userData, err = s.userRepo.CreateUser(userData); err != nil {
				return GoogleSignInRes{}, err
			}


		} else {
			// 3.3 found user by email. update the google info columns
			userData.GoogleID = &googleUserInfo.Sub
			userData.Picture = &googleUserInfo.Picture
			userData.IsVerified = true

			if userData, err = s.userRepo.UpdateUser(userData); err != nil {
				return GoogleSignInRes{}, err
			}
		}

	}

	// 4. CREATE THE TOKEN
	logger.Log.Debug("creating access and refresh token", "user data", userData)
	tokenResult, err := s.CreateAccessRefreshToken(userData.ID, s.cfg)

	if err != nil {
		return GoogleSignInRes{}, err
	}

	return GoogleSignInRes{AccessToken: tokenResult.AccessToken, RefreshToken: tokenResult.RefreshToken, ID: userData.ID, FirstName: userData.FirstName, LastName: userData.LastName, Username: userData.Username}, nil
}