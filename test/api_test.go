package test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	// 配置测试环境
	baseURL          = "http://localhost:8080/api/v1"
	signatureSecret  = "your-signature-secret-key-change-this" // 配置中设置的签名密钥
	timestampFormat  = time.RFC3339
	defaultAdminUser = "Admin"
	defaultAdminPass = "admin123456"
)

// ApiTestSuite 定义测试套件
type ApiTestSuite struct {
	suite.Suite
	client       *http.Client
	accessToken  string
	refreshToken string
	userId       string
	nonce        string
}

// 通用API响应结构
type ApiResponse struct {
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

// NonceResponse 随机数响应
type NonceResponse struct {
	Nonce string `json:"nonce"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	User struct {
		ID        string  `json:"id"`
		Email     string  `json:"email"`
		Username  string  `json:"username"`
		Role      string  `json:"role"`
		Active    bool    `json:"active"`
		AvatarURL *string `json:"avatar_url"`
		CreatedAt string  `json:"created_at"`
		UpdatedAt string  `json:"updated_at"`
	} `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Username  string  `json:"username"`
	Role      string  `json:"role"`
	Active    bool    `json:"active"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// TokenResponse 刷新令牌响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// 随机生成测试邮箱
func generateTestEmail() string {
	return fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
}

// SetupSuite 设置测试套件
func (suite *ApiTestSuite) SetupSuite() {
	// 初始化HTTP客户端
	suite.client = &http.Client{
		Timeout: 10 * time.Second,
	}

	// 获取随机数
	nonce, err := suite.getNonce()
	if err != nil {
		suite.T().Fatalf("Failed to get nonce: %v", err)
	}
	suite.nonce = nonce

	// 尝试登录管理员
	err = suite.login(defaultAdminUser, defaultAdminPass)
	if err != nil {
		// 如果登录失败，可能需要先创建管理员账户
		suite.T().Logf("Admin login failed, will try to create one: %v", err)
	}
}

// 获取随机数
func (suite *ApiTestSuite) getNonce() (string, error) {
	timestamp := time.Now().Format(timestampFormat)

	// 创建请求
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/auth/nonce", baseURL), nil)
	if err != nil {
		return "", err
	}

	// 添加时间戳头
	req.Header.Set("X-Timestamp", timestamp)

	// 发送请求
	resp, err := suite.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 解析响应
	var nonceResp NonceResponse
	if err := json.NewDecoder(resp.Body).Decode(&nonceResp); err != nil {
		return "", err
	}

	return nonceResp.Nonce, nil
}

// 计算签名
func calculateSignature(params map[string]string, secret string) string {
	// 按键排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signStr string
	for _, k := range keys {
		signStr += k + "=" + params[k] + "&"
	}

	// 移除末尾的 "&"
	if len(signStr) > 0 {
		signStr = signStr[:len(signStr)-1]
	}

	// 计算HMAC-SHA256
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signStr))
	return hex.EncodeToString(h.Sum(nil))
}

// 发送带安全头的请求
func (suite *ApiTestSuite) sendSecureRequest(method, path string, body interface{}, authRequired bool) (*http.Response, error) {
	// 获取新的随机数（如果当前没有）
	if suite.nonce == "" {
		nonce, err := suite.getNonce()
		if err != nil {
			return nil, err
		}
		suite.nonce = nonce
	}

	var reqBody []byte
	var err error

	// 序列化请求体（如果有）
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	// 创建请求
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", baseURL, path), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// 设置内容类型
	req.Header.Set("Content-Type", "application/json")

	// 获取当前时间戳
	timestamp := time.Now().Format(timestampFormat)

	// 构建签名参数
	params := map[string]string{
		"timestamp": timestamp,
		"nonce":     suite.nonce,
	}

	// 计算签名
	signature := calculateSignature(params, signatureSecret)

	// 添加安全头
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Nonce", suite.nonce)
	req.Header.Set("X-Sign", signature)

	// 添加认证头（如果需要）
	if authRequired && suite.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+suite.accessToken)
	}

	// 用完一次随机数就清空，下次请求会重新获取
	suite.nonce = ""

	// 发送请求
	return suite.client.Do(req)
}

// 注册测试
func (suite *ApiTestSuite) TestA_Registration() {
	t := suite.T()

	// 创建随机测试用户
	testEmail := generateTestEmail()
	reqBody := map[string]string{
		"email":    testEmail,
		"username": "Test User",
		"password": "password123",
	}

	// 发送注册请求
	resp, err := suite.sendSecureRequest("POST", "/auth/register", reqBody, false)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// 验证响应状态码
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// 解析响应
	var user UserResponse
	err = json.NewDecoder(resp.Body).Decode(&user)
	assert.NoError(t, err)

	// 验证响应内容
	assert.Equal(t, testEmail, user.Email)
	assert.Equal(t, "Test User", user.Username)
	assert.Equal(t, "user", user.Role) // 默认角色应为user
	assert.True(t, user.Active)        // 默认应为激活状态
}

// 登录方法
func (suite *ApiTestSuite) login(email, password string) error {
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}

	// 发送登录请求
	resp, err := suite.sendSecureRequest("POST", "/auth/login", reqBody, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 验证响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	// 解析响应
	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return err
	}

	// 保存令牌和用户ID
	suite.accessToken = authResp.AccessToken
	suite.refreshToken = authResp.RefreshToken
	suite.userId = authResp.User.ID

	return nil
}

// 登录测试
func (suite *ApiTestSuite) TestB_Login() {
	t := suite.T()

	// 创建随机测试用户
	testEmail := generateTestEmail()
	testPassword := "password123"

	// 先注册用户
	reqBody := map[string]string{
		"email":    testEmail,
		"username": "Login Test User",
		"password": testPassword,
	}

	resp, err := suite.sendSecureRequest("POST", "/auth/register", reqBody, false)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// 测试登录
	err = suite.login(testEmail, testPassword)
	assert.NoError(t, err)

	// 验证令牌不为空
	assert.NotEmpty(t, suite.accessToken)
	assert.NotEmpty(t, suite.refreshToken)
	assert.NotEmpty(t, suite.userId)
}

// 刷新令牌测试
func (suite *ApiTestSuite) TestC_RefreshToken() {
	t := suite.T()

	// 确保已登录
	if suite.refreshToken == "" {
		err := suite.login(defaultAdminUser, defaultAdminPass)
		assert.NoError(t, err)
	}

	// 保存当前令牌以便比较
	oldAccessToken := suite.accessToken
	oldRefreshToken := suite.refreshToken

	// 请求体
	reqBody := map[string]string{
		"refresh_token": suite.refreshToken,
	}

	// 发送刷新令牌请求
	resp, err := suite.sendSecureRequest("POST", "/auth/refresh", reqBody, false)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// 验证响应状态码
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 解析响应
	var tokenResp TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	assert.NoError(t, err)

	// 更新令牌
	suite.accessToken = tokenResp.AccessToken
	suite.refreshToken = tokenResp.RefreshToken

	// 验证新令牌不同于旧令牌
	assert.NotEqual(t, oldAccessToken, suite.accessToken)
	assert.NotEqual(t, oldRefreshToken, suite.refreshToken)
}

// 获取当前用户信息测试
func (suite *ApiTestSuite) TestD_GetCurrentUser() {
	t := suite.T()

	// 确保已登录
	if suite.accessToken == "" {
		err := suite.login(defaultAdminUser, defaultAdminPass)
		assert.NoError(t, err)
	}

	// 发送获取用户信息请求
	resp, err := suite.sendSecureRequest("GET", "/users/"+suite.userId, nil, true)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// 验证响应状态码
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 解析响应
	var user UserResponse
	err = json.NewDecoder(resp.Body).Decode(&user)
	assert.NoError(t, err)

	// 验证用户ID
	assert.Equal(t, suite.userId, user.ID)
}

// 更新用户信息测试
func (suite *ApiTestSuite) TestE_UpdateUser() {
	t := suite.T()

	// 确保已登录
	if suite.accessToken == "" {
		err := suite.login(defaultAdminUser, defaultAdminPass)
		assert.NoError(t, err)
	}

	// 新的用户名
	newUsername := "Updated Username " + strconv.FormatInt(time.Now().Unix(), 10)

	// 请求体
	reqBody := map[string]string{
		"username": newUsername,
	}

	// 发送更新用户请求
	resp, err := suite.sendSecureRequest("PUT", "/users/"+suite.userId, reqBody, true)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// 验证响应状态码
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 解析响应
	var user UserResponse
	err = json.NewDecoder(resp.Body).Decode(&user)
	assert.NoError(t, err)

	// 验证用户名已更新
	assert.Equal(t, newUsername, user.Username)
}

// 测试非授权访问
func (suite *ApiTestSuite) TestF_UnauthorizedAccess() {
	t := suite.T()

	// 保存当前令牌
	savedToken := suite.accessToken

	// 清空令牌
	suite.accessToken = ""

	// 尝试访问需要授权的端点
	resp, err := suite.sendSecureRequest("GET", "/users/"+suite.userId, nil, true)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// 验证响应状态码应为未授权
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// 恢复令牌
	suite.accessToken = savedToken
}

// 错误签名测试
func (suite *ApiTestSuite) TestG_InvalidSignature() {
	t := suite.T()

	// 获取随机数
	nonce, err := suite.getNonce()
	assert.NoError(t, err)

	// 创建请求
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/users/%s", baseURL, suite.userId), nil)
	assert.NoError(t, err)

	// 设置内容类型
	req.Header.Set("Content-Type", "application/json")

	// 设置认证
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)

	// 设置时间戳和随机数，但使用错误的签名
	timestamp := time.Now().Format(timestampFormat)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-Sign", "invalid-signature")

	// 发送请求
	resp, err := suite.client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// 验证响应状态码应为错误请求
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// 运行测试套件
func TestAPISuite(t *testing.T) {
	suite.Run(t, new(ApiTestSuite))
}
