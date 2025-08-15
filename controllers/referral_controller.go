package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"miniprogram/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// 使用 models.Referral / models.ReferralUsage / models.Commission，禁止重复定义

//将小程序码scene绑定一个推荐码，然后别人扫了这个码之后，获取别人的推荐码，在购物的时候会得到对应折扣，提供推荐码的人也会得到对应的佣金
// 推荐码的生成规则是：推荐码由6位随机字母和数字组成，前两位是推荐码的类型，后四位是随机字母和数字
// 推荐码的类型有：
// 1. 普通推荐码
// 2. 校代理推荐码
// 3. 区域代理推荐码

// ===== HTTP 处理器 =====

// GenerateUnlimitedQRCodeHandler 获取不限制数量的小程序码（服务端代理微信接口）
func GenerateUnlimitedQRCodeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.UnlimitedQRCodeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 初始化服务
		qrCodeService := NewWechatQRCodeService()

		// 场景值长度与字符集校验
		if err := qrCodeService.ValidateScene(req.Scene); err != nil {
			BadRequestResponse(c, "scene 参数长度不合法，必须为1-32个可见字符", nil)
			return
		}

		// 获取 access_token（使用本服务缓存）
		accessToken, err := GetCachedAccessToken()
		if err != nil {
			InternalServerErrorResponse(c, "获取微信 access_token 失败", err)
			return
		}

		// 组装微信API请求体
		payload := qrCodeService.BuildQRCodePayload(req)

		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			InternalServerErrorResponse(c, "请求序列化失败", err)
			return
		}

		// 调用微信 getwxacodeunlimit 接口
		url := "https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=" + accessToken
		httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
		if err != nil {
			InternalServerErrorResponse(c, "构建微信请求失败", err)
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			InternalServerErrorResponse(c, "调用微信接口失败", err)
			return
		}
		defer resp.Body.Close()

		// 读取响应内容
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			InternalServerErrorResponse(c, "读取微信响应失败", err)
			return
		}

		// 检查微信接口返回的错误
		if resp.Header.Get("Content-Type") == "application/json" {
			// 如果是JSON响应，说明有错误
			var errResp map[string]interface{}
			if err := json.Unmarshal(respBytes, &errResp); err == nil {
				if errCode, exists := errResp["errcode"]; exists && errCode != 0 {
					ErrorResponse(c, http.StatusBadRequest, 400, fmt.Sprintf("微信接口错误: %v", errResp["errmsg"]), nil)
					return
				}
			}
		}

		// 成功获取小程序码，返回Base64编码
		base64Image := base64.StdEncoding.EncodeToString(respBytes)

		SuccessResponse(c, "获取小程序码成功", gin.H{
			"qr_code": "data:image/jpeg;base64," + base64Image,
			"scene":   req.Scene,
		})
	}
}

// GetReferralInfoHandler 获取用户推荐信息处理器
func GetReferralInfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		openID := c.Param("user_id")

		// 初始化服务
		referralService := NewReferralCodeService()
		commissionService := NewCommissionService()

		// 1. 根据用户ID查询推荐信息
		user, err := GetUserByOpenID(openID)
		if err != nil {
			NotFoundResponse(c, "用户不存在", err)
			return
		}

		// 2. 获取推荐记录
		referral, err := referralService.GetReferralByUserID(user.OpenID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				// 如果没有推荐记录，创建默认响应
				SuccessResponse(c, "获取推荐信息成功", gin.H{
					"_id":                  user.ID,
					"openID":               openID,
					"referral_code":        user.ReferralCode,
					"total_referrals":      0,
					"successful_referrals": 0,
					"total_commission":     0.0,
					"pending_commission":   0.0,
					"withdrawn_commission": 0.0,
					"referral_stats": gin.H{
						"this_month":     0,
						"last_month":     0,
						"total_earnings": 0.0,
					},
					"recent_referrals": []gin.H{},
				})
				return
			}
			InternalServerErrorResponse(c, "获取推荐记录失败", err)
			return
		}

		// 3. 获取佣金记录
		commissions, err := commissionService.GetCommissionsByUserID(user.OpenID, "", "referral")
		if err != nil {
			InternalServerErrorResponse(c, "获取佣金记录失败", err)
			return
		}

		// 4. 计算统计信息
		totalAmount, pendingAmount, _, thisMonthTotal, lastMonthTotal := commissionService.CalculateCommissionStats(commissions)

		// 5. 准备最近推荐记录
		var recentReferrals []gin.H
		for _, usage := range referral.UsedBy {
			recentReferrals = append(recentReferrals, gin.H{
				"user_name":  usage.UserName,
				"used_at":    usage.UsedAt.Format(time.RFC3339),
				"order_id":   usage.OrderID,
				"commission": usage.Commission,
				"status":     usage.Status,
			})
		}

		// 6. 构建响应数据
		SuccessResponse(c, "获取推荐信息成功", gin.H{
			"_id":                  user.ID,
			"openID":               openID,
			"referral_code":        user.ReferralCode,
			"total_referrals":      len(referral.UsedBy),
			"successful_referrals": len(referral.UsedBy), // 可以根据状态进一步筛选
			"total_commission":     totalAmount,
			"pending_commission":   pendingAmount,
			"withdrawn_commission": 0.0, // 这里可以从提现记录计算
			"referral_stats": gin.H{
				"this_month":     thisMonthTotal,
				"last_month":     lastMonthTotal,
				"total_earnings": totalAmount,
			},
			"recent_referrals": recentReferrals,
		})
	}
}

// TrackReferralHandler 跟踪推荐关系处理器
func TrackReferralHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.TrackReferralRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 初始化服务
		rewardService := NewReferralRewardService()

		// 1. 验证推荐码是否存在
		referralService := NewReferralCodeService()
		_, err := referralService.GetUserByReferralCode(req.ReferralCode)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				NotFoundResponse(c, "推荐码不存在", err)
			} else {
				InternalServerErrorResponse(c, "验证推荐码失败", err)
			}
			return
		}

		// 2. 验证被推荐用户是否存在
		referredUser, err := GetUserByOpenID(req.ReferredUserID)
		if err != nil {
			NotFoundResponse(c, "被推荐用户不存在", err)
			return
		}

		// 3. 检查用户是否已经有推荐人
		if referredUser.ReferredBy != "" {
			ErrorResponse(c, http.StatusBadRequest, 400, "用户已有推荐人", nil)
			return
		}

		// 4. 处理推荐关系
		err = rewardService.ProcessNewUserReferral(req.ReferredUserID, req.ReferralCode)
		if err != nil {
			InternalServerErrorResponse(c, "建立推荐关系失败", err)
			return
		}

		SuccessResponse(c, "推荐关系建立成功", gin.H{
			"referral_code":    req.ReferralCode,
			"referred_user_id": req.ReferredUserID,
			"status":           "success",
		})
	}
}

// ValidateReferralCodeHandler 验证推荐码处理器
func ValidateReferralCodeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.ValidateReferralRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 初始化服务
		referralService := NewReferralCodeService()

		// 1. 根据推荐码查询推荐人
		referrer, err := referralService.GetUserByReferralCode(req.ReferralCode)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				NotFoundResponse(c, "推荐码不存在", err)
			} else {
				InternalServerErrorResponse(c, "查询推荐码失败", err)
			}
			return
		}

		// 2. 计算折扣率
		discountRate := referralService.CalculateDiscountRate(referrer.AgentLevel)

		// 3. 返回验证结果
		SuccessResponse(c, "推荐码验证成功", gin.H{
			"valid": true,
			"referrer": gin.H{
				"openID":      referrer.OpenID,
				"user_name":   referrer.UserName,
				"agent_level": referrer.AgentLevel,
			},
			"discount_rate": discountRate,
		})
	}
}

// GetCommissionsHandler 查看返现记录处理器
func GetCommissionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		openID := c.Param("user_id")
		status := c.Query("status")
		commissionType := c.Query("type")

		// 初始化服务
		commissionService := NewCommissionService()

		// 1. 根据openID获取用户信息
		user, err := GetUserByOpenID(openID)
		if err != nil {
			NotFoundResponse(c, "用户不存在", err)
			return
		}

		// 2. 根据用户ID查询佣金记录
		commissions, err := commissionService.GetCommissionsByUserID(user.OpenID, status, commissionType)
		if err != nil {
			InternalServerErrorResponse(c, "获取佣金记录失败", err)
			return
		}

		// 3. 计算总金额统计
		totalAmount, pendingAmount, paidAmount, thisMonthTotal, lastMonthTotal := commissionService.CalculateCommissionStats(commissions)

		// 4. 分类统计
		var totalReferralCommission, totalAgentCommission float64
		for _, commission := range commissions {
			switch commission.Type {
			case "referral":
				totalReferralCommission += commission.Amount
			case "agent":
				totalAgentCommission += commission.Amount
			}
		}

		// 5. 准备响应数据
		var commissionsData []gin.H
		for _, commission := range commissions {
			commissionsData = append(commissionsData, gin.H{
				"_id":           commission.ID,
				"commission_id": commission.CommissionID,
				"openID":        openID, // 从参数中获取，保持响应格式一致
				"amount":        commission.Amount,
				"date":          commission.Date.Format(time.RFC3339),
				"status":        commission.Status,
				"type":          commission.Type,
				"description":   commission.Description,
				"order_id":      commission.OrderID,
			})
		}

		SuccessResponse(c, "获取返现记录成功", gin.H{
			"commissions":    commissionsData,
			"total_amount":   totalAmount,
			"pending_amount": pendingAmount,
			"paid_amount":    paidAmount,
			"summary": gin.H{
				"total_referral_commission": totalReferralCommission,
				"total_agent_commission":    totalAgentCommission,
				"this_month_total":          thisMonthTotal,
				"last_month_total":          lastMonthTotal,
			},
		})
	}
}

// ===== 向后兼容的函数别名 =====
// 这些函数保持原有的函数名，但内部调用新的服务层

// GetUserByReferralCode 根据推荐码获取用户信息 (向后兼容)
func GetUserByReferralCode(referralCode string) (*models.User, error) {
	service := NewReferralCodeService()
	return service.GetUserByReferralCode(referralCode)
}

// GetCommissionsByUserID 根据用户openID获取佣金记录 (向后兼容)
func GetCommissionsByUserID(openID string, status string, commissionType string) ([]models.Commission, error) {
	service := NewCommissionService()
	return service.GetCommissionsByUserID(openID, status, commissionType)
}

// CalculateCommissionStats 计算佣金统计信息 (向后兼容)
func CalculateCommissionStats(commissions []models.Commission) (float64, float64, float64, float64, float64) {
	service := NewCommissionService()
	return service.CalculateCommissionStats(commissions)
}

// CreateCommissionRecord 创建佣金记录 (向后兼容)
func CreateCommissionRecord(openID string, amount float64, commissionType string, description string, orderID string) error {
	service := NewCommissionService()
	return service.CreateCommissionRecord(openID, amount, commissionType, description, orderID)
}

// ProcessReferralReward 处理推荐奖励 (向后兼容)
func ProcessReferralReward(referredUserOpenID string, orderID string, orderAmount float64) error {
	service := NewReferralRewardService()
	return service.ProcessReferralReward(referredUserOpenID, orderID, orderAmount)
}
