package services

import (
	"strings"

	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/dto"
	"github.com/Haidarr-h/backend-go/models"
	"github.com/Haidarr-h/backend-go/pkg/jwt"
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

func (s *OAuthService) GoogleSignIn(req dto.GoogleSignInReq) (dto.GoogleSignInRes, error) {

	// 1. verify the token with google api
	googleUserInfo, verifyErr := oauth.VerifyGoogleToken(req.IDToken, s.cfg)

	if verifyErr != nil {
		return dto.GoogleSignInRes{}, ErrInvalidGoogleIDToken
	}

	// 2. find user with google id
	isUserExist, dbErr := s.userRepo.ExistByGoogleID(googleUserInfo.Sub)

	if dbErr != nil {
		return dto.GoogleSignInRes{}, dbErr
	}

	var userData models.User

	// 3. IF User is not exist (we have to create it first)
	if !isUserExist {

		// 3.1 IF email is already occupied
		isEmailExist, emailErr := s.userRepo.ExistByEmail(googleUserInfo.Email)
		if emailErr != nil {
			return dto.GoogleSignInRes{}, emailErr
		}

		// 3.2 user already created (update the user data)
		if isEmailExist {
			userResult, userErr := s.userRepo.FindByEmail(googleUserInfo.Email)

			if userErr != nil {
				return dto.GoogleSignInRes{}, userErr
			}

			userData.GoogleID = &googleUserInfo.Sub
			userData.Picture = &googleUserInfo.Picture
			userData.ID = userResult.ID

			// update
			_, updateErr := s.userRepo.UpdateUser(userData)
			if updateErr != nil {
				return dto.GoogleSignInRes{}, updateErr
			}

		} else {
			// 3.3 USER IS NOT CREATED == CREATE NEW

			baseUsername := strings.Split(googleUserInfo.Email, "@")[0]

			userData.FirstName = googleUserInfo.GivenName
			userData.Username = baseUsername
			userData.Email = googleUserInfo.Email
			userData.GoogleID = &googleUserInfo.Sub
			userData.Picture = &googleUserInfo.Picture

			if googleUserInfo.FamilyName == "" {
				userData.LastName = googleUserInfo.GivenName
			} else {
				userData.LastName = googleUserInfo.FamilyName
			}

			// create
			userDataCreated, createErr := s.userRepo.CreateUser(userData)
			userData.ID = userDataCreated.ID

			if createErr != nil {
				return dto.GoogleSignInRes{}, createErr
			}
		}
	}

	// 4. CREATE THE TOKEN
	tokenResult, err := jwt.CreateAccessRefreshToken(userData.ID, s.cfg)

	if err != nil {
		return dto.GoogleSignInRes{}, ErrFailedCreateToken
	}

	return dto.GoogleSignInRes{AccessToken: tokenResult.AccessToken, RefreshToken: tokenResult.RefreshToken, ID: userData.ID, FirstName: userData.FirstName, LastName: userData.LastName, Username: userData.Username}, nil
}
