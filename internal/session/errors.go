package session

import "errors"

var ErrSessionNotFound = errors.New("session not found")
var InvalidSessionOwnership = errors.New("this session cannot be accessed by this account")
var ErrRoutineNotFound = errors.New("routine not found")
var ErrRoutineNotAccessible = errors.New("this routine cannot be used by this account")
var ErrInvalidTimeRange = errors.New("end time must not be before start time")
var ErrCreateSession = errors.New("failed to create session")
