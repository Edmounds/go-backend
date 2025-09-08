package controllers

import (
	"fmt"
	"log"
	"miniprogram/models"
	"miniprogram/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ===== 推荐服务层 =====

// ReferralCodeService 推荐码服务
type ReferralCodeService struct{}

// NewReferralCodeService 创建推荐码服务实例
func NewReferralCodeService() *ReferralCodeService {
	return &ReferralCodeService{}
}

// GetUserByReferralCode 根据推荐码获取用户信息
func (s *ReferralCodeService) GetUserByReferralCode(referralCode string) (*models.User, error) {
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"referral_code": referralCode}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CalculateDiscountRate 根据代理等级计算折扣率
func (s *ReferralCodeService) CalculateDiscountRate(agentLevel int) float64 {
	switch agentLevel {
	case 0: // 普通用户
		return 0.05 // 5%折扣
	case 1: // 校代理
		return 0.10 // 10%折扣
	case 2: // 区域代理
		return 0.15 // 15%折扣
	default:
		return 0.02 // 默认2%折扣
	}
}

// CreateReferralRecord 创建推荐记录
func (s *ReferralCodeService) CreateReferralRecord(referralCode string, referrerOpenID string) error {
	collection := GetCollection("referrals")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 检查记录是否已存在
	var existingReferral models.Referral
	err := collection.FindOne(ctx, bson.M{"referral_code": referralCode}).Decode(&existingReferral)
	if err == nil {
		return nil // 记录已存在，直接返回
	}
	if err != mongo.ErrNoDocuments {
		return err // 其他错误
	}

	// 创建新的推荐记录
	referral := models.Referral{
		ReferralCode: referralCode,
		UserOpenID:   referrerOpenID,
		UsedBy:       []models.ReferralUsage{},
		CreatedAt:    utils.GetCurrentUTCTime(),
		UpdatedAt:    utils.GetCurrentUTCTime(),
	}

	_, err = collection.InsertOne(ctx, referral)
	return err
}

// GetReferralByCode 根据推荐码获取推荐记录
func (s *ReferralCodeService) GetReferralByCode(referralCode string) (*models.Referral, error) {
	collection := GetCollection("referrals")
	ctx, cancel := CreateDBContext()
	defer cancel()

	var referral models.Referral
	err := collection.FindOne(ctx, bson.M{"referral_code": referralCode}).Decode(&referral)
	if err != nil {
		return nil, err
	}
	return &referral, nil
}

// GetReferralByUserID 根据用户ID获取推荐记录
func (s *ReferralCodeService) GetReferralByUserID(openID string) (*models.Referral, error) {
	collection := GetCollection("referrals")
	ctx, cancel := CreateDBContext()
	defer cancel()

	var referral models.Referral
	err := collection.FindOne(ctx, bson.M{"user_openid": openID}).Decode(&referral)
	if err != nil {
		return nil, err
	}
	return &referral, nil
}

// CommissionService 佣金服务
type CommissionService struct{}

// NewCommissionService 创建佣金服务实例
func NewCommissionService() *CommissionService {
	return &CommissionService{}
}

// GetCommissionsByUserID 根据用户openID获取佣金记录
func (s *CommissionService) GetCommissionsByUserID(openID string, status string, commissionType string) ([]models.Commission, error) {
	collection := GetCollection("commissions")
	ctx, cancel := CreateDBContext()
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
func (s *CommissionService) CalculateCommissionStats(commissions []models.Commission) (float64, float64, float64, float64, float64) {
	var totalAmount, pendingAmount, paidAmount, thisMonthTotal, lastMonthTotal float64

	now := utils.GetCurrentUTCTime()
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

		// 本月佣金统计
		if commission.Date.After(thisMonthStart) {
			thisMonthTotal += commission.Amount
		}

		// 上月佣金统计
		if commission.Date.After(lastMonthStart) && commission.Date.Before(lastMonthEnd) {
			lastMonthTotal += commission.Amount
		}
	}

	return totalAmount, pendingAmount, paidAmount, thisMonthTotal, lastMonthTotal
}

// CreateCommissionRecord 创建佣金记录
func (s *CommissionService) CreateCommissionRecord(openID string, amount float64, commissionType string, description string, orderID string, referredUserOpenID string, referredUserName string) error {
	collection := GetCollection("commissions")
	ctx, cancel := CreateDBContext()
	defer cancel()

	commissionID := GenerateCommissionID()

	commission := models.Commission{
		CommissionID:       commissionID,
		UserOpenID:         openID,
		Amount:             amount,
		Date:               utils.GetCurrentUTCTime(),
		Status:             "pending",
		Type:               commissionType,
		Description:        description,
		OrderID:            orderID,
		ReferredUserOpenID: referredUserOpenID,
		ReferredUserName:   referredUserName,
		CreatedAt:          utils.GetCurrentUTCTime(),
		UpdatedAt:          utils.GetCurrentUTCTime(),
	}

	_, err := collection.InsertOne(ctx, commission)
	return err
}

// ReferralRewardService 推荐奖励服务
type ReferralRewardService struct {
	referralCodeService *ReferralCodeService
	commissionService   *CommissionService
}

// NewReferralRewardService 创建推荐奖励服务实例
func NewReferralRewardService() *ReferralRewardService {
	return &ReferralRewardService{
		referralCodeService: NewReferralCodeService(),
		commissionService:   NewCommissionService(),
	}
}

// ValidateReferralCode 验证推荐码是否存在
func (s *ReferralRewardService) ValidateReferralCode(referralCode string) (*models.User, error) {
	return s.referralCodeService.GetUserByReferralCode(referralCode)
}

// AddReferralUsage 添加推荐使用记录（仅记录注册时的使用关系）
func (s *ReferralRewardService) AddReferralUsage(referralCode string, referredUserOpenID string, referredUserName string) error {
	collection := GetCollection("referrals")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 创建新的使用记录
	usage := models.ReferralUsage{
		UserOpenID: referredUserOpenID,
		UserName:   referredUserName,
		UsedAt:     utils.GetCurrentUTCTime(),
	}

	// 更新推荐记录，添加使用记录
	filter := bson.M{"referral_code": referralCode}
	update := bson.M{
		"$push": bson.M{"used_by": usage},
		"$set":  bson.M{"updated_at": utils.GetCurrentUTCTime()},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// ProcessNewUserReferral 处理新用户的推荐关系
func (s *ReferralRewardService) ProcessNewUserReferral(userOpenID string, referralCode string) error {
	// 1. 获取推荐人信息
	referrer, err := s.referralCodeService.GetUserByReferralCode(referralCode)
	if err != nil {
		return err
	}

	// 2. 确保推荐人有推荐记录
	err = s.referralCodeService.CreateReferralRecord(referralCode, referrer.OpenID)
	if err != nil && err != mongo.ErrNoDocuments {
		// 如果不是"已存在"的错误，则返回错误
		return err
	}

	// 3. 更新被推荐用户的 referred_by 字段
	err = s.UpdateUserReferredBy(userOpenID, referralCode)
	if err != nil {
		return err
	}

	// 4. 获取被推荐用户信息，用于记录使用记录
	referredUser, err := GetUserByOpenID(userOpenID)
	if err != nil {
		return err
	}

	// 5. 添加推荐使用记录到推荐人的 referral 文档中
	err = s.AddReferralUsage(referralCode, userOpenID, referredUser.UserName)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUserReferredBy 更新用户的推荐人信息
func (s *ReferralRewardService) UpdateUserReferredBy(openID string, referralCode string) error {
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	filter := bson.M{"openID": openID}
	update := bson.M{
		"$set": bson.M{
			"referred_by": referralCode,
			"updated_at":  utils.GetCurrentUTCTime(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// ProcessReferralReward 处理推荐奖励 - 仅新用户首次购买时给推荐人4元固定返现
func (s *ReferralRewardService) ProcessReferralReward(referredUserOpenID string, orderID string, orderAmount float64) error {
	// 固定返现金额4元
	const FIXED_REFERRAL_COMMISSION = 4.0

	// 1. 获取被推荐用户信息
	referredUser, err := GetUserByOpenID(referredUserOpenID)
	if err != nil {
		return err
	}

	// 如果用户没有推荐人，直接返回
	if referredUser.ReferredBy == "" {
		return nil
	}

	// 2. 检查用户是否已经使用过推荐优惠
	// 只有新用户首次购买才给推荐人返现
	if referredUser.HasUsedReferralDiscount {
		return nil // 用户已经享受过推荐优惠，推荐人也已经获得过返现
	}

	// 3. 检查是否为用户首次订单
	isFirstOrder, err := s.isUserFirstCompletedOrder(referredUserOpenID, orderID)
	if err != nil {
		return err
	}

	if !isFirstOrder {
		return nil // 不是首次订单，不给推荐人返现
	}

	// 4. 获取推荐人信息
	referrer, err := s.referralCodeService.GetUserByReferralCode(referredUser.ReferredBy)
	if err != nil {
		return err
	}

	// 5. 创建固定金额的佣金记录
	description := "推荐新用户首次购买获得返现"
	err = s.commissionService.CreateCommissionRecord(referrer.OpenID, FIXED_REFERRAL_COMMISSION, "referral", description, orderID, referredUser.OpenID, referredUser.UserName)
	if err != nil {
		return err
	}

	return nil
}

// isUserFirstCompletedOrder 检查这是否是用户首次完成的订单
func (s *ReferralRewardService) isUserFirstCompletedOrder(userOpenID string, currentOrderID string) (bool, error) {
	collection := GetCollection("orders")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 查找该用户所有已完成的订单，排除当前订单
	count, err := collection.CountDocuments(ctx, bson.M{
		"user_openid": userOpenID,
		"status":      "completed",
		"_id":         bson.M{"$ne": currentOrderID}, // 排除当前订单
	})
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// 注意：UpdateReferralUsageStatus 函数已删除
// 因为推荐使用记录和佣金系统已分离，used_by只记录注册关系，不再需要更新状态和佣金

// AgentTieredCommissionService 代理分级提成服务
type AgentTieredCommissionService struct {
	commissionService *CommissionService
}

// NewAgentTieredCommissionService 创建代理分级提成服务实例
func NewAgentTieredCommissionService() *AgentTieredCommissionService {
	return &AgentTieredCommissionService{
		commissionService: NewCommissionService(),
	}
}

// ProcessAgentCommission 处理代理分级提成
func (s *AgentTieredCommissionService) ProcessAgentCommission(userOpenID string, orderAmount float64, orderID string) error {
	// 获取用户信息
	user, err := GetUserByOpenID(userOpenID)
	if err != nil {
		return err
	}

	// 如果用户没有学校信息，无法匹配校代理
	if user.School == "" {
		return nil
	}

	// 1. 处理校代理提成
	err = s.processSchoolAgentCommission(user, orderAmount, orderID)
	if err != nil {
		log.Printf("处理校代理提成失败: %v", err)
	}

	// 2. 处理区代理提成
	err = s.processRegionalAgentCommission(user, orderAmount, orderID)
	if err != nil {
		log.Printf("处理区代理提成失败: %v", err)
	}

	return nil
}

// processSchoolAgentCommission 处理校代理提成
func (s *AgentTieredCommissionService) processSchoolAgentCommission(user *models.User, orderAmount float64, orderID string) error {
	// 查找该学校的校代理
	schoolAgent, err := s.findSchoolAgent(user.School)
	if err != nil || schoolAgent == nil {
		return nil // 没有找到校代理，直接返回
	}

	// 更新校代理的累计销售额
	newAccumulatedSales := schoolAgent.AccumulatedSales + orderAmount
	err = s.updateAgentAccumulatedSales(schoolAgent.OpenID, newAccumulatedSales)
	if err != nil {
		return err
	}

	// 计算校代理分级提成
	commissionRate := s.calculateSchoolAgentCommissionRate(newAccumulatedSales)
	commissionAmount := orderAmount * commissionRate

	// 创建校代理提成记录
	description := fmt.Sprintf("校代理分级提成 - %s学校用户消费", user.School)
	return s.commissionService.CreateCommissionRecord(
		schoolAgent.OpenID,
		commissionAmount,
		"agent",
		description,
		orderID,
		user.OpenID,
		user.UserName,
	)
}

// processRegionalAgentCommission 处理区代理提成
func (s *AgentTieredCommissionService) processRegionalAgentCommission(user *models.User, orderAmount float64, orderID string) error {
	// 查找该区域的区代理
	regionalAgent, err := s.findRegionalAgent(user.City)
	if err != nil || regionalAgent == nil {
		return nil // 没有找到区代理，直接返回
	}

	// 更新区代理的累计销售额
	newAccumulatedSales := regionalAgent.AccumulatedSales + orderAmount
	err = s.updateAgentAccumulatedSales(regionalAgent.OpenID, newAccumulatedSales)
	if err != nil {
		return err
	}

	// 计算区代理分级提成
	commissionRate := s.calculateRegionalAgentCommissionRate(newAccumulatedSales)
	commissionAmount := orderAmount * commissionRate

	// 创建区代理提成记录
	description := fmt.Sprintf("区代理分级提成 - %s地区用户消费", user.City)
	return s.commissionService.CreateCommissionRecord(
		regionalAgent.OpenID,
		commissionAmount,
		"agent",
		description,
		orderID,
		user.OpenID,
		user.UserName,
	)
}

// findSchoolAgent 查找学校的校代理
func (s *AgentTieredCommissionService) findSchoolAgent(school string) (*models.User, error) {
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	var agent models.User
	err := collection.FindOne(ctx, bson.M{
		"is_agent":    true,
		"agent_level": 1, // 校代理
		"school":      school,
	}).Decode(&agent)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // 没有找到校代理
		}
		return nil, err
	}

	return &agent, nil
}

// findRegionalAgent 查找区域的区代理
func (s *AgentTieredCommissionService) findRegionalAgent(city string) (*models.User, error) {
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	var agent models.User
	err := collection.FindOne(ctx, bson.M{
		"is_agent":        true,
		"agent_level":     2, // 区代理
		"managed_regions": bson.M{"$in": []string{city}},
	}).Decode(&agent)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // 没有找到区代理
		}
		return nil, err
	}

	return &agent, nil
}

// updateAgentAccumulatedSales 更新代理的累计销售额
func (s *AgentTieredCommissionService) updateAgentAccumulatedSales(agentOpenID string, newAccumulatedSales float64) error {
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	filter := bson.M{"openID": agentOpenID}
	update := bson.M{
		"$set": bson.M{
			"accumulated_sales": newAccumulatedSales,
			"updated_at":        utils.GetCurrentUTCTime(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// calculateSchoolAgentCommissionRate 计算校代理分级提成率
func (s *AgentTieredCommissionService) calculateSchoolAgentCommissionRate(accumulatedSales float64) float64 {
	switch {
	case accumulatedSales >= 100000: // 10万元及以上
		return 0.10
	case accumulatedSales >= 80000: // 8-10万元
		return 0.09
	case accumulatedSales >= 60000: // 6-8万元
		return 0.08
	case accumulatedSales >= 40000: // 4-6万元
		return 0.07
	case accumulatedSales >= 20000: // 2-4万元
		return 0.06
	default: // 0-2万元
		return 0.05
	}
}

// calculateRegionalAgentCommissionRate 计算区代理分级提成率
func (s *AgentTieredCommissionService) calculateRegionalAgentCommissionRate(accumulatedSales float64) float64 {
	switch {
	case accumulatedSales >= 16000000: // 1600万元及以上
		return 0.15
	case accumulatedSales >= 12000000: // 1200-1600万元
		return 0.13
	case accumulatedSales >= 8000000: // 800-1200万元
		return 0.12
	case accumulatedSales >= 4000000: // 400-800万元
		return 0.11
	default: // 0-400万元
		return 0.10
	}
}

// WechatQRCodeService 微信小程序码服务
type WechatQRCodeService struct{}

// NewWechatQRCodeService 创建微信小程序码服务实例
func NewWechatQRCodeService() *WechatQRCodeService {
	return &WechatQRCodeService{}
}

// ValidateScene 验证场景值参数
func (s *WechatQRCodeService) ValidateScene(scene string) error {
	if len(scene) == 0 || len(scene) > 32 {
		return mongo.ErrNoDocuments // 使用标准错误，调用方可以统一处理
	}
	return nil
}

// BuildQRCodePayload 构建小程序码请求载荷
func (s *WechatQRCodeService) BuildQRCodePayload(req models.UnlimitedQRCodeRequest) map[string]interface{} {
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

	return payload
}
