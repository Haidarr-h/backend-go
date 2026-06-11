package migrate

import (
	"fmt"
	"os"

	"github.com/Haidarr-h/backend-go/internal/auth"
	"github.com/Haidarr-h/backend-go/internal/exercise"
	"github.com/Haidarr-h/backend-go/internal/routine"
	"github.com/Haidarr-h/backend-go/internal/session"
	"github.com/Haidarr-h/backend-go/internal/user"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) {

	if os.Getenv("AUTO_MIGRATE") == "true" {

		err := db.AutoMigrate(
			&user.User{},
			&exercise.Exercise{},
			&routine.Routine{},
			&routine.RoutineExercise{},
			&routine.RoutineExerciseSet{},
			&session.Session{},
			&session.SessionExercise{},
			&session.SessionExerciseSet{},
			&auth.RefreshToken{},
			&auth.OtpVerification{},
		)

		if err != nil {
			fmt.Println("Auto migration error:", err)
		} else {
			fmt.Println("Auto migration success:")
		}

	}
}
