package exercise

import (
	"github.com/Haidarr-h/backend-go/pkg/logger"
	"gorm.io/gorm"
)

type ExerciseRepository struct {
	db *gorm.DB
}

func NewExerciseRepository(db *gorm.DB) *ExerciseRepository {
	return &ExerciseRepository{db: db}
}

// GET ALL
func (r *ExerciseRepository) FindAll(params ExerciseFilterParams) ([]Exercise, error) {

	var exercises []Exercise

	query := "SELECT * FROM Exercises WHERE is_featured = true"
	args := []interface{}{}

	if params.Category != "" {
		query += " AND category = ?"
		args = append(args, params.Category)
	}

	if params.Equipment != "" {
		query += " AND equipment = ?"
		args = append(args, params.Equipment)
	}

	if params.Level != "" {
		query += " AND level = ?"
		args = append(args, params.Level)
	}

	if params.Mechanic != "" {
		query += " AND mechanic = ?"
		args = append(args, params.Mechanic)
	}

	if params.MuscleGroup != "" {
		query += " AND (? = ANY(primary_muscles) OR ? = ANY(secondary_muscles))"
		args = append(args, params.MuscleGroup)
	}

	if params.Limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, params.Limit, params.Offset)
	}

	result := r.db.Raw(query, args...).Scan(&exercises)

	if result.Error != nil {
		logger.Log.Error("FindAll exercise function is error", "error", result.Error)
		return []Exercise{}, result.Error
	}

	return exercises, nil
}

// GET BY ID
func (r *ExerciseRepository) FindByID(id uint) (Exercise, error) {

	var exercise Exercise

	query := "SELECT * FROM Exercises WHERE id = ? LIMIT 1"
	result := r.db.Raw(query, id).Scan(&exercise)

	if result.Error != nil {
		logger.Log.Error("FindByID exercise function is error", "error", result.Error)
		return Exercise{}, result.Error
	}

	return exercise, nil
}
