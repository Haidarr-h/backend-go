package session

import (
	"time"

	"github.com/Haidarr-h/backend-go/internal/exercise"
)

type CreateSessionRequest struct {
	RoutineID uint      `json:"routineId" binding:"required" example:"1"`
	Name      string    `json:"name" example:"Morning Push"`
	Notes     string    `json:"notes" example:"Felt strong today"`
	StartTime time.Time `json:"startTime" binding:"required" example:"2026-06-11T08:00:00Z"`
	EndTime   time.Time `json:"endTime" binding:"required" example:"2026-06-11T09:00:00Z"`
}

type UpdateSessionReq struct {
	Name      *string    `json:"name" example:"Morning Push (Heavy)"`
	Notes     *string    `json:"notes" example:"New PR on bench"`
	StartTime *time.Time `json:"startTime" example:"2026-06-11T08:00:00Z"`
	EndTime   *time.Time `json:"endTime" example:"2026-06-11T09:15:00Z"`
}

type SessionResponse struct {
	ID               uint                      `json:"id"`
	RoutineID        *uint                     `json:"routineId"`
	Name             string                    `json:"name"`
	Notes            string                    `json:"notes"`
	StartTime        time.Time                 `json:"startTime"`
	EndTime          time.Time                 `json:"endTime"`
	Metrics          SessionMetrics            `json:"metrics"`
	SessionExercises []SessionExerciseResponse `json:"sessionExercises"`
	CreatedAt        time.Time                 `json:"createdAt"`
}

type SessionMetrics struct {
	TotalWeight          float64  `json:"totalWeight" example:"4250"`
	TotalDurationSeconds int      `json:"totalDurationSeconds" example:"3600"`
	TotalSets            int      `json:"totalSets" example:"12"`
	TotalReps            int      `json:"totalReps" example:"96"`
	TargetedMuscles      []string `json:"targetedMuscles" example:"chest,triceps,shoulders"`
}

type SessionExerciseResponse struct {
	ID         uint                         `json:"id"`
	ExerciseID uint                         `json:"exerciseId"`
	Order      int                          `json:"order"`
	RestSecond int                          `json:"restSecond"`
	Exercise   exercise.ExerciseResponse    `json:"exercise"`
	Sets       []SessionExerciseSetResponse `json:"sessionExerciseSet"`
}

type SessionExerciseSetResponse struct {
	ID        uint    `json:"id"`
	SetNumber int     `json:"setNumber"`
	Reps      int     `json:"reps"`
	WeightKG  float64 `json:"weightKg"`
}
