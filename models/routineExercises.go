package models

import "gorm.io/gorm"

type RoutineExercises struct {
	gorm.Model
	RoutineID uint
	ExerciseID uint
	Exercise Exercise
	Order int
	Sets int
	Reps int
	WeightKG float64
	RestSecond int
}