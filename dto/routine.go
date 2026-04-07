package dto

type CreateRoutineRequest struct {
	Name             string                    `json:"name" binding:"required"`
	Description      string                    `json:"description"`
	IsPublic         bool                      `json:"isPublic"`
	RoutineExercises []RoutineExercisesRequest `json:"routineExercises" binding:"required"`
}

type CreateRoutineResponse struct {
	ID               uint                    `json:"id"`
	Name             string                    `json:"name"`
	Description      string                    `json:"description"`
	IsPublic         bool                      `json:"isPublic"`
	RoutineExercises []RoutineExerciseResponse `json:"routineExercises"`
}
