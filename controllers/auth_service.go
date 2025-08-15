package controllers

import (
	"encoding/json"
	"fmt"
	"miniprogram/config"
	"miniprogram/models"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// ===== 认证服务层 =====

// AuthService 认证服务
type AuthService struct{}

// GetAuthService 获取认证服务实例
func GetAuthService() *AuthService {
	return &AuthService{}
}

// WechatAuthResult 微信认证结果
type WechatAuthResult struct {
	User       *models.User
	Token      string
	SessionKey string
	IsNewUser  bool
}

// AuthenticateWithWechat 使用微信认证
func (s *AuthService) AuthenticateWithWechat(code string, referralCode string) (*WechatAuthResult, error) {
	// 1. 调用微信API获取openid和session_key
	wechatData, err := s.CallWechatAPI(code)
	if err != nil {
		return nil, fmt.Errorf("微信API调用失败: %w", err)
	}

	// 2. 检查关键字段
	if wechatData.SessionKey == "" || wechatData.OpenID == "" {
		return nil, fmt.Errorf("微信API响应数据不完整: session_key 或 openid 为空")
	}

	// 3. 查找或创建用户
	user, err := GetUserByOpenID(wechatData.OpenID)

	isNewUser := false
	if err != nil {
		// 创建新用户
		user, err = s.createNewUserWithReferral(wechatData.OpenID, referralCode)
		if err != nil {
			return nil, fmt.Errorf("创建用户失败: %w", err)
		}
		isNewUser = true
	} else {
		// 现有用户处理推荐码
		err = s.handleExistingUserReferral(user, referralCode)
		if err != nil {
			return nil, fmt.Errorf("处理推荐关系失败: %w", err)
		}
	}

	// 4. 生成Token
	token, err := s.generateUserToken(user)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}

	return &WechatAuthResult{
		User:       user,
		Token:      token,
		SessionKey: wechatData.SessionKey,
		IsNewUser:  isNewUser,
	}, nil
}

// CallWechatAPI 调用微信API获取用户信息
func (s *AuthService) CallWechatAPI(code string) (*models.WechatLoginResponse, error) {
	cfg := config.GetConfig()

	// 构建微信API URL
	baseURL := cfg.WechatAPIURL + "/sns/jscode2session"
	params := url.Values{}
	params.Add("appid", cfg.WechatAppID)
	params.Add("secret", cfg.WechatAppSecret)
	params.Add("js_code", code)
	params.Add("grant_type", "authorization_code")
	apiURL := baseURL + "?" + params.Encode()

	// 调用微信API
	response, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var responseData models.WechatLoginResponse
	err = json.NewDecoder(response.Body).Decode(&responseData)
	if err != nil {
		return nil, err
	}

	// 检查微信API是否返回错误
	if responseData.ErrCode != 0 {
		return nil, fmt.Errorf("微信API错误: %d - %s", responseData.ErrCode, responseData.ErrMsg)
	}

	return &responseData, nil
}

// createNewUserWithReferral 创建新用户并处理推荐关系
func (s *AuthService) createNewUserWithReferral(openID string, referralCode string) (*models.User, error) {

	// 创建基础用户
	newUser := &models.User{
		OpenID:         openID,
		CollectedCards: []string{},
		Addresses:      []models.Address{},
		Progress: models.Progress{
			LearnedWords: []string{},
		},
		ManagedSchools: []string{},
		ManagedRegions: []string{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 处理推荐码
	if referralCode != "" {
		valid, err := s.ValidateReferralCode(referralCode)
		if err != nil {
			return nil, fmt.Errorf("验证推荐码失败: %w", err)
		}
		if !valid {
			return nil, fmt.Errorf("推荐码无效")
		}
		newUser.ReferredBy = referralCode
	}

	// 创建用户
	err := CreateUser(newUser)
	if err != nil {
		return nil, err
	}

	// 处理推荐关系
	if referralCode != "" {
		referralService := NewReferralRewardService()
		err := referralService.ProcessNewUserReferral(newUser.OpenID, referralCode)
		if err != nil {
			// 推荐关系处理失败不影响用户创建
			// 这里可以记录日志
		}
	}

	return newUser, nil
}

// handleExistingUserReferral 处理现有用户的推荐关系
func (s *AuthService) handleExistingUserReferral(user *models.User, referralCode string) error {
	if user.ReferredBy == "" && referralCode != "" {
		// 验证推荐码
		valid, err := s.ValidateReferralCode(referralCode)
		if err != nil {
			return fmt.Errorf("验证推荐码失败: %w", err)
		}
		if !valid {
			return fmt.Errorf("推荐码无效")
		}

		// 更新用户推荐关系
		referralService := NewReferralRewardService()
		err = referralService.UpdateUserReferredBy(user.OpenID, referralCode)
		if err != nil {
			return fmt.Errorf("设置推荐码失败: %w", err)
		}

		// 处理推荐关系
		err = referralService.ProcessNewUserReferral(user.OpenID, referralCode)
		if err != nil {
			// 推荐关系处理失败不影响登录
		}

		// 更新用户对象
		user.ReferredBy = referralCode
	}
	return nil
}

// ValidateReferralCode 验证推荐码
func (s *AuthService) ValidateReferralCode(referralCode string) (bool, error) {
	referralService := NewReferralCodeService()
	_, err := referralService.GetUserByReferralCode(referralCode)
	if err != nil {
		return false, err
	}
	return true, nil
}

// generateUserToken 生成用户Token
func (s *AuthService) generateUserToken(user *models.User) (string, error) {
	// 这里应该调用token服务，暂时保持原有逻辑
	// 后续可以抽取为独立的TokenService
	tokenService := GetTokenService()
	return tokenService.GenerateTokenForUser(user)
}

// WechatAccessTokenService 微信访问令牌服务
type WechatAccessTokenService struct{}

// accessTokenCache 访问令牌缓存
var accessTokenCache = struct {
	mu       sync.Mutex
	token    string
	expireAt time.Time
}{}

// GetWechatAccessTokenService 获取微信访问令牌服务实例
func GetWechatAccessTokenService() *WechatAccessTokenService {
	return &WechatAccessTokenService{}
}

// GetAccessToken 获取微信访问令牌
func (s *WechatAccessTokenService) GetAccessToken() (string, error) {
	accessTokenCache.mu.Lock()
	defer accessTokenCache.mu.Unlock()

	// 检查缓存是否有效
	if accessTokenCache.token != "" && time.Now().Before(accessTokenCache.expireAt) {
		return accessTokenCache.token, nil
	}

	// 获取新的访问令牌
	token, expireAt, err := s.fetchAccessToken()
	if err != nil {
		return "", err
	}

	// 更新缓存
	accessTokenCache.token = token
	accessTokenCache.expireAt = expireAt
	return token, nil
}

// fetchAccessToken 从微信服务器获取访问令牌
func (s *WechatAccessTokenService) fetchAccessToken() (string, time.Time, error) {
	cfg := config.GetConfig()

	// 构建请求URL
	baseURL := cfg.WechatAPIURL + "/cgi-bin/token"
	params := url.Values{}
	params.Add("grant_type", "client_credential")
	params.Add("appid", cfg.WechatAppID)
	params.Add("secret", cfg.WechatAppSecret)
	apiURL := baseURL + "?" + params.Encode()

	// 发送请求
	response, err := http.Get(apiURL)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("请求微信API失败: %w", err)
	}
	defer response.Body.Close()

	// 解析响应
	var data models.WechatAccessTokenResponse
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("解析微信API响应失败: %w", err)
	}

	// 检查错误
	if data.ErrCode != 0 {
		return "", time.Time{}, fmt.Errorf("微信API错误: %d - %s", data.ErrCode, data.ErrMsg)
	}

	if data.AccessToken == "" || data.ExpiresIn <= 0 {
		return "", time.Time{}, fmt.Errorf("微信access_token响应数据不完整")
	}

	// 计算过期时间（预留5分钟缓冲）
	bufferSeconds := 300
	expireAt := time.Now().Add(time.Duration(data.ExpiresIn-bufferSeconds) * time.Second)

	return data.AccessToken, expireAt, nil
}

// TokenService Token服务
type TokenService struct{}

// GetTokenService 获取Token服务实例
func GetTokenService() *TokenService {
	return &TokenService{}
}

// GenerateTokenForUser 为用户生成Token
func (s *TokenService) GenerateTokenForUser(user *models.User) (string, error) {
	// 暂时返回简单的token，实际使用时会在controller中调用middlewares.GenerateToken
	return fmt.Sprintf("token_%s_%d", user.OpenID, time.Now().Unix()), nil
}

// ===== 向后兼容函数 =====

// GetCachedAccessToken 获取已缓存的访问令牌 (向后兼容)
func GetCachedAccessToken() (string, error) {
	service := GetWechatAccessTokenService()
	return service.GetAccessToken()
}
