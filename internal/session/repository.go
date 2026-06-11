package session

import (
	"github.com/Haidarr-h/backend-go/pkg/logger"
	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// CREATE
// Inserts the session and its snapshotted exercises + sets. GORM cascades the
// nested creates from the association slices, so a single Create handles the
// whole tree inside one transaction.
func (r *SessionRepository) Create(session Session) (Session, error) {

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&session).Error; err != nil {
			logger.Log.Error("create session failed", "error", err)
			return ErrCreateSession
		}
		return nil
	})

	if err != nil {
		return Session{}, err
	}

	// re-fetch so the response carries the preloaded exercise + sets
	return r.FindByID(session.ID)
}

// READ ALL
func (r *SessionRepository) FindAll(userID uint) ([]Session, error) {
	var sessions []Session

	result := r.db.Where("user_id = ?", userID).
		Preload("SessionExercises.Exercise").
		Preload("SessionExercises.SessionExerciseSets").
		Order("start_time desc").
		Find(&sessions)

	if result.Error != nil {
		return nil, result.Error
	}

	return sessions, nil
}

// READ ONE
func (r *SessionRepository) FindByID(id uint) (Session, error) {
	var session Session

	result := r.db.Where("id = ?", id).
		Preload("SessionExercises.Exercise").
		Preload("SessionExercises.SessionExerciseSets").
		First(&session)

	if result.Error != nil {
		return Session{}, result.Error
	}

	return session, nil
}

// UPDATE
// Only the session's own columns change; the snapshotted exercises are immutable.
func (r *SessionRepository) Update(session Session) (Session, error) {

	err := r.db.Model(&Session{}).
		Where("id = ?", session.ID).
		Updates(map[string]interface{}{
			"name":       session.Name,
			"notes":      session.Notes,
			"start_time": session.StartTime,
			"end_time":   session.EndTime,
		}).Error

	if err != nil {
		return Session{}, err
	}

	return r.FindByID(session.ID)
}

// DELETE
func (r *SessionRepository) Delete(id uint, userID uint) error {

	// transaction: delete the tree bottom-up so FK constraints hold.
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. collect the exercise ids so we can clean up their sets
		var exerciseIDs []uint
		if err := tx.Model(&SessionExercise{}).
			Where("session_id = ?", id).
			Pluck("id", &exerciseIDs).Error; err != nil {
			return err
		}

		// 2. delete the sets (grandchildren)
		if len(exerciseIDs) > 0 {
			if err := tx.Where("session_exercise_id IN ?", exerciseIDs).
				Delete(&SessionExerciseSet{}).Error; err != nil {
				return err
			}
		}

		// 3. delete the session exercises (children)
		if err := tx.Where("session_id = ?", id).Delete(&SessionExercise{}).Error; err != nil {
			return err
		}

		// 4. delete the session itself, scoped to the owner
		result := tx.Where("id = ? AND user_id = ?", id, userID).Delete(&Session{})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}
