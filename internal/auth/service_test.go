package auth

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/Haidarr-h/backend-go/internal/user"
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

func (m *mockUserRepo) CreateUser(u user.User) (user.User, error) {
	args := m.Called(u)
	return args.Get(0).(user.User), args.Error(1)
}
func (m *mockUserRepo) UpdateUser(u user.User) (user.User, error) {
	args := m.Called(u)
	return args.Get(0).(user.User), args.Error(1)
}
func (m *mockUserRepo) FindByEmail(email string) (user.User, error) {
	args := m.Called(email)
	return args.Get(0).(user.User), args.Error(1)
}
func (m *mockUserRepo) FindByUsername(username string) (user.User, error) {
	args := m.Called(username)
	return args.Get(0).(user.User), args.Error(1)
}
func (m *mockUserRepo) FindByGoogleID(googleID string) (user.User, error) {
	args := m.Called(googleID)
	return args.Get(0).(user.User), args.Error(1)
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

type mockOtpRepo struct{ mock.Mock }

func (m *mockOtpRepo) Create(otp OtpVerification) (OtpVerification, error) {
	args := m.Called(otp)
	return args.Get(0).(OtpVerification), args.Error(1)
}
func (m *mockOtpRepo) FindByEmail(email string) (OtpVerification, error) {
	args := m.Called(email)
	return args.Get(0).(OtpVerification), args.Error(1)
}
func (m *mockOtpRepo) Update(otp OtpVerification) (OtpVerification, error) {
	args := m.Called(otp)
	return args.Get(0).(OtpVerification), args.Error(1)
}

type mockRefreshRepo struct{ mock.Mock }

func (m *mockRefreshRepo) Create(token *RefreshToken) (*RefreshToken, error) {
	args := m.Called(token)
	return args.Get(0).(*RefreshToken), args.Error(1)
}
func (m *mockRefreshRepo) FindByToken(token string) (*RefreshToken, error) {
	args := m.Called(token)
	return args.Get(0).(*RefreshToken), args.Error(1)
}
func (m *mockRefreshRepo) Update(token *RefreshToken) (*RefreshToken, error) {
	args := m.Called(token)
	return args.Get(0).(*RefreshToken), args.Error(1)
}
func (m *mockRefreshRepo) DeleteByToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestService(ur *mockUserRepo, rr *mockRefreshRepo, or *mockOtpRepo) *AuthService {
	return NewAuthService(ur, &config.Config{
		JWTSecret:     "test-jwt-secret-32-bytes-padding!",
		RefreshSecret: "test-refresh-secret-32-bytes-pad!",
	}, rr, or)
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

		svc := newTestService(ur, &mockRefreshRepo{}, &mockOtpRepo{})
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

		svc := newTestService(ur, &mockRefreshRepo{}, &mockOtpRepo{})
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

		svc := newTestService(ur, &mockRefreshRepo{}, &mockOtpRepo{})
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
		ur.On("FindByEmail", "test@example.com").Return(user.User{}, user.ErrUserNotFound)

		svc := newTestService(ur, &mockRefreshRepo{}, &mockOtpRepo{})
		_, err := svc.SignIn(SignInReq{Identifier: "test@example.com", Password: "password123"})

		assert.ErrorIs(t, err, ErrInvalidCredentials)
		ur.AssertExpectations(t)
	})

	t.Run("user not found by username", func(t *testing.T) {
		ur := &mockUserRepo{}
		ur.On("FindByUsername", "testuser").Return(user.User{}, user.ErrUserNotFound)

		svc := newTestService(ur, &mockRefreshRepo{}, &mockOtpRepo{})
		_, err := svc.SignIn(SignInReq{Identifier: "testuser", Password: "password123"})

		assert.ErrorIs(t, err, ErrInvalidCredentials)
		ur.AssertExpectations(t)
	})

	t.Run("user is google-only account", func(t *testing.T) {
		ur := &mockUserRepo{}
		ur.On("FindByEmail", "google@example.com").Return(user.User{Model: gorm.Model{ID: 1}, Password: nil}, nil)

		svc := newTestService(ur, &mockRefreshRepo{}, &mockOtpRepo{})
		_, err := svc.SignIn(SignInReq{Identifier: "google@example.com", Password: "password123"})

		assert.ErrorIs(t, err, ErrUserGoogleSignIn)
		ur.AssertExpectations(t)
	})

	t.Run("wrong password", func(t *testing.T) {
		pass := hashedTestPassword("correctpassword")
		ur := &mockUserRepo{}
		ur.On("FindByEmail", "test@example.com").Return(user.User{Model: gorm.Model{ID: 1}, Password: &pass}, nil)

		svc := newTestService(ur, &mockRefreshRepo{}, &mockOtpRepo{})
		_, err := svc.SignIn(SignInReq{Identifier: "test@example.com", Password: "wrongpassword"})

		assert.ErrorIs(t, err, ErrInvalidCredentials)
		ur.AssertExpectations(t)
	})

	t.Run("success with email", func(t *testing.T) {
		pass := hashedTestPassword("password123")
		ur := &mockUserRepo{}
		rr := &mockRefreshRepo{}
		ur.On("FindByEmail", "test@example.com").Return(user.User{Model: gorm.Model{ID: 1}, Password: &pass}, nil)
		rr.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(&RefreshToken{Model: gorm.Model{ID: 1}}, nil)

		svc := newTestService(ur, rr, &mockOtpRepo{})
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
		ur.On("FindByUsername", "testuser").Return(user.User{Model: gorm.Model{ID: 2}, Password: &pass}, nil)
		rr.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(&RefreshToken{Model: gorm.Model{ID: 1}}, nil)

		svc := newTestService(ur, rr, &mockOtpRepo{})
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
		ur.On("FindByEmail", "test@example.com").Return(user.User{Model: gorm.Model{ID: 1}, Password: &pass}, nil)
		rr.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(&RefreshToken{}, dbErr)

		svc := newTestService(ur, rr, &mockOtpRepo{})
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
		rr.On("FindByToken", "bad-token").Return(&RefreshToken{}, notFoundErr)

		svc := newTestService(&mockUserRepo{}, rr, &mockOtpRepo{})
		_, err := svc.Refresh(RefreshTokenReq{RefreshToken: "bad-token"})

		assert.Error(t, err)
		rr.AssertExpectations(t)
	})

	t.Run("token is expired", func(t *testing.T) {
		rr := &mockRefreshRepo{}
		expired := &RefreshToken{Model: gorm.Model{ID: 1}, UserID: 1, ExpiresAt: time.Now().Add(-time.Hour)}
		rr.On("FindByToken", "expired-token").Return(expired, nil)

		svc := newTestService(&mockUserRepo{}, rr, &mockOtpRepo{})
		_, err := svc.Refresh(RefreshTokenReq{RefreshToken: "expired-token"})

		assert.ErrorIs(t, err, ErrExpiredToken)
		rr.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		rr := &mockRefreshRepo{}
		valid := &RefreshToken{Model: gorm.Model{ID: 1}, UserID: 1, ExpiresAt: time.Now().Add(time.Hour * 24)}
		rr.On("FindByToken", "valid-token").Return(valid, nil)
		rr.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(&RefreshToken{Model: gorm.Model{ID: 2}}, nil)

		svc := newTestService(&mockUserRepo{}, rr, &mockOtpRepo{})
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

		svc := newTestService(&mockUserRepo{}, rr, &mockOtpRepo{})
		err := svc.DeleteToken(RefreshTokenReq{RefreshToken: "some-token"})

		assert.Error(t, err)
		rr.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		rr := &mockRefreshRepo{}
		rr.On("DeleteByToken", "some-token").Return(nil)

		svc := newTestService(&mockUserRepo{}, rr, &mockOtpRepo{})
		err := svc.DeleteToken(RefreshTokenReq{RefreshToken: "some-token"})

		assert.NoError(t, err)
		rr.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// VerifyOTP
// ---------------------------------------------------------------------------

func TestVerifyOTP(t *testing.T) {
	t.Run("otp record not found", func(t *testing.T) {
		or := &mockOtpRepo{}
		or.On("FindByEmail", "test@example.com").Return(OtpVerification{}, ErrEmailNotFound)

		svc := newTestService(&mockUserRepo{}, &mockRefreshRepo{}, or)
		_, err := svc.VerifyOTP(VerifyOTPreq{Email: "test@example.com", OtpCode: "123456"})

		assert.ErrorIs(t, err, ErrEmailNotFound)
		or.AssertExpectations(t)
	})

	t.Run("too many attempts", func(t *testing.T) {
		or := &mockOtpRepo{}
		or.On("FindByEmail", "test@example.com").Return(OtpVerification{
			Attempts: 5, ExpiresAt: time.Now().Add(time.Hour),
		}, nil)

		svc := newTestService(&mockUserRepo{}, &mockRefreshRepo{}, or)
		_, err := svc.VerifyOTP(VerifyOTPreq{Email: "test@example.com", OtpCode: "123456"})

		assert.ErrorIs(t, err, ErrInvalidOTPAttempts)
		or.AssertExpectations(t)
	})

	t.Run("otp expired", func(t *testing.T) {
		or := &mockOtpRepo{}
		or.On("FindByEmail", "test@example.com").Return(OtpVerification{
			Attempts: 0, ExpiresAt: time.Now().Add(-time.Minute),
		}, nil)

		svc := newTestService(&mockUserRepo{}, &mockRefreshRepo{}, or)
		_, err := svc.VerifyOTP(VerifyOTPreq{Email: "test@example.com", OtpCode: "123456"})

		assert.ErrorIs(t, err, ErrOTPExpired)
		or.AssertExpectations(t)
	})

	t.Run("otp already used", func(t *testing.T) {
		or := &mockOtpRepo{}
		or.On("FindByEmail", "test@example.com").Return(OtpVerification{
			Attempts: 0, ExpiresAt: time.Now().Add(time.Hour), Used: true,
		}, nil)

		svc := newTestService(&mockUserRepo{}, &mockRefreshRepo{}, or)
		_, err := svc.VerifyOTP(VerifyOTPreq{Email: "test@example.com", OtpCode: "123456"})

		assert.ErrorIs(t, err, ErrInvalidOTPUsed)
		or.AssertExpectations(t)
	})

	t.Run("wrong otp code increments attempts", func(t *testing.T) {
		or := &mockOtpRepo{}
		record := OtpVerification{
			Model: gorm.Model{ID: 1}, UserID: 1, OTPHash: "wronghash",
			ExpiresAt: time.Now().Add(time.Hour), Attempts: 0, Used: false,
		}
		updated := record
		updated.Attempts = 1

		or.On("FindByEmail", "test@example.com").Return(record, nil)
		or.On("Update", updated).Return(updated, nil)

		svc := newTestService(&mockUserRepo{}, &mockRefreshRepo{}, or)
		_, err := svc.VerifyOTP(VerifyOTPreq{Email: "test@example.com", OtpCode: "000000"})

		assert.ErrorIs(t, err, ErrInvalidOTP)
		or.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		code := "123456"
		hash := computeOTPHash(code)
		or := &mockOtpRepo{}
		ur := &mockUserRepo{}

		record := OtpVerification{
			Model: gorm.Model{ID: 1}, UserID: 1, OTPHash: hash,
			ExpiresAt: time.Now().Add(time.Hour), Attempts: 0, Used: false,
		}
		updated := record
		updated.Attempts = 1
		updated.Used = true

		or.On("FindByEmail", "test@example.com").Return(record, nil)
		or.On("Update", updated).Return(updated, nil)
		ur.On("Verify", uint(1)).Return(nil)

		svc := newTestService(ur, &mockRefreshRepo{}, or)
		ok, err := svc.VerifyOTP(VerifyOTPreq{Email: "test@example.com", OtpCode: code})

		assert.NoError(t, err)
		assert.True(t, ok)
		or.AssertExpectations(t)
		ur.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// ResendOTP
// ---------------------------------------------------------------------------

func TestResendOTP(t *testing.T) {
	t.Run("email not found", func(t *testing.T) {
		or := &mockOtpRepo{}
		or.On("FindByEmail", "notfound@example.com").Return(OtpVerification{}, ErrEmailNotFound)

		svc := newTestService(&mockUserRepo{}, &mockRefreshRepo{}, or)
		err := svc.ResendOTP(ResendOTPreq{Email: "notfound@example.com"})

		assert.ErrorIs(t, err, ErrEmailNotFound)
		or.AssertExpectations(t)
	})

	t.Run("otp create fails", func(t *testing.T) {
		or := &mockOtpRepo{}
		record := OtpVerification{Model: gorm.Model{ID: 1}, UserID: 1}
		dbErr := errors.New("db error")

		or.On("FindByEmail", "test@example.com").Return(record, nil)
		or.On("Create", mock.AnythingOfType("OtpVerification")).Return(OtpVerification{}, dbErr)

		svc := newTestService(&mockUserRepo{}, &mockRefreshRepo{}, or)
		err := svc.ResendOTP(ResendOTPreq{Email: "test@example.com"})

		assert.Error(t, err)
		or.AssertExpectations(t)
	})
}
