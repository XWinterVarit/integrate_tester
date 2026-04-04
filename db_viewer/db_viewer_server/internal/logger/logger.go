package logger

import (
	"fmt"
	"log"
	"os"
)

// Logger is a lightweight structured logger that prefixes every line with a level and component tag.
type Logger struct {
	component string
	l         *log.Logger
}

// New creates a Logger for the given component name.
func New(component string) *Logger {
	return &Logger{
		component: component,
		l:         log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Info logs a message at INFO level.
func (l *Logger) Info(format string, args ...any) {
	l.l.Printf("[INFO]  [%s] %s", l.component, fmt.Sprintf(format, args...))
}

// Warn logs a message at WARN level.
func (l *Logger) Warn(format string, args ...any) {
	l.l.Printf("[WARN]  [%s] %s", l.component, fmt.Sprintf(format, args...))
}

// Error logs a message at ERROR level.
func (l *Logger) Error(format string, args ...any) {
	l.l.Printf("[ERROR] [%s] %s", l.component, fmt.Sprintf(format, args...))
}

// Fatal logs a message at FATAL level and exits the process.
func (l *Logger) Fatal(format string, args ...any) {
	l.l.Fatalf("[FATAL] [%s] %s", l.component, fmt.Sprintf(format, args...))
}
