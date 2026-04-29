package services

import (
	"errors"
	"strings"

	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/dto"
	"github.com/Haidarr-h/backend-go/pkg/jwt"
	"github.com/Haidarr-h/backend-go/pkg/logger"
	oauth "github.com/Haidarr-h/backend-go/pkg/oAuth"
	"github.com/Haidarr-h/backend-go/repositories"
)

type OAuthService struct {
	userRepo *repositories.UserRepository
	cfg      *config.Config
}

func NewOAuthService(userRepo *repositories.UserRepository, cfg *config.Config) *OAuthService {
	return &OAuthService{userRepo: userRepo, cfg: cfg}
}

// GoogleSignIn sign ins by oauth google
func (s *OAuthService) GoogleSignIn(req dto.GoogleSignInReq) (dto.GoogleSignInRes, error) {

	// 1. verify the token with google api
	googleUserInfo, verifyErr := oauth.VerifyGoogleToken(req.IDToken, s.cfg)

	if verifyErr != nil {
		return dto.GoogleSignInRes{}, ErrInvalidGoogleIDToken
	}

	// 2. find userData with google id
	userData, err := s.userRepo.FindByGoogleID(googleUserInfo.Sub)

	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return dto.GoogleSignInRes{}, err
	}

	// 3. no user found by google id -> create a new one or updates
	if errors.Is(err, repositories.ErrUserNotFound) {

		// 3.1 check user by email
		userData, err = s.userRepo.FindByEmail(googleUserInfo.Email)

		if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
			return dto.GoogleSignInRes{}, err
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
				return dto.GoogleSignInRes{}, err
			}


		} else {
			// 3.3 found user by email. update the google info columns
			userData.GoogleID = &googleUserInfo.Sub
			userData.Picture = &googleUserInfo.Picture
			userData.IsVerified = true

			if userData, err = s.userRepo.UpdateUser(userData); err != nil {
				return dto.GoogleSignInRes{}, err
			}
		}

	}

	// 4. CREATE THE TOKEN
	logger.Log.Debug("creating access and refresh token", "user data", userData)
	tokenResult, err := jwt.CreateAccessRefreshToken(userData.ID, s.cfg)

	if err != nil {
		return dto.GoogleSignInRes{}, err
	}

	return dto.GoogleSignInRes{AccessToken: tokenResult.AccessToken, RefreshToken: tokenResult.RefreshToken, ID: userData.ID, FirstName: userData.FirstName, LastName: userData.LastName, Username: userData.Username}, nil
}
