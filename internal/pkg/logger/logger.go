package logger

import (
	"log"
	"os"
)

// Logger provides a simple wrapper for structured logging.
type Logger struct {
	*log.Logger
}

// NewLogger creates a new Logger instance.
func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "[EMAIL_SERVICE] ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Printf logs a formatted message.
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Logger.Printf(format, v...)
}

// Fatalf logs a formatted message and exits the application.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Logger.Fatalf(format, v...)
}

// Errorf logs a formatted error message.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Logger.Printf("[ERROR] "+format, v...)
}

// Warnf logs a formatted warning message.
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Logger.Printf("[WARN] "+format, v...)
}

// Println logs a message with a newline.
func (l *Logger) Println(v ...interface{}) {
	l.Logger.Println(v...)
}
