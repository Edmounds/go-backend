package controllers

import (
	"encoding/json"
	"fmt"
	"miniprogram/config"
	"miniprogram/middlewares"
	"net/http"
	"net/url"
	"time"

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

		// 构建微信API URL（方法1：使用net/url构建查询参数）
		baseURL := cfg.WechatAPIURL + "/sns/jscode2session"
		params := url.Values{}
		params.Add("appid", cfg.WechatAppID)
		params.Add("secret", cfg.WechatAppSecret)
		params.Add("js_code", code)
		params.Add("grant_type", "authorization_code")
		apiURL := baseURL + "?" + params.Encode()

		// fmt.Println(apiURL)

		// 调用微信API，错误会被全局错误处理中间件捕获
		response, err := http.Get(apiURL)

		// fmt.Println(response)

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

		// unionID := responseData.UnionID

		// 根据openid查找用户
		user, _ := GetUserByOpenID(openID)

		if user == nil {
			// 如果用户不存在，创建一个只有openid的新用户
			newUser := &User{
				OpenID:    openID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := CreateSimpleUser(newUser)
			if err != nil {
				panic("创建用户失败: " + err.Error())
			}

			user = newUser
		}

		// 用户已存在，生成JWT token
		tokenUser := middlewares.User{
			UserName:     user.UserName,
			UserId:       user.ID.Hex(),
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
				User:  user.ToUserProfileResponse(),
			},
		})
	}
}
