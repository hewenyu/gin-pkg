#!/bin/bash

# 默认值
API_URL="http://localhost:8080/api/v1"
API_SECRET="your-signature-secret-key-change-this"
TEST_FILTER=""
VERBOSE=false

# 显示帮助信息
function show_help {
    echo "Usage: $0 [options]"
    echo
    echo "Options:"
    echo "  -u, --url URL      指定API基础URL (默认: $API_URL)"
    echo "  -s, --secret KEY   指定API签名密钥"
    echo "  -t, --test NAME    指定要运行的测试 (例如: TestA_Registration)"
    echo "  -v, --verbose      显示详细输出"
    echo "  -h, --help         显示此帮助信息"
    echo
    exit 0
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        -u|--url)
            API_URL="$2"
            shift
            shift
            ;;
        -s|--secret)
            API_SECRET="$2"
            shift
            shift
            ;;
        -t|--test)
            TEST_FILTER="$2"
            shift
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            show_help
            ;;
        *)
            echo "未知选项: $1"
            show_help
            ;;
    esac
done

# 检查API服务是否可用
echo "正在检查API服务可用性..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$API_URL/auth/nonce")

if [[ $HTTP_CODE != 2* ]] && [[ $HTTP_CODE != 4* ]]; then
    echo "错误: 无法连接到API服务: $API_URL"
    echo "请确保API服务正在运行，或使用--url参数指定正确的API URL。"
    exit 1
fi

echo "API服务可用: $API_URL ✓"

# 准备运行测试
echo "准备运行API测试..."

# 构建测试命令
TEST_CMD="go test"

if [ "$VERBOSE" = true ]; then
    TEST_CMD="$TEST_CMD -v"
fi

if [ -n "$TEST_FILTER" ]; then
    TEST_CMD="$TEST_CMD -run $TEST_FILTER"
fi

# 设置环境变量
export API_BASE_URL="$API_URL"
export API_SIGNATURE_SECRET="$API_SECRET"

# 显示测试信息
echo "API URL: $API_URL"
echo "测试过滤器: ${TEST_FILTER:-全部测试}"
echo "详细模式: $VERBOSE"
echo

# 执行测试
echo "开始执行测试..."
START_TIME=$(date +%s)

eval "$TEST_CMD"
TEST_RESULT=$?

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo
echo "测试执行完成，用时: ${DURATION}s"

if [ $TEST_RESULT -ne 0 ]; then
    echo "测试失败!"
    exit 1
else
    echo "测试成功! ✓"
fi 