package bodymetrics

import "time"

// UpsertBodyMetricRequest is the body for POST /body-metrics. BMI is computed
// server-side from height + weight, so it is not part of the request.
type UpsertBodyMetricRequest struct {
	HeightCM float64 `json:"heightCm" binding:"required,gt=0" example:"178"`
	WeightKG float64 `json:"weightKg" binding:"required,gt=0" example:"75"`
	Age      int     `json:"age" binding:"required,gt=0" example:"25"`
	Gender   string  `json:"gender" binding:"required,oneof=male female other" example:"male"`
	BodyFat  float64 `json:"bodyFat" binding:"omitempty,gte=0,lte=100" example:"15"`
	Goal     string  `json:"goal" binding:"required" example:"build_muscle"`
}

// BodyMetricResponse is what GET/POST /body-metrics return, including derived BMI.
type BodyMetricResponse struct {
	HeightCM  float64   `json:"heightCm" example:"178"`
	WeightKG  float64   `json:"weightKg" example:"75"`
	BMI       float64   `json:"bmi" example:"23.7"`
	Age       int       `json:"age" example:"25"`
	Gender    string    `json:"gender" example:"male"`
	BodyFat   float64   `json:"bodyFat" example:"15"`
	Goal      string    `json:"goal" example:"build_muscle"`
	UpdatedAt time.Time `json:"updatedAt"`
}
