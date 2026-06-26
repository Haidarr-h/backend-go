package stats

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// Mock Repo
// ---------------------------------------------------------------------------

type mockRepo struct{ mock.Mock }

func (m *mockRepo) CareerTotals(userID uint) (CareerTotals, error) {
	args := m.Called(userID)
	return args.Get(0).(CareerTotals), args.Error(1)
}
func (m *mockRepo) PersonalRecords(userID uint) ([]PersonalRecordResponse, error) {
	args := m.Called(userID)
	return args.Get(0).([]PersonalRecordResponse), args.Error(1)
}
func (m *mockRepo) VolumeOverTime(userID uint, trunc string) ([]VolumePointResponse, error) {
	args := m.Called(userID, trunc)
	return args.Get(0).([]VolumePointResponse), args.Error(1)
}
func (m *mockRepo) WorkoutDates(userID uint) ([]time.Time, error) {
	args := m.Called(userID)
	return args.Get(0).([]time.Time), args.Error(1)
}
func (m *mockRepo) MuscleBreakdown(userID uint) ([]MuscleVolumeResponse, error) {
	args := m.Called(userID)
	return args.Get(0).([]MuscleVolumeResponse), args.Error(1)
}

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// ---------------------------------------------------------------------------
// computeLevel
// ---------------------------------------------------------------------------

func TestComputeLevel(t *testing.T) {
	t.Run("zero volume is level 0 with no progress", func(t *testing.T) {
		level, xp, into, forNext, progress := computeLevel(0)
		assert.Equal(t, 0, level)
		assert.Equal(t, 0.0, xp)
		assert.Equal(t, 0.0, into)
		assert.Equal(t, 100.0, forNext) // 100*1^2 - 100*0^2
		assert.Equal(t, 0.0, progress)
	})

	t.Run("xp 12750 -> level 11", func(t *testing.T) {
		// volume * 0.3 = xp  =>  volume = 42500 -> xp 12750
		level, xp, _, _, _ := computeLevel(42500)
		assert.Equal(t, 12750.0, xp)
		assert.Equal(t, 11, level) // floor(sqrt(12750/100)) = floor(11.29) = 11
	})

	t.Run("exact level boundary sits at start of the level", func(t *testing.T) {
		// xp = 100*5^2 = 2500 -> volume = 2500/0.3
		level, _, into, _, progress := computeLevel(2500 / xpPerKg)
		assert.Equal(t, 5, level)
		assert.InDelta(t, 0.0, into, 1e-6)
		assert.InDelta(t, 0.0, progress, 1e-6)
	})

	t.Run("negative volume clamps xp to 0", func(t *testing.T) {
		level, xp, _, _, _ := computeLevel(-100)
		assert.Equal(t, 0, level)
		assert.Equal(t, 0.0, xp)
	})
}

// ---------------------------------------------------------------------------
// normalizePeriod
// ---------------------------------------------------------------------------

func TestNormalizePeriod(t *testing.T) {
	cases := map[string]struct {
		want string
		err  error
	}{
		"":      {"week", nil},
		"week":  {"week", nil},
		"month": {"month", nil},
		"day":   {"", ErrInvalidPeriod},
		"year":  {"", ErrInvalidPeriod},
	}
	for in, exp := range cases {
		got, err := normalizePeriod(in)
		assert.Equal(t, exp.want, got, "period %q", in)
		assert.ErrorIs(t, err, exp.err, "period %q", in)
	}
}

// ---------------------------------------------------------------------------
// computeStreak (rest-day tolerant: up to 2 rest days keeps the chain alive)
// ---------------------------------------------------------------------------

func TestComputeStreak(t *testing.T) {
	t.Run("no workouts", func(t *testing.T) {
		s := computeStreak(nil, day(2026, 6, 26))
		assert.Equal(t, StreakResponse{}, s)
	})

	t.Run("single workout today", func(t *testing.T) {
		s := computeStreak([]time.Time{day(2026, 6, 26)}, day(2026, 6, 26))
		assert.Equal(t, 1, s.CurrentDays)
		assert.Equal(t, 1, s.LongestDays)
		assert.True(t, s.IsActive)
		assert.Equal(t, "2026-06-26", s.LastWorkout)
		assert.Equal(t, 2, s.DaysUntilReset)
	})

	t.Run("2 rest days between workouts continues the chain", func(t *testing.T) {
		// 20th, then 23rd (gap of 3 days = 2 rest days) -> still one chain
		dates := []time.Time{day(2026, 6, 20), day(2026, 6, 23)}
		s := computeStreak(dates, day(2026, 6, 23))
		assert.Equal(t, 2, s.CurrentDays)
		assert.Equal(t, 2, s.LongestDays)
		assert.True(t, s.IsActive)
	})

	t.Run("3 rest days resets the chain", func(t *testing.T) {
		// 20th, then 24th (gap of 4 days = 3 rest days) -> chain breaks
		dates := []time.Time{day(2026, 6, 20), day(2026, 6, 24)}
		s := computeStreak(dates, day(2026, 6, 24))
		assert.Equal(t, 1, s.CurrentDays) // only the 24th
		assert.Equal(t, 1, s.LongestDays)
		assert.True(t, s.IsActive)
	})

	t.Run("stale streak reports 0 current but keeps longest", func(t *testing.T) {
		// last workout was 4 days before today -> current chain is dead
		dates := []time.Time{day(2026, 6, 18), day(2026, 6, 19), day(2026, 6, 20)}
		s := computeStreak(dates, day(2026, 6, 24))
		assert.Equal(t, 0, s.CurrentDays)
		assert.False(t, s.IsActive)
		assert.Equal(t, 3, s.LongestDays)
		assert.Equal(t, "2026-06-20", s.LastWorkout)
		assert.Equal(t, 0, s.DaysUntilReset)
	})

	t.Run("daysUntilReset counts down as rest days pass", func(t *testing.T) {
		last := day(2026, 6, 20)
		assert.Equal(t, 2, computeStreak([]time.Time{last}, day(2026, 6, 20)).DaysUntilReset) // same day
		assert.Equal(t, 2, computeStreak([]time.Time{last}, day(2026, 6, 21)).DaysUntilReset) // 1 day later
		assert.Equal(t, 1, computeStreak([]time.Time{last}, day(2026, 6, 22)).DaysUntilReset) // 1 rest day passed
		assert.Equal(t, 0, computeStreak([]time.Time{last}, day(2026, 6, 23)).DaysUntilReset) // 2 rest days passed
	})

	t.Run("longest reflects the best historical chain not the current one", func(t *testing.T) {
		// chain A: 1,2,3,4 (4) ; big gap ; chain B (current): 20,21 (2)
		dates := []time.Time{
			day(2026, 6, 1), day(2026, 6, 2), day(2026, 6, 3), day(2026, 6, 4),
			day(2026, 6, 20), day(2026, 6, 21),
		}
		s := computeStreak(dates, day(2026, 6, 21))
		assert.Equal(t, 2, s.CurrentDays)
		assert.Equal(t, 4, s.LongestDays)
	})
}

// ---------------------------------------------------------------------------
// Service methods
// ---------------------------------------------------------------------------

func TestGetCareerStats(t *testing.T) {
	t.Run("maps totals and derives level", func(t *testing.T) {
		repo := &mockRepo{}
		repo.On("CareerTotals", uint(1)).Return(CareerTotals{
			TotalVolumeKg: 42500, TotalReps: 9600, TotalSets: 1200, TotalSessions: 100,
		}, nil)

		svc := NewStatsService(repo)
		res, err := svc.GetCareerStats(1)

		assert.NoError(t, err)
		assert.Equal(t, 42500.0, res.TotalVolumeKg)
		assert.Equal(t, 100, res.TotalSessions)
		assert.Equal(t, 12750.0, res.XP)
		assert.Equal(t, 11, res.Level)
		repo.AssertExpectations(t)
	})

	t.Run("repo error propagates", func(t *testing.T) {
		repo := &mockRepo{}
		dbErr := errors.New("db down")
		repo.On("CareerTotals", uint(1)).Return(CareerTotals{}, dbErr)

		svc := NewStatsService(repo)
		_, err := svc.GetCareerStats(1)

		assert.ErrorIs(t, err, dbErr)
		repo.AssertExpectations(t)
	})
}

func TestGetVolumeOverTime(t *testing.T) {
	t.Run("empty period defaults to week and returns non-nil", func(t *testing.T) {
		repo := &mockRepo{}
		repo.On("VolumeOverTime", uint(1), "week").Return([]VolumePointResponse(nil), nil)

		svc := NewStatsService(repo)
		res, err := svc.GetVolumeOverTime(1, "")

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Empty(t, res)
		repo.AssertExpectations(t)
	})

	t.Run("invalid period rejected before hitting repo", func(t *testing.T) {
		repo := &mockRepo{}
		svc := NewStatsService(repo)

		_, err := svc.GetVolumeOverTime(1, "decade")

		assert.ErrorIs(t, err, ErrInvalidPeriod)
		repo.AssertNotCalled(t, "VolumeOverTime", mock.Anything, mock.Anything)
	})
}

func TestGetStreakService(t *testing.T) {
	repo := &mockRepo{}
	repo.On("WorkoutDates", uint(1)).Return([]time.Time{day(2026, 6, 1)}, nil)

	svc := NewStatsService(repo)
	res, err := svc.GetStreak(1)

	assert.NoError(t, err)
	assert.Equal(t, "2026-06-01", res.LastWorkout)
	repo.AssertExpectations(t)
}

func TestSliceGettersNeverNil(t *testing.T) {
	repo := &mockRepo{}
	repo.On("PersonalRecords", uint(1)).Return([]PersonalRecordResponse(nil), nil)
	repo.On("MuscleBreakdown", uint(1)).Return([]MuscleVolumeResponse(nil), nil)

	svc := NewStatsService(repo)

	prs, err := svc.GetPersonalRecords(1)
	assert.NoError(t, err)
	assert.NotNil(t, prs)

	muscles, err := svc.GetMuscleBreakdown(1)
	assert.NoError(t, err)
	assert.NotNil(t, muscles)

	repo.AssertExpectations(t)
}
