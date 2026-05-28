package auth

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/Haidarr-h/backend-go/internal/user"
	jwtlocal "github.com/Haidarr-h/backend-go/pkg/jwtLocal"
	"github.com/Haidarr-h/backend-go/pkg/logger"
	oauth "github.com/Haidarr-h/backend-go/pkg/oAuth"
	"github.com/Haidarr-h/backend-go/pkg/otp"
	"github.com/Haidarr-h/backend-go/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo    user.Repository
	refreshRepo RefreshRepo
	otpRepo     OtpRepo
	cfg         *config.Config
}

func NewAuthService(userRepo user.Repository, cfg *config.Config, refreshRepo RefreshRepo, otpRepo OtpRepo) *AuthService {
	return &AuthService{userRepo: userRepo, cfg: cfg, refreshRepo: refreshRepo, otpRepo: otpRepo}
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
	newUser := user.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		Password:  &hashedPassword,
	}

	// 5. generate OTP
	plainOTP, hashedOTP, otpErr := otp.GenerateOtp()

	if otpErr != nil {
		return SignUpResponse{}, otpErr
	}

	// 6. Create user and otp table
	var result user.User

	txErr := s.cfg.DB.Transaction(func(tx *gorm.DB) error {
		var err error

		result, err = user.NewUserRepository(tx).CreateUser(newUser)

		if err != nil {
			return err
		}

		otpData := OtpVerification{
			UserID:    result.ID,
			OTPHash:   hashedOTP,
			ExpiresAt: time.Now().Add(time.Minute * 60),
			Attempts:  0,
			Used:      false,
		}

		if _, err = NewOTPRepository(tx).Create(otpData); err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		return SignUpResponse{}, txErr
	}

	// 8. send OTP to the email
	if err := otp.SendOTP(newUser.Email, plainOTP, s.cfg); err != nil {
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

	var foundUser user.User
	var err error

	// 2. check if user exist
	if isEmail {
		foundUser, err = s.userRepo.FindByEmail(req.Identifier)
	} else {
		foundUser, err = s.userRepo.FindByUsername(req.Identifier)
	}

	// 3. Error checks
	if err != nil {

		if errors.Is(err, user.ErrUserNotFound) {
			return SignInRes{}, ErrInvalidCredentials
		}

		return SignInRes{}, err
	}

	if foundUser.ID == 0 {
		return SignInRes{}, ErrInvalidCredentials
	}

	if foundUser.Password == nil {
		return SignInRes{}, ErrUserGoogleSignIn
	}

	// 4. Make sure password match
	if passwordErr := bcrypt.CompareHashAndPassword([]byte(*foundUser.Password), []byte(req.Password)); passwordErr != nil {
		return SignInRes{}, ErrInvalidCredentials
	}

	// 5. Create refresh token
	tokenResult, errTokenCreate := s.createAccessRefreshToken(foundUser.ID)

	if errTokenCreate != nil {
		return SignInRes{}, errTokenCreate
	}

	// 8. send token and success
	return SignInRes{RefreshToken: tokenResult.RefreshToken, AccessToken: tokenResult.AccessToken}, nil
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
		return RefreshTokenRes{}, ErrExpiredToken
	}

	// 3. create a new refresh token
	tokenResult, errTokenCreate := s.createAccessRefreshToken(result.UserID)

	if errTokenCreate != nil {
		return RefreshTokenRes{}, errTokenCreate
	}

	return RefreshTokenRes{AccessToken: tokenResult.AccessToken, RefreshToken: tokenResult.RefreshToken}, nil
}

// DELETE
func (s *AuthService) DeleteToken(req RefreshTokenReq) error {

	// 1. Delete it
	if err := s.refreshRepo.DeleteByToken(req.RefreshToken); err != nil {
		return err
	}

	return nil
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

// createAccessRefreshToken generates access and refresh tokens
// used for: sign in manual, sign in by google, and refresh
func (s *AuthService) createAccessRefreshToken(userID uint) (Tokens, error) {

	// 1. Generate the refresh token
	refreshTokenString, refreshErr := jwtlocal.CreateToken(24*30, userID, true, s.cfg)

	if refreshErr != nil {
		return Tokens{}, refreshErr
	}

	// 2. create new row in the refresh token table
	refreshTokenModel := RefreshToken{
		UserID:    userID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
	}

	if _, createErr := s.refreshRepo.Create(&refreshTokenModel); createErr != nil {
		return Tokens{}, createErr
	}

	// 3. create the access token
	accessTokenString, accessErr := jwtlocal.CreateToken(12, userID, false, s.cfg)

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

	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return GoogleSignInRes{}, err
	}

	// 3. no user found by google id -> create a new one or updates
	if errors.Is(err, user.ErrUserNotFound) {

		// 3.1 check user by email
		userData, err = s.userRepo.FindByEmail(googleUserInfo.Email)

		if err != nil && !errors.Is(err, user.ErrUserNotFound) {
			return GoogleSignInRes{}, err
		}

		// 3.2 no user found by email == CREATE A NEW ONE
		if errors.Is(err, user.ErrUserNotFound) {

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
	tokenResult, err := s.createAccessRefreshToken(userData.ID)

	if err != nil {
		return GoogleSignInRes{}, err
	}

	return GoogleSignInRes{AccessToken: tokenResult.AccessToken, RefreshToken: tokenResult.RefreshToken, ID: userData.ID, FirstName: userData.FirstName, LastName: userData.LastName, Username: userData.Username}, nil
}
