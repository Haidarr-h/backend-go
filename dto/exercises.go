package dto

type ExerciseRequest struct {
	Name        string `json:"name"`
	MuscleGroup string `json:"muscleGroup"`
	Equipment   string `json:"equipment"`
	Category    string `json:"category"`
}

type ExerciseResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	MuscleGroup string `json:"muscleGroup"`
	Equipment   string `json:"equipment"`
	Category    string `json:"category"`
}
