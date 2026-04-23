package repositories

import "errors"

// user repo
var ErrUserNotFound = errors.New("user not found")

// refresh token repo
var ErrGenerateToken = errors.New("failed to insert refresh token data")