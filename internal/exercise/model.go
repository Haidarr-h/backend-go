package exercise

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Exercise struct {
	gorm.Model
	Name             string         `json:"name"`
	Force            string         `json:"force"`
	Level            string         `json:"level"`
	Mechanic         string         `json:"mechanic"`
	Equipment        string         `json:"equipment"`
	PrimaryMuscles   pq.StringArray `gorm:"column:primary_muscles;type:text[]"   json:"primaryMuscles"`
	SecondaryMuscles pq.StringArray `gorm:"column:secondary_muscles;type:text[]" json:"secondaryMuscles"`
	Instructions     pq.StringArray `gorm:"column:instructions;type:text[]"      json:"instructions"`
	Images           pq.StringArray `gorm:"column:images;type:text[]"            json:"images"`
	Category         string         `json:"category"`
	VideoURL         string         `json:"videoUrl"`
}
