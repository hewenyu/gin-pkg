package logger

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinLoggerMiddleware 返回一个Gin中间件，使用我们的zap日志记录请求
func GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		// 请求路径
		path := c.Request.URL.Path
		// 请求方法
		method := c.Request.Method
		// 客户端IP
		clientIP := c.ClientIP()
		// 请求ID（如果有）
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = c.GetHeader("Request-ID")
		}

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		// 延迟时间
		latency := end.Sub(start)
		// 状态码
		statusCode := c.Writer.Status()
		// 错误信息
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 构建日志字段
		fields := []zap.Field{
			zap.String("client_ip", clientIP),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
		}

		// 添加请求ID（如果有）
		if requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		// 添加错误信息（如果有）
		if errorMessage != "" {
			fields = append(fields, zap.String("error", errorMessage))
		}

		// 根据状态码选择日志级别
		var logFunc func(msg string, fields ...zap.Field)

		// 获取zap记录器
		zapLogger := std.(*ZapLogger).logger

		switch {
		case statusCode >= http.StatusInternalServerError:
			logFunc = zapLogger.Error
		case statusCode >= http.StatusBadRequest:
			logFunc = zapLogger.Warn
		default:
			logFunc = zapLogger.Info
		}

		// 记录请求日志
		logFunc("HTTP Request", fields...)
	}
}

// WithZapLogger 配置Gin使用zap进行日志记录
func WithZapLogger(r *gin.Engine) {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 禁用Gin的控制台颜色
	gin.DisableConsoleColor()

	// 使用我们的zap日志中间件
	r.Use(GinLoggerMiddleware())
}

// GetGinEngine 创建一个配置了zap日志的Gin引擎
func GetGinEngine() *gin.Engine {
	// 创建一个默认的gin引擎但不包括默认的Logger和Recovery中间件
	r := gin.New()

	// 使用我们自己的Logger
	r.Use(GinLoggerMiddleware())

	// 使用Gin的Recovery中间件来捕获panic
	r.Use(gin.Recovery())

	return r
}
