package stats

import (
	"time"

	"gorm.io/gorm"
)

type StatsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// CareerTotals carries the raw lifetime aggregates before level/XP is derived.
type CareerTotals struct {
	TotalVolumeKg float64
	TotalReps     int
	TotalSets     int
	TotalSessions int
}

// CareerTotals sums volume (reps*weight), reps, sets and counts sessions across
// every non-deleted session the user owns. LEFT JOINs keep empty sessions in the
// session count, and COALESCE keeps the sums at 0 for a brand-new user.
func (r *StatsRepository) CareerTotals(userID uint) (CareerTotals, error) {
	var totals CareerTotals

	result := r.db.Raw(`
		SELECT
			COALESCE(SUM(ses.reps * ses.weight_kg), 0) AS total_volume_kg,
			COALESCE(SUM(ses.reps), 0)                  AS total_reps,
			COUNT(ses.id)                               AS total_sets,
			COUNT(DISTINCT s.id)                        AS total_sessions
		FROM sessions s
		LEFT JOIN session_exercises se
			ON se.session_id = s.id AND se.deleted_at IS NULL
		LEFT JOIN session_exercise_sets ses
			ON ses.session_exercise_id = se.id AND ses.deleted_at IS NULL
		WHERE s.user_id = ? AND s.deleted_at IS NULL
	`, userID).Scan(&totals)

	if result.Error != nil {
		return CareerTotals{}, result.Error
	}

	return totals, nil
}

// PersonalRecords returns the heaviest single set and best single-set volume per
// exercise the user has performed, ordered by heaviest weight first.
func (r *StatsRepository) PersonalRecords(userID uint) ([]PersonalRecordResponse, error) {
	var records []PersonalRecordResponse

	result := r.db.Raw(`
		SELECT
			se.exercise_id              AS exercise_id,
			e.name                      AS exercise_name,
			MAX(ses.weight_kg)          AS max_weight_kg,
			MAX(ses.reps * ses.weight_kg) AS best_set_volume_kg
		FROM sessions s
		JOIN session_exercises se
			ON se.session_id = s.id AND se.deleted_at IS NULL
		JOIN session_exercise_sets ses
			ON ses.session_exercise_id = se.id AND ses.deleted_at IS NULL
		JOIN exercises e
			ON e.id = se.exercise_id
		WHERE s.user_id = ? AND s.deleted_at IS NULL
		GROUP BY se.exercise_id, e.name
		ORDER BY max_weight_kg DESC
	`, userID).Scan(&records)

	if result.Error != nil {
		return nil, result.Error
	}

	return records, nil
}

// VolumeOverTime buckets volume/reps/sessions by the given date_trunc unit
// ("week" or "month"), oldest bucket first. The caller validates trunc.
func (r *StatsRepository) VolumeOverTime(userID uint, trunc string) ([]VolumePointResponse, error) {
	var points []VolumePointResponse

	result := r.db.Raw(`
		SELECT
			date_trunc(?, s.start_time)                 AS period,
			COALESCE(SUM(ses.reps * ses.weight_kg), 0)  AS volume_kg,
			COALESCE(SUM(ses.reps), 0)                  AS reps,
			COUNT(DISTINCT s.id)                        AS sessions
		FROM sessions s
		LEFT JOIN session_exercises se
			ON se.session_id = s.id AND se.deleted_at IS NULL
		LEFT JOIN session_exercise_sets ses
			ON ses.session_exercise_id = se.id AND ses.deleted_at IS NULL
		WHERE s.user_id = ? AND s.deleted_at IS NULL
		GROUP BY period
		ORDER BY period
	`, trunc, userID).Scan(&points)

	if result.Error != nil {
		return nil, result.Error
	}

	return points, nil
}

// WorkoutDates returns the distinct calendar dates (ascending) on which the user
// started a session. Streak math is done in the service from this list.
func (r *StatsRepository) WorkoutDates(userID uint) ([]time.Time, error) {
	var dates []time.Time

	result := r.db.Raw(`
		SELECT DISTINCT date(s.start_time) AS d
		FROM sessions s
		WHERE s.user_id = ? AND s.deleted_at IS NULL
		ORDER BY d
	`, userID).Scan(&dates)

	if result.Error != nil {
		return nil, result.Error
	}

	return dates, nil
}

// MuscleBreakdown sums volume per primary muscle, unnesting the exercise's
// primary_muscles array, ordered by most-trained muscle first.
func (r *StatsRepository) MuscleBreakdown(userID uint) ([]MuscleVolumeResponse, error) {
	var muscles []MuscleVolumeResponse

	result := r.db.Raw(`
		SELECT
			muscle                                      AS muscle,
			COALESCE(SUM(ses.reps * ses.weight_kg), 0)  AS volume_kg
		FROM sessions s
		JOIN session_exercises se
			ON se.session_id = s.id AND se.deleted_at IS NULL
		JOIN session_exercise_sets ses
			ON ses.session_exercise_id = se.id AND ses.deleted_at IS NULL
		JOIN exercises e
			ON e.id = se.exercise_id
		CROSS JOIN LATERAL unnest(e.primary_muscles) AS muscle
		WHERE s.user_id = ? AND s.deleted_at IS NULL
		GROUP BY muscle
		ORDER BY volume_kg DESC
	`, userID).Scan(&muscles)

	if result.Error != nil {
		return nil, result.Error
	}

	return muscles, nil
}
