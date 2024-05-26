package errors

import "errors"

var (
	ErrTypeAssertionFailed = errors.New("type assertion failed")
	ErrNoAffectedRows      = errors.New("no affected rows")
)
