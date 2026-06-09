package routine

import (
	"github.com/Haidarr-h/backend-go/internal/exercise"
	"gorm.io/gorm"
)

type Routine struct {
	gorm.Model
	UserID           *uint `gorm:"default:null"`
	Name             string
	Description      string
	IsPublic         bool
	RoutineExercises []RoutineExercise
}

type RoutineExercise struct {
	gorm.Model
	RoutineID           uint
	ExerciseID          uint
	Exercise            exercise.Exercise
	Order               int
	RestSecond          int
	RoutineExerciseSets []RoutineExerciseSet
}

type RoutineExerciseSet struct {
	gorm.Model
	RoutineExerciseID uint
	SetNumber         int
	Reps              int
	WeightKG          float64
}
