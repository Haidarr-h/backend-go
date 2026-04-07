package repositories

import (
	"github.com/Haidarr-h/backend-go/models"
	"gorm.io/gorm"
)

type RoutineRepository struct {
	db *gorm.DB
}

func NewRoutineRepository(db *gorm.DB) *RoutineRepository {
	return &RoutineRepository{db: db}
}

func (r *RoutineRepository) Create(routine models.Routine) (models.Routine, error) {
	result := r.db.Create(&routine)

	if result.Error != nil {
		return models.Routine{}, result.Error
	}

	return routine, nil
}

func (r *RoutineRepository) FindAll(userID uint) ([]models.Routine, error) {
	var routine []models.Routine

	result := r.db.Where("user_id = ?", userID).Preload("RoutineExercises.Exercise").Find(&routine)

	if result.Error != nil {
		return nil, result.Error
	}

	return routine, nil
}

func (r *RoutineRepository) Update(routine models.Routine) (models.Routine, error) {

	result := r.db.Save(&routine)

	if result.Error != nil {
		return models.Routine{}, result.Error
	}

	return routine, nil
}

func (r *RoutineRepository) FindByID(id uint) (models.Routine, error) {
	var routine models.Routine

	result := r.db.Where("id = ?", id).Preload("RoutineExercises.Exercise").Find(&routine)

	if result.Error != nil {
		return models.Routine{}, result.Error
	}

	return routine, nil
}
