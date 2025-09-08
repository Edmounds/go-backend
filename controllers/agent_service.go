package controllers

import (
	"fmt"
	"miniprogram/models"
	"miniprogram/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ===== 代理服务层 =====

// AgentUserService 代理用户管理服务
type AgentUserService struct{}

// NewAgentUserService 创建代理用户服务实例
func NewAgentUserService() *AgentUserService {
	return &AgentUserService{}
}

// IsValidAgent 验证用户是否为代理
func (s *AgentUserService) IsValidAgent(openID string) (*models.User, error) {
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
func (s *AgentUserService) GetManagedUsers(agent *models.User, school string, region string) ([]models.User, error) {
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
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
func (s *AgentUserService) GetUserOrderStats(openID string) (int, float64, error) {
	collection := GetCollection("orders")
	ctx, cancel := CreateDBContext()
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
func (s *AgentUserService) GetAgentStatistics(agent *models.User) (map[string]interface{}, error) {
	managedUsers, err := s.GetManagedUsers(agent, "", "")
	if err != nil {
		return nil, err
	}

	activeUsers := 0
	newUsersThisMonth := 0
	totalRevenue := 0.0
	totalOrders := 0

	now := utils.GetCurrentUTCTime()
	thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	for _, user := range managedUsers {
		// 统计活跃用户（本月有订单）
		orders, spent, err := s.GetUserOrderStats(user.OpenID)
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

// AgentCommissionService 代理佣金服务
type AgentCommissionService struct{}

// NewAgentCommissionService 创建代理佣金服务实例
func NewAgentCommissionService() *AgentCommissionService {
	return &AgentCommissionService{}
}

// GetCommissionsByDateRange 根据时间范围获取佣金记录
func (s *AgentCommissionService) GetCommissionsByDateRange(openID string, startDate, endDate time.Time) ([]models.Commission, error) {
	collection := GetCollection("commissions")
	ctx, cancel := CreateDBContext()
	defer cancel()

	filter := bson.M{
		"user_openid": openID,
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
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

// GetCommissionDashboard 获取代理佣金仪表板数据
func (s *AgentCommissionService) GetCommissionDashboard(openID string) (map[string]interface{}, error) {
	now := utils.GetCurrentUTCTime()

	// 今日开始时间
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 本月开始时间
	thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// 上月开始时间
	lastMonthStart := thisMonthStart.AddDate(0, -1, 0)
	lastMonthEnd := thisMonthStart.Add(-time.Second)

	// 获取今日佣金
	todayCommissions, err := s.GetCommissionsByDateRange(openID, todayStart, now)
	if err != nil {
		return nil, err
	}

	todayIncome := 0.0
	for _, commission := range todayCommissions {
		if commission.Status == "paid" {
			todayIncome += commission.Amount
		}
	}

	// 获取本月佣金
	thisMonthCommissions, err := s.GetCommissionsByDateRange(openID, thisMonthStart, now)
	if err != nil {
		return nil, err
	}

	thisMonthIncome := 0.0
	for _, commission := range thisMonthCommissions {
		if commission.Status == "paid" {
			thisMonthIncome += commission.Amount
		}
	}

	// 获取上月佣金
	lastMonthCommissions, err := s.GetCommissionsByDateRange(openID, lastMonthStart, lastMonthEnd)
	if err != nil {
		return nil, err
	}

	lastMonthIncome := 0.0
	for _, commission := range lastMonthCommissions {
		if commission.Status == "paid" {
			lastMonthIncome += commission.Amount
		}
	}

	// 计算月度对比
	monthComparison := 0.0
	if lastMonthIncome > 0 {
		monthComparison = ((thisMonthIncome - lastMonthIncome) / lastMonthIncome) * 100
	} else if thisMonthIncome > 0 {
		monthComparison = 100.0
	}

	// 获取可提取佣金
	withdrawService := NewAgentWithdrawService()
	availableCommission, err := withdrawService.GetAvailableCommission(openID)
	if err != nil {
		availableCommission = 0.0
	}

	return map[string]interface{}{
		"today_income":         todayIncome,
		"this_month_income":    thisMonthIncome,
		"last_month_income":    lastMonthIncome,
		"month_comparison":     monthComparison,
		"available_commission": availableCommission,
		"total_commissions":    len(thisMonthCommissions),
	}, nil
}

// GetCommissionDetails 获取代理佣金明细数据（按月统计）
func (s *AgentCommissionService) GetCommissionDetails(openID string, months int) (map[string]interface{}, error) {
	now := utils.GetCurrentUTCTime()
	var monthlyData []map[string]interface{}

	for i := months - 1; i >= 0; i-- {
		// 计算每个月的开始和结束时间
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

		// 获取该月佣金
		commissions, err := s.GetCommissionsByDateRange(openID, monthStart, monthEnd)
		if err != nil {
			continue
		}

		monthIncome := 0.0
		var commissionDetails []map[string]interface{}

		for _, commission := range commissions {
			if commission.Status == "paid" {
				monthIncome += commission.Amount
			}

			// 添加详细的佣金记录信息（移除敏感的openID信息）
			commissionDetails = append(commissionDetails, map[string]interface{}{
				"commission_id":      commission.CommissionID,
				"amount":             commission.Amount,
				"status":             commission.Status,
				"type":               commission.Type,
				"description":        commission.Description,
				"order_id":           commission.OrderID,
				"referred_user_name": commission.ReferredUserName, // 只保留用户名，移除openID
				"date":               commission.Date.Format(time.RFC3339),
				"created_at":         commission.CreatedAt.Format(time.RFC3339),
			})
		}

		monthlyData = append(monthlyData, map[string]interface{}{
			"month":              monthStart.Format("2006年1月"),
			"month_code":         monthStart.Format("2006-01"),
			"income":             monthIncome,
			"commissions_count":  len(commissions),
			"commission_details": commissionDetails,
		})
	}

	// 计算总收入
	totalIncome := 0.0
	totalCommissions := 0
	for _, data := range monthlyData {
		totalIncome += data["income"].(float64)
		totalCommissions += data["commissions_count"].(int)
	}

	return map[string]interface{}{
		"monthly_data":      monthlyData,
		"total_income":      totalIncome,
		"total_commissions": totalCommissions,
		"period_months":     months,
	}, nil
}

// AgentWithdrawService 代理提现服务
type AgentWithdrawService struct{}

// NewAgentWithdrawService 创建代理提现服务实例
func NewAgentWithdrawService() *AgentWithdrawService {
	return &AgentWithdrawService{}
}

// CheckPendingWithdrawRecord 检查是否存在待处理的提现记录
func (s *AgentWithdrawService) CheckPendingWithdrawRecord(openID string) (*models.WithdrawRecord, error) {
	collection := GetCollection("withdrawals")
	ctx, cancel := CreateDBContext()
	defer cancel()

	var record models.WithdrawRecord
	err := collection.FindOne(ctx, bson.M{
		"user_openid": openID,
		"status":      bson.M{"$in": []string{"pending", "processing"}},
	}).Decode(&record)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // 没有找到待处理的记录
		}
		return nil, err
	}

	return &record, nil
}

// CreateWithdrawRecord 创建提现记录
func (s *AgentWithdrawService) CreateWithdrawRecord(openID string, amount float64) (*models.WithdrawRecord, error) {
	collection := GetCollection("withdrawals")
	ctx, cancel := CreateDBContext()
	defer cancel()

	withdrawID := GenerateWithdrawID()

	record := models.WithdrawRecord{
		WithdrawID:     withdrawID,
		UserOpenID:     openID,
		Amount:         amount,
		WithdrawMethod: "wechat", // 微信支付企业转账
		Status:         "pending",
		OutBillNo:      withdrawID, // 使用withdrawID作为商户单号，确保唯一性
		CreatedAt:      utils.GetCurrentUTCTime(),
		UpdatedAt:      utils.GetCurrentUTCTime(),
	}

	result, err := collection.InsertOne(ctx, record)
	if err != nil {
		return nil, err
	}

	record.ID = result.InsertedID.(primitive.ObjectID)
	return &record, nil
}

// GetWithdrawRecords 获取提现记录
func (s *AgentWithdrawService) GetWithdrawRecords(openID string) ([]models.WithdrawRecord, error) {
	collection := GetCollection("withdrawals")
	ctx, cancel := CreateDBContext()
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

// GetWithdrawRecordByID 根据withdraw_id获取单个提取记录
func (s *AgentWithdrawService) GetWithdrawRecordByID(withdrawID string) (*models.WithdrawRecord, error) {
	collection := GetCollection("withdrawals")
	ctx, cancel := CreateDBContext()
	defer cancel()

	var record models.WithdrawRecord
	err := collection.FindOne(ctx, bson.M{"withdraw_id": withdrawID}).Decode(&record)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("提取记录不存在")
		}
		return nil, err
	}

	return &record, nil
}

// GetAvailableCommission 获取可提取佣金余额
func (s *AgentWithdrawService) GetAvailableCommission(openID string) (float64, error) {
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
	withdrawRecords, err := s.GetWithdrawRecords(openID)
	if err != nil {
		return 0, err
	}

	withdrawnAmount := 0.0
	for _, record := range withdrawRecords {
		if record.Status == "completed" {
			withdrawnAmount += record.Amount
		}
	}

	availableAmount := totalCommission - withdrawnAmount

	// 添加调试日志
	fmt.Printf("[DEBUG] 用户 %s 可提取佣金计算: 总佣金=%.6f, 已提取=%.6f, 可提取=%.6f\n",
		openID, totalCommission, withdrawnAmount, availableAmount)
	fmt.Printf("[DEBUG] 佣金记录数量: %d, 提取记录数量: %d\n",
		len(commissions), len(withdrawRecords))

	return availableAmount, nil
}

// AgentSalesService 代理销售服务
type AgentSalesService struct{}

// NewAgentSalesService 创建代理销售服务实例
func NewAgentSalesService() *AgentSalesService {
	return &AgentSalesService{}
}

// GetSalesData 获取代理销售数据
func (s *AgentSalesService) GetSalesData(agent *models.User, startDate, endDate string) (map[string]interface{}, error) {
	collection := GetCollection("orders")
	ctx, cancel := CreateDBContext()
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
			now := utils.GetCurrentUTCTime()
			start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		}
	} else {
		// 默认为本月第一天
		now := utils.GetCurrentUTCTime()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			end = utils.GetCurrentUTCTime()
		} else {
			// 设置为当天结束时间
			end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())
		}
	} else {
		end = utils.GetCurrentUTCTime()
	}

	filter["created_at"] = bson.M{
		"$gte": start,
		"$lte": end,
	}

	// 根据代理等级添加用户筛选条件
	userService := NewAgentUserService()
	managedUsers, err := userService.GetManagedUsers(agent, "", "")
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

	return s.processSalesData(cursor, agent, start, end)
}

// processSalesData 处理销售数据统计
func (s *AgentSalesService) processSalesData(cursor *mongo.Cursor, agent *models.User, start, end time.Time) (map[string]interface{}, error) {
	ctx, cancel := CreateDBContext()
	defer cancel()

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
	commissionRate := CalculateCommissionRate(agent.AgentLevel)
	totalCommission := totalSales * commissionRate

	response := s.buildSalesResponse(totalSales, totalOrders, totalCommission, start, end, productSales, monthlySales, userSales, commissionRate)
	return response, nil
}

// buildSalesResponse 构建销售数据响应
func (s *AgentSalesService) buildSalesResponse(totalSales float64, totalOrders int, totalCommission float64, start, end time.Time,
	productSales map[string]struct {
		Name         string
		QuantitySold int
		TotalRevenue float64
	},
	monthlySales map[string]struct {
		Sales  float64
		Orders int
	},
	userSales map[string]struct {
		UserName    string
		TotalSpent  float64
		OrdersCount int
	},
	commissionRate float64) map[string]interface{} {

	// 构建销售摘要
	salesSummary := map[string]interface{}{
		"total_sales":      totalSales,
		"total_orders":     totalOrders,
		"total_commission": totalCommission,
		"period": map[string]interface{}{
			"start_date": start.Format("2006-01-02"),
			"end_date":   end.Format("2006-01-02"),
		},
	}

	// 构建产品销售数据
	var salesByProduct []map[string]interface{}
	for productID, data := range productSales {
		commission := data.TotalRevenue * commissionRate
		salesByProduct = append(salesByProduct, map[string]interface{}{
			"product_id":    productID,
			"product_name":  data.Name,
			"quantity_sold": data.QuantitySold,
			"total_revenue": data.TotalRevenue,
			"commission":    commission,
		})
	}

	// 构建月度趋势数据
	var monthlyTrend []map[string]interface{}
	for month, data := range monthlySales {
		commission := data.Sales * commissionRate
		monthlyTrend = append(monthlyTrend, map[string]interface{}{
			"month":      month,
			"sales":      data.Sales,
			"orders":     data.Orders,
			"commission": commission,
		})
	}

	// 构建顶级客户数据
	var topPerformers []map[string]interface{}
	for userID, data := range userSales {
		topPerformers = append(topPerformers, map[string]interface{}{
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
	}
}
