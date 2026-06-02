package routine

import (
	"github.com/Haidarr-h/backend-go/pkg/logger"
	"gorm.io/gorm"
)

type RoutineRepository struct {
	db *gorm.DB
}

func NewRoutineRepository(db *gorm.DB) *RoutineRepository {
	return &RoutineRepository{db: db}
}

// CREATE
func (r *RoutineRepository) Create(routine Routine) (Routine, error) {

	// 1. insert to routine table first
	queryRoutine := "INSERT INTO routines (user_id, name, description, is_public, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW()) RETURNING *"

	// 2. for inserting to routine exercises join table
	queryRoutineExercise := "INSERT INTO routine_exercises (routine_id, exercise_id, \"order\", sets, reps, weight_kg, rest_second, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW()) RETURNING *"

	var routineModel Routine
	var routineExerciseModel []RoutineExercises

	// 3. performing transactions since it is affecting more than one table
	errCreate := r.db.Transaction(func(tx *gorm.DB) error {

		// 3.1. create the routines
		routineResult := tx.Raw(queryRoutine, routine.UserId, routine.Name, routine.Description, routine.IsPublic).Scan(&routineModel)

		if routineResult.Error != nil {
			logger.Log.Error("create routine function failed", "error", routineResult.Error)
			return routineResult.Error
		}

		if routineResult.RowsAffected == 0 {
			logger.Log.Error("create routine function failed - rows affected is 0", "error", ErrCreateRoutine)
			return ErrCreateRoutine
		}

		// 3.2. create the routine exercises
		for _, e := range routine.RoutineExercises {

			var routineExercises RoutineExercises

			routineExerciseResult := tx.Raw(queryRoutineExercise, routineModel.ID, e.ExerciseID, e.Order, e.Sets, e.Reps, e.WeightKG, e.RestSecond).Scan(&routineExercises)

			if routineExerciseResult.Error != nil {
				logger.Log.Error("create routine exercise failed", "error", routineExerciseResult.Error)
				return routineExerciseResult.Error
			}

			if routineExerciseResult.RowsAffected == 0 {
				logger.Log.Error("create routine exercise failed - rows affected is 0", "error", ErrCreateRoutine)
				return ErrCreateRoutine
			}

			routineExerciseModel = append(routineExerciseModel, routineExercises)
		}

		// 3.3. map it
		routineModel.RoutineExercises = routineExerciseModel

		return nil
	})

	if errCreate != nil {
		return Routine{}, ErrCreateRoutine
	}

	return routineModel, nil
}

// READ
func (r *RoutineRepository) FindAll(userID uint) ([]Routine, error) {
	var routine []Routine

	result := r.db.Where("user_id = ?", userID).Preload("RoutineExercises.Exercise").Find(&routine)

	if result.Error != nil {
		return nil, result.Error
	}

	return routine, nil
}

// UPDATE
func (r *RoutineRepository) Update(routine Routine) (Routine, error) {

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. delete existing exercises
		if err := tx.Where("routine_id = ?", routine.ID).Delete(&RoutineExercises{}).Error; err != nil {
			return err
		}

		// 2. save routine with new exercises
		if err := tx.Save(&routine).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return Routine{}, err
	}

	return routine, nil
}

// GET one
func (r *RoutineRepository) FindByID(id uint) (Routine, error) {
	var routine Routine

	result := r.db.Where("id = ?", id).Preload("RoutineExercises.Exercise").First(&routine)

	if result.Error != nil {
		return Routine{}, result.Error
	}

	return routine, nil
}

// DELETE
func (r *RoutineRepository) Delete(id uint, userId uint) error {

	// we need transaction because we will modify 2 tables (routine and routine exercises)
	// 1. Transactions
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1.1 Delete the routine exercises first (its the routine children)
		if err := tx.Where("routine_id = ?", id).Delete(&RoutineExercises{}).Error; err != nil {
			return err
		}

		// 1.2 Delete the routine
		result := tx.Where("id = ? AND user_id = ?", id, userId).Delete(&Routine{})
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
