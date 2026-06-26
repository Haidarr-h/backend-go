package stats

import "errors"

// ErrInvalidPeriod is returned when the volume-over-time period is not week or month.
var ErrInvalidPeriod = errors.New("invalid period: must be 'week' or 'month'")
