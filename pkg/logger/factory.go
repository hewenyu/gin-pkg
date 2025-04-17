package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 日志输出目标类型
type LogOutput int

const (
	// 控制台输出
	ConsoleOutput LogOutput = iota
	// 文件输出
	FileOutput
	// 同时输出到控制台和文件
	BothOutput
)

// LogConfig 日志配置
type LogConfig struct {
	// 日志级别
	Level Level
	// 是否为开发模式
	Development bool
	// 日志输出目标
	Output LogOutput
	// 日志文件路径（当输出类型包含文件时使用）
	FilePath string
	// 最大日志文件大小，单位MB
	MaxSize int
	// 保留的旧日志文件最大数量
	MaxBackups int
	// 保留的日志文件最大天数
	MaxAge int
	// 是否压缩旧日志文件
	Compress bool
}

// DefaultLogConfig 返回默认的日志配置
func DefaultLogConfig() LogConfig {
	return LogConfig{
		Level:       InfoLevel,
		Development: false,
		Output:      ConsoleOutput,
		FilePath:    "logs/app.log",
		MaxSize:     100,
		MaxBackups:  3,
		MaxAge:      28,
		Compress:    true,
	}
}

// NewLogger 创建一个新的日志记录器
func NewLogger(config LogConfig) Logger {
	// 构建编码器配置
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

	// 创建不同的编码器用于控制台和文件
	var consoleEncoder, fileEncoder zapcore.Encoder

	// 开发模式使用控制台编码器，生产模式使用JSON编码器
	if config.Development {
		consoleEncoder = zapcore.NewConsoleEncoder(encoderConfig)
		fileEncoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		consoleEncoder = zapcore.NewConsoleEncoder(encoderConfig)
		fileEncoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 设置日志级别
	zapLevel := toZapLevel(config.Level)

	// 创建输出
	var cores []zapcore.Core

	// 控制台输出
	if config.Output == ConsoleOutput || config.Output == BothOutput {
		cores = append(cores, zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			zapLevel,
		))
	}

	// 文件输出
	if config.Output == FileOutput || config.Output == BothOutput {
		// 确保日志目录存在
		logDir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			// 如果无法创建目录，回退到控制台日志
			cores = []zapcore.Core{
				zapcore.NewCore(
					consoleEncoder,
					zapcore.AddSync(os.Stdout),
					zapLevel,
				),
			}

			// 输出错误信息
			consoleLog := zap.New(
				zapcore.NewCore(
					consoleEncoder,
					zapcore.AddSync(os.Stderr),
					zapcore.ErrorLevel,
				),
			)
			consoleLog.Error("无法创建日志目录", zap.String("dir", logDir), zap.Error(err))
			consoleLog.Sync()
		} else {
			// 创建文件输出
			fileWriter := zapcore.AddSync(&lumberjack.Logger{
				Filename:   config.FilePath,
				MaxSize:    config.MaxSize,
				MaxBackups: config.MaxBackups,
				MaxAge:     config.MaxAge,
				Compress:   config.Compress,
			})

			cores = append(cores, zapcore.NewCore(
				fileEncoder,
				fileWriter,
				zapLevel,
			))
		}
	}

	// 组合多个输出
	core := zapcore.NewTee(cores...)

	// 创建zap日志记录器
	var zapOptions []zap.Option
	zapOptions = append(zapOptions, zap.AddCaller(), zap.AddCallerSkip(2))

	// 开发模式下添加堆栈跟踪
	if config.Development {
		zapOptions = append(zapOptions, zap.Development())
	}

	// 初始化zap日志记录器
	zapLogger := zap.New(core, zapOptions...)

	// 返回我们的日志记录器
	return &ZapLogger{
		logger: zapLogger,
		sugar:  zapLogger.Sugar(),
		level:  config.Level,
	}
}

// GetFileLogger 创建一个文件日志记录器
func GetFileLogger(filePath string, level Level) Logger {
	config := DefaultLogConfig()
	config.FilePath = filePath
	config.Level = level
	config.Output = FileOutput
	return NewLogger(config)
}

// GetConsoleLogger 创建一个控制台日志记录器
func GetConsoleLogger(level Level, development bool) Logger {
	config := DefaultLogConfig()
	config.Level = level
	config.Development = development
	config.Output = ConsoleOutput
	return NewLogger(config)
}

// GetDualLogger 创建一个同时输出到控制台和文件的日志记录器
func GetDualLogger(filePath string, level Level, development bool) Logger {
	config := DefaultLogConfig()
	config.FilePath = filePath
	config.Level = level
	config.Development = development
	config.Output = BothOutput
	return NewLogger(config)
}
