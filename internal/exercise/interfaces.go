package exercise

type Service interface {
	GetExercises(params ExerciseFilterParams) ([]ExerciseResponse, error)
	GetExercise(req ExerciseRequest) (ExerciseResponse, error)
}

type Repo interface {
	FindAll(params ExerciseFilterParams) ([]Exercise, error)
	FindByID(id uint) (Exercise, error)
}