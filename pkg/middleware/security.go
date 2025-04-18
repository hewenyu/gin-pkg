package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewenyu/gin-pkg/pkg/auth/security"
)

// SecurityMiddleware validates request timestamps, nonces, and signatures
func SecurityMiddleware(securityService security.SecurityService, timestampWindow time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract parameters (from headers or query params)
		timestamp := getParameter(c, "timestamp", "X-Timestamp")
		nonce := getParameter(c, "nonce", "X-Nonce")
		signature := getParameter(c, "sign", "X-Sign")

		// Skip validation for the nonce endpoint
		if c.FullPath() == "/api/v1/auth/nonce" {
			// For nonce endpoint, only validate timestamp
			if timestamp == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "timestamp is required"})
				c.Abort()
				return
			}

			if err := securityService.ValidateTimestamp(timestamp, timestampWindow); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}

			c.Next()
			return
		}

		// For all other endpoints, validate all security parameters
		if timestamp == "" || nonce == "" || signature == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "timestamp, nonce, and signature are required"})
			c.Abort()
			return
		}

		// Validate timestamp
		if err := securityService.ValidateTimestamp(timestamp, timestampWindow); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Validate nonce
		if err := securityService.ValidateNonce(nonce); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Build parameters map for signature validation
		params := make(map[string]string)

		// Add query parameters
		for k, v := range c.Request.URL.Query() {
			if len(v) > 0 && k != "sign" {
				params[k] = v[0]
			}
		}

		// Add form parameters if POST/PUT/PATCH with form data
		if c.Request.Method != http.MethodGet {
			if err := c.Request.ParseForm(); err == nil {
				for k, v := range c.Request.PostForm {
					if len(v) > 0 && k != "sign" {
						params[k] = v[0]
					}
				}
			}
		}

		// 为非GET请求尝试从JSON请求体中获取参数
		if c.Request.Method != http.MethodGet && c.Request.Header.Get("Content-Type") == "application/json" {
			// 保存请求体
			requestBody, err := c.GetRawData()
			if err == nil && len(requestBody) > 0 {
				// 重新设置请求体，以便后续处理
				c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

				// 解析JSON请求体
				var bodyMap map[string]interface{}
				if json.Unmarshal(requestBody, &bodyMap) == nil {
					// 将JSON请求体中的参数添加到签名验证参数中
					for k, v := range bodyMap {
						// 只添加字符串值
						if strValue, ok := v.(string); ok {
							params[k] = strValue
						}
					}
				}
			}
		}

		// 检查时间戳和随机数是否来自请求头
		// 如果是，则使用适合签名计算的参数名添加到params
		if c.GetHeader("X-Timestamp") != "" {
			params["timestamp"] = timestamp
		}
		if c.GetHeader("X-Nonce") != "" {
			params["nonce"] = nonce
		}

		// Validate signature
		if err := securityService.ValidateSignature(params, signature); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

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
