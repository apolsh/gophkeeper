package backend_client

import "errors"

var (
	// ErrServerIsNotAvailable appears when server is not available
	ErrServerIsNotAvailable = errors.New("server is not available")
)
