package repositories

import "errors"

// user repo
var ErrUserNotFound = errors.New("user not found")
var ErrUserCreation = errors.New("failed to create user")

// refresh token repo
var ErrGenerateToken = errors.New("failed to insert refresh token data")