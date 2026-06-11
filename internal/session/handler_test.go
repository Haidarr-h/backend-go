package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// Mock Service
// ---------------------------------------------------------------------------

type mockService struct{ mock.Mock }

func (m *mockService) CreateSession(userID uint, req CreateSessionRequest) (SessionResponse, error) {
	args := m.Called(userID, req)
	return args.Get(0).(SessionResponse), args.Error(1)
}
func (m *mockService) GetSessions(userID uint) ([]SessionResponse, error) {
	args := m.Called(userID)
	return args.Get(0).([]SessionResponse), args.Error(1)
}
func (m *mockService) GetSession(userID uint, sessionID uint) (SessionResponse, error) {
	args := m.Called(userID, sessionID)
	return args.Get(0).(SessionResponse), args.Error(1)
}
func (m *mockService) UpdateSession(id, userID uint, req UpdateSessionReq) (SessionResponse, error) {
	args := m.Called(id, userID, req)
	return args.Get(0).(SessionResponse), args.Error(1)
}
func (m *mockService) DeleteSession(sessionID uint, userID uint) error {
	args := m.Called(sessionID, userID)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestRouter(svc Service, userID uint) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if userID != 0 {
			c.Set("userID", userID)
		}
		c.Next()
	})

	h := NewSessionHandler(svc)
	h.RegisterRoutes(r.Group("/api/v1"), &config.Config{})
	return r
}

func doJSON(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func doRaw(r *gin.Engine, method, path, rawBody string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(rawBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func validBody() CreateSessionRequest {
	return CreateSessionRequest{RoutineID: 1, Name: "Morning Push", StartTime: startTime, EndTime: endTime}
}

// ---------------------------------------------------------------------------
// CreateSession handler
// ---------------------------------------------------------------------------

func TestCreateSessionHandler(t *testing.T) {
	t.Run("unauthorized - no userID", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(svc, 0)
		w := doJSON(r, http.MethodPost, "/api/v1/sessions/", validBody())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("bad request - invalid body", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPost, "/api/v1/sessions/", map[string]string{})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bad request - routine not found", func(t *testing.T) {
		svc := &mockService{}
		body := validBody()
		svc.On("CreateSession", uint(1), body).Return(SessionResponse{}, ErrRoutineNotFound)
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPost, "/api/v1/sessions/", body)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("forbidden - routine not accessible", func(t *testing.T) {
		svc := &mockService{}
		body := validBody()
		svc.On("CreateSession", uint(1), body).Return(SessionResponse{}, ErrRoutineNotAccessible)
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPost, "/api/v1/sessions/", body)
		assert.Equal(t, http.StatusForbidden, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("bad request - invalid time range", func(t *testing.T) {
		svc := &mockService{}
		body := validBody()
		svc.On("CreateSession", uint(1), body).Return(SessionResponse{}, ErrInvalidTimeRange)
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPost, "/api/v1/sessions/", body)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		body := validBody()
		svc.On("CreateSession", uint(1), body).Return(SessionResponse{}, errors.New("db error"))
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPost, "/api/v1/sessions/", body)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		body := validBody()
		svc.On("CreateSession", uint(1), body).Return(SessionResponse{ID: 5, Name: "Morning Push"}, nil)
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPost, "/api/v1/sessions/", body)
		assert.Equal(t, http.StatusCreated, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// GetSessions handler
// ---------------------------------------------------------------------------

func TestGetSessionsHandler(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(svc, 0)
		w := doJSON(r, http.MethodGet, "/api/v1/sessions/", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		svc.AssertNotCalled(t, "GetSessions", mock.Anything)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetSessions", uint(1)).Return([]SessionResponse{}, errors.New("db error"))
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodGet, "/api/v1/sessions/", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetSessions", uint(1)).Return([]SessionResponse{{ID: 1}}, nil)
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodGet, "/api/v1/sessions/", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// GetSession handler
// ---------------------------------------------------------------------------

func TestGetSessionHandler(t *testing.T) {
	t.Run("bad request - invalid id", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodGet, "/api/v1/sessions/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetSession", uint(1), uint(5)).Return(SessionResponse{}, ErrSessionNotFound)
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodGet, "/api/v1/sessions/5", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("forbidden", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetSession", uint(1), uint(5)).Return(SessionResponse{}, InvalidSessionOwnership)
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodGet, "/api/v1/sessions/5", nil)
		assert.Equal(t, http.StatusForbidden, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetSession", uint(1), uint(5)).Return(SessionResponse{ID: 5}, nil)
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodGet, "/api/v1/sessions/5", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// UpdateSession handler
// ---------------------------------------------------------------------------

func TestUpdateSessionHandler(t *testing.T) {
	t.Run("bad request - invalid id", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPatch, "/api/v1/sessions/abc", UpdateSessionReq{})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bad request - malformed json", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(svc, 1)
		w := doRaw(r, http.MethodPatch, "/api/v1/sessions/5", "{not json")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		svc := &mockService{}
		req := UpdateSessionReq{Name: ptr("New")}
		svc.On("UpdateSession", uint(5), uint(1), req).Return(SessionResponse{}, ErrSessionNotFound)
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPatch, "/api/v1/sessions/5", req)
		assert.Equal(t, http.StatusNotFound, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("forbidden", func(t *testing.T) {
		svc := &mockService{}
		req := UpdateSessionReq{Name: ptr("New")}
		svc.On("UpdateSession", uint(5), uint(1), req).Return(SessionResponse{}, InvalidSessionOwnership)
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPatch, "/api/v1/sessions/5", req)
		assert.Equal(t, http.StatusForbidden, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("bad request - invalid time range", func(t *testing.T) {
		svc := &mockService{}
		req := UpdateSessionReq{EndTime: ptr(startTime.Add(-time.Hour))}
		svc.On("UpdateSession", uint(5), uint(1), req).Return(SessionResponse{}, ErrInvalidTimeRange)
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPatch, "/api/v1/sessions/5", req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		req := UpdateSessionReq{Name: ptr("New")}
		svc.On("UpdateSession", uint(5), uint(1), req).Return(SessionResponse{ID: 5, Name: "New"}, nil)
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPatch, "/api/v1/sessions/5", req)
		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// DeleteSession handler
// ---------------------------------------------------------------------------

func TestDeleteSessionHandler(t *testing.T) {
	t.Run("bad request - invalid id", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodDelete, "/api/v1/sessions/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		svc := &mockService{}
		svc.On("DeleteSession", uint(5), uint(1)).Return(ErrSessionNotFound)
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodDelete, "/api/v1/sessions/5", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		svc.On("DeleteSession", uint(5), uint(1)).Return(errors.New("db error"))
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodDelete, "/api/v1/sessions/5", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("DeleteSession", uint(5), uint(1)).Return(nil)
		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodDelete, "/api/v1/sessions/5", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}
