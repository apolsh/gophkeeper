package app_errors

import "errors"

var (
	// ErrServerIsNotAvailable appears when server is not available.
	ErrServerIsNotAvailable = errors.New("server is not available")
)
