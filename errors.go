package goinsta

import "errors"

// ErrNotFound is returned if the request responds with a 404 status code
// i.e a non existent user
var ErrNotFound = errors.New("The specified data wasn't found.")

// ErrLoggedOut is returned if the request responds with a 400 status code
var ErrLoggedOut = errors.New("The account is logged out")

var ErrBadPassword = errors.New("The Instagram password you provided is incorrect")

var ErrChallenge = errors.New("Instagram needs you to authorize our system logging into your account. Please check your email for a challenge link we have sent you or login to your instagram app and click “This is me”")
var ErrPrivate = errors.New("The account is not viewable")
