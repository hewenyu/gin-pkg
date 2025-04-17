package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Logger defines the interface for logging operations
type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
}

// Level represents a logging level
type Level int

const (
	// DebugLevel defines debug log level
	DebugLevel Level = iota
	// InfoLevel defines info log level
	InfoLevel
	// WarnLevel defines warn log level
	WarnLevel
	// ErrorLevel defines error log level
	ErrorLevel
	// FatalLevel defines fatal log level
	FatalLevel
)

// DefaultLogger implements Logger interface with standard log
type DefaultLogger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
	fatal *log.Logger
	level Level
}

// NewDefaultLogger creates a new logger instance
func NewDefaultLogger(out io.Writer, level Level) Logger {
	if out == nil {
		out = os.Stdout
	}

	return &DefaultLogger{
		debug: log.New(out, formatPrefix("DEBUG"), log.Ldate|log.Ltime|log.Lshortfile),
		info:  log.New(out, formatPrefix("INFO"), log.Ldate|log.Ltime|log.Lshortfile),
		warn:  log.New(out, formatPrefix("WARN"), log.Ldate|log.Ltime|log.Lshortfile),
		error: log.New(out, formatPrefix("ERROR"), log.Ldate|log.Ltime|log.Lshortfile),
		fatal: log.New(out, formatPrefix("FATAL"), log.Ldate|log.Ltime|log.Lshortfile),
		level: level,
	}
}

// formatPrefix formats a log level prefix
func formatPrefix(level string) string {
	return fmt.Sprintf("[%s] ", level)
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(v ...interface{}) {
	if l.level <= DebugLevel {
		l.debug.Output(2, fmt.Sprint(v...))
	}
}

// Debugf logs a formatted debug message
func (l *DefaultLogger) Debugf(format string, v ...interface{}) {
	if l.level <= DebugLevel {
		l.debug.Output(2, fmt.Sprintf(format, v...))
	}
}

// Info logs an info message
func (l *DefaultLogger) Info(v ...interface{}) {
	if l.level <= InfoLevel {
		l.info.Output(2, fmt.Sprint(v...))
	}
}

// Infof logs a formatted info message
func (l *DefaultLogger) Infof(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.info.Output(2, fmt.Sprintf(format, v...))
	}
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(v ...interface{}) {
	if l.level <= WarnLevel {
		l.warn.Output(2, fmt.Sprint(v...))
	}
}

// Warnf logs a formatted warning message
func (l *DefaultLogger) Warnf(format string, v ...interface{}) {
	if l.level <= WarnLevel {
		l.warn.Output(2, fmt.Sprintf(format, v...))
	}
}

// Error logs an error message
func (l *DefaultLogger) Error(v ...interface{}) {
	if l.level <= ErrorLevel {
		l.error.Output(2, fmt.Sprint(v...))
	}
}

// Errorf logs a formatted error message
func (l *DefaultLogger) Errorf(format string, v ...interface{}) {
	if l.level <= ErrorLevel {
		l.error.Output(2, fmt.Sprintf(format, v...))
	}
}

// Fatal logs a fatal message and exits the program
func (l *DefaultLogger) Fatal(v ...interface{}) {
	if l.level <= FatalLevel {
		l.fatal.Output(2, fmt.Sprint(v...))
		os.Exit(1)
	}
}

// Fatalf logs a formatted fatal message and exits the program
func (l *DefaultLogger) Fatalf(format string, v ...interface{}) {
	if l.level <= FatalLevel {
		l.fatal.Output(2, fmt.Sprintf(format, v...))
		os.Exit(1)
	}
}

// Global logger instance
var (
	std = NewDefaultLogger(os.Stdout, InfoLevel)
)

// SetDefaultLogger sets the default global logger
func SetDefaultLogger(logger Logger) {
	std = logger
}

// Debug logs a debug message using the default logger
func Debug(v ...interface{}) {
	std.Debug(v...)
}

// Debugf logs a formatted debug message using the default logger
func Debugf(format string, v ...interface{}) {
	std.Debugf(format, v...)
}

// Info logs an info message using the default logger
func Info(v ...interface{}) {
	std.Info(v...)
}

// Infof logs a formatted info message using the default logger
func Infof(format string, v ...interface{}) {
	std.Infof(format, v...)
}

// Warn logs a warning message using the default logger
func Warn(v ...interface{}) {
	std.Warn(v...)
}

// Warnf logs a formatted warning message using the default logger
func Warnf(format string, v ...interface{}) {
	std.Warnf(format, v...)
}

// Error logs an error message using the default logger
func Error(v ...interface{}) {
	std.Error(v...)
}

// Errorf logs a formatted error message using the default logger
func Errorf(format string, v ...interface{}) {
	std.Errorf(format, v...)
}

// Fatal logs a fatal message and exits the program using the default logger
func Fatal(v ...interface{}) {
	std.Fatal(v...)
}

// Fatalf logs a formatted fatal message and exits the program using the default logger
func Fatalf(format string, v ...interface{}) {
	std.Fatalf(format, v...)
}
