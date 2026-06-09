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
	queryRoutineExercise := "INSERT INTO routine_exercises (routine_id, exercise_id, \"order\", rest_second, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW()) RETURNING *"

	// 3. inserting to routine exercise set join table
	queryRoutineExerciseSet := "INSERT INTO routine_exercise_sets (routine_exercise_id, set_number, reps, weight_kg, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())"

	var routineModel Routine

	// 3. performing transactions since it is affecting more than one table
	errCreate := r.db.Transaction(func(tx *gorm.DB) error {

		// 3.1. create the routines
		routineResult := tx.Raw(queryRoutine, routine.UserID, routine.Name, routine.Description, routine.IsPublic).Scan(&routineModel)

		if routineResult.Error != nil {
			logger.Log.Error("create routine function failed", "error", routineResult.Error)
			return ErrCreateRoutine
		}

		if routineResult.RowsAffected == 0 {
			logger.Log.Error("create routine function failed - rows affected is 0", "error", ErrCreateRoutine)
			return ErrCreateRoutine
		}

		// 3.2. create the routine exercises
		for _, e := range routine.RoutineExercises {

			var routineExercise RoutineExercise

			if routineExerciseErr := tx.Raw(queryRoutineExercise, routineModel.ID, e.ExerciseID, e.Order, e.RestSecond).Scan(&routineExercise); routineExerciseErr.Error != nil {
				logger.Log.Error("create routine exercise failed", "error", routineExerciseErr)
				return ErrCreateRoutine
			}

			// loop through each set
			for _, s := range e.RoutineExerciseSets {
				if err := tx.Exec(queryRoutineExerciseSet, routineExercise.ID, s.SetNumber, s.Reps, s.WeightKG).Error; err != nil {
					logger.Log.Error("create routine exercise set failed", "error", err)
					return ErrCreateRoutine
				}
			}
		}

		return nil
	})

	if errCreate != nil {
		return Routine{}, ErrCreateRoutine
	}

	return r.FindByID(routineModel.ID)
}

// READ
func (r *RoutineRepository) FindAll(userID uint) ([]Routine, error) {
	var routine []Routine

	result := r.db.Where("user_id = ?", userID).
		Preload("RoutineExercises.Exercise").
		Preload("RoutineExercises.RoutineExerciseSets").
		Find(&routine)

	if result.Error != nil {
		return nil, result.Error
	}

	return routine, nil
}

// UPDATE
func (r *RoutineRepository) Update(routine Routine) (Routine, error) {

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. collect the existing exercise ids for this routine
		var exerciseIDs []uint
		if err := tx.Model(&RoutineExercise{}).
			Where("routine_id = ?", routine.ID).
			Pluck("id", &exerciseIDs).Error; err != nil {
			return err
		}

		// 2. delete the sets belonging to those exercises (otherwise they orphan)
		if len(exerciseIDs) > 0 {
			if err := tx.Where("routine_exercise_id IN ?", exerciseIDs).
				Delete(&RoutineExerciseSet{}).Error; err != nil {
				return err
			}
		}

		// 3. delete the existing exercises
		if err := tx.Where("routine_id = ?", routine.ID).Delete(&RoutineExercise{}).Error; err != nil {
			return err
		}

		// 4. save routine with new exercises + nested sets (GORM cascades the creates)
		if err := tx.Save(&routine).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return Routine{}, err
	}

	// re-fetch so the response carries the preloaded exercise + sets, like Create
	return r.FindByID(routine.ID)
}

// GET one
func (r *RoutineRepository) FindByID(id uint) (Routine, error) {
	var routine Routine

	result := r.db.Where("id = ?", id).
		Preload("RoutineExercises.Exercise").
		Preload("RoutineExercises.RoutineExerciseSets").
		First(&routine)

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
		// 1.1 collect the exercise ids so we can clean up their sets
		var exerciseIDs []uint
		if err := tx.Model(&RoutineExercise{}).
			Where("routine_id = ?", id).
			Pluck("id", &exerciseIDs).Error; err != nil {
			return err
		}

		// 1.2 Delete the sets belonging to those exercises (grandchildren)
		if len(exerciseIDs) > 0 {
			if err := tx.Where("routine_exercise_id IN ?", exerciseIDs).
				Delete(&RoutineExerciseSet{}).Error; err != nil {
				return err
			}
		}

		// 1.3 Delete the routine exercises (its the routine children)
		if err := tx.Where("routine_id = ?", id).Delete(&RoutineExercise{}).Error; err != nil {
			return err
		}

		// 1.4 Delete the routine
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
