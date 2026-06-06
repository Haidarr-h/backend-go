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
	Name             *string                    `json:"name" example:"Push Day (Heavy)"`
	Description      *string                    `json:"description" example:"Updated chest and triceps focus"`
	IsPublic         *bool                      `json:"isPublic" example:"true"`
	RoutineExercises []RoutineExercisesRequest `json:"routineExercises"`
}

type RoutineExercisesRequest struct {
	ExerciseID uint    `json:"exerciseId" binding:"required" example:"12"`
	Order      int     `json:"order" binding:"required" example:"1"`
	Sets       int     `json:"sets" example:"3"`
	Reps       int     `json:"reps" example:"10"`
	WeightKG   float64 `json:"weightKg" example:"60.5"`
	RestSecond int     `json:"restSecond" example:"90"`
}

type UpdateRoutineExercisesReq struct {
	ExerciseID *uint    `json:"exerciseId" example:"12"`
	Order      int     `json:"order" example:"1"`
	Sets       int     `json:"sets" example:"4"`
	Reps       int     `json:"reps" example:"8"`
	WeightKG   float64 `json:"weightKg" example:"65"`
	RestSecond int     `json:"restSecond" example:"120"`
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