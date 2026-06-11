package routine

import (
	"time"

	"github.com/Haidarr-h/backend-go/internal/exercise"
)

type CreateRoutineRequest struct {
	Name             string                    `json:"name" binding:"required" example:"Push Day"`
	Description      string                    `json:"description" example:"Chest, shoulders and triceps"`
	IsPublic         bool                      `json:"isPublic" example:"false"`
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
	Name             *string                   `json:"name" example:"Push Day (Heavy)"`
	Description      *string                   `json:"description" example:"Updated chest and triceps focus"`
	IsPublic         *bool                     `json:"isPublic" example:"true"`
	RoutineExercises []RoutineExercisesRequest `json:"routineExercises"`
}

type RoutineExercisesRequest struct {
	ExerciseID uint                        `json:"exerciseId" binding:"required" example:"12"`
	Order      int                         `json:"order" binding:"required" example:"1"`
	RestSecond int                         `json:"restSecond" example:"90"`
	Sets       []RoutineExerciseSetRequest `json:"routineExerciseSet"`
}

type UpdateRoutineExercisesReq struct {
	ExerciseID *uint `json:"exerciseId" example:"12" binding:"required"`
	Order      int   `json:"order" example:"1" binding:"required"`
	RestSecond int   `json:"restSecond" example:"120" binding:"required"`
}

type RoutineExerciseResponse struct {
	ID         uint                         `json:"id"`
	ExerciseID uint                         `json:"exerciseId"`
	Order      int                          `json:"order"`
	RestSecond int                          `json:"restSecond"`
	Exercise   exercise.ExerciseResponse    `json:"exercise"`
	Sets       []RoutineExerciseSetResponse `json:"routineExerciseSet"`
}

type RoutineExerciseSetRequest struct {
	SetNumber int     `json:"setNumber" binding:"required" example:"1"`
	Reps      int     `json:"reps" binding:"required" example:"10"`
	WeightKG  float64 `json:"weightKg" binding:"required" example:"40"`
}

type RoutineExerciseSetResponse struct {
	ID        uint    `json:"id"`
	SetNumber int     `json:"setNumber"`
	Reps      int     `json:"reps"`
	WeightKG  float64 `json:"weightKg"`
}
