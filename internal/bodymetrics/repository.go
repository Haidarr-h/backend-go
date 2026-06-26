package bodymetrics

import (
	"errors"

	"gorm.io/gorm"
)

type BodyMetricRepository struct {
	db *gorm.DB
}

func NewBodyMetricRepository(db *gorm.DB) *BodyMetricRepository {
	return &BodyMetricRepository{db: db}
}

// Upsert creates the user's body metrics row, or overwrites the existing one.
// One row per user is guaranteed by the unique index on user_id.
func (r *BodyMetricRepository) Upsert(metric BodyMetric) (BodyMetric, error) {
	var existing BodyMetric

	err := r.db.Where("user_id = ?", metric.UserID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if createErr := r.db.Create(&metric).Error; createErr != nil {
				return BodyMetric{}, createErr
			}
			return metric, nil
		}
		return BodyMetric{}, err
	}

	existing.HeightCM = metric.HeightCM
	existing.WeightKG = metric.WeightKG
	existing.Age = metric.Age
	existing.Gender = metric.Gender
	existing.BodyFat = metric.BodyFat
	existing.Goal = metric.Goal

	if saveErr := r.db.Save(&existing).Error; saveErr != nil {
		return BodyMetric{}, saveErr
	}

	return existing, nil
}

// FindByUserID returns the user's body metrics, or gorm.ErrRecordNotFound.
func (r *BodyMetricRepository) FindByUserID(userID uint) (BodyMetric, error) {
	var metric BodyMetric

	if err := r.db.Where("user_id = ?", userID).First(&metric).Error; err != nil {
		return BodyMetric{}, err
	}

	return metric, nil
}
