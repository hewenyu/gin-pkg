package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hewenyu/gin-pkg/internal/app"
	"github.com/hewenyu/gin-pkg/pkg/logger"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config/default.yaml", "path to configuration file")
	debugMode := flag.Bool("debug", false, "enable debug logging")
	logPath := flag.String("log", "logs/app.log", "path to log file")
	flag.Parse()

	// 设置日志级别
	logLevel := logger.InfoLevel
	if *debugMode {
		logLevel = logger.DebugLevel
	}

	// 获取可执行文件所在目录
	execDir, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get executable directory: %v\n", err)
		execDir = "."
	}

	// 构建日志文件的绝对路径
	logFilePath := *logPath
	if !filepath.IsAbs(logFilePath) {
		logFilePath = filepath.Join(filepath.Dir(execDir), logFilePath)
	}

	// 配置同时输出到控制台和文件的日志记录器
	log := logger.GetDualLogger(logFilePath, logLevel, *debugMode)
	logger.SetDefaultLogger(log)

	// 确保在程序退出时刷新日志缓冲
	defer logger.Sync()

	// 输出启动信息
	logger.Info("Starting application...")
	logger.Infof("Log level: %v, Debug mode: %v", logLevel, *debugMode)
	logger.Infof("Log file: %s", logFilePath)

	// Create new application
	application, err := app.NewApp(*configPath)
	if err != nil {
		logger.Fatalf("Failed to create application: %v", err)
	}

	// Initialize application
	if err := application.Initialize(); err != nil {
		logger.Fatalf("Failed to initialize application: %v", err)
	}

	// Ensure resources are cleaned up
	defer application.Cleanup()

	// Run application
	if err := application.Run(); err != nil {
		logger.Fatalf("Application error: %v", err)
	}
}
