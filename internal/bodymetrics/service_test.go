package bodymetrics

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Mock Repo
// ---------------------------------------------------------------------------

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Upsert(metric BodyMetric) (BodyMetric, error) {
	args := m.Called(metric)
	return args.Get(0).(BodyMetric), args.Error(1)
}
func (m *mockRepo) FindByUserID(userID uint) (BodyMetric, error) {
	args := m.Called(userID)
	return args.Get(0).(BodyMetric), args.Error(1)
}

// ---------------------------------------------------------------------------
// computeBMI
// ---------------------------------------------------------------------------

func TestComputeBMI(t *testing.T) {
	assert.Equal(t, 23.7, computeBMI(178, 75)) // 75 / 1.78^2 = 23.67 -> 23.7
	assert.Equal(t, 25.0, computeBMI(200, 100))
	assert.Equal(t, 0.0, computeBMI(0, 75))  // guard against divide-by-zero
	assert.Equal(t, 0.0, computeBMI(-1, 75)) // guard against negative height
}

// ---------------------------------------------------------------------------
// UpsertBodyMetric
// ---------------------------------------------------------------------------

func TestUpsertBodyMetric(t *testing.T) {
	t.Run("saves and returns derived BMI", func(t *testing.T) {
		repo := &mockRepo{}
		repo.On("Upsert", mock.MatchedBy(func(m BodyMetric) bool {
			return m.UserID == 1 && m.HeightCM == 178 && m.WeightKG == 75 && m.Gender == "male"
		})).Return(BodyMetric{
			Model: gorm.Model{ID: 9}, UserID: 1, HeightCM: 178, WeightKG: 75,
			Age: 25, Gender: "male", BodyFat: 15, Goal: "build_muscle",
		}, nil)

		svc := NewBodyMetricService(repo)
		res, err := svc.UpsertBodyMetric(1, UpsertBodyMetricRequest{
			HeightCM: 178, WeightKG: 75, Age: 25, Gender: "male", BodyFat: 15, Goal: "build_muscle",
		})

		assert.NoError(t, err)
		assert.Equal(t, 23.7, res.BMI)
		assert.Equal(t, "build_muscle", res.Goal)
		repo.AssertExpectations(t)
	})

	t.Run("repo error propagates", func(t *testing.T) {
		repo := &mockRepo{}
		dbErr := errors.New("db down")
		repo.On("Upsert", mock.Anything).Return(BodyMetric{}, dbErr)

		svc := NewBodyMetricService(repo)
		_, err := svc.UpsertBodyMetric(1, UpsertBodyMetricRequest{HeightCM: 178, WeightKG: 75})

		assert.ErrorIs(t, err, dbErr)
		repo.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// GetBodyMetric
// ---------------------------------------------------------------------------

func TestGetBodyMetric(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockRepo{}
		repo.On("FindByUserID", uint(1)).Return(BodyMetric{
			UserID: 1, HeightCM: 178, WeightKG: 75, Gender: "male",
		}, nil)

		svc := NewBodyMetricService(repo)
		res, err := svc.GetBodyMetric(1)

		assert.NoError(t, err)
		assert.Equal(t, 23.7, res.BMI)
		repo.AssertExpectations(t)
	})

	t.Run("not found maps to ErrBodyMetricNotFound", func(t *testing.T) {
		repo := &mockRepo{}
		repo.On("FindByUserID", uint(1)).Return(BodyMetric{}, gorm.ErrRecordNotFound)

		svc := NewBodyMetricService(repo)
		_, err := svc.GetBodyMetric(1)

		assert.ErrorIs(t, err, ErrBodyMetricNotFound)
		repo.AssertExpectations(t)
	})
}
