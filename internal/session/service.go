package session

import (
	"errors"
	"sort"

	"github.com/Haidarr-h/backend-go/internal/exercise"
	"github.com/Haidarr-h/backend-go/internal/routine"
	"gorm.io/gorm"
)

type SessionService struct {
	sessionRepo Repo
	routineRepo routine.Repo
}

func NewSessionService(sessionRepo Repo, routineRepo routine.Repo) *SessionService {
	return &SessionService{sessionRepo: sessionRepo, routineRepo: routineRepo}
}

// CREATE
// Snapshots the chosen routine's exercises + sets into a new session.
func (s *SessionService) CreateSession(userID uint, req CreateSessionRequest) (SessionResponse, error) {

	// 1. validate the time range
	if req.EndTime.Before(req.StartTime) {
		return SessionResponse{}, ErrInvalidTimeRange
	}

	// 2. load the source routine
	sourceRoutine, err := s.routineRepo.FindByID(req.RoutineID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SessionResponse{}, ErrRoutineNotFound
		}
		return SessionResponse{}, err
	}

	// 3. access check: only the owner or a public routine can be started
	if !sourceRoutine.IsPublic && (sourceRoutine.UserID == nil || *sourceRoutine.UserID != userID) {
		return SessionResponse{}, ErrRoutineNotAccessible
	}

	// 4. build the session, snapshotting the routine's exercise tree
	name := req.Name
	if name == "" {
		name = sourceRoutine.Name
	}

	routineID := req.RoutineID
	session := Session{
		UserID:           userID,
		RoutineID:        &routineID,
		Name:             name,
		Notes:            req.Notes,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		SessionExercises: buildSessionExercises(sourceRoutine.RoutineExercises),
	}

	// 5. persist + map
	created, err := s.sessionRepo.Create(session)
	if err != nil {
		return SessionResponse{}, err
	}

	return mapToSessionResponse(created), nil
}

// GET ALL
func (s *SessionService) GetSessions(userID uint) ([]SessionResponse, error) {
	sessions, err := s.sessionRepo.FindAll(userID)
	if err != nil {
		return nil, err
	}

	// non-nil so the API returns [] not null
	response := make([]SessionResponse, 0, len(sessions))
	for _, sess := range sessions {
		response = append(response, mapToSessionResponse(sess))
	}

	return response, nil
}

// GET ONE
func (s *SessionService) GetSession(userID uint, sessionID uint) (SessionResponse, error) {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SessionResponse{}, ErrSessionNotFound
		}
		return SessionResponse{}, err
	}

	if session.UserID != userID {
		return SessionResponse{}, InvalidSessionOwnership
	}

	return mapToSessionResponse(session), nil
}

// UPDATE (non-exercise fields only)
func (s *SessionService) UpdateSession(id, userID uint, req UpdateSessionReq) (SessionResponse, error) {
	// 1. fetch existing
	existing, err := s.sessionRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SessionResponse{}, ErrSessionNotFound
		}
		return SessionResponse{}, err
	}

	// 2. ownership
	if existing.UserID != userID {
		return SessionResponse{}, InvalidSessionOwnership
	}

	// 3. apply only the provided fields
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}
	if req.StartTime != nil {
		existing.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		existing.EndTime = *req.EndTime
	}

	// 4. validate the resulting time range
	if existing.EndTime.Before(existing.StartTime) {
		return SessionResponse{}, ErrInvalidTimeRange
	}

	// 5. persist + map
	updated, err := s.sessionRepo.Update(existing)
	if err != nil {
		return SessionResponse{}, err
	}

	return mapToSessionResponse(updated), nil
}

// DELETE
func (s *SessionService) DeleteSession(sessionID uint, userID uint) error {
	err := s.sessionRepo.Delete(sessionID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		return err
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildSessionExercises copies a routine's exercise tree into session models.
// Primary keys / foreign keys are left zero so GORM inserts fresh rows.
func buildSessionExercises(routineExercises []routine.RoutineExercise) []SessionExercise {
	var exercises []SessionExercise

	for _, re := range routineExercises {
		se := SessionExercise{
			ExerciseID: re.ExerciseID,
			Order:      re.Order,
			RestSecond: re.RestSecond,
		}

		for _, set := range re.RoutineExerciseSets {
			se.SessionExerciseSets = append(se.SessionExerciseSets, SessionExerciseSet{
				SetNumber: set.SetNumber,
				Reps:      set.Reps,
				WeightKG:  set.WeightKG,
			})
		}

		exercises = append(exercises, se)
	}

	return exercises
}

// computeMetrics derives the aggregate stats shown when reading a session.
func computeMetrics(s Session) SessionMetrics {
	var metrics SessionMetrics

	muscleCount := map[string]int{}

	for _, se := range s.SessionExercises {
		for _, muscle := range se.Exercise.PrimaryMuscles {
			muscleCount[muscle]++
		}

		for _, set := range se.SessionExerciseSets {
			metrics.TotalSets++
			metrics.TotalReps += set.Reps
			metrics.TotalWeight += float64(set.Reps) * set.WeightKG
		}
	}

	// duration, clamped to >= 0
	if seconds := int(s.EndTime.Sub(s.StartTime).Seconds()); seconds > 0 {
		metrics.TotalDurationSeconds = seconds
	}

	metrics.TargetedMuscles = rankMuscles(muscleCount)

	return metrics
}

// rankMuscles orders muscles by frequency desc, then name asc for stability.
func rankMuscles(muscleCount map[string]int) []string {
	muscles := make([]string, 0, len(muscleCount))
	for muscle := range muscleCount {
		muscles = append(muscles, muscle)
	}

	sort.Slice(muscles, func(i, j int) bool {
		if muscleCount[muscles[i]] != muscleCount[muscles[j]] {
			return muscleCount[muscles[i]] > muscleCount[muscles[j]]
		}
		return muscles[i] < muscles[j]
	})

	return muscles
}

// mapToSessionResponse maps the model tree to the response DTO (with metrics).
func mapToSessionResponse(s Session) SessionResponse {
	var exercises []SessionExerciseResponse

	for _, se := range s.SessionExercises {
		var sets []SessionExerciseSetResponse
		for _, set := range se.SessionExerciseSets {
			sets = append(sets, SessionExerciseSetResponse{
				ID:        set.ID,
				SetNumber: set.SetNumber,
				Reps:      set.Reps,
				WeightKG:  set.WeightKG,
			})
		}

		exercises = append(exercises, SessionExerciseResponse{
			ID:         se.ID,
			ExerciseID: se.ExerciseID,
			Order:      se.Order,
			RestSecond: se.RestSecond,
			Sets:       sets,
			Exercise: exercise.ExerciseResponse{
				ID:               se.Exercise.ID,
				Name:             se.Exercise.Name,
				Force:            se.Exercise.Force,
				Level:            se.Exercise.Level,
				Mechanic:         se.Exercise.Mechanic,
				Equipment:        se.Exercise.Equipment,
				PrimaryMuscles:   se.Exercise.PrimaryMuscles,
				SecondaryMuscles: se.Exercise.SecondaryMuscles,
				Instructions:     se.Exercise.Instructions,
				Images:           se.Exercise.Images,
				Category:         se.Exercise.Category,
				VideoURL:         se.Exercise.VideoURL,
			},
		})
	}

	return SessionResponse{
		ID:               s.ID,
		RoutineID:        s.RoutineID,
		Name:             s.Name,
		Notes:            s.Notes,
		StartTime:        s.StartTime,
		EndTime:          s.EndTime,
		Metrics:          computeMetrics(s),
		SessionExercises: exercises,
		CreatedAt:        s.CreatedAt,
	}
}
