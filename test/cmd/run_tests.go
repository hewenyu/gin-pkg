package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	// 定义命令行参数
	apiURL := flag.String("url", "http://localhost:8080/api/v1", "API基础URL")
	signSecret := flag.String("secret", "", "API签名密钥")
	testFilter := flag.String("test", "", "指定要运行的测试(例如: TestA_Registration)")
	verbose := flag.Bool("v", false, "显示详细输出")

	flag.Parse()

	// 检查API服务是否可用
	fmt.Printf("正在检查API服务可用性...\n")
	serviceAvailable := checkServiceAvailability(*apiURL)
	if !serviceAvailable {
		fmt.Printf("错误: 无法连接到API服务: %s\n", *apiURL)
		fmt.Printf("请确保API服务正在运行，或使用--url参数指定正确的API URL。\n")
		os.Exit(1)
	}
	fmt.Printf("API服务可用: %s ✓\n", *apiURL)

	// 构建测试命令
	fmt.Printf("准备运行API测试...\n")

	// 构建测试参数
	var testArgs []string
	testArgs = append(testArgs, "test")

	if *verbose {
		testArgs = append(testArgs, "-v")
	}

	if *testFilter != "" {
		testArgs = append(testArgs, "-run", *testFilter)
	}

	// 添加环境变量
	env := os.Environ()
	if *apiURL != "" {
		env = append(env, fmt.Sprintf("API_BASE_URL=%s", *apiURL))
	}
	if *signSecret != "" {
		env = append(env, fmt.Sprintf("API_SIGNATURE_SECRET=%s", *signSecret))
	}

	// 执行测试
	fmt.Printf("开始执行测试...\n")
	startTime := time.Now()

	cmd := exec.Command("go", testArgs...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	duration := time.Since(startTime).Round(time.Millisecond)
	fmt.Printf("\n测试执行完成，用时: %v\n", duration)

	if err != nil {
		fmt.Printf("测试失败: %v\n", err)
		os.Exit(1)
	}
}

// 检查API服务是否可用
func checkServiceAvailability(baseURL string) bool {
	// 构建测试用的URL (例如尝试获取nonce端点)
	url := fmt.Sprintf("%s/auth/nonce", baseURL)

	// 使用ping命令检查服务是否可达
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "-X", "GET", url)
	output, err := cmd.Output()

	if err != nil {
		return false
	}

	// 检查HTTP状态码，2xx或4xx都表示服务在运行
	// (4xx可能是因为缺少安全头，但至少说明服务在响应)
	statusCode := string(output)
	return statusCode[0] == '2' || statusCode[0] == '4'
}
