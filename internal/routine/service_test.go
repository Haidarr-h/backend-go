package routine

import (
	"errors"
	"testing"

	"github.com/Haidarr-h/backend-go/internal/exercise"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockRoutineRepo struct{ mock.Mock }

func (m *mockRoutineRepo) Create(routine Routine) (Routine, error) {
	args := m.Called(routine)
	return args.Get(0).(Routine), args.Error(1)
}
func (m *mockRoutineRepo) FindAll(userID uint) ([]Routine, error) {
	args := m.Called(userID)
	return args.Get(0).([]Routine), args.Error(1)
}
func (m *mockRoutineRepo) FindByID(id uint) (Routine, error) {
	args := m.Called(id)
	return args.Get(0).(Routine), args.Error(1)
}
func (m *mockRoutineRepo) Update(routine Routine, replaceExercises bool) (Routine, error) {
	args := m.Called(routine, replaceExercises)
	return args.Get(0).(Routine), args.Error(1)
}
func (m *mockRoutineRepo) Delete(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

type mockExerciseRepo struct{ mock.Mock }

func (m *mockExerciseRepo) FindAll(params exercise.ExerciseFilterParams) ([]exercise.Exercise, error) {
	args := m.Called(params)
	return args.Get(0).([]exercise.Exercise), args.Error(1)
}
func (m *mockExerciseRepo) FindByID(id uint) (exercise.Exercise, error) {
	args := m.Called(id)
	return args.Get(0).(exercise.Exercise), args.Error(1)
}
func (m *mockExerciseRepo) ExistsByIDs(ids []uint) ([]uint, error) {
	args := m.Called(ids)
	return args.Get(0).([]uint), args.Error(1)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func ptr[T any](v T) *T { return &v }

// sampleCreateReq is a valid create request referencing exercise id 12.
func sampleCreateReq() CreateRoutineRequest {
	return CreateRoutineRequest{
		Name:        "Push Day",
		Description: "Chest and triceps",
		IsPublic:    false,
		RoutineExercises: []RoutineExercisesRequest{
			{
				ExerciseID: 12,
				Order:      1,
				RestSecond: 90,
				Sets: []RoutineExerciseSetRequest{
					{SetNumber: 1, Reps: 10, WeightKG: 40},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// CreateRoutine
// ---------------------------------------------------------------------------

func TestCreateRoutine(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		er := &mockExerciseRepo{}

		er.On("ExistsByIDs", []uint{12}).Return([]uint{}, nil)
		rr.On("Create", mock.AnythingOfType("routine.Routine")).
			Return(Routine{Model: gorm.Model{ID: 1}, Name: "Push Day"}, nil)

		svc := NewRoutineService(rr, er)
		res, err := svc.CreateRoutine(1, sampleCreateReq())

		assert.NoError(t, err)
		assert.Equal(t, uint(1), res.ID)
		assert.Equal(t, "Push Day", res.Name)
		er.AssertExpectations(t)
		rr.AssertExpectations(t)
	})

	t.Run("exercise does not exist", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		er := &mockExerciseRepo{}

		er.On("ExistsByIDs", []uint{12}).Return([]uint{12}, nil)

		svc := NewRoutineService(rr, er)
		_, err := svc.CreateRoutine(1, sampleCreateReq())

		assert.ErrorIs(t, err, ErrExerciseNotFound)
		er.AssertExpectations(t)
		// repo.Create must not be called when validation fails
		rr.AssertNotCalled(t, "Create", mock.Anything)
	})

	t.Run("exercise repo error propagates", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		er := &mockExerciseRepo{}
		dbErr := errors.New("db error")

		er.On("ExistsByIDs", []uint{12}).Return([]uint{}, dbErr)

		svc := NewRoutineService(rr, er)
		_, err := svc.CreateRoutine(1, sampleCreateReq())

		assert.ErrorIs(t, err, dbErr)
		er.AssertExpectations(t)
	})

	t.Run("routine repo error propagates", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		er := &mockExerciseRepo{}
		dbErr := errors.New("insert failed")

		er.On("ExistsByIDs", []uint{12}).Return([]uint{}, nil)
		rr.On("Create", mock.AnythingOfType("routine.Routine")).Return(Routine{}, dbErr)

		svc := NewRoutineService(rr, er)
		_, err := svc.CreateRoutine(1, sampleCreateReq())

		assert.ErrorIs(t, err, dbErr)
		er.AssertExpectations(t)
		rr.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// GetRoutines
// ---------------------------------------------------------------------------

func TestGetRoutines(t *testing.T) {
	t.Run("success returns mapped list", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		rr.On("FindAll", uint(1)).Return([]Routine{
			{Model: gorm.Model{ID: 1}, Name: "A"},
			{Model: gorm.Model{ID: 2}, Name: "B"},
		}, nil)

		svc := NewRoutineService(rr, &mockExerciseRepo{})
		res, err := svc.GetRoutines(1)

		assert.NoError(t, err)
		assert.Len(t, res, 2)
		rr.AssertExpectations(t)
	})

	t.Run("no routines returns non-nil empty slice", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		rr.On("FindAll", uint(1)).Return([]Routine{}, nil)

		svc := NewRoutineService(rr, &mockExerciseRepo{})
		res, err := svc.GetRoutines(1)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Empty(t, res)
		rr.AssertExpectations(t)
	})

	t.Run("repo error propagates", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		dbErr := errors.New("db error")
		rr.On("FindAll", uint(1)).Return([]Routine{}, dbErr)

		svc := NewRoutineService(rr, &mockExerciseRepo{})
		_, err := svc.GetRoutines(1)

		assert.ErrorIs(t, err, dbErr)
		rr.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// GetRoutine
// ---------------------------------------------------------------------------

func TestGetRoutine(t *testing.T) {
	t.Run("public routine accessible by anyone", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		owner := uint(99)
		rr.On("FindByID", uint(5)).Return(Routine{
			Model: gorm.Model{ID: 5}, UserID: &owner, IsPublic: true,
		}, nil)

		svc := NewRoutineService(rr, &mockExerciseRepo{})
		res, err := svc.GetRoutine(1, 5)

		assert.NoError(t, err)
		assert.Equal(t, uint(5), res.ID)
		rr.AssertExpectations(t)
	})

	t.Run("private routine accessible by owner", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		owner := uint(1)
		rr.On("FindByID", uint(5)).Return(Routine{
			Model: gorm.Model{ID: 5}, UserID: &owner, IsPublic: false,
		}, nil)

		svc := NewRoutineService(rr, &mockExerciseRepo{})
		res, err := svc.GetRoutine(1, 5)

		assert.NoError(t, err)
		assert.Equal(t, uint(5), res.ID)
		rr.AssertExpectations(t)
	})

	t.Run("private routine rejected for non-owner", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		owner := uint(99)
		rr.On("FindByID", uint(5)).Return(Routine{
			Model: gorm.Model{ID: 5}, UserID: &owner, IsPublic: false,
		}, nil)

		svc := NewRoutineService(rr, &mockExerciseRepo{})
		_, err := svc.GetRoutine(1, 5)

		assert.ErrorIs(t, err, InvalidRoutineOwnership)
		rr.AssertExpectations(t)
	})

	t.Run("not found maps to ErrRoutineNotFound", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		rr.On("FindByID", uint(5)).Return(Routine{}, gorm.ErrRecordNotFound)

		svc := NewRoutineService(rr, &mockExerciseRepo{})
		_, err := svc.GetRoutine(1, 5)

		assert.ErrorIs(t, err, ErrRoutineNotFound)
		rr.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// UpdateRoutine
// ---------------------------------------------------------------------------

func TestUpdateRoutine(t *testing.T) {
	existingOwned := func() Routine {
		owner := uint(1)
		return Routine{Model: gorm.Model{ID: 5}, UserID: &owner, Name: "Old"}
	}

	t.Run("not found", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		rr.On("FindByID", uint(5)).Return(Routine{}, gorm.ErrRecordNotFound)

		svc := NewRoutineService(rr, &mockExerciseRepo{})
		_, err := svc.UpdateRoutine(5, 1, UpdateRoutineReq{Name: ptr("New")})

		assert.ErrorIs(t, err, ErrRoutineNotFound)
		rr.AssertExpectations(t)
	})

	t.Run("not owner", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		owner := uint(99)
		rr.On("FindByID", uint(5)).Return(Routine{
			Model: gorm.Model{ID: 5}, UserID: &owner,
		}, nil)

		svc := NewRoutineService(rr, &mockExerciseRepo{})
		_, err := svc.UpdateRoutine(5, 1, UpdateRoutineReq{Name: ptr("New")})

		assert.ErrorIs(t, err, InvalidRoutineOwnership)
		rr.AssertExpectations(t)
		rr.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("scalar-only update keeps exercises (replace=false, no validation)", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		er := &mockExerciseRepo{}
		rr.On("FindByID", uint(5)).Return(existingOwned(), nil)
		rr.On("Update", mock.AnythingOfType("routine.Routine"), false).
			Return(Routine{Model: gorm.Model{ID: 5}, Name: "New"}, nil)

		svc := NewRoutineService(rr, er)
		res, err := svc.UpdateRoutine(5, 1, UpdateRoutineReq{Name: ptr("New")})

		assert.NoError(t, err)
		assert.Equal(t, "New", res.Name)
		// omitted exercises => exercise existence is never checked
		er.AssertNotCalled(t, "ExistsByIDs", mock.Anything)
		rr.AssertExpectations(t)
	})

	t.Run("providing exercises validates and replaces (replace=true)", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		er := &mockExerciseRepo{}
		rr.On("FindByID", uint(5)).Return(existingOwned(), nil)
		er.On("ExistsByIDs", []uint{12}).Return([]uint{}, nil)
		rr.On("Update", mock.AnythingOfType("routine.Routine"), true).
			Return(Routine{Model: gorm.Model{ID: 5}, Name: "Old"}, nil)

		svc := NewRoutineService(rr, er)
		_, err := svc.UpdateRoutine(5, 1, UpdateRoutineReq{
			RoutineExercises: []RoutineExercisesRequest{
				{ExerciseID: 12, Order: 1},
			},
		})

		assert.NoError(t, err)
		er.AssertExpectations(t)
		rr.AssertExpectations(t)
	})

	t.Run("providing unknown exercise fails before repo update", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		er := &mockExerciseRepo{}
		rr.On("FindByID", uint(5)).Return(existingOwned(), nil)
		er.On("ExistsByIDs", []uint{77}).Return([]uint{77}, nil)

		svc := NewRoutineService(rr, er)
		_, err := svc.UpdateRoutine(5, 1, UpdateRoutineReq{
			RoutineExercises: []RoutineExercisesRequest{
				{ExerciseID: 77, Order: 1},
			},
		})

		assert.ErrorIs(t, err, ErrExerciseNotFound)
		er.AssertExpectations(t)
		rr.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("explicit empty list clears exercises (replace=true)", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		er := &mockExerciseRepo{}
		rr.On("FindByID", uint(5)).Return(existingOwned(), nil)
		rr.On("Update", mock.AnythingOfType("routine.Routine"), true).
			Return(Routine{Model: gorm.Model{ID: 5}}, nil)

		svc := NewRoutineService(rr, er)
		_, err := svc.UpdateRoutine(5, 1, UpdateRoutineReq{
			RoutineExercises: []RoutineExercisesRequest{},
		})

		assert.NoError(t, err)
		// empty (but non-nil) list => no ids to validate, repo still replaces
		er.AssertNotCalled(t, "ExistsByIDs", mock.Anything)
		rr.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// DeleteRoutine
// ---------------------------------------------------------------------------

func TestDeleteRoutine(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		rr.On("Delete", uint(5), uint(1)).Return(nil)

		svc := NewRoutineService(rr, &mockExerciseRepo{})
		err := svc.DeleteRoutine(5, 1)

		assert.NoError(t, err)
		rr.AssertExpectations(t)
	})

	t.Run("not found maps to ErrRoutineNotFound", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		rr.On("Delete", uint(5), uint(1)).Return(gorm.ErrRecordNotFound)

		svc := NewRoutineService(rr, &mockExerciseRepo{})
		err := svc.DeleteRoutine(5, 1)

		assert.ErrorIs(t, err, ErrRoutineNotFound)
		rr.AssertExpectations(t)
	})

	t.Run("other error propagates", func(t *testing.T) {
		rr := &mockRoutineRepo{}
		dbErr := errors.New("db error")
		rr.On("Delete", uint(5), uint(1)).Return(dbErr)

		svc := NewRoutineService(rr, &mockExerciseRepo{})
		err := svc.DeleteRoutine(5, 1)

		assert.ErrorIs(t, err, dbErr)
		rr.AssertExpectations(t)
	})
}
