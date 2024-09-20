package models

import "errors"

var ErrDuplicateEmail = errors.New("models: duplicate email")
var ErrDuplicateUsername = errors.New("models: duplicate username")
var ErrInvalidCredentials = errors.New("models: invalid credentials")
var ErrNoSessionsFound = errors.New("models: no sessions matching the specified token or id where found")
var ErrSessionExpired = errors.New("models: the session has expired")
var ErrInvalidSession = errors.New("models: session token is invalid")
var ErrRelationshipExists = errors.New("models: the users are already friends")
var ErrSameUser = errors.New("models: the user IDs are the same")
var ErrUserNotFound = errors.New("models: user not found")
var ErrMaxFriends = errors.New("models: the user has reached the maximum number of friends")
var ErrRecipientHasBlockedUser = errors.New("models: the recipient has blocked the client user")
var ErrNothingToUpdate = errors.New("models: nothing to update")
