package services

// import (
// 	"errors"

// 	"github.com/Haidarr-h/backend-go/dto"
// 	"github.com/Haidarr-h/backend-go/repositories"
// 	"gorm.io/gorm"
// )

// type RoutineService struct {
// 	routineRepo *repositories.RoutineRepository
// }

// func NewRoutineService(routineRepo *repositories.RoutineRepository) *RoutineService {
// 	return &RoutineService{routineRepo: routineRepo}
// }

// // CREATE
// func (s *RoutineService) CreateRoutine(userID uint, req dto.CreateRoutineRequest) (dto.RoutineResponse, error) {
// 	// 1. map DTO to model
// 	routine := models.Routine{
// 		Name:        req.Name,
// 		Description: req.Description,
// 		IsPublic:    req.IsPublic,
// 		UserId:      &userID,
// 	}

// 	// 2. Map nested exercise DTO to model
// 	for _, e := range req.RoutineExercises {
// 		routine.RoutineExercises = append(routine.RoutineExercises, models.RoutineExercises{
// 			ExerciseID: e.ExerciseID,
// 			Order:      e.Order,
// 			Sets:       e.Sets,
// 			Reps:       e.Reps,
// 			WeightKG:   e.WeightKG,
// 			RestSecond: e.RestSecond,
// 		})
// 	}

// 	// 3. pass model to repository
// 	created, err := s.routineRepo.Create(routine)

// 	if err != nil {
// 		return dto.RoutineResponse{}, err
// 	}

// 	// 4. map model to response dto
// 	return mapToRoutineResponse(created), nil

// }

// // GET ALL ROUTINES
// func (s *RoutineService) GetRoutines(userID uint) ([]dto.RoutineResponse, error) {
// 	// run the repo function
// 	routines, err := s.routineRepo.FindAll(userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// map each model to dto response
// 	var response []dto.RoutineResponse
// 	for _, r := range routines {
// 		response = append(response, mapToRoutineResponse(r))
// 	}

// 	return response, nil
// }

// // UPDATE
// func (s *RoutineService) UpdateRoutine(id uint, req dto.UpdateRoutineReq) (dto.RoutineResponse, error) {
// 	// 1. Fetch existing, so we have a complete model
// 	existing, err := s.routineRepo.FindByID(id)
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return dto.RoutineResponse{}, errors.New("routine not found")
// 		}
// 		return dto.RoutineResponse{}, err
// 	}

// 	// 2. Apply the only changes
// 	if req.Name != nil {
// 		existing.Name = *req.Name
// 	}
// 	if req.Description != nil {
// 		existing.Description = *req.Description
// 	}
// 	if req.IsPublic != nil {
// 		existing.IsPublic = *req.IsPublic
// 	}
// 	if req.RoutineExercises != nil {
// 		var exercises []models.RoutineExercises

// 		// map exercises dto to model
// 		for _, e := range req.RoutineExercises {
// 			exercises = append(exercises, models.RoutineExercises{
// 				ExerciseID: e.ExerciseID,
// 				RoutineID:  id,
// 				Order:      e.Order,
// 				Sets:       e.Sets,
// 				Reps:       e.Reps,
// 				WeightKG:   e.WeightKG,
// 				RestSecond: e.RestSecond,
// 			})
// 		}

// 		existing.RoutineExercises = exercises
// 	}

// 	// 3. save the full model
// 	updated, err := s.routineRepo.Update(existing)
// 	if err != nil {
// 		return dto.RoutineResponse{}, err
// 	}

// 	return mapToRoutineResponse(updated), nil
// }

// // DELETE
// func (s *RoutineService) DeleteRoutine(routineId uint, userId uint) error {

// 	err := s.routineRepo.Delete(routineId, userId)

// 	if err != nil {

// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return errors.New("routine not found")
// 		}

// 		return err
// 	}

// 	return nil
// }

// // HELPER MAPPING
// func mapToRoutineResponse(r models.Routine) dto.RoutineResponse {
// 	var exercises []dto.RoutineExerciseResponse

// 	for _, e := range r.RoutineExercises {
// 		exercises = append(exercises, dto.RoutineExerciseResponse{
// 			ID:         e.ID,
// 			ExerciseID: e.ExerciseID,
// 			Order:      e.Order,
// 			Sets:       e.Sets,
// 			Reps:       e.Reps,
// 			WeightKG:   e.WeightKG,
// 			RestSecond: e.RestSecond,
// 			Exercise: dto.ExerciseResponse{
// 				ID:          e.Exercise.ID,
// 				Name:        e.Exercise.Name,
// 				MuscleGroup: e.Exercise.MuscleGroup,
// 				Equipment:   e.Exercise.Equipment,
// 				Category:    e.Exercise.Category,
// 			},
// 		})
// 	}

// 	return dto.RoutineResponse{
// 		ID:               r.ID,
// 		Name:             r.Name,
// 		Description:      r.Description,
// 		IsPublic:         r.IsPublic,
// 		RoutineExercises: exercises,
// 		CreatedAt:        r.CreatedAt,
// 	}
// }
