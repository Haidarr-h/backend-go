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
	"github.com/Haidarr-h/backend-go/pkg/cache"
	"github.com/Haidarr-h/backend-go/pkg/logger"
	oauth "github.com/Haidarr-h/backend-go/pkg/oAuth"
	"github.com/Haidarr-h/backend-go/pkg/otp"
	"github.com/Haidarr-h/backend-go/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo          user.Repository
	refreshRepo       RefreshRepo
	registrationCache *cache.RegistrationCache
	cfg               *config.Config
}

func NewAuthService(userRepo user.Repository, cfg *config.Config, refreshRepo RefreshRepo, registrationCache *cache.RegistrationCache) *AuthService {
	return &AuthService{userRepo: userRepo, cfg: cfg, refreshRepo: refreshRepo, registrationCache: registrationCache}
}

// SIGN UP
func (s *AuthService) SignUp(req SignUpRequest) (SignUpResponse, error) {

	// 1. Check if user email already exist in DB
	isEmailExist, emailErr := s.userRepo.ExistByEmail(req.Email)
	if emailErr != nil {
		return SignUpResponse{}, emailErr
	}

	if isEmailExist {
		return SignUpResponse{}, ErrEmailIsExists
	}

	// 2. Check if username already exist in DB
	isUsernameExist, usernameErr := s.userRepo.ExistByUsername(req.Username)
	if usernameErr != nil {
		return SignUpResponse{}, usernameErr
	}

	if isUsernameExist {
		return SignUpResponse{}, ErrUsernameIsExists
	}

	// 3. Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return SignUpResponse{}, err
	}

	// 4. Generate OTP
	plainOTP, hashedOTP, otpErr := otp.GenerateOtp()
	if otpErr != nil {
		return SignUpResponse{}, otpErr
	}

	// 5. Store pending registration in cache (overwrites any previous attempt for this email)
	s.registrationCache.Set(req.Email, cache.PendingRegistration{
		Email:          req.Email,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Username:       req.Username,
		HashedPassword: string(hash),
		OTPHash:        hashedOTP,
		ExpiresAt:      time.Now().Add(time.Minute * 60),
		Attempts:       0,
	})

	// 6. Send OTP to the email
	if err := otp.SendOTP(req.Email, plainOTP, s.cfg); err != nil {
		s.registrationCache.Delete(req.Email)
		return SignUpResponse{}, err
	}

	// 7. Return response — no DB row yet, so ID is 0
	return SignUpResponse{
		ID:         0,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Username:   req.Username,
		IsVerified: false,
	}, nil
}

// SIGN IN
func (s *AuthService) SignIn(req SignInReq) (SignInRes, error) {

	// 1. check sign in by email or username
	isEmail := utils.IsEmail(req.Identifier)

	var foundUser models.User
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

	// 6. send token and success
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

// VERIFY OTP
func (s *AuthService) VerifyOTP(req VerifyOTPreq) (bool, error) {

	// 1. Find pending registration in cache
	pending, found := s.registrationCache.Get(req.Email)
	if !found {
		return false, ErrPendingRegistrationNotFound
	}

	// 2. Check attempts and expiry
	if pending.Attempts >= 5 {
		return false, ErrInvalidOTPAttempts
	}

	if pending.ExpiresAt.Before(time.Now()) {
		s.registrationCache.Delete(req.Email)
		return false, ErrOTPExpired
	}

	// 3. Compare the OTP hash
	hashOTPreq := fmt.Sprintf("%x", sha256.Sum256([]byte(req.OtpCode)))

	if hashOTPreq != pending.OTPHash {
		logger.Log.Error("failed to compare otp code")
		pending.Attempts++
		s.registrationCache.Update(req.Email, pending)
		return false, ErrInvalidOTP
	}

	// 4. OTP valid — insert user to DB with IsVerified = true
	newUser := models.User{
		Email:      pending.Email,
		FirstName:  pending.FirstName,
		LastName:   pending.LastName,
		Username:   pending.Username,
		Password:   &pending.HashedPassword,
		IsVerified: true,
	}

	if _, err := s.userRepo.CreateUser(newUser); err != nil {
		return false, err
	}

	// 5. Remove from cache
	s.registrationCache.Delete(req.Email)

	return true, nil
}

// RESEND OTP
func (s *AuthService) ResendOTP(req ResendOTPreq) error {
	// 1. Find pending registration in cache
	pending, found := s.registrationCache.Get(req.Email)
	if !found {
		return ErrPendingRegistrationNotFound
	}

	// 2. Generate new OTP
	plainOTP, hashedOTP, otpErr := otp.GenerateOtp()
	if otpErr != nil {
		return otpErr
	}

	// 3. Update cache entry with new OTP, reset attempts and expiry
	pending.OTPHash = hashedOTP
	pending.ExpiresAt = time.Now().Add(time.Minute * 60)
	pending.Attempts = 0
	s.registrationCache.Update(req.Email, pending)

	// 4. Send new OTP to email
	if err := otp.SendOTP(req.Email, plainOTP, s.cfg); err != nil {
		return err
	}

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
	refreshTokenModel := models.RefreshToken{
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
