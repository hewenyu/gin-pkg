package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewenyu/gin-pkg/pkg/auth/security"
	"github.com/hewenyu/gin-pkg/pkg/logger"
)

// SecurityMiddleware validates request timestamps, nonces, and signatures
func SecurityMiddleware(securityService security.SecurityService, timestampWindow time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("【请求签名验证】-------------------------开始验证-------------------------")

		// Extract parameters (from headers or query params)
		timestamp := getParameter(c, "timestamp", "X-Timestamp")
		nonce := getParameter(c, "nonce", "X-Nonce")
		signature := getParameter(c, "sign", "X-Sign")

		if signature == "" {
			logger.Info("【请求签名验证】未找到签名参数")
		} else {
			logger.Infof("【请求签名验证】接收到的签名: %s", signature)
		}

		// Skip validation for the nonce endpoint
		if c.FullPath() == "/api/v1/auth/nonce" {
			// For nonce endpoint, only validate timestamp
			if timestamp == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "timestamp is required"})
				c.Abort()
				logger.Info("【请求签名验证】-------------------------结束验证-------------------------")
				return
			}

			if err := securityService.ValidateTimestamp(timestamp, timestampWindow); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				logger.Info("【请求签名验证】-------------------------结束验证-------------------------")
				return
			}

			c.Next()
			logger.Info("【请求签名验证】-------------------------结束验证-------------------------")
			return
		}

		// For all other endpoints, validate all security parameters
		if timestamp == "" || nonce == "" || signature == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "timestamp, nonce, and signature are required"})
			c.Abort()
			logger.Info("【请求签名验证】-------------------------结束验证-------------------------")
			return
		}

		// Validate timestamp
		if err := securityService.ValidateTimestamp(timestamp, timestampWindow); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			logger.Info("【请求签名验证】-------------------------结束验证-------------------------")
			return
		}

		// Validate nonce
		if err := securityService.ValidateNonce(nonce); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			logger.Info("【请求签名验证】-------------------------结束验证-------------------------")
			return
		}

		// Build parameters map for signature validation
		params := make(map[string]string)

		// Add query parameters
		logger.Info("【请求签名验证】收集URL查询参数")
		for k, v := range c.Request.URL.Query() {
			if len(v) > 0 && k != "sign" {
				params[k] = v[0]
				logger.Infof("【请求签名验证】URL参数: %s = %s", k, v[0])
			}
		}

		// Add form parameters if POST/PUT/PATCH with form data
		if c.Request.Method != http.MethodGet {
			if err := c.Request.ParseForm(); err == nil {
				logger.Info("【请求签名验证】收集表单参数")
				for k, v := range c.Request.PostForm {
					if len(v) > 0 && k != "sign" {
						params[k] = v[0]
						logger.Infof("【请求签名验证】表单参数: %s = %s", k, v[0])
					}
				}
			}
		}

		// 为非GET请求尝试从JSON请求体中获取参数
		if c.Request.Method != http.MethodGet && c.Request.Header.Get("Content-Type") == "application/json" {
			// 保存请求体
			requestBody, err := c.GetRawData()
			if err == nil && len(requestBody) > 0 {
				// 记录原始请求体
				logger.Infof("【请求签名验证】请求体原始数据: %s", string(requestBody))

				// 重新设置请求体，以便后续处理
				c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

				// 解析JSON请求体
				var bodyMap map[string]interface{}
				if err := json.Unmarshal(requestBody, &bodyMap); err == nil {
					logger.Info("【请求签名验证】JSON解析成功，提取参数")
					// 将JSON请求体中的参数添加到签名验证参数中
					for k, v := range bodyMap {
						// 记录参数的类型和值
						logger.Infof("【请求签名验证】JSON参数: %s = %v (原始类型: %T)", k, v, v)

						// 处理字符串类型
						if strValue, ok := v.(string); ok {
							params[k] = strValue
						} else if v != nil {
							// 对于复杂类型，记录详细信息
							logger.Infof("【请求签名验证】复杂对象详情 %s: %+v", k, v)
						}
					}
				}
			}
		}

		// 检查时间戳和随机数是否来自请求头
		// 如果是，则使用适合签名计算的参数名添加到params
		if c.GetHeader("X-Timestamp") != "" {
			params["timestamp"] = timestamp
			logger.Infof("【请求签名验证】请求头参数: timestamp = %s", timestamp)
		}
		if c.GetHeader("X-Nonce") != "" {
			params["nonce"] = nonce
			logger.Infof("【请求签名验证】请求头参数: nonce = %s", nonce)
		}

		logger.Infof("【请求签名验证】最终收集到的所有参数: %v", params)
		logger.Infof("【请求签名验证】服务器接收到的签名: %s", signature)

		// 计算期望的签名（用于日志）
		h := security.GenerateSignature(params, securityService.GetSignatureSecret())
		logger.Infof("【请求签名验证】服务器计算的签名: %s", h)
		logger.Infof("【请求签名验证】服务器使用的API密钥: %s", securityService.GetSignatureSecret())

		// Validate signature
		err := securityService.ValidateSignature(params, signature)
		logger.Infof("【请求签名验证】签名是否匹配: %v", err == nil)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			logger.Infof("【请求签名验证】最终验证结果: %v", false)
			logger.Info("【请求签名验证】-------------------------结束验证-------------------------")
			return
		}

		logger.Infof("【请求签名验证】最终验证结果: %v", true)
		logger.Info("【请求签名验证】-------------------------结束验证-------------------------")
		c.Next()
	}
}

// getParameter gets a parameter from either the query string or header
func getParameter(c *gin.Context, paramName, headerName string) string {
	// First try to get from query parameters
	if param := c.Query(paramName); param != "" {
		return param
	}

	// Then try to get from header
	if header := c.GetHeader(headerName); header != "" {
		return header
	}

	// For POST/PUT/PATCH methods, try to get from form data
	if c.Request.Method != http.MethodGet {
		if err := c.Request.ParseForm(); err == nil {
			if param := c.Request.PostForm.Get(paramName); param != "" {
				return param
			}
		}
	}

	return ""
}
