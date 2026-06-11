package exercise

import "github.com/lib/pq"

type ExerciseRequest struct {
	ID uint `json:"id"`
}

type ExerciseResponse struct {
	ID               uint           `json:"id"`
	Name             string         `json:"name"`
	Force            string         `json:"force"`
	Level            string         `json:"level"`
	Mechanic         string         `json:"mechanic"`
	Equipment        string         `json:"equipment"`
	PrimaryMuscles   pq.StringArray `json:"primaryMuscles"`
	SecondaryMuscles pq.StringArray `json:"secondaryMuscles"`
	Instructions     pq.StringArray `json:"instructions"`
	Images           pq.StringArray `json:"images"`
	Category         string         `json:"category"`
	VideoURL         string         `json:"videoUrl"`
}

type ExerciseFilterParams struct {
	MuscleGroup string
	Category    string
	Level       string
	Equipment   string
	Mechanic    string
	Limit       int
	Offset      int
}
