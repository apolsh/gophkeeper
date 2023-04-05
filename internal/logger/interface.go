package logger

// ComponentKey application component.
const ComponentKey = "component"

// Interface logger interface.
type Interface interface {
	// Debug write message with debug level.
	Debug(message string, args ...interface{})
	// Info write message with level info.
	Info(message string, args ...interface{})
	// Warn write message with warn level.
	Warn(message string, args ...interface{})
	// Error write message with error level.
	Error(err error)
	// Fatal write message with fatal level and exit.
	Fatal(err error)
}
