package controllers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"miniprogram/middlewares"
	"miniprogram/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TrackReferralRequest 跟踪推荐关系请求
type TrackReferralRequest struct {
	ReferralCode   string `json:"referral_code" binding:"required"`
	ReferredUserID string `json:"referred_user_id" binding:"required"`
}

// ValidateReferralRequest 验证推荐码请求
type ValidateReferralRequest struct {
	ReferralCode string `json:"referral_code" binding:"required"`
}

// 使用 models.Referral / models.ReferralUsage / models.Commission，禁止重复定义

//将小程序码scene绑定一个推荐码，然后别人扫了这个码之后，获取别人的推荐码，在购物的时候会得到对应折扣，提供推荐码的人也会得到对应的佣金
// 推荐码的生成规则是：推荐码由6位随机字母和数字组成，前两位是推荐码的类型，后四位是随机字母和数字
// 推荐码的类型有：
// 1. 普通推荐码
// 2. 校代理推荐码
// 3. 区域代理推荐码

// GenerateMiniprogramCode 生成小程序码

// GenerateUnlimitedQRCodeHandler 获取不限制数量的小程序码（服务端代理微信接口）
func GenerateUnlimitedQRCodeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.UnlimitedQRCodeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 场景值长度与字符集校验由微信侧严格校验，这里仅做最基本长度保护
		if len(req.Scene) == 0 || len(req.Scene) > 32 {
			BadRequestResponse(c, "scene 参数长度不合法，必须为1-32个可见字符", nil)
			return
		}

		// 获取 access_token（使用本服务缓存）
		accessToken, err := GetCachedAccessToken()
		if err != nil {
			InternalServerErrorResponse(c, "获取微信 access_token 失败", err)
			return
		}

		// 组装微信API请求体（字段命名遵循微信接口）
		payload := map[string]interface{}{
			"scene": req.Scene,
		}
		if req.Page != "" {
			payload["page"] = req.Page
		}
		if req.CheckPath != nil {
			payload["check_path"] = *req.CheckPath
		}
		if req.EnvVersion != "" {
			payload["env_version"] = req.EnvVersion
		}
		if req.Width > 0 {
			payload["width"] = req.Width
		}
		if req.AutoColor != nil {
			payload["auto_color"] = *req.AutoColor
		}
		if req.LineColor != nil {
			payload["line_color"] = map[string]int{"r": req.LineColor.R, "g": req.LineColor.G, "b": req.LineColor.B}
		}
		if req.IsHyaline != nil {
			payload["is_hyaline"] = *req.IsHyaline
		}

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

		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			InternalServerErrorResponse(c, "读取微信响应失败", err)
			return
		}

		// 判断是否为JSON错误响应
		var wxErr struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		if err := json.Unmarshal(respBytes, &wxErr); err == nil && wxErr.ErrCode != 0 {
			BadRequestResponse(c, "微信接口错误: "+wxErr.ErrMsg, fmt.Errorf("errcode=%d", wxErr.ErrCode))
			return
		}

		// 正常为图片二进制，转为base64返回
		encoded := base64.StdEncoding.EncodeToString(respBytes)
		SuccessResponse(c, "获取小程序码成功", gin.H{
			"image_base64": encoded,
			"content_type": resp.Header.Get("Content-Type"),
			"scene":        req.Scene,
			"page":         req.Page,
		})
	}
}

// GetUserByReferralCode 根据推荐码获取用户信息
func GetUserByReferralCode(referralCode string) (*models.User, error) {
	collection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"referral_code": referralCode}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// calculateDiscountRate 根据代理等级计算折扣率
func calculateDiscountRate(agentLevel int) float64 {
	switch agentLevel {
	case 0: // 普通用户
		return 0.02 // 2%折扣
	case 1: // 校代理
		return 0.05 // 5%折扣
	case 2: // 区域代理
		return 0.08 // 8%折扣
	default:
		return 0.02 // 默认2%折扣
	}
}

// calculateCommissionRate 根据代理等级计算佣金率
func calculateCommissionRate(agentLevel int) float64 {
	switch agentLevel {
	case 0: // 普通用户
		return 0.01 // 1%佣金
	case 1: // 校代理
		return 0.03 // 3%佣金
	case 2: // 区域代理
		return 0.05 // 5%佣金
	default:
		return 0.01 // 默认1%佣金
	}
}

// CreateReferralRecord 创建推荐记录
func CreateReferralRecord(referralCode string, referrerOpenID string) error {
	collection := GetCollection("referrals")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 检查是否已存在推荐记录
	var existingReferral models.Referral
	err := collection.FindOne(ctx, bson.M{"user_openid": referrerOpenID}).Decode(&existingReferral)
	if err == mongo.ErrNoDocuments {
		// 创建新的推荐记录
		newReferral := models.Referral{
			ReferralCode: referralCode,
			UserOpenID:   referrerOpenID, // 直接使用openID
			UsedBy:       []models.ReferralUsage{},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		_, err = collection.InsertOne(ctx, newReferral)
		return err
	}
	return err // 如果已存在或有其他错误，返回错误
}

// AddReferralUsage 添加推荐使用记录
func AddReferralUsage(referralCode string, referredUserOpenID string, referredUserName string, orderID string, commissionAmount float64) error {
	collection := GetCollection("referrals")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usage := models.ReferralUsage{
		UserOpenID: referredUserOpenID, // 直接使用openID
		UserName:   referredUserName,
		UsedAt:     time.Now(),
		OrderID:    orderID,
		Commission: commissionAmount,
		Status:     "pending", // 初始状态为待处理
	}

	filter := bson.M{"referral_code": referralCode}
	update := bson.M{
		"$push": bson.M{"used_by": usage},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// CreateCommissionRecord 创建佣金记录
func CreateCommissionRecord(openID string, amount float64, commissionType string, description string, orderID string) error {
	collection := GetCollection("commissions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 生成佣金ID
	commissionID := generateCommissionID()

	commission := models.Commission{
		CommissionID: commissionID,
		UserOpenID:   openID, // 直接使用openID
		Amount:       amount,
		Date:         time.Now(),
		Status:       "pending",
		Type:         commissionType,
		Description:  description,
		OrderID:      orderID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := collection.InsertOne(ctx, commission)
	return err
}

// generateCommissionID 生成佣金ID
func generateCommissionID() string {
	return "COMM" + time.Now().Format("20060102150405") + generateRandomString(4)
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ProcessNewUserReferral 处理新用户的推荐关系
func ProcessNewUserReferral(userOpenID string, referralCode string) error {
	// 1. 获取推荐人信息
	referrer, err := GetUserByReferralCode(referralCode)
	if err != nil {
		return err
	}

	// 2. 确保推荐人有推荐记录
	err = CreateReferralRecord(referralCode, referrer.OpenID)
	if err != nil && err != mongo.ErrNoDocuments {
		// 如果不是"已存在"的错误，则返回错误
		return err
	}

	// 3. 获取被推荐用户信息
	referredUser, err := GetUserByOpenID(userOpenID)
	if err != nil {
		return err
	}

	// 4. 添加推荐使用记录（暂时不创建佣金记录，等待订单完成时再创建）
	err = AddReferralUsage(referralCode, referredUser.OpenID, referredUser.UserName, "", 0.0)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUserReferredBy 更新用户的推荐人信息
func UpdateUserReferredBy(openID string, referralCode string) error {
	collection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"openID": openID}
	update := bson.M{
		"$set": bson.M{
			"referred_by": referralCode,
			"updated_at":  time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// ProcessReferralReward 处理推荐奖励 - 当被推荐用户完成订单时调用
func ProcessReferralReward(referredUserOpenID string, orderID string, orderAmount float64) error {
	// 1. 获取被推荐用户信息
	referredUser, err := GetUserByOpenID(referredUserOpenID)
	if err != nil {
		return err
	}

	// 2. 检查用户是否有推荐人
	if referredUser.ReferredBy == "" {
		return nil // 没有推荐人，不处理奖励
	}

	// 3. 获取推荐人信息
	referrer, err := GetUserByReferralCode(referredUser.ReferredBy)
	if err != nil {
		return err
	}

	// 4. 计算佣金金额
	commissionRate := calculateCommissionRate(referrer.AgentLevel)
	commissionAmount := orderAmount * commissionRate

	// 5. 创建佣金记录
	description := "推荐佣金 - 用户 " + referredUser.UserName + " 完成订单"
	err = CreateCommissionRecord(referrer.OpenID, commissionAmount, "referral", description, orderID)
	if err != nil {
		return err
	}

	// 6. 更新推荐使用记录状态
	err = UpdateReferralUsageStatus(referredUser.ReferredBy, referredUser.OpenID, orderID, commissionAmount, "completed")
	if err != nil {
		return err
	}

	return nil
}

// UpdateReferralUsageStatus 更新推荐使用记录状态
func UpdateReferralUsageStatus(referralCode string, referredUserOpenID string, orderID string, commissionAmount float64, status string) error {
	collection := GetCollection("referrals")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"referral_code":       referralCode,
		"used_by.user_openid": referredUserOpenID, // 直接使用openID
	}
	update := bson.M{
		"$set": bson.M{
			"used_by.$.order_id":   orderID,
			"used_by.$.commission": commissionAmount,
			"used_by.$.status":     status,
			"updated_at":           time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// ===== 数据库查询函数 =====

// GetReferralByCode 根据推荐码获取推荐信息
func GetReferralByCode(referralCode string) (*models.Referral, error) {
	collection := GetCollection("referrals")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var referral models.Referral
	err := collection.FindOne(ctx, bson.M{"referral_code": referralCode}).Decode(&referral)
	if err != nil {
		return nil, err
	}
	return &referral, nil
}

// GetReferralByUserID 根据用户openID获取推荐信息
func GetReferralByUserID(openID string) (*models.Referral, error) {
	collection := GetCollection("referrals")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var referral models.Referral
	err := collection.FindOne(ctx, bson.M{"user_openid": openID}).Decode(&referral)
	if err != nil {
		return nil, err
	}
	return &referral, nil
}

// GetCommissionsByUserID 根据用户openID获取佣金记录
func GetCommissionsByUserID(openID string, status string, commissionType string) ([]models.Commission, error) {
	collection := GetCollection("commissions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_openid": openID} // 直接使用openID
	if status != "" {
		filter["status"] = status
	}
	if commissionType != "" {
		filter["type"] = commissionType
	}

	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var commissions []models.Commission
	if err = cursor.All(ctx, &commissions); err != nil {
		return nil, err
	}

	return commissions, nil
}

// CalculateCommissionStats 计算佣金统计信息
func CalculateCommissionStats(commissions []models.Commission) (float64, float64, float64, float64, float64) {
	var totalAmount, pendingAmount, paidAmount, thisMonthTotal, lastMonthTotal float64

	now := time.Now()
	thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonthStart := thisMonthStart.AddDate(0, -1, 0)
	lastMonthEnd := thisMonthStart.Add(-time.Second)

	for _, commission := range commissions {
		totalAmount += commission.Amount

		switch commission.Status {
		case "pending":
			pendingAmount += commission.Amount
		case "paid":
			paidAmount += commission.Amount
		}

		// 本月统计
		if commission.Date.After(thisMonthStart) || commission.Date.Equal(thisMonthStart) {
			thisMonthTotal += commission.Amount
		}

		// 上月统计
		if commission.Date.After(lastMonthStart) && commission.Date.Before(lastMonthEnd) {
			lastMonthTotal += commission.Amount
		}
	}

	return totalAmount, pendingAmount, paidAmount, thisMonthTotal, lastMonthTotal
}

// GetReferralInfoHandler 获取用户推荐信息处理器
func GetReferralInfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		openID := c.Param("user_id")

		// 1. 根据用户ID查询推荐信息
		user, err := GetUserByOpenID(openID)
		if err != nil {
			if middlewares.HandleError(err, "获取用户信息失败", false) {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "用户不存在",
					"error":   err.Error(),
				})
				return
			}
		}

		// 2. 获取推荐记录
		referral, err := GetReferralByUserID(user.OpenID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				// 如果没有推荐记录，创建默认响应
				c.JSON(http.StatusOK, gin.H{
					"code":    200,
					"message": "获取推荐信息成功",
					"data": gin.H{
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
					},
				})
				return
			}

			if middlewares.HandleError(err, "获取推荐记录失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "获取推荐记录失败",
					"error":   err.Error(),
				})
				return
			}
		}

		// 3. 获取佣金记录
		commissions, err := GetCommissionsByUserID(user.OpenID, "", "referral")
		if err != nil {
			if middlewares.HandleError(err, "获取佣金记录失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "获取佣金记录失败",
					"error":   err.Error(),
				})
				return
			}
		}

		// 4. 计算统计数据
		totalCommission, pendingCommission, withdrawnCommission, thisMonth, lastMonth := CalculateCommissionStats(commissions)

		// 计算推荐人数统计
		totalReferrals := len(referral.UsedBy)
		successfulReferrals := 0
		for _, usage := range referral.UsedBy {
			if usage.Status == "completed" {
				successfulReferrals++
			}
		}

		// 5. 准备最近推荐记录（最多显示10条）
		var recentReferrals []gin.H
		for i, usage := range referral.UsedBy {
			if i >= 10 { // 只显示最近10条
				break
			}
			recentReferrals = append(recentReferrals, gin.H{
				"referred_user":     usage.UserName,
				"referral_date":     usage.UsedAt.Format(time.RFC3339),
				"commission_earned": usage.Commission,
				"status":            usage.Status,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "获取推荐信息成功",
			"data": gin.H{
				"_id":                  user.ID,
				"openID":               openID,
				"referral_code":        user.ReferralCode,
				"total_referrals":      totalReferrals,
				"successful_referrals": successfulReferrals,
				"total_commission":     totalCommission,
				"pending_commission":   pendingCommission,
				"withdrawn_commission": withdrawnCommission,
				"referral_stats": gin.H{
					"this_month":     thisMonth,
					"last_month":     lastMonth,
					"total_earnings": totalCommission,
				},
				"recent_referrals": recentReferrals,
			},
		})
	}
}

// TrackReferralHandler 跟踪推荐关系处理器
func TrackReferralHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TrackReferralRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
				"error":   err.Error(),
			})
			return
		}

		// 1. 验证推荐码是否有效
		isValid, err := ValidateReferralCode(req.ReferralCode)
		if err != nil {
			if middlewares.HandleError(err, "验证推荐码失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "验证推荐码失败",
					"error":   err.Error(),
				})
				return
			}
		}

		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "推荐码无效",
				"error":   "推荐码不存在或已过期",
			})
			return
		}

		// 2. 验证被推荐用户是否存在
		referredUser, err := GetUserByOpenID(req.ReferredUserID)
		if err != nil {
			if middlewares.HandleError(err, "获取被推荐用户信息失败", false) {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "被推荐用户不存在",
					"error":   err.Error(),
				})
				return
			}
		}

		// 3. 检查用户是否已经被推荐过
		if referredUser.ReferredBy != "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "用户已被推荐过",
				"error":   "该用户已经使用过其他推荐码",
			})
			return
		}

		// 4. 获取推荐人信息
		referrer, err := GetUserByReferralCode(req.ReferralCode)
		if err != nil {
			if middlewares.HandleError(err, "获取推荐人信息失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "获取推荐人信息失败",
					"error":   err.Error(),
				})
				return
			}
		}

		// 防止自己推荐自己
		if referrer.ID == referredUser.ID {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "不能推荐自己",
				"error":   "推荐码属于自己，无法使用",
			})
			return
		}

		// 5. 更新被推荐用户的推荐人信息
		err = UpdateUserReferredBy(referredUser.OpenID, req.ReferralCode)
		if err != nil {
			if middlewares.HandleError(err, "更新用户推荐信息失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "更新用户推荐信息失败",
					"error":   err.Error(),
				})
				return
			}
		}

		// 6. 确保推荐人有推荐记录
		err = CreateReferralRecord(req.ReferralCode, referrer.OpenID)
		if err != nil && err != mongo.ErrNoDocuments {
			// 如果不是"已存在"的错误，则记录日志但不影响流程
			middlewares.HandleError(err, "创建推荐记录失败", false)
		}

		// 7. 添加推荐使用记录（暂时不创建佣金记录，等待订单完成时再创建）
		err = AddReferralUsage(req.ReferralCode, referredUser.OpenID, referredUser.UserName, "", 0.0)
		if err != nil {
			if middlewares.HandleError(err, "添加推荐使用记录失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "添加推荐使用记录失败",
					"error":   err.Error(),
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "推荐关系记录成功",
			"data": gin.H{
				"referral_code":    req.ReferralCode,
				"referred_user_id": req.ReferredUserID,
				"referrer_info": gin.H{
					"user_id":     referrer.OpenID,
					"user_name":   referrer.UserName,
					"school":      referrer.School,
					"agent_level": referrer.AgentLevel,
				},
				"discount_rate": calculateDiscountRate(referrer.AgentLevel),
			},
		})
	}
}

// ValidateReferralCodeHandler 验证推荐码处理器
func ValidateReferralCodeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ValidateReferralRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
				"error":   err.Error(),
			})
			return
		}

		// 1. 验证推荐码是否有效
		isValid, err := ValidateReferralCode(req.ReferralCode)
		if err != nil {
			if middlewares.HandleError(err, "验证推荐码失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "验证推荐码失败",
					"error":   err.Error(),
				})
				return
			}
		}

		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "推荐码无效",
				"error":   "推荐码不存在或已过期",
			})
			return
		}

		// 2. 获取推荐人信息
		referrer, err := GetUserByReferralCode(req.ReferralCode)
		if err != nil {
			if middlewares.HandleError(err, "获取推荐人信息失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "获取推荐人信息失败",
					"error":   err.Error(),
				})
				return
			}
		}

		// 3. 计算折扣率（基于代理等级）
		discountRate := calculateDiscountRate(referrer.AgentLevel)

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "推荐码有效",
			"data": gin.H{
				"referral_code": req.ReferralCode,
				"referrer_info": gin.H{
					"user_id":     referrer.OpenID,
					"user_name":   referrer.UserName,
					"school":      referrer.School,
					"agent_level": referrer.AgentLevel,
				},
				"discount_rate": discountRate,
			},
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

		// 1. 根据openID获取用户信息
		user, err := GetUserByOpenID(openID)
		if err != nil {
			if middlewares.HandleError(err, "获取用户信息失败", false) {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "用户不存在",
					"error":   err.Error(),
				})
				return
			}
		}

		// 2. 根据用户ID查询佣金记录
		commissions, err := GetCommissionsByUserID(user.OpenID, status, commissionType)
		if err != nil {
			if middlewares.HandleError(err, "获取佣金记录失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "获取佣金记录失败",
					"error":   err.Error(),
				})
				return
			}
		}

		// 3. 计算总金额统计
		totalAmount, pendingAmount, paidAmount, thisMonthTotal, lastMonthTotal := CalculateCommissionStats(commissions)

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

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "获取返现记录成功",
			"data": gin.H{
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
			},
		})
	}
}
