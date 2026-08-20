package user

import "errors"

var ErrIncorrectPassword = errors.New("incorrect password")
var ErrSessionProblems = errors.New("session problems")
