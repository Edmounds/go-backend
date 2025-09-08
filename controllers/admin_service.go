package controllers

import (
	"errors"
	"math"
	"miniprogram/models"
	"miniprogram/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ===== 管理员服务层 =====

// AdminService 管理员服务
type AdminService struct{}

// GetAdminService 获取管理员服务实例
func GetAdminService() *AdminService {
	return &AdminService{}
}

// UpdateUserAgentLevel 更新用户代理等级
func (s *AdminService) UpdateUserAgentLevel(openID string, agentLevel int) error {
	// 验证用户是否存在
	userService := GetUserService()
	_, err := userService.FindUserByOpenID(openID)
	if err != nil {
		return err
	}

	// 更新代理等级
	updates := map[string]interface{}{
		"agent_level": agentLevel,
		"is_agent":    agentLevel > 0,
		"updated_at":  utils.GetCurrentUTCTime(),
	}

	_, err = userService.UpdateUser(openID, updates)
	return err
}

// ===== 用户管理服务 =====

// GetAllUsers 获取所有用户列表（分页）
func (s *AdminService) GetAllUsers(req models.AdminUserListRequest) (*models.AdminUserListResponse, error) {
	ctx, cancel := CreateDBContext()
	defer cancel()

	usersCollection := GetCollection("users")

	// 构建筛选条件
	filter := bson.M{}

	// 学校筛选
	if req.School != "" {
		filter["school"] = bson.M{"$regex": req.School, "$options": "i"}
	}

	// 代理筛选
	if req.IsAgent != nil {
		filter["is_agent"] = *req.IsAgent
	}

	// 管理员筛选
	if req.IsAdmin != nil {
		filter["is_admin"] = *req.IsAdmin
	}

	// 关键词搜索（用户名、手机号）
	if req.Keyword != "" {
		filter["$or"] = []bson.M{
			{"user_name": bson.M{"$regex": req.Keyword, "$options": "i"}},
			{"phone": bson.M{"$regex": req.Keyword, "$options": "i"}},
		}
	}

	// 计算分页
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	skip := (req.Page - 1) * req.Limit

	// 获取总数
	total, err := usersCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 获取用户列表
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(req.Limit)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := usersCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	// 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	return &models.AdminUserListResponse{
		Users: users,
		Pagination: models.Pagination{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}, nil
}

// GetUserDetail 获取用户详细信息
func (s *AdminService) GetUserDetail(openID string) (*models.User, error) {
	userService := GetUserService()
	return userService.FindUserByOpenID(openID)
}

// UpdateUserAdminStatus 设置/取消用户管理员权限
func (s *AdminService) UpdateUserAdminStatus(openID string, isAdmin bool) error {
	// 验证用户是否存在
	userService := GetUserService()
	_, err := userService.FindUserByOpenID(openID)
	if err != nil {
		return err
	}

	// 更新管理员状态
	updates := map[string]interface{}{
		"is_admin":   isAdmin,
		"updated_at": utils.GetCurrentUTCTime(),
	}

	_, err = userService.UpdateUser(openID, updates)
	return err
}

// GetUserOrders 获取指定用户的订单列表
func (s *AdminService) GetUserOrders(openID string, page, limit int) ([]models.Order, *models.Pagination, error) {
	ctx, cancel := CreateDBContext()
	defer cancel()

	ordersCollection := GetCollection("orders")

	// 构建筛选条件
	filter := bson.M{"user_openid": openID}

	// 计算分页
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	skip := (page - 1) * limit

	// 获取总数
	total, err := ordersCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	// 获取订单列表
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := ordersCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, nil, err
	}

	// 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	pagination := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return orders, pagination, nil
}

// ===== 代理管理服务扩展 =====

// UpdateAgentSchools 设置校代理管理的学校
func (s *AdminService) UpdateAgentSchools(openID string, schools []string) error {
	// 验证用户是否存在且为校代理
	userService := GetUserService()
	user, err := userService.FindUserByOpenID(openID)
	if err != nil {
		return err
	}

	if user.AgentLevel != 1 {
		return errors.New("用户不是校代理")
	}

	// 更新管理的学校
	updates := map[string]interface{}{
		"managed_schools": schools,
		"updated_at":      utils.GetCurrentUTCTime(),
	}

	_, err = userService.UpdateUser(openID, updates)
	return err
}

// UpdateAgentRegions 设置区代理管理的区域
func (s *AdminService) UpdateAgentRegions(openID string, regions []string) error {
	// 验证用户是否存在且为区代理
	userService := GetUserService()
	user, err := userService.FindUserByOpenID(openID)
	if err != nil {
		return err
	}

	if user.AgentLevel != 2 {
		return errors.New("用户不是区代理")
	}

	// 更新管理的区域
	updates := map[string]interface{}{
		"managed_regions": regions,
		"updated_at":      utils.GetCurrentUTCTime(),
	}

	_, err = userService.UpdateUser(openID, updates)
	return err
}

// GetAgentStats 获取代理统计信息
func (s *AdminService) GetAgentStats(openID string) (*models.AgentStatsResponse, error) {
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 获取代理信息
	userService := GetUserService()
	agentUser, err := userService.FindUserByOpenID(openID)
	if err != nil {
		return nil, err
	}

	if !agentUser.IsAgent {
		return nil, errors.New("用户不是代理")
	}

	// 获取下属用户
	usersCollection := GetCollection("users")
	var subordinateUsers []models.User
	var userFilter bson.M

	if agentUser.AgentLevel == 1 { // 校代理
		userFilter = bson.M{
			"school": bson.M{"$in": agentUser.ManagedSchools},
		}
	} else if agentUser.AgentLevel == 2 { // 区代理
		// 获取该区域下的所有校代理和普通用户
		userFilter = bson.M{
			"$or": []bson.M{
				{"belongs_to_region": bson.M{"$in": agentUser.ManagedRegions}},
				{"city": bson.M{"$in": agentUser.ManagedRegions}},
			},
		}
	}

	cursor, err := usersCollection.Find(ctx, userFilter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &subordinateUsers); err != nil {
		return nil, err
	}

	// 计算销售统计
	ordersCollection := GetCollection("orders")
	var subordinateOpenIDs []string
	for _, user := range subordinateUsers {
		subordinateOpenIDs = append(subordinateOpenIDs, user.OpenID)
	}

	// 获取下属用户的订单
	orderFilter := bson.M{
		"user_openid": bson.M{"$in": subordinateOpenIDs},
		"status":      "paid",
	}

	orderCursor, err := ordersCollection.Find(ctx, orderFilter)
	if err != nil {
		return nil, err
	}
	defer orderCursor.Close(ctx)

	var orders []models.Order
	if err = orderCursor.All(ctx, &orders); err != nil {
		return nil, err
	}

	// 计算总销售额
	var totalSales float64
	monthlyStatsMap := make(map[string]*models.MonthlyAgentStats)

	for _, order := range orders {
		totalSales += order.TotalAmount

		// 按月统计
		if !order.CreatedAt.IsZero() {
			month := order.CreatedAt.Format("2006-01")
			if _, exists := monthlyStatsMap[month]; !exists {
				monthlyStatsMap[month] = &models.MonthlyAgentStats{
					Month:      month,
					Sales:      0,
					UserCount:  0,
					OrderCount: 0,
				}
			}
			monthlyStatsMap[month].Sales += order.TotalAmount
			monthlyStatsMap[month].OrderCount++
		}
	}

	// 转换月度统计为切片
	var monthlyStats []models.MonthlyAgentStats
	for _, stats := range monthlyStatsMap {
		monthlyStats = append(monthlyStats, *stats)
	}

	return &models.AgentStatsResponse{
		AgentInfo:        *agentUser,
		TotalUsers:       len(subordinateUsers),
		TotalSales:       totalSales,
		MonthlyStats:     monthlyStats,
		SubordinateUsers: subordinateUsers,
	}, nil
}

// ===== 商品管理服务 =====

// CreateProduct 创建商品
func (s *AdminService) CreateProduct(req models.CreateProductRequest) (*models.Product, error) {
	ctx, cancel := CreateDBContext()
	defer cancel()

	productsCollection := GetCollection("products")

	// 处理BookID
	var bookID primitive.ObjectID
	var err error
	if req.BookID != "" {
		bookID, err = primitive.ObjectIDFromHex(req.BookID)
		if err != nil {
			return nil, errors.New("无效的书籍ID格式")
		}
	}

	// 创建商品对象
	product := models.Product{
		ProductID:      primitive.NewObjectID().Hex(), // 生成唯一商品ID字符串
		Name:           req.Name,
		Price:          req.Price,
		Description:    req.Description,
		Stock:          req.Stock,
		ProductType:    req.ProductType,
		ProductVersion: req.ProductVersion,
		BookID:         bookID,
		Images:         req.Images,
		CreatedAt:      utils.GetCurrentUTCTime(),
		UpdatedAt:      utils.GetCurrentUTCTime(),
	}

	// 插入数据库
	result, err := productsCollection.InsertOne(ctx, product)
	if err != nil {
		return nil, err
	}

	product.ID = result.InsertedID.(primitive.ObjectID)
	return &product, nil
}

// UpdateProduct 更新商品信息
func (s *AdminService) UpdateProduct(productID string, req models.UpdateProductRequest) (*models.Product, error) {
	ctx, cancel := CreateDBContext()
	defer cancel()

	productsCollection := GetCollection("products")

	// 构建更新字段
	updates := bson.M{
		"updated_at": utils.GetCurrentUTCTime(),
	}

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Stock != nil {
		updates["stock"] = *req.Stock
	}
	if req.ProductType != nil {
		updates["product_type"] = *req.ProductType
	}
	if req.ProductVersion != nil {
		updates["product_version"] = *req.ProductVersion
	}
	if req.BookID != nil {
		if *req.BookID != "" {
			bookID, err := primitive.ObjectIDFromHex(*req.BookID)
			if err != nil {
				return nil, errors.New("无效的书籍ID格式")
			}
			updates["book_id"] = bookID
		}
	}
	if req.Images != nil {
		updates["images"] = *req.Images
	}

	// 更新商品
	filter := bson.M{"product_id": productID}
	_, err := productsCollection.UpdateOne(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		return nil, err
	}

	// 获取更新后的商品
	var product models.Product
	err = productsCollection.FindOne(ctx, filter).Decode(&product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

// DeleteProduct 删除商品
func (s *AdminService) DeleteProduct(productID string) error {
	ctx, cancel := CreateDBContext()
	defer cancel()

	productsCollection := GetCollection("products")

	// 删除商品
	filter := bson.M{"product_id": productID}
	result, err := productsCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("商品不存在")
	}

	return nil
}

// UpdateProductStatus 更新商品状态（上架/下架）
func (s *AdminService) UpdateProductStatus(productID string, status string) error {
	ctx, cancel := CreateDBContext()
	defer cancel()

	productsCollection := GetCollection("products")

	// 验证状态值
	if status != "active" && status != "inactive" {
		return errors.New("无效的状态值")
	}

	// 更新商品状态
	filter := bson.M{"product_id": productID}
	updates := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": utils.GetCurrentUTCTime(),
		},
	}

	result, err := productsCollection.UpdateOne(ctx, filter, updates)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("商品不存在")
	}

	return nil
}

// ===== 仪表盘相关服务 =====

// GetDashboardStats 获取仪表盘统计数据
func (s *AdminService) GetDashboardStats() (*models.DashboardStats, error) {
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 获取各个集合
	usersCollection := GetCollection("users")
	ordersCollection := GetCollection("orders")
	productsCollection := GetCollection("products")
	refundsCollection := GetCollection("refund_records")

	// 统计总用户数
	totalUsers, err := usersCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// 统计总订单数
	totalOrders, err := ordersCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// 统计总收入（已支付订单）
	pipeline := []bson.M{
		{"$match": bson.M{"status": "paid"}},
		{"$group": bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$total_amount"},
		}},
	}

	cursor, err := ordersCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var totalRevenue float64
	if cursor.Next(ctx) {
		var result struct {
			Total float64 `bson:"total"`
		}
		if err := cursor.Decode(&result); err == nil {
			totalRevenue = result.Total
		}
	}

	// 统计商品总数
	totalProducts, err := productsCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// 今日统计（UTC时间）
	todayStart := utils.GetTodayStartUTC()
	todayEnd := utils.GetTodayEndUTC()

	// 今日订单数
	todayOrders, err := ordersCollection.CountDocuments(ctx, bson.M{
		"created_at": bson.M{
			"$gte": todayStart,
			"$lt":  todayEnd,
		},
	})
	if err != nil {
		return nil, err
	}

	// 今日收入（已支付订单）
	todayPipeline := []bson.M{
		{"$match": bson.M{
			"status": "paid",
			"created_at": bson.M{
				"$gte": todayStart,
				"$lt":  todayEnd,
			},
		}},
		{"$group": bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$total_amount"},
		}},
	}

	todayCursor, err := ordersCollection.Aggregate(ctx, todayPipeline)
	if err != nil {
		return nil, err
	}
	defer todayCursor.Close(ctx)

	var todayRevenue float64
	if todayCursor.Next(ctx) {
		var result struct {
			Total float64 `bson:"total"`
		}
		if err := todayCursor.Decode(&result); err == nil {
			todayRevenue = result.Total
		}
	}

	// 统计活跃代理数（本月有销售的代理）
	monthStart := utils.GetCurrentMonthStartUTC()
	monthEnd := utils.GetCurrentMonthEndUTC()

	activeAgentsPipeline := []bson.M{
		{"$match": bson.M{
			"status": "paid",
			"created_at": bson.M{
				"$gte": monthStart,
				"$lt":  monthEnd,
			},
			"referrer_openid": bson.M{"$ne": ""},
		}},
		{"$group": bson.M{
			"_id": "$referrer_openid",
		}},
		{"$count": "active_agents"},
	}

	agentsCursor, err := ordersCollection.Aggregate(ctx, activeAgentsPipeline)
	if err != nil {
		return nil, err
	}
	defer agentsCursor.Close(ctx)

	var activeAgents int
	if agentsCursor.Next(ctx) {
		var result struct {
			ActiveAgents int `bson:"active_agents"`
		}
		if err := agentsCursor.Decode(&result); err == nil {
			activeAgents = result.ActiveAgents
		}
	}

	// 统计待处理退款数
	pendingRefunds, err := refundsCollection.CountDocuments(ctx, bson.M{
		"status": bson.M{"$in": []string{"PROCESSING", "ABNORMAL"}},
	})
	if err != nil {
		return nil, err
	}

	return &models.DashboardStats{
		TotalUsers:     int(totalUsers),
		TotalOrders:    int(totalOrders),
		TotalRevenue:   totalRevenue,
		TotalProducts:  int(totalProducts),
		TodayOrders:    int(todayOrders),
		TodayRevenue:   todayRevenue,
		ActiveAgents:   activeAgents,
		PendingRefunds: int(pendingRefunds),
	}, nil
}

// GetRecentOrders 获取最近订单列表
func (s *AdminService) GetRecentOrders(limit int) ([]models.RecentOrderInfo, error) {
	ctx, cancel := CreateDBContext()
	defer cancel()

	ordersCollection := GetCollection("orders")

	// 设置默认限制
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// 聚合查询，关联用户信息
	pipeline := []bson.M{
		{"$sort": bson.M{"created_at": -1}},
		{"$limit": limit},
		{"$lookup": bson.M{
			"from":         "users",
			"localField":   "user_openid",
			"foreignField": "openID",
			"as":           "user_info",
		}},
		{"$project": bson.M{
			"_id":          1,
			"user_openid":  1,
			"total_amount": 1,
			"status":       1,
			"created_at":   1,
			"user_name": bson.M{
				"$ifNull": []interface{}{
					bson.M{"$arrayElemAt": []interface{}{"$user_info.user_name", 0}},
					"未知用户",
				},
			},
		}},
	}

	cursor, err := ordersCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.RecentOrderInfo
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, err
	}

	return orders, nil
}

// GetSalesTrend 获取销售趋势数据
func (s *AdminService) GetSalesTrend(days int) ([]models.SalesTrendData, error) {
	ctx, cancel := CreateDBContext()
	defer cancel()

	ordersCollection := GetCollection("orders")

	// 设置默认天数
	if days <= 0 || days > 365 {
		days = 30
	}

	// 计算开始和结束日期
	endDate := utils.GetTodayEndUTC()
	startDate := endDate.AddDate(0, 0, -days)

	// 聚合查询，按日期分组统计
	pipeline := []bson.M{
		{"$match": bson.M{
			"status": "paid",
			"created_at": bson.M{
				"$gte": startDate,
				"$lt":  endDate,
			},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{
					"format": "%Y-%m-%d",
					"date":   "$created_at",
				},
			},
			"sales":       bson.M{"$sum": "$total_amount"},
			"order_count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"_id": 1}},
	}

	cursor, err := ordersCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var trendData []models.SalesTrendData
	for cursor.Next(ctx) {
		var result struct {
			Date       string  `bson:"_id"`
			Sales      float64 `bson:"sales"`
			OrderCount int     `bson:"order_count"`
		}
		if err := cursor.Decode(&result); err == nil {
			trendData = append(trendData, models.SalesTrendData{
				Date:       result.Date,
				Sales:      result.Sales,
				OrderCount: result.OrderCount,
			})
		}
	}

	return trendData, nil
}

// GetUserGrowth 获取用户增长趋势数据
func (s *AdminService) GetUserGrowth(days int) ([]models.UserGrowthData, error) {
	ctx, cancel := CreateDBContext()
	defer cancel()

	usersCollection := GetCollection("users")

	// 设置默认天数
	if days <= 0 || days > 365 {
		days = 30
	}

	// 计算开始和结束日期
	endDate := utils.GetTodayEndUTC()
	startDate := endDate.AddDate(0, 0, -days)

	// 聚合查询，按日期分组统计新用户
	pipeline := []bson.M{
		{"$match": bson.M{
			"created_at": bson.M{
				"$gte": startDate,
				"$lt":  endDate,
			},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{
					"format": "%Y-%m-%d",
					"date":   "$created_at",
				},
			},
			"new_users": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"_id": 1}},
	}

	cursor, err := usersCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 获取总用户数用于计算累计数
	totalUsers, err := usersCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var growthData []models.UserGrowthData
	runningTotal := int(totalUsers)

	for cursor.Next(ctx) {
		var result struct {
			Date     string `bson:"_id"`
			NewUsers int    `bson:"new_users"`
		}
		if err := cursor.Decode(&result); err == nil {
			// 计算当日的累计用户数（从总数倒推）
			growthData = append(growthData, models.UserGrowthData{
				Date:       result.Date,
				NewUsers:   result.NewUsers,
				TotalUsers: runningTotal,
			})
		}
	}

	// 由于我们是按日期正序排列，需要重新计算累计用户数
	if len(growthData) > 0 {
		// 倒序计算累计用户数
		for i := len(growthData) - 2; i >= 0; i-- {
			growthData[i].TotalUsers = growthData[i+1].TotalUsers - growthData[i+1].NewUsers
		}
	}

	return growthData, nil
}

// GetAllOrders 管理员获取所有订单列表
func (s *AdminService) GetAllOrders(req models.AdminOrderListRequest) (*models.AdminOrderListResponse, error) {
	ctx, cancel := CreateDBContext()
	defer cancel()

	ordersCollection := GetCollection("orders")

	// 构建筛选条件
	filter := bson.M{}

	// 状态筛选
	if req.Status != "" && req.Status != "all" {
		filter["status"] = req.Status
	}

	// 用户筛选
	if req.UserID != "" {
		filter["user_openid"] = req.UserID
	}

	// 日期筛选
	if req.DateFrom != "" || req.DateTo != "" {
		dateFilter := bson.M{}
		if req.DateFrom != "" {
			if startTime, err := utils.ParseDateString(req.DateFrom); err == nil {
				dateFilter["$gte"] = startTime
			}
		}
		if req.DateTo != "" {
			if endTime, err := utils.ParseDateString(req.DateTo); err == nil {
				dateFilter["$lt"] = endTime.AddDate(0, 0, 1) // 包含当天
			}
		}
		if len(dateFilter) > 0 {
			filter["created_at"] = dateFilter
		}
	}

	// 计算分页
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	skip := (req.Page - 1) * req.Limit

	// 获取总数
	total, err := ordersCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 聚合查询，关联用户信息
	pipeline := []bson.M{
		{"$match": filter},
		{"$sort": bson.M{"created_at": -1}},
		{"$skip": skip},
		{"$limit": req.Limit},
		{"$lookup": bson.M{
			"from":         "users",
			"localField":   "user_openid",
			"foreignField": "openID",
			"as":           "user_info",
		}},
		{"$project": bson.M{
			"_id":          1,
			"user_openid":  1,
			"total_amount": 1,
			"status":       1,
			"created_at":   1,
			"user_name": bson.M{
				"$ifNull": []interface{}{
					bson.M{"$arrayElemAt": []interface{}{"$user_info.user_name", 0}},
					"未知用户",
				},
			},
		}},
	}

	cursor, err := ordersCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.RecentOrderInfo
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, err
	}

	// 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	return &models.AdminOrderListResponse{
		Orders: orders,
		Pagination: models.Pagination{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}, nil
}
