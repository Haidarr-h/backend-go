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

// CREATE
func (r *RoutineRepository) Create(routine models.Routine) (models.Routine, error) {
	result := r.db.Create(&routine)

	if result.Error != nil {
		return models.Routine{}, result.Error
	}

	return routine, nil
}

// READ
func (r *RoutineRepository) FindAll(userID uint) ([]models.Routine, error) {
	var routine []models.Routine

	result := r.db.Where("user_id = ?", userID).Preload("RoutineExercises.Exercise").Find(&routine)

	if result.Error != nil {
		return nil, result.Error
	}

	return routine, nil
}

// UPDATE
func (r *RoutineRepository) Update(routine models.Routine) (models.Routine, error) {

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. delete existing exercises
		if err := tx.Where("routine_id = ?", routine.ID).Delete(&models.RoutineExercises{}).Error; err != nil {
			return err
		}

		// 2. save routine with new exercises
		if err := tx.Save(&routine).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return models.Routine{}, err
	}

	return routine, nil
}

// GET one
func (r *RoutineRepository) FindByID(id uint) (models.Routine, error) {
	var routine models.Routine

	result := r.db.Where("id = ?", id).Preload("RoutineExercises.Exercise").First(&routine)

	if result.Error != nil {
		return models.Routine{}, result.Error
	}

	return routine, nil
}

// DELETE
func (r *RoutineRepository) Delete(id uint, userId uint) error {

	// we need transaction because we will modify 2 tables (routine and routine exercises)
	// 1. Transactions
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1.1 Delete the routine exercises first (its the routine children)
		if err := tx.Where("routine_id = ?", id).Delete(&models.RoutineExercises{}).Error; err != nil {
			return err
		}

		// 1.2 Delete the routine
		result := tx.Where("id = ? AND user_id = ?", id, userId).Delete(&models.Routine{})
		if result.Error != nil {
			return result.Error
		}

		// 1.3 check if there is even that row ?
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})

	return err
}
