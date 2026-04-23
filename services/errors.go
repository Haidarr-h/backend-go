package services

import "errors"

// auth - sign up
var ErrEmailUsernameExists = errors.New("email or username already exist")

// auth - sign in
var ErrInvalidCredentials = errors.New("invalid email or username or password")
var ErrUserGoogleSignIn = errors.New("user uses different way to sign in")
var ErrFailedCreateToken = errors.New("failed to create token")
