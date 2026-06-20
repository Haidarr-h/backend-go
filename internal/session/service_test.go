package session

import (
	"errors"
	"testing"
	"time"

	"github.com/Haidarr-h/backend-go/internal/exercise"
	"github.com/Haidarr-h/backend-go/internal/routine"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockSessionRepo struct{ mock.Mock }

func (m *mockSessionRepo) Create(s Session) (Session, error) {
	args := m.Called(s)
	return args.Get(0).(Session), args.Error(1)
}
func (m *mockSessionRepo) FindAll(userID uint) ([]Session, error) {
	args := m.Called(userID)
	return args.Get(0).([]Session), args.Error(1)
}
func (m *mockSessionRepo) FindByID(id uint) (Session, error) {
	args := m.Called(id)
	return args.Get(0).(Session), args.Error(1)
}
func (m *mockSessionRepo) Update(s Session) (Session, error) {
	args := m.Called(s)
	return args.Get(0).(Session), args.Error(1)
}
func (m *mockSessionRepo) Delete(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

type mockRoutineRepo struct{ mock.Mock }

func (m *mockRoutineRepo) Create(r routine.Routine) (routine.Routine, error) {
	args := m.Called(r)
	return args.Get(0).(routine.Routine), args.Error(1)
}
func (m *mockRoutineRepo) FindAll(userID uint) ([]routine.Routine, error) {
	args := m.Called(userID)
	return args.Get(0).([]routine.Routine), args.Error(1)
}
func (m *mockRoutineRepo) FindAllPublic() ([]routine.Routine, error) {
	args := m.Called()
	return args.Get(0).([]routine.Routine), args.Error(1)
}
func (m *mockRoutineRepo) FindByID(id uint) (routine.Routine, error) {
	args := m.Called(id)
	return args.Get(0).(routine.Routine), args.Error(1)
}
func (m *mockRoutineRepo) Update(r routine.Routine, replaceExercises bool) (routine.Routine, error) {
	args := m.Called(r, replaceExercises)
	return args.Get(0).(routine.Routine), args.Error(1)
}
func (m *mockRoutineRepo) Delete(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func ptr[T any](v T) *T { return &v }

var (
	startTime = time.Date(2026, 6, 11, 8, 0, 0, 0, time.UTC)
	endTime   = time.Date(2026, 6, 11, 9, 0, 0, 0, time.UTC) // +1h = 3600s
)

// sampleRoutine is an owned routine with one exercise (chest/triceps) and 2 sets.
func sampleRoutine(owner uint) routine.Routine {
	return routine.Routine{
		Model:    gorm.Model{ID: 1},
		UserID:   &owner,
		Name:     "Push Day",
		IsPublic: false,
		RoutineExercises: []routine.RoutineExercise{
			{
				Model:      gorm.Model{ID: 10},
				ExerciseID: 12,
				Order:      1,
				RestSecond: 90,
				Exercise: exercise.Exercise{
					Model:          gorm.Model{ID: 12},
					Name:           "Bench Press",
					PrimaryMuscles: pq.StringArray{"chest", "triceps"},
				},
				RoutineExerciseSets: []routine.RoutineExerciseSet{
					{SetNumber: 1, Reps: 10, WeightKG: 40},
					{SetNumber: 2, Reps: 8, WeightKG: 45},
				},
			},
		},
	}
}

func validCreateReq() CreateSessionRequest {
	return CreateSessionRequest{
		RoutineID: 1,
		Name:      "Morning Push",
		StartTime: startTime,
		EndTime:   endTime,
	}
}

// ---------------------------------------------------------------------------
// CreateSession
// ---------------------------------------------------------------------------

func TestCreateSession(t *testing.T) {
	t.Run("success snapshots routine exercises", func(t *testing.T) {
		sr := &mockSessionRepo{}
		rr := &mockRoutineRepo{}
		rr.On("FindByID", uint(1)).Return(sampleRoutine(1), nil)

		// capture what gets persisted to assert the snapshot
		sr.On("Create", mock.MatchedBy(func(s Session) bool {
			return s.UserID == 1 &&
				len(s.SessionExercises) == 1 &&
				s.SessionExercises[0].ExerciseID == 12 &&
				len(s.SessionExercises[0].SessionExerciseSets) == 2
		})).Return(Session{
			Model:            gorm.Model{ID: 5},
			UserID:           1,
			Name:             "Morning Push",
			StartTime:        startTime,
			EndTime:          endTime,
			SessionExercises: sampleSnapshotExercises(),
		}, nil)

		svc := NewSessionService(sr, rr)
		res, err := svc.CreateSession(1, validCreateReq())

		assert.NoError(t, err)
		assert.Equal(t, uint(5), res.ID)
		assert.Equal(t, 3600, res.Metrics.TotalDurationSeconds)
		assert.Equal(t, float64(10*40+8*45), res.Metrics.TotalWeight)
		rr.AssertExpectations(t)
		sr.AssertExpectations(t)
	})

	t.Run("routine not found", func(t *testing.T) {
		sr := &mockSessionRepo{}
		rr := &mockRoutineRepo{}
		rr.On("FindByID", uint(1)).Return(routine.Routine{}, gorm.ErrRecordNotFound)

		svc := NewSessionService(sr, rr)
		_, err := svc.CreateSession(1, validCreateReq())

		assert.ErrorIs(t, err, ErrRoutineNotFound)
		sr.AssertNotCalled(t, "Create", mock.Anything)
	})

	t.Run("routine not accessible (private, not owner)", func(t *testing.T) {
		sr := &mockSessionRepo{}
		rr := &mockRoutineRepo{}
		rr.On("FindByID", uint(1)).Return(sampleRoutine(99), nil) // owned by someone else

		svc := NewSessionService(sr, rr)
		_, err := svc.CreateSession(1, validCreateReq())

		assert.ErrorIs(t, err, ErrRoutineNotAccessible)
		sr.AssertNotCalled(t, "Create", mock.Anything)
	})

	t.Run("public routine is usable by anyone", func(t *testing.T) {
		sr := &mockSessionRepo{}
		rr := &mockRoutineRepo{}
		pub := sampleRoutine(99)
		pub.IsPublic = true
		rr.On("FindByID", uint(1)).Return(pub, nil)
		sr.On("Create", mock.AnythingOfType("session.Session")).Return(Session{Model: gorm.Model{ID: 5}, UserID: 1}, nil)

		svc := NewSessionService(sr, rr)
		_, err := svc.CreateSession(1, validCreateReq())

		assert.NoError(t, err)
		sr.AssertExpectations(t)
	})

	t.Run("invalid time range rejected before touching repos", func(t *testing.T) {
		sr := &mockSessionRepo{}
		rr := &mockRoutineRepo{}

		req := validCreateReq()
		req.EndTime = startTime.Add(-time.Hour)

		svc := NewSessionService(sr, rr)
		_, err := svc.CreateSession(1, req)

		assert.ErrorIs(t, err, ErrInvalidTimeRange)
		rr.AssertNotCalled(t, "FindByID", mock.Anything)
		sr.AssertNotCalled(t, "Create", mock.Anything)
	})
}

// ---------------------------------------------------------------------------
// GetSessions
// ---------------------------------------------------------------------------

func TestGetSessions(t *testing.T) {
	t.Run("success maps list", func(t *testing.T) {
		sr := &mockSessionRepo{}
		sr.On("FindAll", uint(1)).Return([]Session{
			{Model: gorm.Model{ID: 1}, UserID: 1},
			{Model: gorm.Model{ID: 2}, UserID: 1},
		}, nil)

		svc := NewSessionService(sr, &mockRoutineRepo{})
		res, err := svc.GetSessions(1)

		assert.NoError(t, err)
		assert.Len(t, res, 2)
		sr.AssertExpectations(t)
	})

	t.Run("empty returns non-nil slice", func(t *testing.T) {
		sr := &mockSessionRepo{}
		sr.On("FindAll", uint(1)).Return([]Session{}, nil)

		svc := NewSessionService(sr, &mockRoutineRepo{})
		res, err := svc.GetSessions(1)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Empty(t, res)
		sr.AssertExpectations(t)
	})

	t.Run("repo error propagates", func(t *testing.T) {
		sr := &mockSessionRepo{}
		dbErr := errors.New("db error")
		sr.On("FindAll", uint(1)).Return([]Session{}, dbErr)

		svc := NewSessionService(sr, &mockRoutineRepo{})
		_, err := svc.GetSessions(1)

		assert.ErrorIs(t, err, dbErr)
		sr.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// GetSession
// ---------------------------------------------------------------------------

func TestGetSession(t *testing.T) {
	t.Run("success with metrics", func(t *testing.T) {
		sr := &mockSessionRepo{}
		sr.On("FindByID", uint(5)).Return(Session{
			Model:            gorm.Model{ID: 5},
			UserID:           1,
			StartTime:        startTime,
			EndTime:          endTime,
			SessionExercises: sampleSnapshotExercises(),
		}, nil)

		svc := NewSessionService(sr, &mockRoutineRepo{})
		res, err := svc.GetSession(1, 5)

		assert.NoError(t, err)
		assert.Equal(t, float64(10*40+8*45), res.Metrics.TotalWeight)
		assert.Equal(t, 3600, res.Metrics.TotalDurationSeconds)
		assert.Equal(t, []string{"chest", "triceps"}, res.Metrics.TargetedMuscles)
		sr.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		sr := &mockSessionRepo{}
		sr.On("FindByID", uint(5)).Return(Session{}, gorm.ErrRecordNotFound)

		svc := NewSessionService(sr, &mockRoutineRepo{})
		_, err := svc.GetSession(1, 5)

		assert.ErrorIs(t, err, ErrSessionNotFound)
		sr.AssertExpectations(t)
	})

	t.Run("forbidden for non-owner", func(t *testing.T) {
		sr := &mockSessionRepo{}
		sr.On("FindByID", uint(5)).Return(Session{Model: gorm.Model{ID: 5}, UserID: 99}, nil)

		svc := NewSessionService(sr, &mockRoutineRepo{})
		_, err := svc.GetSession(1, 5)

		assert.ErrorIs(t, err, InvalidSessionOwnership)
		sr.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// UpdateSession
// ---------------------------------------------------------------------------

func TestUpdateSession(t *testing.T) {
	existingOwned := func() Session {
		return Session{Model: gorm.Model{ID: 5}, UserID: 1, Name: "Old", StartTime: startTime, EndTime: endTime}
	}

	t.Run("not found", func(t *testing.T) {
		sr := &mockSessionRepo{}
		sr.On("FindByID", uint(5)).Return(Session{}, gorm.ErrRecordNotFound)

		svc := NewSessionService(sr, &mockRoutineRepo{})
		_, err := svc.UpdateSession(5, 1, UpdateSessionReq{Name: ptr("New")})

		assert.ErrorIs(t, err, ErrSessionNotFound)
		sr.AssertExpectations(t)
	})

	t.Run("forbidden for non-owner", func(t *testing.T) {
		sr := &mockSessionRepo{}
		sr.On("FindByID", uint(5)).Return(Session{Model: gorm.Model{ID: 5}, UserID: 99}, nil)

		svc := NewSessionService(sr, &mockRoutineRepo{})
		_, err := svc.UpdateSession(5, 1, UpdateSessionReq{Name: ptr("New")})

		assert.ErrorIs(t, err, InvalidSessionOwnership)
		sr.AssertNotCalled(t, "Update", mock.Anything)
	})

	t.Run("invalid resulting time range", func(t *testing.T) {
		sr := &mockSessionRepo{}
		sr.On("FindByID", uint(5)).Return(existingOwned(), nil)

		svc := NewSessionService(sr, &mockRoutineRepo{})
		_, err := svc.UpdateSession(5, 1, UpdateSessionReq{EndTime: ptr(startTime.Add(-time.Hour))})

		assert.ErrorIs(t, err, ErrInvalidTimeRange)
		sr.AssertNotCalled(t, "Update", mock.Anything)
	})

	t.Run("scalar update success", func(t *testing.T) {
		sr := &mockSessionRepo{}
		sr.On("FindByID", uint(5)).Return(existingOwned(), nil)
		sr.On("Update", mock.MatchedBy(func(s Session) bool {
			return s.Name == "New" && s.Notes == "felt strong"
		})).Return(Session{Model: gorm.Model{ID: 5}, UserID: 1, Name: "New", Notes: "felt strong"}, nil)

		svc := NewSessionService(sr, &mockRoutineRepo{})
		res, err := svc.UpdateSession(5, 1, UpdateSessionReq{Name: ptr("New"), Notes: ptr("felt strong")})

		assert.NoError(t, err)
		assert.Equal(t, "New", res.Name)
		assert.Equal(t, "felt strong", res.Notes)
		sr.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// DeleteSession
// ---------------------------------------------------------------------------

func TestDeleteSession(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sr := &mockSessionRepo{}
		sr.On("Delete", uint(5), uint(1)).Return(nil)

		svc := NewSessionService(sr, &mockRoutineRepo{})
		err := svc.DeleteSession(5, 1)

		assert.NoError(t, err)
		sr.AssertExpectations(t)
	})

	t.Run("not found maps to ErrSessionNotFound", func(t *testing.T) {
		sr := &mockSessionRepo{}
		sr.On("Delete", uint(5), uint(1)).Return(gorm.ErrRecordNotFound)

		svc := NewSessionService(sr, &mockRoutineRepo{})
		err := svc.DeleteSession(5, 1)

		assert.ErrorIs(t, err, ErrSessionNotFound)
		sr.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// computeMetrics
// ---------------------------------------------------------------------------

func TestComputeMetrics(t *testing.T) {
	s := Session{
		StartTime: startTime,
		EndTime:   endTime,
		SessionExercises: []SessionExercise{
			{
				Exercise:            exercise.Exercise{PrimaryMuscles: pq.StringArray{"chest", "triceps"}},
				SessionExerciseSets: []SessionExerciseSet{{Reps: 10, WeightKG: 40}, {Reps: 8, WeightKG: 45}},
			},
			{
				Exercise:            exercise.Exercise{PrimaryMuscles: pq.StringArray{"chest"}},
				SessionExerciseSets: []SessionExerciseSet{{Reps: 12, WeightKG: 20}},
			},
		},
	}

	m := computeMetrics(s)

	assert.Equal(t, 3, m.TotalSets)
	assert.Equal(t, 30, m.TotalReps)
	assert.Equal(t, float64(10*40+8*45+12*20), m.TotalWeight)
	assert.Equal(t, 3600, m.TotalDurationSeconds)
	// chest appears twice, triceps once -> chest first, then triceps
	assert.Equal(t, []string{"chest", "triceps"}, m.TargetedMuscles)
}

func TestComputeMetricsNegativeDurationClamped(t *testing.T) {
	m := computeMetrics(Session{StartTime: endTime, EndTime: startTime})
	assert.Equal(t, 0, m.TotalDurationSeconds)
}

// sampleSnapshotExercises is the snapshot tree as it would come back from the repo.
func sampleSnapshotExercises() []SessionExercise {
	return []SessionExercise{
		{
			Model:      gorm.Model{ID: 100},
			ExerciseID: 12,
			Order:      1,
			RestSecond: 90,
			Exercise: exercise.Exercise{
				Model:          gorm.Model{ID: 12},
				Name:           "Bench Press",
				PrimaryMuscles: pq.StringArray{"chest", "triceps"},
			},
			SessionExerciseSets: []SessionExerciseSet{
				{Model: gorm.Model{ID: 1}, SetNumber: 1, Reps: 10, WeightKG: 40},
				{Model: gorm.Model{ID: 2}, SetNumber: 2, Reps: 8, WeightKG: 45},
			},
		},
	}
}
