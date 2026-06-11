package session

import (
	"time"

	"github.com/Haidarr-h/backend-go/internal/exercise"
	"gorm.io/gorm"
)

type Session struct {
	gorm.Model
	UserID uint
	// RoutineID is a plain nullable pointer with NO Routine association field on
	// purpose: we do not want GORM to create an FK constraint
	// sessions.routine_id -> routines.id. The session snapshots the routine's
	// exercises at create time, so it must survive the routine being edited or
	// deleted (routine/user deletes cascade elsewhere). It only records provenance.
	RoutineID        *uint
	Name             string
	Notes            string
	StartTime        time.Time
	EndTime          time.Time
	SessionExercises []SessionExercise
}

type SessionExercise struct {
	gorm.Model
	SessionID           uint
	ExerciseID          uint
	Exercise            exercise.Exercise
	Order               int
	RestSecond          int
	SessionExerciseSets []SessionExerciseSet
}

type SessionExerciseSet struct {
	gorm.Model
	SessionExerciseID uint
	SetNumber         int
	Reps              int
	WeightKG          float64
}
