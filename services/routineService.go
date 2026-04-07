package services

import (
	"github.com/Haidarr-h/backend-go/dto"
	"github.com/Haidarr-h/backend-go/models"
	"github.com/Haidarr-h/backend-go/repositories"
)

type RoutineService struct {
	routineRepo *repositories.RoutineRepository
}

func NewRoutineService(routineRepo *repositories.RoutineRepository) *RoutineService {
	return &RoutineService{routineRepo: routineRepo}
}

func (s *RoutineService) CreateRoutine(userID uint, req dto.CreateRoutineRequest) (dto.RoutineResponse, error) {
	// 1. map DTO to model
	routine := models.Routine{
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		UserId:      &userID,
	}

	// 2. Map nested exercise DTO to model
	for _, e := range req.RoutineExercises {
		routine.RoutineExercises = append(routine.RoutineExercises, models.RoutineExercises{
			ExerciseID: e.ExerciseID,
			Order:      e.Order,
			Sets:       e.Sets,
			Reps:       e.Reps,
			WeightKG:   e.WeightKG,
			RestSecond: e.RestSecond,
		})
	}

	// 3. pass model to repository
	created, err := s.routineRepo.Create(routine)

	if err != nil {
		return dto.RoutineResponse{}, err
	}

	// 4. map model to response dto
	return mapToRoutineResponse(created), nil

}

func (s *RoutineService) GetRoutines(userID uint) ([]dto.RoutineResponse, error) {
	// run the repo function
	routines, err := s.routineRepo.FindAll(userID)
	if err != nil {
		return nil, err
	}

	// map each model to dto response
	var response []dto.RoutineResponse
	for _, r := range routines {
		response = append(response, mapToRoutineResponse(r))
	}

	return response, nil
}

func mapToRoutineResponse(r models.Routine) dto.RoutineResponse {
	var exercises []dto.RoutineExerciseResponse

	for _, e := range r.RoutineExercises {
		exercises = append(exercises, dto.RoutineExerciseResponse{
			ID:         e.ID,
			ExerciseID: e.ExerciseID,
			Order:      e.Order,
			Sets:       e.Sets,
			Reps:       e.Reps,
			WeightKG:   e.WeightKG,
			RestSecond: e.RestSecond,
			Exercise: dto.ExerciseResponse{
				ID:          e.Exercise.ID,
				Name:        e.Exercise.Name,
				MuscleGroup: e.Exercise.MuscleGroup,
				Equipment:   e.Exercise.Equipment,
				Category:    e.Exercise.Category,
			},
		})
	}

	return dto.RoutineResponse{
		ID:               r.ID,
		Name:             r.Name,
		Description:      r.Description,
		IsPublic:         r.IsPublic,
		RoutineExercises: exercises,
		CreatedAt:        r.CreatedAt,
	}
}
