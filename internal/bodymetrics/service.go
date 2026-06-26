package bodymetrics

import (
	"errors"
	"math"

	"gorm.io/gorm"
)

type BodyMetricService struct {
	repo Repo
}

func NewBodyMetricService(repo Repo) *BodyMetricService {
	return &BodyMetricService{repo: repo}
}

// UpsertBodyMetric saves (creating or overwriting) the user's body metrics.
func (s *BodyMetricService) UpsertBodyMetric(userID uint, req UpsertBodyMetricRequest) (BodyMetricResponse, error) {
	saved, err := s.repo.Upsert(BodyMetric{
		UserID:   userID,
		HeightCM: req.HeightCM,
		WeightKG: req.WeightKG,
		Age:      req.Age,
		Gender:   req.Gender,
		BodyFat:  req.BodyFat,
		Goal:     req.Goal,
	})
	if err != nil {
		return BodyMetricResponse{}, err
	}

	return mapToResponse(saved), nil
}

// GetBodyMetric returns the user's saved body metrics.
func (s *BodyMetricService) GetBodyMetric(userID uint) (BodyMetricResponse, error) {
	metric, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return BodyMetricResponse{}, ErrBodyMetricNotFound
		}
		return BodyMetricResponse{}, err
	}

	return mapToResponse(metric), nil
}

// computeBMI returns weight(kg) / height(m)^2, rounded to 1 decimal. 0 if height <= 0.
func computeBMI(heightCM, weightKG float64) float64 {
	if heightCM <= 0 {
		return 0
	}
	heightM := heightCM / 100
	bmi := weightKG / (heightM * heightM)
	return math.Round(bmi*10) / 10
}

func mapToResponse(m BodyMetric) BodyMetricResponse {
	return BodyMetricResponse{
		HeightCM:  m.HeightCM,
		WeightKG:  m.WeightKG,
		BMI:       computeBMI(m.HeightCM, m.WeightKG),
		Age:       m.Age,
		Gender:    m.Gender,
		BodyFat:   m.BodyFat,
		Goal:      m.Goal,
		UpdatedAt: m.UpdatedAt,
	}
}
