package stats

import "time"

// CareerStatsResponse is the lifetime summary returned by GET /stats/career.
type CareerStatsResponse struct {
	TotalVolumeKg  float64 `json:"totalVolumeKg" example:"127500"`
	TotalReps      int     `json:"totalReps" example:"9600"`
	TotalSets      int     `json:"totalSets" example:"1200"`
	TotalSessions  int     `json:"totalSessions" example:"100"`
	Level          int     `json:"level" example:"35"`
	XP             float64 `json:"xp" example:"38250"`
	XPIntoLevel    float64 `json:"xpIntoLevel" example:"5750"`    // xp - 100*level^2
	XPForNextLevel float64 `json:"xpForNextLevel" example:"7100"` // 100*(level+1)^2 - 100*level^2
	Progress       float64 `json:"progress" example:"0.81"`       // 0..1 toward next level
}

// PersonalRecordResponse is one row of GET /stats/records: a user's best lifts per exercise.
type PersonalRecordResponse struct {
	ExerciseID      uint    `json:"exerciseId" example:"12"`
	ExerciseName    string  `json:"exerciseName" example:"Bench Press"`
	MaxWeightKg     float64 `json:"maxWeightKg" example:"100"`
	BestSetVolumeKg float64 `json:"bestSetVolumeKg" example:"800"`
}

// VolumePointResponse is one time bucket of GET /stats/volume.
type VolumePointResponse struct {
	Period   time.Time `json:"period" example:"2026-06-08T00:00:00Z"`
	VolumeKg float64   `json:"volumeKg" example:"4250"`
	Reps     int       `json:"reps" example:"96"`
	Sessions int       `json:"sessions" example:"3"`
}

// StreakResponse is the rest-day-tolerant streak returned by GET /stats/streak.
type StreakResponse struct {
	CurrentDays    int    `json:"currentDays" example:"5"`          // workout days in the active chain (0 if broken)
	LongestDays    int    `json:"longestDays" example:"12"`         // best chain ever
	IsActive       bool   `json:"isActive" example:"true"`          // false once the rest-day budget is exceeded
	LastWorkout    string `json:"lastWorkout" example:"2026-06-25"` // YYYY-MM-DD of most recent workout, "" if none
	DaysUntilReset int    `json:"daysUntilReset" example:"2"`       // rest days left before the current streak resets
}

// MuscleVolumeResponse is one row of GET /stats/muscles.
type MuscleVolumeResponse struct {
	Muscle   string  `json:"muscle" example:"chest"`
	VolumeKg float64 `json:"volumeKg" example:"32000"`
}
