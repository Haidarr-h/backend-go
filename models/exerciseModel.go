package models

import "gorm.io/gorm"

type Exercise struct {
	gorm.Model
	Name        string `json: "name"`
	MuscleGroup string `json: "muscleGroup"`
	Equipment   string `json: "equipment"`
	Category    string `json: "category"`
}
