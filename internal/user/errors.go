package user

import "errors"

// user repo
var ErrUserNotFound = errors.New("user not found")
var ErrUserCreation = errors.New("failed to create user")
var ErrUserUpdate = errors.New("failed to update user")
