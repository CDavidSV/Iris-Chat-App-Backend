package models

import "errors"

var ErrDuplicateEmail = errors.New("models: duplicate email")
var ErrDuplicateUsername = errors.New("models: duplicate username")
var ErrInvalidCredentials = errors.New("models: invalid credentials")
var ErrNoSessionsFound = errors.New("models: no sessions matching the specified token or id where found")
var ErrSessionExpired = errors.New("models: the session has expired")
var ErrInvalidSession = errors.New("models: session token is invalid")
