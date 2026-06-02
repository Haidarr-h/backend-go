package routine

import (
	"time"

	"github.com/Haidarr-h/backend-go/internal/exercise"
)

type CreateRoutineRequest struct {
	Name             string                    `json:"name" binding:"required"`
	Description      string                    `json:"description"`
	IsPublic         bool                      `json:"isPublic"`
	RoutineExercises []RoutineExercisesRequest `json:"routineExercises" binding:"required"`
}

type RoutineResponse struct {
	ID               uint                      `json:"id"`
	Name             string                    `json:"name"`
	Description      string                    `json:"description"`
	IsPublic         bool                      `json:"isPublic"`
	RoutineExercises []RoutineExerciseResponse `json:"routineExercises"`
	CreatedAt        time.Time                 `json:"createdAt"`
}

type UpdateRoutineReq struct {
	Name             *string                    `json:"name"`
	Description      *string                    `json:"description"`
	IsPublic         *bool                      `json:"isPublic"`
	RoutineExercises []RoutineExercisesRequest `json:"routineExercises"`
}

type RoutineExercisesRequest struct {
	ExerciseID uint    `json:"exerciseId" binding:"required"`
	Order      int     `json:"order" binding:"required"`
	Sets       int     `json:"sets"`
	Reps       int     `json:"reps"`
	WeightKG   float64 `json:"weightKg"`
	RestSecond int     `json:"restSecond"`
}

type UpdateRoutineExercisesReq struct {
	ExerciseID *uint    `json:"exerciseId"`
	Order      int     `json:"order"`
	Sets       int     `json:"sets"`
	Reps       int     `json:"reps"`
	WeightKG   float64 `json:"weightKg"`
	RestSecond int     `json:"restSecond"`
}

type RoutineExerciseResponse struct {
	ID         uint             `json:"id"`
	ExerciseID uint             `json:"exerciseId" binding:"required"`
	Order      int              `json:"order" binding:"required"`
	Sets       int              `json:"sets"`
	Reps       int              `json:"reps"`
	WeightKG   float64          `json:"weightKg"`
	RestSecond int              `json:"restSecond"`
	Exercise   exercise.ExerciseResponse `json:"exercise"`
}