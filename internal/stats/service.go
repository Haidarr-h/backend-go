package stats

import (
	"math"
	"time"
)

type StatsService struct {
	repo Repo
}

func NewStatsService(repo Repo) *StatsService {
	return &StatsService{repo: repo}
}

// xpPerKg is how much XP each kilogram of lifted volume is worth. Kept as a
// single constant so other XP sources can be layered in later without touching
// the level curve.
const xpPerKg = 0.3

// GetCareerStats returns lifetime totals plus the derived level/XP/progress.
func (s *StatsService) GetCareerStats(userID uint) (CareerStatsResponse, error) {
	totals, err := s.repo.CareerTotals(userID)
	if err != nil {
		return CareerStatsResponse{}, err
	}

	level, xp, intoLevel, forNext, progress := computeLevel(totals.TotalVolumeKg)

	return CareerStatsResponse{
		TotalVolumeKg:  totals.TotalVolumeKg,
		TotalReps:      totals.TotalReps,
		TotalSets:      totals.TotalSets,
		TotalSessions:  totals.TotalSessions,
		Level:          level,
		XP:             xp,
		XPIntoLevel:    intoLevel,
		XPForNextLevel: forNext,
		Progress:       progress,
	}, nil
}

// GetPersonalRecords returns per-exercise PRs (never nil).
func (s *StatsService) GetPersonalRecords(userID uint) ([]PersonalRecordResponse, error) {
	records, err := s.repo.PersonalRecords(userID)
	if err != nil {
		return nil, err
	}
	if records == nil {
		records = []PersonalRecordResponse{}
	}
	return records, nil
}

// GetVolumeOverTime returns time-bucketed volume. period defaults to "week".
func (s *StatsService) GetVolumeOverTime(userID uint, period string) ([]VolumePointResponse, error) {
	trunc, err := normalizePeriod(period)
	if err != nil {
		return nil, err
	}

	points, err := s.repo.VolumeOverTime(userID, trunc)
	if err != nil {
		return nil, err
	}
	if points == nil {
		points = []VolumePointResponse{}
	}
	return points, nil
}

// GetStreak computes the rest-day-tolerant streak from the user's workout dates.
func (s *StatsService) GetStreak(userID uint) (StreakResponse, error) {
	dates, err := s.repo.WorkoutDates(userID)
	if err != nil {
		return StreakResponse{}, err
	}

	today := truncateToDay(time.Now().UTC())
	return computeStreak(normalizeDates(dates), today), nil
}

// GetMuscleBreakdown returns volume per primary muscle (never nil).
func (s *StatsService) GetMuscleBreakdown(userID uint) ([]MuscleVolumeResponse, error) {
	muscles, err := s.repo.MuscleBreakdown(userID)
	if err != nil {
		return nil, err
	}
	if muscles == nil {
		muscles = []MuscleVolumeResponse{}
	}
	return muscles, nil
}

// ---------------------------------------------------------------------------
// Pure helpers (unit-tested)
// ---------------------------------------------------------------------------

// computeLevel derives the progressive RPG-style level from lifted volume.
//
//	xp    = volumeKg * 0.3
//	level = floor(sqrt(xp / 100))   -> XP to reach level L is 100*L^2
func computeLevel(volumeKg float64) (level int, xp, intoLevel, forNext, progress float64) {
	xp = volumeKg * xpPerKg
	if xp < 0 {
		xp = 0
	}

	level = int(math.Floor(math.Sqrt(xp / 100)))

	currBase := 100 * float64(level*level)
	nextBase := 100 * float64((level+1)*(level+1))

	intoLevel = xp - currBase
	forNext = nextBase - currBase
	if forNext > 0 {
		progress = intoLevel / forNext
	}

	return level, xp, intoLevel, forNext, progress
}

// normalizePeriod maps the request param to a Postgres date_trunc unit.
func normalizePeriod(period string) (string, error) {
	switch period {
	case "", "week":
		return "week", nil
	case "month":
		return "month", nil
	default:
		return "", ErrInvalidPeriod
	}
}

// computeStreak applies the rest-day-tolerant rule: up to 2 rest days between
// workouts keep a chain alive; a 3rd rest day (4+ day gap) resets it.
//
// dates must be ascending, distinct and day-truncated; today is day-truncated.
func computeStreak(dates []time.Time, today time.Time) StreakResponse {
	if len(dates) == 0 {
		return StreakResponse{}
	}

	// longest chain across all history
	longest, run := 1, 1
	for i := 1; i < len(dates); i++ {
		if dayDiff(dates[i-1], dates[i]) <= 3 {
			run++
		} else {
			run = 1
		}
		if run > longest {
			longest = run
		}
	}

	// current chain: walk back from the most recent workout while it stays linked
	currentRun := 1
	for i := len(dates) - 1; i > 0; i-- {
		if dayDiff(dates[i-1], dates[i]) <= 3 {
			currentRun++
		} else {
			break
		}
	}

	last := dates[len(dates)-1]
	resp := StreakResponse{
		LongestDays: longest,
		LastWorkout: last.Format("2006-01-02"),
	}

	// the chain is only "current" if the last workout is still within the budget
	if diff := dayDiff(last, today); diff <= 3 {
		resp.IsActive = true
		resp.CurrentDays = currentRun

		restElapsed := diff - 1 // full rest days that have already passed
		if restElapsed < 0 {
			restElapsed = 0
		}
		if daysLeft := 2 - restElapsed; daysLeft > 0 {
			resp.DaysUntilReset = daysLeft
		}
	}

	return resp
}

// dayDiff returns the number of whole calendar days from a to b (both day-truncated).
func dayDiff(a, b time.Time) int {
	return int(math.Round(b.Sub(a).Hours() / 24))
}

// truncateToDay drops the time-of-day, normalizing to UTC midnight.
func truncateToDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// normalizeDates day-truncates every date so dayDiff is exact regardless of the
// driver's returned time-of-day or zone.
func normalizeDates(dates []time.Time) []time.Time {
	out := make([]time.Time, len(dates))
	for i, d := range dates {
		out[i] = truncateToDay(d)
	}
	return out
}
