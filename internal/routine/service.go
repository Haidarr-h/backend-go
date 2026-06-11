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

	// 1. map DTO to model (routine id is unknown until insert, so pass 0)
	routine := Routine{
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		UserID:      &userID,
	}

	exercises, exerciseIDs := buildRoutineExercises(0, req.RoutineExercises)
	routine.RoutineExercises = exercises

	// 2. make sure every referenced exercise actually exists
	if err := s.assertExercisesExist(exerciseIDs); err != nil {
		return RoutineResponse{}, err
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

	// map each model to dto response (non-nil so the API returns [] not null)
	response := make([]RoutineResponse, 0, len(routines))
	for _, r := range routines {
		response = append(response, mapToRoutineResponse(r))
	}

	return response, nil
}

// GET ONE ROUTINE
func (s *RoutineService) GetRoutine(userID uint, routineID uint) (RoutineResponse, error) {

	// 1. fetch the
	routineData, err := s.routineRepo.FindByID(routineID)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return RoutineResponse{}, ErrRoutineNotFound
		}

		return RoutineResponse{}, err
	}

	// if routine public = everyone can access
	// if not public = only the owner can access

	// 2. check if public
	if routineData.IsPublic {
		return mapToRoutineResponse(routineData), nil
	}

	// 3. if not public, only the owner can access it (nil-safe)
	if routineData.UserID != nil && *routineData.UserID == userID {
		return mapToRoutineResponse(routineData), nil
	}

	return RoutineResponse{}, InvalidRoutineOwnership
}

// UPDATE
func (s *RoutineService) UpdateRoutine(id, userID uint, req UpdateRoutineReq) (RoutineResponse, error) {
	// 1. Fetch existing, so we have a complete model
	existing, err := s.routineRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return RoutineResponse{}, ErrRoutineNotFound
		}
		return RoutineResponse{}, err
	}

	// make sure the one who updates it is the owner (nil-safe)
	if existing.UserID == nil || *existing.UserID != userID {
		return RoutineResponse{}, InvalidRoutineOwnership
	}

	// 2. Apply only the provided scalar fields
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.IsPublic != nil {
		existing.IsPublic = *req.IsPublic
	}

	// 3. Exercises: omitted (nil) keeps them as-is, an explicit list (even
	// empty) replaces the whole set. Only validate + rebuild when provided.
	replaceExercises := req.RoutineExercises != nil
	if replaceExercises {
		exercises, exerciseIDs := buildRoutineExercises(id, req.RoutineExercises)

		if err := s.assertExercisesExist(exerciseIDs); err != nil {
			return RoutineResponse{}, err
		}

		existing.RoutineExercises = exercises
	}

	// 4. persist
	updated, err := s.routineRepo.Update(existing, replaceExercises)
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
			return ErrRoutineNotFound
		}

		return err
	}

	return nil
}

// buildRoutineExercises maps the request DTOs into RoutineExercise models and
// collects the referenced exercise ids. routineID is set on each child (it is
// ignored by Create, which only learns the routine id after insert).
func buildRoutineExercises(routineID uint, reqs []RoutineExercisesRequest) ([]RoutineExercise, []uint) {
	var exercises []RoutineExercise
	var exerciseIDs []uint

	for _, e := range reqs {
		re := RoutineExercise{
			RoutineID:  routineID,
			ExerciseID: e.ExerciseID,
			Order:      e.Order,
			RestSecond: e.RestSecond,
		}

		for _, set := range e.Sets {
			re.RoutineExerciseSets = append(re.RoutineExerciseSets, RoutineExerciseSet{
				SetNumber: set.SetNumber,
				Reps:      set.Reps,
				WeightKG:  set.WeightKG,
			})
		}

		exercises = append(exercises, re)
		exerciseIDs = append(exerciseIDs, e.ExerciseID)
	}

	return exercises, exerciseIDs
}

// assertExercisesExist returns ErrExerciseNotFound (wrapping the missing ids)
// when any of the given exercise ids does not exist. An empty list is a no-op.
func (s *RoutineService) assertExercisesExist(exerciseIDs []uint) error {
	if len(exerciseIDs) == 0 {
		return nil
	}

	missingIDs, err := s.exerciseRepo.ExistsByIDs(exerciseIDs)
	if err != nil {
		return err
	}

	if len(missingIDs) > 0 {
		return fmt.Errorf("%w: %v", ErrExerciseNotFound, missingIDs)
	}

	return nil
}

// HELPER MAPPING
func mapToRoutineResponse(r Routine) RoutineResponse {
	var exercises []RoutineExerciseResponse

	for _, e := range r.RoutineExercises {
		var exerciseSet []RoutineExerciseSetResponse

		for _, s := range e.RoutineExerciseSets {
			exerciseSet = append(exerciseSet, RoutineExerciseSetResponse{
				ID:        s.ID,
				SetNumber: s.SetNumber,
				Reps:      s.Reps,
				WeightKG:  s.WeightKG,
			})
		}

		exercises = append(exercises, RoutineExerciseResponse{
			ID:         e.ID,
			ExerciseID: e.ExerciseID,
			Order:      e.Order,
			RestSecond: e.RestSecond,
			Sets:       exerciseSet,
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
