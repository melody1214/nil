package auth

import "errors"

var ErrInternal = errors.New("internal error")
var ErrIncorrectKey = errors.New("incorrect secret key")
var ErrNoSuchKey = errors.New("no such key")
