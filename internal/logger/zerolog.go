package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// SetGlobalLevel set global logging level.
func SetGlobalLevel(level string) {
	var lvl zerolog.Level

	switch strings.ToLower(level) {
	case "error":
		lvl = zerolog.ErrorLevel
	case "warn":
		lvl = zerolog.WarnLevel
	case "info":
		lvl = zerolog.InfoLevel
	case "debug":
		lvl = zerolog.DebugLevel
	default:
		lvl = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(lvl)
}

// Logger zerolog logger.
type Logger struct {
	logger *zerolog.Logger
}

// Debug write message with debug level.
func (l *Logger) Debug(message string, args ...interface{}) {
	l.logger.Debug().Msgf(message, args...)
}

// Info write message with level info.
func (l *Logger) Info(message string, args ...interface{}) {
	l.logger.Info().Msgf(message, args...)
}

// Warn write message with warn level.
func (l *Logger) Warn(message string, args ...interface{}) {
	l.logger.Warn().Msgf(message, args...)
}

// Error write message with error level.
func (l *Logger) Error(err error) {
	l.logger.Error().Stack().Err(err).Msg("")
}

// Fatal write message with fatal level and exit.
func (l *Logger) Fatal(err error) {
	l.logger.Error().Stack().Err(err).Msg("")
	os.Exit(1)
}

// LoggerOfComponent get component logger.
func LoggerOfComponent(component string) Interface {
	logger := log.With().Str(ComponentKey, component).Logger()
	return &Logger{logger: &logger}

}
