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
func (r *ExerciseRepository) FindAll() ([]Exercise, error) {

	var exercises []Exercise

	query := "SELECT * FROM Exercises"
	result := r.db.Raw(query).Scan(&exercises)

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


