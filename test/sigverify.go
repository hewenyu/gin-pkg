package test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
