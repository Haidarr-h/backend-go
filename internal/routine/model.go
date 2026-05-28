package routine

import (
	"github.com/Haidarr-h/backend-go/internal/exercise"
	"gorm.io/gorm"
)

type Routine struct {
	gorm.Model
	UserId           *uint `gorm:"default:null"`
	Name             string
	Description      string
	IsPublic         bool
	RoutineExercises []RoutineExercises
}

type RoutineExercises struct {
	gorm.Model
	RoutineID  uint
	ExerciseID uint
	Exercise   exercise.Exercise
	Order      int
	Sets       int
	Reps       int
	WeightKG   float64
	RestSecond int
}
