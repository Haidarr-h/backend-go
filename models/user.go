package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email     string  `gorm:"uniqueIndex;not null"`
	FirstName string  `gorm:"not null"`
	LastName  string  `gorm:"not null"`
	Username  string  `gorm:"uniqueIndex;not null"`
	GoogleID  *string `gorm:"default:null"`
	Picture   *string
	Password  *string
	Routines  []Routine
}
