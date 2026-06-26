package bodymetrics

type Service interface {
	UpsertBodyMetric(userID uint, req UpsertBodyMetricRequest) (BodyMetricResponse, error)
	GetBodyMetric(userID uint) (BodyMetricResponse, error)
}

type Repo interface {
	Upsert(metric BodyMetric) (BodyMetric, error)
	FindByUserID(userID uint) (BodyMetric, error)
}
