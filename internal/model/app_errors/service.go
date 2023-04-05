package app_errors

import "errors"

var (
	// ErrorEmptyValue error when empty values are not allowed.
	ErrorEmptyValue = errors.New("empty values are not allowed")
	// ErrorInvalidPassword  error when password is invalid.
	ErrorInvalidPassword = errors.New("invalid password")
)
