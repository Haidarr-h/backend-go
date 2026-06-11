package user

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// Mock Service
// ---------------------------------------------------------------------------

type mockService struct{ mock.Mock }

func (m *mockService) GetMyData(userID uint) (UserResponse, error) {
	args := m.Called(userID)
	return args.Get(0).(UserResponse), args.Error(1)
}
func (m *mockService) Delete(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newTestRouter wires the user routes behind a middleware that injects the
// given userID (0 simulates a missing/invalid token).
func newTestRouter(svc Service, userID uint) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if userID != 0 {
			c.Set("userID", userID)
		}
		c.Next()
	})

	h := NewUserHandler(svc)
	h.RegisterUserRoutes(r.Group("/api/v1"))
	return r
}

func do(r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---------------------------------------------------------------------------
// GetMyData handler
// ---------------------------------------------------------------------------

func TestGetMyDataHandler(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetMyData", uint(1)).Return(UserResponse{}, ErrUserNotFound)

		r := newTestRouter(svc, 1)
		w := do(r, http.MethodGet, "/api/v1/users/me")

		assert.Equal(t, http.StatusNotFound, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetMyData", uint(1)).Return(UserResponse{}, errors.New("db error"))

		r := newTestRouter(svc, 1)
		w := do(r, http.MethodGet, "/api/v1/users/me")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetMyData", uint(1)).Return(UserResponse{Email: "jane@example.com"}, nil)

		r := newTestRouter(svc, 1)
		w := do(r, http.MethodGet, "/api/v1/users/me")

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// DeleteUser handler
// ---------------------------------------------------------------------------

func TestDeleteUserHandler(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc := &mockService{}
		svc.On("Delete", uint(1)).Return(ErrUserNotFound)

		r := newTestRouter(svc, 1)
		w := do(r, http.MethodDelete, "/api/v1/users/me")

		assert.Equal(t, http.StatusNotFound, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		svc.On("Delete", uint(1)).Return(errors.New("db error"))

		r := newTestRouter(svc, 1)
		w := do(r, http.MethodDelete, "/api/v1/users/me")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("Delete", uint(1)).Return(nil)

		r := newTestRouter(svc, 1)
		w := do(r, http.MethodDelete, "/api/v1/users/me")

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}
