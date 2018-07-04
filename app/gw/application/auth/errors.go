package auth

import "errors"

// ErrInternal means that internal error is occured.
var ErrInternal = errors.New("internal error")

// ErrIncorrectKey means that the provided key is not correct.
var ErrIncorrectKey = errors.New("incorrect secret key")

// ErrNoSuchKey means that the provided key is not exist.
var ErrNoSuchKey = errors.New("no such key")
