package exercise

import "github.com/Haidarr-h/backend-go/internal/config"

type ExerciseService struct {
	repo Repo
	cfg  *config.Config
}

func NewExerciseService(repo Repo, cfg *config.Config) *ExerciseService {
	return &ExerciseService{repo: repo, cfg: cfg}
}

func (s *ExerciseService) GetExercises() ([]ExerciseResponse, error) {

	exercises, err := s.repo.FindAll()

	if err != nil {
		return []ExerciseResponse{}, err
	}

	var result []ExerciseResponse
	for _, e := range exercises {
		result = append(result, ExerciseResponse{
			ID:               e.ID,
			Name:             e.Name,
			Force:            e.Force,
			Level:            e.Level,
			Mechanic:         e.Mechanic,
			Equipment:        e.Equipment,
			PrimaryMuscles:   e.PrimaryMuscles,
			SecondaryMuscles: e.SecondaryMuscles,
			Instructions:     e.Instructions,
			Images:           e.Images,
			Category:         e.Category,
			VideoURL:         e.VideoURL,
		})
	}

	return result, nil
}

func (s *ExerciseService) GetExercise(id ExerciseRequest) (ExerciseResponse, error) {

	exercise, err := s.repo.FindByID(id.ID)

	if err != nil {
		return ExerciseResponse{}, err
	}

	result := ExerciseResponse{
		ID:               exercise.ID,
		Name:             exercise.Name,
		Force:            exercise.Force,
		Level:            exercise.Level,
		Mechanic:         exercise.Mechanic,
		Equipment:        exercise.Equipment,
		PrimaryMuscles:   exercise.PrimaryMuscles,
		SecondaryMuscles: exercise.SecondaryMuscles,
		Instructions:     exercise.Instructions,
		Images:           exercise.Images,
		Category:         exercise.Category,
		VideoURL:         exercise.VideoURL,
	}

	return result, nil
}
