package goinsta

import "errors"

// ErrNotFound is returned if the request responds with a 404 status code
// i.e a non existent user
var ErrNotFound = errors.New("The specified data wasn't found.")

// ErrChallengeRequired is returned if Instagram requires a challenge code
var ErrChallengeOptionsRequired = errors.New("Account Locked. Need to verify account via Email or SMS. Pass goinsta.Challenge struct with goinsta.New()")

// ErrChallengeCodeRequired is returned when the client requires a goinsta.Challenge struct with code
var ErrChallengeCodeRequired = errors.New("Account Locked. You need to provide a challenge code.")

// ErrChallengeCodeInvalid is returned when the client passes a challenge code and it is not accepted by Instagram
var ErrChallengeCodeInvalid = errors.New("Account Locked. The challenge code provided is not valid for this user. Please check email/SMS for the correct 6-digit code.")
