# API 测试框架

这是一个用于测试 Gin-Pkg API 服务的自动化测试框架。它提供了完整的 API 测试功能，包括安全验证、身份认证和用户管理功能的测试。

## 功能特点

- 自动处理安全验证流程（时间戳、随机数和签名）
- 支持 JWT 认证
- 测试 API 端点的完整流程
- 使用 testify 库进行断言和测试套件管理
- 按顺序执行测试（命名以字母前缀排序）

## 使用方法

### 使用测试脚本运行（推荐）

我们提供了一个方便的脚本来运行测试，支持各种参数：

```bash
# 使用默认参数运行所有测试
./run_test.sh

# 运行特定测试
./run_test.sh -t TestA_Registration

# 指定API URL和密钥
./run_test.sh -u http://api.example.com/api/v1 -s "your-signature-secret"

# 显示详细输出
./run_test.sh -v

# 显示帮助信息
./run_test.sh -h
```

### 使用Go工具直接运行

```bash
# 运行所有测试
go test -v

# 运行特定测试
go test -v -run TestA_Registration

# 指定环境变量
API_BASE_URL=http://localhost:8080/api/v1 API_SIGNATURE_SECRET=your-secret go test -v
```

## 环境变量配置

你可以通过以下环境变量来配置测试：

- `API_BASE_URL` - API服务的基础URL
- `API_SIGNATURE_SECRET` - 用于签名的密钥
- `API_ADMIN_USER` - 默认管理员用户的邮箱
- `API_ADMIN_PASS` - 默认管理员用户的密码

## 测试流程

测试套件按以下顺序执行测试：

1. 用户注册功能
2. 用户登录功能
3. 令牌刷新功能
4. 获取用户信息功能
5. 更新用户信息功能
6. 未授权访问测试
7. 无效签名测试

## 配置说明

在 `api_test.go` 文件开头的常量部分可以调整配置：

```go
var (
    baseURL          = "http://localhost:8080/api/v1"  // API基础URL
    signatureSecret  = "your-signature-secret-key-change-this"  // 签名密钥，需要与服务器配置一致
    timestampFormat  = time.RFC3339  // 时间戳格式
    defaultAdminUser = "admin@example.com"  // 默认管理员用户名
    defaultAdminPass = "adminpassword"  // 默认管理员密码
)
```

请确保 `signatureSecret` 与服务器的配置一致，否则所有请求都会因签名验证失败而被拒绝。

## 扩展测试

要添加新的测试，只需在 `ApiTestSuite` 中添加新的测试方法。为了确保测试顺序，建议使用字母前缀命名（如 `TestH_YourNewTest`）。

```go
func (suite *ApiTestSuite) TestH_YourNewTest() {
    t := suite.T()
    
    // 使用 sendSecureRequest 方法发送带安全验证的请求
    resp, err := suite.sendSecureRequest("GET", "/your/endpoint", nil, true)
    assert.NoError(t, err)
    defer resp.Body.Close()
    
    // 进行断言
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // 解析和验证响应
    // ...
}
``` 