package controllers

import (
	"encoding/json"
	"fmt"
	"miniprogram/config"
	"miniprogram/middlewares"
	"miniprogram/models"
	"net/http"
	"net/url"
	"time"

	"sync"

	"github.com/gin-gonic/gin"
)

// WechatLoginRequest 微信登录请求
type WechatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

type WechatLoginResponse struct {
	SessionKey string `json:"session_key"`
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token string      `json:"token"`
		User  interface{} `json:"user"`
	} `json:"data"`
}

// WechatAuthHandler 微信认证处理器
func WechatAuthHandler() gin.HandlerFunc {
	cfg := config.GetConfig()

	return func(c *gin.Context) {
		var req WechatLoginRequest

		// 绑定请求参数，错误会被全局错误处理中间件捕获
		if err := c.ShouldBindJSON(&req); err != nil {
			panic("请求参数错误: " + err.Error())
		}

		code := req.Code

		// 构建微信API URL
		baseURL := cfg.WechatAPIURL + "/sns/jscode2session"
		params := url.Values{}
		params.Add("appid", cfg.WechatAppID)
		params.Add("secret", cfg.WechatAppSecret)
		params.Add("js_code", code)
		params.Add("grant_type", "authorization_code")
		apiURL := baseURL + "?" + params.Encode()

		// 调用微信API，错误会被全局错误处理中间件捕获
		response, err := http.Get(apiURL)

		middlewares.HandleError(err, "调用微信API失败", false)
		defer response.Body.Close()
		var responseData WechatLoginResponse
		err = json.NewDecoder(response.Body).Decode(&responseData)
		middlewares.HandleError(err, "解析微信API响应失败", false)

		// 检查微信API是否返回错误
		if responseData.ErrCode != 0 {
			panic(fmt.Sprintf("微信API错误: %d - %s", responseData.ErrCode, responseData.ErrMsg))
		}

		// 解析微信API响应
		sessionKey := responseData.SessionKey
		openID := responseData.OpenID

		// 检查关键字段是否为空
		if sessionKey == "" || openID == "" {
			panic("微信API响应数据不完整: session_key 或 openid 为空")
		}

		// sessionKey 在本实现中暂未使用，预留给未来的会话管理功能
		_ = sessionKey

		// 根据openid查找用户
		user, _ := GetUserByOpenID(openID)

		if user == nil {
			// 如果用户不存在，创建一个只有openid的新用户
			newUser := &models.User{
				OpenID:    openID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := CreateUser(newUser)
			if err != nil {
				panic("创建用户失败: " + err.Error())
			}

			user = newUser
		}

		// 用户已存在，生成JWT token
		tokenUser := middlewares.User{
			UserName:     user.UserName,
			UserId:       user.OpenID,
			UserPassword: user.UserPassword,
			OpenID:       user.OpenID,
		}

		token, err := middlewares.GenerateToken(tokenUser)
		middlewares.HandleError(err, "生成token失败", false)

		// 返回登录成功响应
		c.JSON(http.StatusOK, LoginResponse{
			Code:    200,
			Message: "登录成功",
			Data: struct {
				Token string      `json:"token"`
				User  interface{} `json:"user"`
			}{
				Token: token,
				User:  user,
			},
		})
	}
}

// wechatAccessTokenResponse 用于解析微信 access_token 接口响应
type wechatAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// accessTokenCache 用于在进程内缓存 access_token，避免频繁请求
var accessTokenCache struct {
	mu       sync.Mutex
	token    string
	expireAt time.Time
}

// getAccessToken 从微信服务器获取新的 access_token（不使用缓存）
func getAccessToken() (string, time.Time, error) {
	cfg := config.GetConfig()
	baseURL := cfg.WechatAPIURL + "/cgi-bin/token"
	params := url.Values{}
	params.Add("grant_type", "client_credential")
	params.Add("appid", cfg.WechatAppID)
	params.Add("secret", cfg.WechatAppSecret)
	apiURL := baseURL + "?" + params.Encode()

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("调用微信获取access_token失败: %w", err)
	}
	defer resp.Body.Close()

	var data wechatAccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", time.Time{}, fmt.Errorf("解析微信access_token响应失败: %w", err)
	}

	if data.ErrCode != 0 {
		return "", time.Time{}, fmt.Errorf("微信API错误: %d - %s", data.ErrCode, data.ErrMsg)
	}

	if data.AccessToken == "" || data.ExpiresIn <= 0 {
		return "", time.Time{}, fmt.Errorf("微信access_token响应数据不完整")
	}

	// 预留5分钟缓冲，避免临界过期
	bufferSeconds := 300
	expireAt := time.Now().Add(time.Duration(data.ExpiresIn-bufferSeconds) * time.Second)
	return data.AccessToken, expireAt, nil
}

// GetCachedAccessToken 获取已缓存的 access_token；若无或过期则自动刷新
func GetCachedAccessToken() (string, error) {
	accessTokenCache.mu.Lock()
	defer accessTokenCache.mu.Unlock()

	if accessTokenCache.token != "" && time.Now().Before(accessTokenCache.expireAt) {
		return accessTokenCache.token, nil
	}

	token, expireAt, err := getAccessToken()
	if err != nil {
		return "", err
	}

	accessTokenCache.token = token
	accessTokenCache.expireAt = expireAt
	return token, nil
}
