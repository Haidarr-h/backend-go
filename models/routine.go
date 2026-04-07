package models

import "gorm.io/gorm"

type Routine struct {
	gorm.Model
	UserId           *uint `gorm:"default:null"`
	Name             string
	Description      string
	IsPublic         bool
	RoutineExercises []RoutineExercises
}
