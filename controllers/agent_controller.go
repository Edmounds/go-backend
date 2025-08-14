package controllers

import (
	"context"
	"miniprogram/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// WithdrawRequest 提取佣金请求
type WithdrawRequest struct {
	Amount         float64            `json:"amount" binding:"required,min=0.01"`
	WithdrawMethod string             `json:"withdraw_method" binding:"required"`
	AccountInfo    models.AccountInfo `json:"account_info"`
}

// ===== 辅助函数 =====

// generateWithdrawID 生成提现ID
func generateWithdrawID() string {
	return "WD" + time.Now().Format("20060102150405") + generateRandomString(4)
}

// IsValidAgent 验证用户是否为代理
func IsValidAgent(openID string) (*models.User, error) {
	user, err := GetUserByOpenID(openID)
	if err != nil {
		return nil, err
	}

	if !user.IsAgent || user.AgentLevel < 1 {
		return nil, mongo.ErrNoDocuments // 不是代理，返回错误
	}

	return user, nil
}

// GetManagedUsers 根据代理等级获取管理的用户
func GetManagedUsers(agent *models.User, school string, region string) ([]models.User, error) {
	collection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构建查询条件
	filter := bson.M{}

	// 根据代理等级设置不同的查询条件
	switch agent.AgentLevel {
	case 1: // 校代理
		// 校代理只能管理同一学校的用户
		filter["school"] = agent.School
	case 2: // 区域代理
		// 区域代理可以管理指定区域的用户
		if len(agent.ManagedRegions) > 0 {
			filter["city"] = bson.M{"$in": agent.ManagedRegions}
		}
	}

	// 添加额外的筛选条件
	if school != "" {
		filter["school"] = school
	}
	if region != "" {
		filter["city"] = region
	}

	// 排除代理自己
	filter["openID"] = bson.M{"$ne": agent.OpenID}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetUserOrderStats 获取用户订单统计
func GetUserOrderStats(openID string) (int, float64, error) {
	collection := GetCollection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 根据用户openID查询订单
	filter := bson.M{"user_openid": openID, "status": "completed"}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(ctx)

	totalOrders := 0
	totalSpent := 0.0

	for cursor.Next(ctx) {
		var order struct {
			TotalAmount float64 `bson:"total_amount"`
		}
		if err := cursor.Decode(&order); err != nil {
			continue
		}
		totalOrders++
		totalSpent += order.TotalAmount
	}

	return totalOrders, totalSpent, nil
}

// GetAgentStatistics 获取代理统计信息
func GetAgentStatistics(agent *models.User) (map[string]interface{}, error) {
	managedUsers, err := GetManagedUsers(agent, "", "")
	if err != nil {
		return nil, err
	}

	activeUsers := 0
	newUsersThisMonth := 0
	totalRevenue := 0.0
	totalOrders := 0

	now := time.Now()
	thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	for _, user := range managedUsers {
		// 统计活跃用户（本月有订单）
		orders, spent, err := GetUserOrderStats(user.OpenID)
		if err == nil {
			if orders > 0 {
				activeUsers++
			}
			totalOrders += orders
			totalRevenue += spent
		}

		// 统计本月新用户
		if user.CreatedAt.After(thisMonthStart) {
			newUsersThisMonth++
		}
	}

	return map[string]interface{}{
		"active_users":    activeUsers,
		"new_users_month": newUsersThisMonth,
		"total_revenue":   totalRevenue,
		"total_orders":    totalOrders,
	}, nil
}

// ===== 数据库查询函数 =====

// CreateWithdrawRecord 创建提现记录
func CreateWithdrawRecord(openID string, amount float64, withdrawMethod string, accountInfo models.AccountInfo) (*models.WithdrawRecord, error) {
	collection := GetCollection("withdrawals")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	withdrawID := generateWithdrawID()
	processingFee := amount * 0.01 // 1%手续费
	actualAmount := amount - processingFee
	estimatedArrival := time.Now().Add(48 * time.Hour) // 2个工作日

	record := models.WithdrawRecord{
		WithdrawID:       withdrawID,
		UserOpenID:       openID,
		Amount:           amount,
		WithdrawMethod:   withdrawMethod,
		AccountInfo:      accountInfo,
		Status:           "pending",
		ProcessingFee:    processingFee,
		ActualAmount:     actualAmount,
		EstimatedArrival: estimatedArrival,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	result, err := collection.InsertOne(ctx, record)
	if err != nil {
		return nil, err
	}

	record.ID = result.InsertedID.(primitive.ObjectID)
	return &record, nil
}

// GetWithdrawRecords 获取提现记录
func GetWithdrawRecords(openID string) ([]models.WithdrawRecord, error) {
	collection := GetCollection("withdrawals")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_openid": openID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []models.WithdrawRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}

	return records, nil
}

// GetAvailableCommission 获取可提取佣金余额
func GetAvailableCommission(openID string) (float64, error) {
	// 获取所有已完成的佣金
	commissions, err := GetCommissionsByUserID(openID, "paid", "")
	if err != nil {
		return 0, err
	}

	totalCommission := 0.0
	for _, commission := range commissions {
		totalCommission += commission.Amount
	}

	// 获取已提取的金额
	withdrawRecords, err := GetWithdrawRecords(openID)
	if err != nil {
		return 0, err
	}

	withdrawnAmount := 0.0
	for _, record := range withdrawRecords {
		if record.Status == "completed" {
			withdrawnAmount += record.Amount
		}
	}

	return totalCommission - withdrawnAmount, nil
}

// GetAgentSalesData 获取代理销售数据
func GetAgentSalesData(agent *models.User, startDate, endDate string) (map[string]interface{}, error) {
	collection := GetCollection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构建时间过滤条件
	filter := bson.M{"status": "completed"}

	// 解析时间参数
	var start, end time.Time
	var err error

	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			// 默认为本月第一天
			now := time.Now()
			start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		}
	} else {
		// 默认为本月第一天
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			end = time.Now()
		} else {
			// 设置为当天结束时间
			end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())
		}
	} else {
		end = time.Now()
	}

	filter["created_at"] = bson.M{
		"$gte": start,
		"$lte": end,
	}

	// 根据代理等级添加用户筛选条件
	managedUsers, err := GetManagedUsers(agent, "", "")
	if err != nil {
		return nil, err
	}

	// 获取管理用户的openID列表
	var userOpenIDs []string
	for _, user := range managedUsers {
		userOpenIDs = append(userOpenIDs, user.OpenID)
	}

	if len(userOpenIDs) > 0 {
		filter["user_openid"] = bson.M{"$in": userOpenIDs}
	}

	// 查询订单
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 统计数据
	totalSales := 0.0
	totalOrders := 0
	productSales := make(map[string]struct {
		Name         string
		QuantitySold int
		TotalRevenue float64
	})
	monthlySales := make(map[string]struct {
		Sales  float64
		Orders int
	})
	userSales := make(map[string]struct {
		UserName    string
		TotalSpent  float64
		OrdersCount int
	})

	// 处理订单数据
	for cursor.Next(ctx) {
		var order struct {
			TotalAmount float64 `bson:"total_amount"`
			UserOpenID  string  `bson:"user_openid"`
			Items       []struct {
				ProductID string  `bson:"product_id"`
				Quantity  int     `bson:"quantity"`
				Price     float64 `bson:"price"`
			} `bson:"items"`
			CreatedAt time.Time `bson:"created_at"`
		}

		if err := cursor.Decode(&order); err != nil {
			continue
		}

		totalSales += order.TotalAmount
		totalOrders++

		// 按月统计
		monthKey := order.CreatedAt.Format("2006-01")
		monthData := monthlySales[monthKey]
		monthData.Sales += order.TotalAmount
		monthData.Orders++
		monthlySales[monthKey] = monthData

		// 按用户统计
		if user, err := GetUserByOpenID(order.UserOpenID); err == nil {
			userData := userSales[order.UserOpenID]
			userData.UserName = user.UserName
			userData.TotalSpent += order.TotalAmount
			userData.OrdersCount++
			userSales[order.UserOpenID] = userData
		}

		// 按产品统计
		for _, item := range order.Items {
			productData := productSales[item.ProductID]
			productData.Name = "产品 " + item.ProductID // 这里可以从产品表获取真实名称
			productData.QuantitySold += item.Quantity
			productData.TotalRevenue += item.Price * float64(item.Quantity)
			productSales[item.ProductID] = productData
		}
	}

	// 计算佣金
	commissionRate := calculateCommissionRate(agent.AgentLevel)
	totalCommission := totalSales * commissionRate

	// 构建销售摘要
	salesSummary := gin.H{
		"total_sales":      totalSales,
		"total_orders":     totalOrders,
		"total_commission": totalCommission,
		"period": gin.H{
			"start_date": start.Format("2006-01-02"),
			"end_date":   end.Format("2006-01-02"),
		},
	}

	// 构建产品销售数据
	var salesByProduct []gin.H
	for productID, data := range productSales {
		commission := data.TotalRevenue * commissionRate
		salesByProduct = append(salesByProduct, gin.H{
			"product_id":    productID,
			"product_name":  data.Name,
			"quantity_sold": data.QuantitySold,
			"total_revenue": data.TotalRevenue,
			"commission":    commission,
		})
	}

	// 构建月度趋势数据
	var monthlyTrend []gin.H
	for month, data := range monthlySales {
		commission := data.Sales * commissionRate
		monthlyTrend = append(monthlyTrend, gin.H{
			"month":      month,
			"sales":      data.Sales,
			"orders":     data.Orders,
			"commission": commission,
		})
	}

	// 构建顶级客户数据
	var topPerformers []gin.H
	for userID, data := range userSales {
		topPerformers = append(topPerformers, gin.H{
			"user_id":      userID,
			"user_name":    data.UserName,
			"total_spent":  data.TotalSpent,
			"orders_count": data.OrdersCount,
		})
	}

	return map[string]interface{}{
		"sales_summary":    salesSummary,
		"sales_by_product": salesByProduct,
		"monthly_trend":    monthlyTrend,
		"top_performers":   topPerformers,
	}, nil
}

// GetAgentUsersHandler 获取代理管理的用户列表处理器
func GetAgentUsersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		openID := c.Param("user_id")
		school := c.Query("school")
		region := c.Query("region")

		// 1. 验证用户是否为代理
		agent, err := IsValidAgent(openID)
		if err != nil {
			ErrorResponse(c, http.StatusForbidden, 403, "用户不是代理或权限不足", err)
			return
		}

		// 2. 根据代理等级查询管理的用户
		managedUsers, err := GetManagedUsers(agent, school, region)
		if err != nil {
			InternalServerErrorResponse(c, "获取管理用户列表失败", err)
			return
		}

		// 3. 为每个用户添加订单统计信息
		var usersData []gin.H
		for _, user := range managedUsers {
			totalOrders, totalSpent, err := GetUserOrderStats(user.OpenID)
			if err != nil {
				totalOrders = 0
				totalSpent = 0.0
			}

			// 确定代理类型显示
			agentType := "普通用户"
			if user.IsAgent {
				switch user.AgentLevel {
				case 1:
					agentType = "校代理"
				case 2:
					agentType = "区域代理"
				}
			}

			usersData = append(usersData, gin.H{
				"_id":           user.ID,
				"openID":        user.OpenID,
				"user_name":     user.UserName,
				"school":        user.School,
				"city":          user.City,
				"age":           user.Age,
				"phone":         user.Phone,
				"agent_level":   user.AgentLevel,
				"agent_type":    agentType,
				"referral_code": user.ReferralCode,
				"created_at":    user.CreatedAt.Format(time.RFC3339),
				"updated_at":    user.UpdatedAt.Format(time.RFC3339),
				"total_orders":  totalOrders,
				"total_spent":   totalSpent,
			})
		}

		// 4. 获取代理统计信息
		statistics, err := GetAgentStatistics(agent)
		if err != nil {
			// 如果获取统计失败，使用默认值
			statistics = map[string]interface{}{
				"active_users":    0,
				"new_users_month": 0,
				"total_revenue":   0.0,
				"total_orders":    0,
			}
		}

		// 5. 构建代理信息
		agentTypeStr := "普通用户"
		switch agent.AgentLevel {
		case 1:
			agentTypeStr = "校代理"
		case 2:
			agentTypeStr = "区域代理"
		}

		SuccessResponse(c, "获取管理用户列表成功", gin.H{
			"users":       usersData,
			"total_users": len(managedUsers),
			"agent_info": gin.H{
				"agent_level":     agent.AgentLevel,
				"agent_type":      agentTypeStr,
				"managed_schools": agent.ManagedSchools,
				"managed_regions": agent.ManagedRegions,
			},
			"statistics": statistics,
		})
	}
}

// GetAgentSalesHandler 获取代理销售数据处理器
func GetAgentSalesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		openID := c.Param("user_id")
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")

		// 1. 验证用户是否为代理
		agent, err := IsValidAgent(openID)
		if err != nil {
			ErrorResponse(c, http.StatusForbidden, 403, "用户不是代理或权限不足", err)
			return
		}

		// 2. 根据时间范围查询销售数据
		salesData, err := GetAgentSalesData(agent, startDate, endDate)
		if err != nil {
			InternalServerErrorResponse(c, "获取销售数据失败", err)
			return
		}

		SuccessResponse(c, "获取销售数据成功", salesData)
	}
}

// WithdrawCommissionHandler 提取佣金处理器
func WithdrawCommissionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		openID := c.Param("user_id")
		var req WithdrawRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 1. 验证用户是否为代理
		agent, err := IsValidAgent(openID)
		if err != nil {
			ErrorResponse(c, http.StatusForbidden, 403, "用户不是代理或权限不足", err)
			return
		}

		// 验证提取方式
		validMethods := map[string]bool{
			"wechat":        true,
			"alipay":        true,
			"bank_transfer": true,
		}

		if !validMethods[req.WithdrawMethod] {
			BadRequestResponse(c, "不支持的提取方式", nil)
			return
		}

		// 2. 检查可提取佣金余额
		availableAmount, err := GetAvailableCommission(agent.OpenID)
		if err != nil {
			InternalServerErrorResponse(c, "获取可提取佣金失败", err)
			return
		}

		// 3. 验证提取金额
		if req.Amount > availableAmount {
			ErrorResponse(c, http.StatusBadRequest, 400, "提取金额不能超过可提取佣金余额", nil)
			return
		}

		// 最小提取金额检查
		if req.Amount < 10.0 {
			BadRequestResponse(c, "最小提取金额为10元", nil)
			return
		}

		// 4. 创建提取记录
		withdrawRecord, err := CreateWithdrawRecord(agent.OpenID, req.Amount, req.WithdrawMethod, req.AccountInfo)
		if err != nil {
			InternalServerErrorResponse(c, "创建提取记录失败", err)
			return
		}

		// 5. TODO: 在实际应用中，这里应该调用第三方支付接口处理提取
		// 例如：微信支付、支付宝等

		SuccessResponse(c, "提取申请提交成功", gin.H{
			"withdraw_id":       withdrawRecord.WithdrawID,
			"openID":            openID,
			"amount":            withdrawRecord.Amount,
			"withdraw_method":   withdrawRecord.WithdrawMethod,
			"account_info":      withdrawRecord.AccountInfo,
			"status":            withdrawRecord.Status,
			"estimated_arrival": withdrawRecord.EstimatedArrival.Format(time.RFC3339),
			"created_at":        withdrawRecord.CreatedAt.Format(time.RFC3339),
			"processing_fee":    withdrawRecord.ProcessingFee,
			"actual_amount":     withdrawRecord.ActualAmount,
		})
	}
}
