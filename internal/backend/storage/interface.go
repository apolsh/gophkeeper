package storage

import (
	"errors"

	"github.com/apolsh/yapr-gophkeeper/internal/logger"
)

var log = logger.LoggerOfComponent("storage")

var (
	// ErrorLoginIsAlreadyUsed used when login is already used.
	ErrorLoginIsAlreadyUsed = errors.New("login is already used")
	// ErrItemNotFound used when requested element not found.
	ErrItemNotFound = errors.New("requested element not found")
	// ErrUnknownDatabase used when unexpected database error appears.
	ErrUnknownDatabase = errors.New("unknown database error")
)

// HandleUnknownDatabaseError handle unexpected database error.
func HandleUnknownDatabaseError(err error) error {
	log.Error(err)
	return ErrUnknownDatabase
}
