package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Haidarr-h/backend-go/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	logger.Init()
	m.Run()
}

// ---------------------------------------------------------------------------
// Mock Service
// ---------------------------------------------------------------------------

type mockService struct{ mock.Mock }

func (m *mockService) SignUp(req SignUpRequest) (SignUpResponse, error) {
	args := m.Called(req)
	return args.Get(0).(SignUpResponse), args.Error(1)
}
func (m *mockService) SignIn(req SignInReq) (SignInRes, error) {
	args := m.Called(req)
	return args.Get(0).(SignInRes), args.Error(1)
}
func (m *mockService) Refresh(req RefreshTokenReq) (RefreshTokenRes, error) {
	args := m.Called(req)
	return args.Get(0).(RefreshTokenRes), args.Error(1)
}
func (m *mockService) DeleteToken(req RefreshTokenReq) error {
	args := m.Called(req)
	return args.Error(0)
}
func (m *mockService) VerifyOTP(req VerifyOTPreq) (bool, error) {
	args := m.Called(req)
	return args.Bool(0), args.Error(1)
}
func (m *mockService) ResendOTP(req ResendOTPreq) error {
	args := m.Called(req)
	return args.Error(0)
}
func (m *mockService) GoogleSignIn(req GoogleSignInReq) (GoogleSignInRes, error) {
	args := m.Called(req)
	return args.Get(0).(GoogleSignInRes), args.Error(1)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestHandler(svc Service) *AuthHandler {
	return NewAuthHandler(svc)
}

func newTestRouter(h *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")
	auth.POST("/signup", h.SignUp)
	auth.POST("/signin", h.SignIn)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/signout", h.SignOut)
	auth.POST("/verifyOTP", h.VerifyOtp)
	auth.POST("/resendOTP", h.ResendOTP)
	auth.POST("/google/mobile", h.GoogleMobileSignIn)
	return r
}

func postJSON(r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---------------------------------------------------------------------------
// SignUp handler
// ---------------------------------------------------------------------------

func TestSignUpHandler(t *testing.T) {
	validBody := SignUpRequest{
		Email: "test@example.com", Password: "password123",
		Username: "testuser", FirstName: "Test", LastName: "User",
	}

	t.Run("bad request - invalid body", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(newTestHandler(svc))

		w := postJSON(r, "/api/v1/auth/signup", map[string]string{"email": "not-an-email"})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("conflict - email exists", func(t *testing.T) {
		svc := &mockService{}
		svc.On("SignUp", validBody).Return(SignUpResponse{}, ErrEmailIsExists)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/signup", validBody)

		assert.Equal(t, http.StatusConflict, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("conflict - username exists", func(t *testing.T) {
		svc := &mockService{}
		svc.On("SignUp", validBody).Return(SignUpResponse{}, ErrUsernameIsExists)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/signup", validBody)

		assert.Equal(t, http.StatusConflict, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		svc.On("SignUp", validBody).Return(SignUpResponse{}, errors.New("db error"))

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/signup", validBody)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("SignUp", validBody).Return(SignUpResponse{ID: 1, Username: "testuser"}, nil)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/signup", validBody)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// SignIn handler
// ---------------------------------------------------------------------------

func TestSignInHandler(t *testing.T) {
	validBody := SignInReq{Identifier: "test@example.com", Password: "password123"}

	t.Run("bad request - invalid body", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(newTestHandler(svc))

		w := postJSON(r, "/api/v1/auth/signin", map[string]string{})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized - invalid credentials", func(t *testing.T) {
		svc := &mockService{}
		svc.On("SignIn", validBody).Return(SignInRes{}, ErrInvalidCredentials)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/signin", validBody)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("bad request - google user", func(t *testing.T) {
		svc := &mockService{}
		svc.On("SignIn", validBody).Return(SignInRes{}, ErrUserGoogleSignIn)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/signin", validBody)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error - failed create token", func(t *testing.T) {
		svc := &mockService{}
		svc.On("SignIn", validBody).Return(SignInRes{}, ErrFailedCreateToken)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/signin", validBody)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("SignIn", validBody).Return(SignInRes{AccessToken: "aaa", RefreshToken: "bbb"}, nil)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/signin", validBody)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// Refresh handler
// ---------------------------------------------------------------------------

func TestRefreshHandler(t *testing.T) {
	validBody := RefreshTokenReq{RefreshToken: "valid-refresh-token"}

	t.Run("bad request - invalid body", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(newTestHandler(svc))

		w := postJSON(r, "/api/v1/auth/refresh", map[string]string{})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bad request - expired token", func(t *testing.T) {
		svc := &mockService{}
		svc.On("Refresh", validBody).Return(RefreshTokenRes{}, ErrExpiredToken)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/refresh", validBody)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		svc.On("Refresh", validBody).Return(RefreshTokenRes{}, errors.New("db error"))

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/refresh", validBody)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("Refresh", validBody).Return(RefreshTokenRes{AccessToken: "aaa", RefreshToken: "bbb"}, nil)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/refresh", validBody)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// SignOut handler
// ---------------------------------------------------------------------------

func TestSignOutHandler(t *testing.T) {
	validBody := RefreshTokenReq{RefreshToken: "valid-refresh-token"}

	t.Run("bad request - invalid body", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(newTestHandler(svc))

		w := postJSON(r, "/api/v1/auth/signout", map[string]string{})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		svc.On("DeleteToken", validBody).Return(errors.New("db error"))

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/signout", validBody)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("DeleteToken", validBody).Return(nil)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/signout", validBody)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// VerifyOtp handler
// ---------------------------------------------------------------------------

func TestVerifyOtpHandler(t *testing.T) {
	validBody := VerifyOTPreq{Email: "test@example.com", OtpCode: "123456"}

	t.Run("bad request - invalid body", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(newTestHandler(svc))

		w := postJSON(r, "/api/v1/auth/verifyOTP", map[string]string{})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bad request - too many attempts", func(t *testing.T) {
		svc := &mockService{}
		svc.On("VerifyOTP", validBody).Return(false, ErrInvalidOTPAttempts)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/verifyOTP", validBody)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("gone - otp expired", func(t *testing.T) {
		svc := &mockService{}
		svc.On("VerifyOTP", validBody).Return(false, ErrOTPExpired)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/verifyOTP", validBody)

		assert.Equal(t, http.StatusGone, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("bad request - already used", func(t *testing.T) {
		svc := &mockService{}
		svc.On("VerifyOTP", validBody).Return(false, ErrInvalidOTPUsed)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/verifyOTP", validBody)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("bad request - wrong code", func(t *testing.T) {
		svc := &mockService{}
		svc.On("VerifyOTP", validBody).Return(false, ErrInvalidOTP)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/verifyOTP", validBody)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		svc.On("VerifyOTP", validBody).Return(false, errors.New("db error"))

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/verifyOTP", validBody)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("VerifyOTP", validBody).Return(true, nil)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/verifyOTP", validBody)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// ResendOTP handler
// ---------------------------------------------------------------------------

func TestResendOTPHandler(t *testing.T) {
	validBody := ResendOTPreq{Email: "test@example.com"}

	t.Run("bad request - invalid body", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(newTestHandler(svc))

		w := postJSON(r, "/api/v1/auth/resendOTP", map[string]string{})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		svc.On("ResendOTP", validBody).Return(errors.New("db error"))

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/resendOTP", validBody)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("ResendOTP", validBody).Return(nil)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/resendOTP", validBody)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// GoogleMobileSignIn handler
// ---------------------------------------------------------------------------

func TestGoogleMobileSignInHandler(t *testing.T) {
	validBody := GoogleSignInReq{IDToken: "google-id-token"}

	t.Run("bad request - invalid body", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(newTestHandler(svc))

		w := postJSON(r, "/api/v1/auth/google/mobile", map[string]string{})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bad request - invalid google token", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GoogleSignIn", validBody).Return(GoogleSignInRes{}, ErrInvalidGoogleIDToken)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/google/mobile", validBody)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GoogleSignIn", validBody).Return(GoogleSignInRes{}, errors.New("unexpected error"))

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/google/mobile", validBody)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GoogleSignIn", validBody).Return(GoogleSignInRes{
			ID: 1, AccessToken: "aaa", RefreshToken: "bbb",
		}, nil)

		r := newTestRouter(newTestHandler(svc))
		w := postJSON(r, "/api/v1/auth/google/mobile", validBody)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// keep the mock package import used
var _ = mock.AnythingOfType
