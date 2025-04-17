package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	// 增加sync方法用于刷新缓冲日志
	Sync() error
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

// ZapLogger implements Logger interface with zap
type ZapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	level  Level
}

// toZapLevel converts our Level to zapcore.Level
func toZapLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// NewZapLogger creates a new logger instance using zap
func NewZapLogger(level Level, development bool) Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	zapLevel := toZapLevel(level)

	var core zapcore.Core
	if development {
		// 在开发模式下使用控制台输出
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		core = zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapLevel)
	} else {
		// 在生产模式下使用JSON输出
		jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)
		core = zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), zapLevel)
	}

	// 添加调用者信息
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))

	return &ZapLogger{
		logger: logger,
		sugar:  logger.Sugar(),
		level:  level,
	}
}

// Debug logs a debug message
func (l *ZapLogger) Debug(v ...interface{}) {
	l.sugar.Debug(v...)
}

// Debugf logs a formatted debug message
func (l *ZapLogger) Debugf(format string, v ...interface{}) {
	l.sugar.Debugf(format, v...)
}

// Info logs an info message
func (l *ZapLogger) Info(v ...interface{}) {
	l.sugar.Info(v...)
}

// Infof logs a formatted info message
func (l *ZapLogger) Infof(format string, v ...interface{}) {
	l.sugar.Infof(format, v...)
}

// Warn logs a warning message
func (l *ZapLogger) Warn(v ...interface{}) {
	l.sugar.Warn(v...)
}

// Warnf logs a formatted warning message
func (l *ZapLogger) Warnf(format string, v ...interface{}) {
	l.sugar.Warnf(format, v...)
}

// Error logs an error message
func (l *ZapLogger) Error(v ...interface{}) {
	l.sugar.Error(v...)
}

// Errorf logs a formatted error message
func (l *ZapLogger) Errorf(format string, v ...interface{}) {
	l.sugar.Errorf(format, v...)
}

// Fatal logs a fatal message and exits the program
func (l *ZapLogger) Fatal(v ...interface{}) {
	l.sugar.Fatal(v...)
}

// Fatalf logs a formatted fatal message and exits the program
func (l *ZapLogger) Fatalf(format string, v ...interface{}) {
	l.sugar.Fatalf(format, v...)
}

// Sync flushes any buffered log entries
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

// For compatibility with the original logger
// DefaultLogger is now an alias for ZapLogger
type DefaultLogger = ZapLogger

// NewDefaultLogger creates a new logger instance with console output
func NewDefaultLogger(out zapcore.WriteSyncer, level Level) Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if out == nil {
		out = zapcore.AddSync(os.Stdout)
	}

	zapLevel := toZapLevel(level)

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		out,
		zapLevel,
	)

	// 添加调用者信息
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))

	return &ZapLogger{
		logger: logger,
		sugar:  logger.Sugar(),
		level:  level,
	}
}

// Global logger instance
var (
	std = NewZapLogger(InfoLevel, true)
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

// Sync flushes any buffered log entries
func Sync() error {
	return std.Sync()
}
