package routine

import "errors"

var ErrCreateRoutine = errors.New("failed to create routine")
var ErrExerciseNotFound = errors.New("exercise not found")
var InvalidRoutineOwnership = errors.New("this routine cannot be accessed by public")
var ErrRoutineNotFound = errors.New("routine not found")