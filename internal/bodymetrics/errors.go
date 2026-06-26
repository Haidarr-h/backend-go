package bodymetrics

import "errors"

// ErrBodyMetricNotFound is returned when the user has no body metrics saved yet.
var ErrBodyMetricNotFound = errors.New("body metrics not found")
