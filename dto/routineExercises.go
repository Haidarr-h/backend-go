package dto

type RoutineExercisesRequest struct {
	ExerciseID uint    `json:"exerciseId" binding:"required"`
	Order      int     `json:"order" binding:"required"`
	Sets       int     `json:"sets"`
	Reps       int     `json:"reps"`
	WeightKG   float64 `json:"weightKg"`
	RestSecond int     `json:"restSecond"`
}

type UpdateRoutineExercisesReq struct {
	ExerciseID *uint    `json:"exerciseId"`
	Order      int     `json:"order"`
	Sets       int     `json:"sets"`
	Reps       int     `json:"reps"`
	WeightKG   float64 `json:"weightKg"`
	RestSecond int     `json:"restSecond"`
}

type RoutineExerciseResponse struct {
	ID         uint             `json:"id"`
	ExerciseID uint             `json:"exerciseId" binding:"required"`
	Order      int              `json:"order" binding:"required"`
	Sets       int              `json:"sets"`
	Reps       int              `json:"reps"`
	WeightKG   float64          `json:"weightKg"`
	RestSecond int              `json:"restSecond"`
	Exercise   ExerciseResponse `json:"exercise"`
}
