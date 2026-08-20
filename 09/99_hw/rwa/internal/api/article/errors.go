package article

import "errors"

var ErrIncorrectPassword = errors.New("incorrect password")
var ErrSessionProblems = errors.New("session problems")
