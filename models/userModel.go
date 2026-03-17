package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `gorm:"unique"`
	FullName string
	Username string `gorm:"unique"`
	// Add to your existing User struct
	GoogleID string `gorm:"default:null" json:"google_id,omitempty"`
	Name     string `json:"name"`
	Picture  string `json:"picture,omitempty"`
	// Make sure Password is nullable since Google users won't have one
	Password *string `json:"password,omitempty"` // pointer = nullable
}
