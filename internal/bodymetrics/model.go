package bodymetrics

import "gorm.io/gorm"

// BodyMetric is a user's current body profile. One row per user (enforced by the
// unique index on UserID); POST upserts it. BMI is derived on read, not stored.
type BodyMetric struct {
	gorm.Model
	UserID   uint `gorm:"uniqueIndex;not null"`
	HeightCM float64
	WeightKG float64
	Age      int
	Gender   string
	BodyFat  float64
	Goal     string
}
