package routine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// Mock Service
// ---------------------------------------------------------------------------

type mockService struct{ mock.Mock }

func (m *mockService) CreateRoutine(userID uint, req CreateRoutineRequest) (RoutineResponse, error) {
	args := m.Called(userID, req)
	return args.Get(0).(RoutineResponse), args.Error(1)
}
func (m *mockService) GetRoutines(userID uint) ([]RoutineResponse, error) {
	args := m.Called(userID)
	return args.Get(0).([]RoutineResponse), args.Error(1)
}
func (m *mockService) GetTemplates() ([]RoutineResponse, error) {
	args := m.Called()
	return args.Get(0).([]RoutineResponse), args.Error(1)
}
func (m *mockService) GetRoutine(userID uint, routineID uint) (RoutineResponse, error) {
	args := m.Called(userID, routineID)
	return args.Get(0).(RoutineResponse), args.Error(1)
}
func (m *mockService) UpdateRoutine(id, userID uint, req UpdateRoutineReq) (RoutineResponse, error) {
	args := m.Called(id, userID, req)
	return args.Get(0).(RoutineResponse), args.Error(1)
}
func (m *mockService) DeleteRoutine(routineID uint, userID uint) error {
	args := m.Called(routineID, userID)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newTestRouter wires the routine routes behind a middleware that injects the
// given userID into the context (userID 0 simulates a missing/invalid token).
func newTestRouter(svc Service, userID uint) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if userID != 0 {
			c.Set("userID", userID)
		}
		c.Next()
	})

	h := NewRoutineHandler(svc)
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

func validCreateBody() CreateRoutineRequest {
	return CreateRoutineRequest{
		Name: "Push Day",
		RoutineExercises: []RoutineExercisesRequest{
			{ExerciseID: 12, Order: 1},
		},
	}
}

// ---------------------------------------------------------------------------
// CreateRoutine handler
// ---------------------------------------------------------------------------

func TestCreateRoutineHandler(t *testing.T) {
	t.Run("bad request - invalid body", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(svc, 1)

		w := doJSON(r, http.MethodPost, "/api/v1/routines/", map[string]string{})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bad request - exercise not found", func(t *testing.T) {
		svc := &mockService{}
		body := validCreateBody()
		svc.On("CreateRoutine", uint(1), body).
			Return(RoutineResponse{}, fmt.Errorf("%w: %v", ErrExerciseNotFound, []uint{12}))

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPost, "/api/v1/routines/", body)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		body := validCreateBody()
		svc.On("CreateRoutine", uint(1), body).Return(RoutineResponse{}, errors.New("db error"))

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPost, "/api/v1/routines/", body)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		body := validCreateBody()
		svc.On("CreateRoutine", uint(1), body).Return(RoutineResponse{ID: 1, Name: "Push Day"}, nil)

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPost, "/api/v1/routines/", body)

		assert.Equal(t, http.StatusCreated, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// GetRoutines handler
// ---------------------------------------------------------------------------

func TestGetRoutinesHandler(t *testing.T) {
	t.Run("unauthorized - no userID", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(svc, 0)

		w := doJSON(r, http.MethodGet, "/api/v1/routines/", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		svc.AssertNotCalled(t, "GetRoutines", mock.Anything)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetRoutines", uint(1)).Return([]RoutineResponse{}, errors.New("db error"))

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodGet, "/api/v1/routines/", nil)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetRoutines", uint(1)).Return([]RoutineResponse{{ID: 1}}, nil)

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodGet, "/api/v1/routines/", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// GetTemplates handler
// ---------------------------------------------------------------------------

func TestGetTemplatesHandler(t *testing.T) {
	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetTemplates").Return([]RoutineResponse{}, errors.New("db error"))

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodGet, "/api/v1/routines/templates", nil)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetTemplates").Return([]RoutineResponse{{ID: 1, IsPublic: true}}, nil)

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodGet, "/api/v1/routines/templates", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// GetRoutine handler
// ---------------------------------------------------------------------------

func TestGetRoutineHandler(t *testing.T) {
	t.Run("bad request - invalid id", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(svc, 1)

		w := doJSON(r, http.MethodGet, "/api/v1/routines/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetRoutine", uint(1), uint(5)).Return(RoutineResponse{}, ErrRoutineNotFound)

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodGet, "/api/v1/routines/5", nil)

		assert.Equal(t, http.StatusNotFound, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("forbidden - not owner", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetRoutine", uint(1), uint(5)).Return(RoutineResponse{}, InvalidRoutineOwnership)

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodGet, "/api/v1/routines/5", nil)

		assert.Equal(t, http.StatusForbidden, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetRoutine", uint(1), uint(5)).Return(RoutineResponse{}, errors.New("db error"))

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodGet, "/api/v1/routines/5", nil)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("GetRoutine", uint(1), uint(5)).Return(RoutineResponse{ID: 5}, nil)

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodGet, "/api/v1/routines/5", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// UpdateRoutine handler
// ---------------------------------------------------------------------------

func TestUpdateRoutineHandler(t *testing.T) {
	t.Run("bad request - invalid id", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(svc, 1)

		w := doJSON(r, http.MethodPatch, "/api/v1/routines/abc", UpdateRoutineReq{})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bad request - malformed json", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(svc, 1)

		w := doRaw(r, http.MethodPatch, "/api/v1/routines/5", "{not json")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("forbidden - not owner", func(t *testing.T) {
		svc := &mockService{}
		req := UpdateRoutineReq{Name: ptr("New")}
		svc.On("UpdateRoutine", uint(5), uint(1), req).Return(RoutineResponse{}, InvalidRoutineOwnership)

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPatch, "/api/v1/routines/5", req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		svc := &mockService{}
		req := UpdateRoutineReq{Name: ptr("New")}
		svc.On("UpdateRoutine", uint(5), uint(1), req).Return(RoutineResponse{}, ErrRoutineNotFound)

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPatch, "/api/v1/routines/5", req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("bad request - exercise not found", func(t *testing.T) {
		svc := &mockService{}
		req := UpdateRoutineReq{RoutineExercises: []RoutineExercisesRequest{{ExerciseID: 77, Order: 1}}}
		svc.On("UpdateRoutine", uint(5), uint(1), req).
			Return(RoutineResponse{}, fmt.Errorf("%w: %v", ErrExerciseNotFound, []uint{77}))

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPatch, "/api/v1/routines/5", req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		req := UpdateRoutineReq{Name: ptr("New")}
		svc.On("UpdateRoutine", uint(5), uint(1), req).Return(RoutineResponse{}, errors.New("db error"))

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPatch, "/api/v1/routines/5", req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		req := UpdateRoutineReq{Name: ptr("New")}
		svc.On("UpdateRoutine", uint(5), uint(1), req).Return(RoutineResponse{ID: 5, Name: "New"}, nil)

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodPatch, "/api/v1/routines/5", req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// DeleteRoutine handler
// ---------------------------------------------------------------------------

func TestDeleteRoutineHandler(t *testing.T) {
	t.Run("bad request - invalid id", func(t *testing.T) {
		svc := &mockService{}
		r := newTestRouter(svc, 1)

		w := doJSON(r, http.MethodDelete, "/api/v1/routines/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		svc := &mockService{}
		svc.On("DeleteRoutine", uint(5), uint(1)).Return(ErrRoutineNotFound)

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodDelete, "/api/v1/routines/5", nil)

		assert.Equal(t, http.StatusNotFound, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc := &mockService{}
		svc.On("DeleteRoutine", uint(5), uint(1)).Return(errors.New("db error"))

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodDelete, "/api/v1/routines/5", nil)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockService{}
		svc.On("DeleteRoutine", uint(5), uint(1)).Return(nil)

		r := newTestRouter(svc, 1)
		w := doJSON(r, http.MethodDelete, "/api/v1/routines/5", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}
