package routine

import (
	"errors"
	"fmt"

	"github.com/Haidarr-h/backend-go/internal/exercise"
	"gorm.io/gorm"
)

type RoutineService struct {
	routineRepo  Repo
	exerciseRepo exercise.Repo
}

func NewRoutineService(routineRepo Repo, exerciseRepo exercise.Repo) *RoutineService {
	return &RoutineService{routineRepo: routineRepo, exerciseRepo: exerciseRepo}
}

// CREATE
func (s *RoutineService) CreateRoutine(userID uint, req CreateRoutineRequest) (RoutineResponse, error) {

	var exerciseIDs []uint

	// 1. map DTO to model
	routine := Routine{
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		UserId:      &userID,
	}

	// 2. Map nested exercise DTO to model
	for _, e := range req.RoutineExercises {
		routine.RoutineExercises = append(routine.RoutineExercises, RoutineExercises{
			ExerciseID: e.ExerciseID,
			Order:      e.Order,
			Sets:       e.Sets,
			Reps:       e.Reps,
			WeightKG:   e.WeightKG,
			RestSecond: e.RestSecond,
		})

		exerciseIDs = append(exerciseIDs, e.ExerciseID)
	}

	// check if all exercise ids are valid
	missingIDs, errMissingIDs := s.exerciseRepo.ExistsByIDs(exerciseIDs)

	if errMissingIDs != nil {
		return RoutineResponse{}, errMissingIDs
	}

	if len(missingIDs) > 0 {
		return RoutineResponse{}, fmt.Errorf("%w: %v", ErrExerciseNotFound, missingIDs)
	}

	// 3. pass model to repository
	created, err := s.routineRepo.Create(routine)

	if err != nil {
		return RoutineResponse{}, err
	}

	// 4. map model to response dto
	return mapToRoutineResponse(created), nil
}

// GET ALL ROUTINES
func (s *RoutineService) GetRoutines(userID uint) ([]RoutineResponse, error) {
	// run the repo function
	routines, err := s.routineRepo.FindAll(userID)
	if err != nil {
		return nil, err
	}

	// map each model to dto response
	var response []RoutineResponse
	for _, r := range routines {
		response = append(response, mapToRoutineResponse(r))
	}

	return response, nil
}

// UPDATE
func (s *RoutineService) UpdateRoutine(id uint, req UpdateRoutineReq) (RoutineResponse, error) {
	// 1. Fetch existing, so we have a complete model
	existing, err := s.routineRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return RoutineResponse{}, errors.New("routine not found")
		}
		return RoutineResponse{}, err
	}

	// 2. Apply the only changes
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.IsPublic != nil {
		existing.IsPublic = *req.IsPublic
	}
	if req.RoutineExercises != nil {
		var exercises []RoutineExercises

		// map exercises dto to model
		for _, e := range req.RoutineExercises {
			exercises = append(exercises, RoutineExercises{
				ExerciseID: e.ExerciseID,
				RoutineID:  id,
				Order:      e.Order,
				Sets:       e.Sets,
				Reps:       e.Reps,
				WeightKG:   e.WeightKG,
				RestSecond: e.RestSecond,
			})
		}

		existing.RoutineExercises = exercises
	}

	// 3. save the full model
	updated, err := s.routineRepo.Update(existing)
	if err != nil {
		return RoutineResponse{}, err
	}

	return mapToRoutineResponse(updated), nil
}

// DELETE
func (s *RoutineService) DeleteRoutine(routineId uint, userId uint) error {

	err := s.routineRepo.Delete(routineId, userId)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("routine not found")
		}

		return err
	}

	return nil
}

// HELPER MAPPING
func mapToRoutineResponse(r Routine) RoutineResponse {
	var exercises []RoutineExerciseResponse

	for _, e := range r.RoutineExercises {
		exercises = append(exercises, RoutineExerciseResponse{
			ID:         e.ID,
			ExerciseID: e.ExerciseID,
			Order:      e.Order,
			Sets:       e.Sets,
			Reps:       e.Reps,
			WeightKG:   e.WeightKG,
			RestSecond: e.RestSecond,
			Exercise: exercise.ExerciseResponse{
				ID:               e.Exercise.ID,
				Name:             e.Exercise.Name,
				PrimaryMuscles:   e.Exercise.PrimaryMuscles,
				Force:            e.Exercise.Force,
				Level:            e.Exercise.Level,
				Mechanic:         e.Exercise.Mechanic,
				SecondaryMuscles: e.Exercise.SecondaryMuscles,
				Instructions:     e.Exercise.Instructions,
				Equipment:        e.Exercise.Equipment,
				Category:         e.Exercise.Category,
			},
		})
	}

	return RoutineResponse{
		ID:               r.ID,
		Name:             r.Name,
		Description:      r.Description,
		IsPublic:         r.IsPublic,
		RoutineExercises: exercises,
		CreatedAt:        r.CreatedAt,
	}
}
