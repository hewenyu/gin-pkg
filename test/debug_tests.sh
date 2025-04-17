#!/bin/bash

# 脚本颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认值
API_URL="http://localhost:8080/api/v1"
API_SECRET="your-signature-secret-key-change-this"
TEST_FILTER=""

# 读取配置文件中的安全密钥
if [ -f "../config/default.yaml" ]; then
    echo -e "${BLUE}尝试从配置文件读取签名密钥...${NC}"
    CONFIG_SECRET=$(grep -E "signatureSecret:\s+\"[^\"]+\"" ../config/default.yaml | awk -F'"' '{print $2}')
    if [ -n "$CONFIG_SECRET" ]; then
        API_SECRET="$CONFIG_SECRET"
        echo -e "${GREEN}成功从配置文件读取签名密钥!${NC}"
    else
        echo -e "${YELLOW}警告: 未在配置文件中找到签名密钥，使用默认值${NC}"
    fi
else
    echo -e "${YELLOW}警告: 未找到配置文件 ../config/default.yaml${NC}"
fi

# 显示当前配置
echo -e "${BLUE}调试信息:${NC}"
echo -e "API URL: ${GREEN}$API_URL${NC}"
echo -e "API SECRET: ${GREEN}$API_SECRET${NC}"

# 生成当前时间戳用于请求 (标准Unix时间戳，秒级)
TIMESTAMP=$(date +%s)
# 转换为毫秒级时间戳
TIMESTAMP="${TIMESTAMP}000"
echo -e "当前时间戳(毫秒): ${GREEN}$TIMESTAMP${NC}"

# 检查API服务是否可用
echo -e "\n${BLUE}正在检查API服务可用性...${NC}"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$API_URL/auth/nonce?timestamp=$TIMESTAMP" -H "X-Timestamp: $TIMESTAMP")

if [[ $HTTP_CODE != 2* ]] && [[ $HTTP_CODE != 4* ]]; then
    echo -e "${RED}错误: 无法连接到API服务: $API_URL${NC}"
    echo -e "${YELLOW}请确保API服务正在运行，或检查URL是否正确${NC}"
    exit 1
fi

if [[ $HTTP_CODE == 2* ]]; then
    echo -e "${GREEN}API服务可用: $API_URL ✓${NC}"
else
    echo -e "${YELLOW}警告: API服务返回状态码 $HTTP_CODE，可能存在配置问题${NC}"
    
    # 尝试获取更详细的错误信息
    ERROR_RESPONSE=$(curl -s "$API_URL/auth/nonce?timestamp=$TIMESTAMP" -H "X-Timestamp: $TIMESTAMP")
    echo -e "${YELLOW}获取nonce的响应: $ERROR_RESPONSE${NC}"
fi

# 准备运行测试
echo -e "\n${BLUE}准备运行API测试，详细模式...${NC}"

# 设置环境变量
export API_BASE_URL="$API_URL"
export API_SIGNATURE_SECRET="$API_SECRET"

# 测试签名计算
echo -e "\n${BLUE}测试签名计算:${NC}"
cat <<EOF | go run - "$API_SECRET"
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("请提供签名密钥作为参数")
		os.Exit(1)
	}
	
	secret := os.Args[1]
	// 获取标准Unix时间戳并转换为毫秒级
	timestamp := fmt.Sprintf("%d000", time.Now().Unix())
	nonce := "test-nonce-123456"
	
	params := map[string]string{
		"timestamp": timestamp,
		"nonce":     nonce,
	}
	
	signature := calculateSignature(params, secret)
	
	fmt.Printf("毫秒级时间戳: %s\n", timestamp)
	fmt.Printf("随机数: %s\n", nonce)
	fmt.Printf("签名: %s\n", signature)
	
	// 输出完整的请求示例
	fmt.Printf("\n完整请求示例:\n")
	fmt.Printf("curl -v \"http://localhost:8080/api/v1/auth/nonce?timestamp=%s&nonce=%s\" \\\n", timestamp, nonce)
	fmt.Printf("  -H \"X-Timestamp: %s\" \\\n", timestamp)
	fmt.Printf("  -H \"X-Nonce: %s\" \\\n", nonce)
	fmt.Printf("  -H \"X-Sign: %s\"\n", signature)
}

func calculateSignature(params map[string]string, secret string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var signStr string
	for _, k := range keys {
		signStr += k + "=" + params[k] + "&"
	}
	if len(signStr) > 0 {
		signStr = signStr[:len(signStr)-1]
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signStr))
	return hex.EncodeToString(h.Sum(nil))
}
EOF

# 运行测试
echo -e "\n${BLUE}执行API测试，输出详细信息...${NC}"
TEST_CMD="go test -v"

if [ -n "$TEST_FILTER" ]; then
    TEST_CMD="$TEST_CMD -run $TEST_FILTER"
fi

# 添加调试输出的环境变量
export API_DEBUG=true

# 执行测试
START_TIME=$(date +%s)

YELLOW=$YELLOW NC=$NC TEST_RESULT=0
echo -e "${YELLOW}执行命令: $TEST_CMD${NC}"
eval "$TEST_CMD"
TEST_RESULT=$?

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo
echo -e "${BLUE}测试执行完成，用时: ${DURATION}s${NC}"

if [ $TEST_RESULT -ne 0 ]; then
    echo -e "${RED}测试失败!${NC}"
    
    # 显示调试提示
    echo -e "\n${YELLOW}调试提示:${NC}"
    echo -e "1. 确保API服务正在运行"
    echo -e "2. 验证签名密钥是否正确，当前使用: $API_SECRET"
    echo -e "3. 检查默认管理员账户是否已创建并可以登录"
    echo -e "4. 尝试手动发送一个请求到API服务进行测试:"
    
    echo -e "\n${BLUE}示例请求:${NC}"
    TIMESTAMP=$(date +%s)
    TIMESTAMP="${TIMESTAMP}000"
    echo -e "curl -v -X GET \"$API_URL/auth/nonce?timestamp=$TIMESTAMP\" -H \"X-Timestamp: $TIMESTAMP\""
    
    exit 1
else
    echo -e "${GREEN}测试成功! ✓${NC}"
fi 