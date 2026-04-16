package services

import (
	"errors"
	"os"
	"time"

	"github.com/Haidarr-h/backend-go/dto"
	"github.com/Haidarr-h/backend-go/models"
	"github.com/Haidarr-h/backend-go/repositories"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	authRepo *repositories.UserRepository
}

func NewAuthService(authRepo *repositories.UserRepository) *AuthService {
	return &AuthService{authRepo: authRepo}
}

// CREATE
func (s *AuthService) SignUp(req dto.SignUpRequest) (dto.SignUpResponse, error) {

	// 1. Check if user email already exist
	_, emailErr := s.authRepo.FindByEmail(req.Email)
	if emailErr == nil {
		return dto.SignUpResponse{}, errors.New("email already exist")
	}

	// 2. Check if username already exist
	_, usernameErr := s.authRepo.FindByUsername(req.Username)
	if usernameErr == nil {
		return dto.SignUpResponse{}, errors.New("username already exist")
	}

	// 3. Hash the passowrd 
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return dto.SignUpResponse{}, err
	}

	// 4. Create the user model
	hashedPassword := string(hash)
	user := models.User{
		Email: req.Email,
		FullName: req.FullName,
		Username: req.Username,
		Password: &hashedPassword,
	}

	// 5. create the user
	user_result, err := s.authRepo.CreateUser(user)
	if err != nil {
		return dto.SignUpResponse{}, err
	}

	response := dto.SignUpResponse{
		ID: user_result.ID,
		FullName: user_result.FullName,
		Username: user_result.Username,
	}

	return response, nil
}

func (s *AuthService) SignIn(req dto.SignInReq) (dto.SignInRes, error) {

	// 1. check if user exist
	user, err := s.authRepo.FindByEmail(req.Email)
	
	if err != nil {
		return dto.SignInRes{}, err
	}

	if user.ID == 0 {
		return dto.SignInRes{}, errors.New("Invalid email or password")
	}

	if user.Password == nil {
		return dto.SignInRes{}, errors.New("This user uses google account sign in")
	}

	// 2. Make sure password match
	passwordErr := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password))

	if passwordErr != nil {
		return dto.SignInRes{}, errors.New("Invalid email or password")
	}

	// 3. Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		return dto.SignInRes{}, errors.New("Failed to create token")
	}

	// 4. send token and success
	return dto.SignInRes{Token: tokenString}, nil
}