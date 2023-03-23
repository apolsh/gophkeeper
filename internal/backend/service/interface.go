package service

import "errors"

var (
	// ErrorEmptyValue error when empty values is not allowed
	ErrorEmptyValue = errors.New("empty values is not allowed")
	// ErrorInvalidPassword  error when password is invalid
	ErrorInvalidPassword = errors.New("invalid password")
)
