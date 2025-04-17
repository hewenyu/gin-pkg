package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"
)

// 计算签名
func calculateSignature(params map[string]string, secret string) string {
	// 按键排序
	var keys []string
	for k := range params {
		if k != "sign" { // 排除sign参数本身
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 构建签名字符串
	var paramPairs []string
	for _, k := range keys {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", k, params[k]))
	}
	paramString := strings.Join(paramPairs, "&")

	// 输出调试信息
	fmt.Printf("签名字符串: %s\n", paramString)

	// 计算HMAC-SHA256
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(paramString))
	return hex.EncodeToString(h.Sum(nil))
}

func main() {
	// 获取环境变量中的签名密钥
	secretKey := "your-signature-secret-key-change-this"
	if envSecret := os.Getenv("API_SIGNATURE_SECRET"); envSecret != "" {
		secretKey = envSecret
	}

	// 测试参数
	params := map[string]string{
		"timestamp": "1744904535696",
		"nonce":     "89757250-fe28-47ae-b94a-b8230325797f",
		"email":     "admin@example.com",
		"password":  "admin123456",
	}

	// 计算签名
	signature := calculateSignature(params, secretKey)
	fmt.Printf("计算签名: %s\n", signature)

	// 尝试不同的密钥
	alternateKey := strings.TrimSpace(secretKey)
	if alternateKey != secretKey {
		fmt.Printf("\n尝试去除空格的密钥...\n")
		signature = calculateSignature(params, alternateKey)
		fmt.Printf("计算签名: %s\n", signature)
	}

	// 提示如何使用此工具
	fmt.Printf("\n使用方法: API_SIGNATURE_SECRET=你的签名密钥 go run sigcheck.go\n")
	fmt.Printf("将结果中的签名与测试中输出的签名对比，应该相同\n")
}
