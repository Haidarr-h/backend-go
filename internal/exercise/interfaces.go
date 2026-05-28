package exercise

type Service interface {
	GetExercises() ([]ExerciseResponse, error)
	GetExercise(req ExerciseRequest) (ExerciseResponse, error)
}

type Repo interface {
	FindAll() ([]Exercise, error)
	FindByID(id uint) (Exercise, error)
}