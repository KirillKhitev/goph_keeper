package errs

import "errors"

// Ошибки.
var (
	ErrNotFound     = errors.New("not found")
	ErrAlreadyExist = errors.New("already exist")
)
