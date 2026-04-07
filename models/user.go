package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `gorm:"unique"`
	FullName string
	Username string  `gorm:"unique"`
	GoogleID string  `gorm:"default:null" json:"google_id,omitempty"`
	Picture  string  `json:"picture,omitempty"`
	Password *string `json:"password,omitempty"`
	Routines []Routine
}
