package stats

import "time"

type Service interface {
	GetCareerStats(userID uint) (CareerStatsResponse, error)
	GetPersonalRecords(userID uint) ([]PersonalRecordResponse, error)
	GetVolumeOverTime(userID uint, period string) ([]VolumePointResponse, error)
	GetStreak(userID uint) (StreakResponse, error)
	GetMuscleBreakdown(userID uint) ([]MuscleVolumeResponse, error)
}

type Repo interface {
	CareerTotals(userID uint) (CareerTotals, error)
	PersonalRecords(userID uint) ([]PersonalRecordResponse, error)
	VolumeOverTime(userID uint, trunc string) ([]VolumePointResponse, error)
	WorkoutDates(userID uint) ([]time.Time, error)
	MuscleBreakdown(userID uint) ([]MuscleVolumeResponse, error)
}
