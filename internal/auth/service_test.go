package auth

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/internal/user"
	"github.com/Haidarr-h/backend-go/models"
	"github.com/Haidarr-h/backend-go/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// computeOTPHash mirrors the hashing logic used in VerifyOTP.
func computeOTPHash(code string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(code)))
}

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) CreateUser(u models.User) (models.User, error) {
	args := m.Called(u)
	return args.Get(0).(models.User), args.Error(1)
}
func (m *mockUserRepo) UpdateUser(u models.User) (models.User, error) {
	args := m.Called(u)
	return args.Get(0).(models.User), args.Error(1)
}
func (m *mockUserRepo) FindByEmail(email string) (models.User, error) {
	args := m.Called(email)
	return args.Get(0).(models.User), args.Error(1)
}
func (m *mockUserRepo) FindByUsername(username string) (models.User, error) {
	args := m.Called(username)
	return args.Get(0).(models.User), args.Error(1)
}
func (m *mockUserRepo) FindByGoogleID(googleID string) (models.User, error) {
	args := m.Called(googleID)
	return args.Get(0).(models.User), args.Error(1)
}
func (m *mockUserRepo) ExistByEmail(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}
func (m *mockUserRepo) ExistByUsername(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}
func (m *mockUserRepo) Verify(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

type mockRefreshRepo struct{ mock.Mock }

func (m *mockRefreshRepo) Create(token *models.RefreshToken) (*models.RefreshToken, error) {
	args := m.Called(token)
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}
func (m *mockRefreshRepo) FindByToken(token string) (*models.RefreshToken, error) {
	args := m.Called(token)
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}
func (m *mockRefreshRepo) Update(token *models.RefreshToken) (*models.RefreshToken, error) {
	args := m.Called(token)
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}
func (m *mockRefreshRepo) DeleteByToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestService(ur *mockUserRepo, rr *mockRefreshRepo, rc *cache.RegistrationCache) *AuthService {
	if rc == nil {
		rc = cache.NewRegistrationCache()
	}
	return NewAuthService(ur, &config.Config{
		JWTSecret:     "test-jwt-secret-32-bytes-padding!",
		RefreshSecret: "test-refresh-secret-32-bytes-pad!",
	}, rr, rc)
}

func hashedTestPassword(plain string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	return string(h)
}

// ---------------------------------------------------------------------------
// SignUp
// ---------------------------------------------------------------------------

func TestSignUp(t *testing.T) {
	t.Run("email already exists", func(t *testing.T) {
		ur := &mockUserRepo{}
		ur.On("ExistByEmail", "test@example.com").Return(true, nil)

		svc := newTestService(ur, &mockRefreshRepo{}, nil)
		_, err := svc.SignUp(SignUpRequest{
			Email: "test@example.com", Password: "password123",
			Username: "testuser", FirstName: "Test", LastName: "User",
		})

		assert.ErrorIs(t, err, ErrEmailIsExists)
		ur.AssertExpectations(t)
	})

	t.Run("username already exists", func(t *testing.T) {
		ur := &mockUserRepo{}
		ur.On("ExistByEmail", "test@example.com").Return(false, nil)
		ur.On("ExistByUsername", "testuser").Return(true, nil)

		svc := newTestService(ur, &mockRefreshRepo{}, nil)
		_, err := svc.SignUp(SignUpRequest{
			Email: "test@example.com", Password: "password123",
			Username: "testuser", FirstName: "Test", LastName: "User",
		})

		assert.ErrorIs(t, err, ErrUsernameIsExists)
		ur.AssertExpectations(t)
	})

	t.Run("repo error on ExistByEmail", func(t *testing.T) {
		ur := &mockUserRepo{}
		dbErr := errors.New("db error")
		ur.On("ExistByEmail", "test@example.com").Return(false, dbErr)

		svc := newTestService(ur, &mockRefreshRepo{}, nil)
		_, err := svc.SignUp(SignUpRequest{
			Email: "test@example.com", Password: "password123",
			Username: "testuser", FirstName: "Test", LastName: "User",
		})

		assert.ErrorIs(t, err, dbErr)
		ur.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// SignIn
// ---------------------------------------------------------------------------

func TestSignIn(t *testing.T) {
	t.Run("user not found by email", func(t *testing.T) {
		ur := &mockUserRepo{}
		ur.On("FindByEmail", "test@example.com").Return(models.User{}, user.ErrUserNotFound)

		svc := newTestService(ur, &mockRefreshRepo{}, nil)
		_, err := svc.SignIn(SignInReq{Identifier: "test@example.com", Password: "password123"})

		assert.ErrorIs(t, err, ErrInvalidCredentials)
		ur.AssertExpectations(t)
	})

	t.Run("user not found by username", func(t *testing.T) {
		ur := &mockUserRepo{}
		ur.On("FindByUsername", "testuser").Return(models.User{}, user.ErrUserNotFound)

		svc := newTestService(ur, &mockRefreshRepo{}, nil)
		_, err := svc.SignIn(SignInReq{Identifier: "testuser", Password: "password123"})

		assert.ErrorIs(t, err, ErrInvalidCredentials)
		ur.AssertExpectations(t)
	})

	t.Run("user is google-only account", func(t *testing.T) {
		ur := &mockUserRepo{}
		ur.On("FindByEmail", "google@example.com").Return(models.User{Model: gorm.Model{ID: 1}, Password: nil}, nil)

		svc := newTestService(ur, &mockRefreshRepo{}, nil)
		_, err := svc.SignIn(SignInReq{Identifier: "google@example.com", Password: "password123"})

		assert.ErrorIs(t, err, ErrUserGoogleSignIn)
		ur.AssertExpectations(t)
	})

	t.Run("wrong password", func(t *testing.T) {
		pass := hashedTestPassword("correctpassword")
		ur := &mockUserRepo{}
		ur.On("FindByEmail", "test@example.com").Return(models.User{Model: gorm.Model{ID: 1}, Password: &pass}, nil)

		svc := newTestService(ur, &mockRefreshRepo{}, nil)
		_, err := svc.SignIn(SignInReq{Identifier: "test@example.com", Password: "wrongpassword"})

		assert.ErrorIs(t, err, ErrInvalidCredentials)
		ur.AssertExpectations(t)
	})

	t.Run("success with email", func(t *testing.T) {
		pass := hashedTestPassword("password123")
		ur := &mockUserRepo{}
		rr := &mockRefreshRepo{}
		ur.On("FindByEmail", "test@example.com").Return(models.User{Model: gorm.Model{ID: 1}, Password: &pass}, nil)
		rr.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(&models.RefreshToken{Model: gorm.Model{ID: 1}}, nil)

		svc := newTestService(ur, rr, nil)
		res, err := svc.SignIn(SignInReq{Identifier: "test@example.com", Password: "password123"})

		assert.NoError(t, err)
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
		ur.AssertExpectations(t)
		rr.AssertExpectations(t)
	})

	t.Run("success with username", func(t *testing.T) {
		pass := hashedTestPassword("password123")
		ur := &mockUserRepo{}
		rr := &mockRefreshRepo{}
		ur.On("FindByUsername", "testuser").Return(models.User{Model: gorm.Model{ID: 2}, Password: &pass}, nil)
		rr.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(&models.RefreshToken{Model: gorm.Model{ID: 1}}, nil)

		svc := newTestService(ur, rr, nil)
		res, err := svc.SignIn(SignInReq{Identifier: "testuser", Password: "password123"})

		assert.NoError(t, err)
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
		ur.AssertExpectations(t)
		rr.AssertExpectations(t)
	})

	t.Run("refresh token creation fails", func(t *testing.T) {
		pass := hashedTestPassword("password123")
		ur := &mockUserRepo{}
		rr := &mockRefreshRepo{}
		dbErr := errors.New("db error")
		ur.On("FindByEmail", "test@example.com").Return(models.User{Model: gorm.Model{ID: 1}, Password: &pass}, nil)
		rr.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(&models.RefreshToken{}, dbErr)

		svc := newTestService(ur, rr, nil)
		_, err := svc.SignIn(SignInReq{Identifier: "test@example.com", Password: "password123"})

		assert.Error(t, err)
		ur.AssertExpectations(t)
		rr.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// Refresh
// ---------------------------------------------------------------------------

func TestRefresh(t *testing.T) {
	t.Run("token not found", func(t *testing.T) {
		rr := &mockRefreshRepo{}
		notFoundErr := errors.New("failed to find refresh token")
		rr.On("FindByToken", "bad-token").Return(&models.RefreshToken{}, notFoundErr)

		svc := newTestService(&mockUserRepo{}, rr, nil)
		_, err := svc.Refresh(RefreshTokenReq{RefreshToken: "bad-token"})

		assert.Error(t, err)
		rr.AssertExpectations(t)
	})

	t.Run("token is expired", func(t *testing.T) {
		rr := &mockRefreshRepo{}
		expired := &models.RefreshToken{Model: gorm.Model{ID: 1}, UserID: 1, ExpiresAt: time.Now().Add(-time.Hour)}
		rr.On("FindByToken", "expired-token").Return(expired, nil)

		svc := newTestService(&mockUserRepo{}, rr, nil)
		_, err := svc.Refresh(RefreshTokenReq{RefreshToken: "expired-token"})

		assert.ErrorIs(t, err, ErrExpiredToken)
		rr.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		rr := &mockRefreshRepo{}
		valid := &models.RefreshToken{Model: gorm.Model{ID: 1}, UserID: 1, ExpiresAt: time.Now().Add(time.Hour * 24)}
		rr.On("FindByToken", "valid-token").Return(valid, nil)
		rr.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(&models.RefreshToken{Model: gorm.Model{ID: 2}}, nil)

		svc := newTestService(&mockUserRepo{}, rr, nil)
		res, err := svc.Refresh(RefreshTokenReq{RefreshToken: "valid-token"})

		assert.NoError(t, err)
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
		rr.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// DeleteToken
// ---------------------------------------------------------------------------

func TestDeleteToken(t *testing.T) {
	t.Run("repo error", func(t *testing.T) {
		rr := &mockRefreshRepo{}
		rr.On("DeleteByToken", "some-token").Return(errors.New("not found"))

		svc := newTestService(&mockUserRepo{}, rr, nil)
		err := svc.DeleteToken(RefreshTokenReq{RefreshToken: "some-token"})

		assert.Error(t, err)
		rr.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		rr := &mockRefreshRepo{}
		rr.On("DeleteByToken", "some-token").Return(nil)

		svc := newTestService(&mockUserRepo{}, rr, nil)
		err := svc.DeleteToken(RefreshTokenReq{RefreshToken: "some-token"})

		assert.NoError(t, err)
		rr.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// VerifyOTP
// ---------------------------------------------------------------------------

func TestVerifyOTP(t *testing.T) {
	t.Run("no pending registration in cache", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{}, &mockRefreshRepo{}, nil)
		_, err := svc.VerifyOTP(VerifyOTPreq{Email: "test@example.com", OtpCode: "123456"})

		assert.ErrorIs(t, err, ErrPendingRegistrationNotFound)
	})

	t.Run("too many attempts", func(t *testing.T) {
		rc := cache.NewRegistrationCache()
		rc.Set("test@example.com", cache.PendingRegistration{
			Email: "test@example.com", OTPHash: "somehash",
			ExpiresAt: time.Now().Add(time.Hour), Attempts: 5,
		})

		svc := newTestService(&mockUserRepo{}, &mockRefreshRepo{}, rc)
		_, err := svc.VerifyOTP(VerifyOTPreq{Email: "test@example.com", OtpCode: "123456"})

		assert.ErrorIs(t, err, ErrInvalidOTPAttempts)
	})

	t.Run("otp expired", func(t *testing.T) {
		rc := cache.NewRegistrationCache()
		rc.Set("test@example.com", cache.PendingRegistration{
			Email: "test@example.com", OTPHash: "somehash",
			ExpiresAt: time.Now().Add(-time.Minute), Attempts: 0,
		})

		svc := newTestService(&mockUserRepo{}, &mockRefreshRepo{}, rc)
		_, err := svc.VerifyOTP(VerifyOTPreq{Email: "test@example.com", OtpCode: "123456"})

		assert.ErrorIs(t, err, ErrOTPExpired)
	})

	t.Run("wrong otp code increments attempts", func(t *testing.T) {
		rc := cache.NewRegistrationCache()
		rc.Set("test@example.com", cache.PendingRegistration{
			Email: "test@example.com", OTPHash: "wronghash",
			ExpiresAt: time.Now().Add(time.Hour), Attempts: 0,
		})

		svc := newTestService(&mockUserRepo{}, &mockRefreshRepo{}, rc)
		_, err := svc.VerifyOTP(VerifyOTPreq{Email: "test@example.com", OtpCode: "000000"})

		assert.ErrorIs(t, err, ErrInvalidOTP)

		// attempts should have been incremented in cache
		pending, _ := rc.Get("test@example.com")
		assert.Equal(t, 1, pending.Attempts)
	})

	t.Run("success creates user in db and removes from cache", func(t *testing.T) {
		code := "123456"
		hash := computeOTPHash(code)
		hashedPw := hashedTestPassword("password123")

		rc := cache.NewRegistrationCache()
		rc.Set("test@example.com", cache.PendingRegistration{
			Email: "test@example.com", FirstName: "Test", LastName: "User",
			Username: "testuser", HashedPassword: hashedPw,
			OTPHash: hash, ExpiresAt: time.Now().Add(time.Hour), Attempts: 0,
		})

		ur := &mockUserRepo{}
		ur.On("CreateUser", mock.MatchedBy(func(u models.User) bool {
			return u.Email == "test@example.com" && u.IsVerified
		})).Return(models.User{Model: gorm.Model{ID: 1}}, nil)

		svc := newTestService(ur, &mockRefreshRepo{}, rc)
		ok, err := svc.VerifyOTP(VerifyOTPreq{Email: "test@example.com", OtpCode: code})

		assert.NoError(t, err)
		assert.True(t, ok)
		ur.AssertExpectations(t)

		// entry should be removed from cache
		_, found := rc.Get("test@example.com")
		assert.False(t, found)
	})
}

// ---------------------------------------------------------------------------
// ResendOTP
// ---------------------------------------------------------------------------

func TestResendOTP(t *testing.T) {
	t.Run("no pending registration in cache", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{}, &mockRefreshRepo{}, nil)
		err := svc.ResendOTP(ResendOTPreq{Email: "notfound@example.com"})

		assert.ErrorIs(t, err, ErrPendingRegistrationNotFound)
	})
}
