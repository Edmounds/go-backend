package controllers

import (
	"miniprogram/middlewares"
	"miniprogram/models"

	"github.com/gin-gonic/gin"
)

// ===== HTTP 处理器 =====

// WechatAuthHandler 微信认证处理器
func WechatAuthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.WechatLoginRequest

		// 绑定请求参数
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 初始化认证服务
		authService := GetAuthService()

		// 执行微信认证
		result, err := authService.AuthenticateWithWechat(req.Code, req.ReferralCode)
		if err != nil {
			InternalServerErrorResponse(c, "认证失败", err)
			return
		}

		// 返回认证结果
		SuccessResponse(c, "登录成功", gin.H{
			"token":       result.Token,
			"user":        result.User,
			"is_new_user": result.IsNewUser,
		})
	}
}

// DevLoginHandler 开发环境登录处理器 - 仅用于测试
func DevLoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.DevLoginRequest

		// 绑定请求参数
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 初始化服务
		authService := GetAuthService()

		// 查找用户
		user, err := GetUserByOpenID(req.OpenID)

		isNewUser := false
		if err != nil {
			// 创建新用户
			user, err = authService.createNewUserWithReferral(req.OpenID, req.ReferralCode)
			if err != nil {
				InternalServerErrorResponse(c, "创建用户失败", err)
				return
			}
			isNewUser = true
		} else {
			// 处理现有用户的推荐关系
			err = authService.handleExistingUserReferral(user, req.ReferralCode)
			if err != nil {
				InternalServerErrorResponse(c, "处理推荐关系失败", err)
				return
			}
		}

		// 生成JWT token - 使用原有的middlewares方法
		tokenUser := middlewares.User{
			UserName:     user.UserName,
			UserId:       user.OpenID,
			UserPassword: user.UserPassword,
			OpenID:       user.OpenID,
		}

		token, err := middlewares.GenerateToken(tokenUser)
		if err != nil {
			InternalServerErrorResponse(c, "生成token失败", err)
			return
		}

		// 返回结果
		SuccessResponse(c, "开发环境登录成功", gin.H{
			"token":       token,
			"user":        user,
			"is_new_user": isNewUser,
		})
	}
}
