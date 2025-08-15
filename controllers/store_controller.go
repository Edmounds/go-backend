package controllers

import (
	"context"
	"miniprogram/middlewares"
	"miniprogram/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AddToCartRequest 添加到购物车请求
type AddToCartRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// UpdateCartRequest 更新购物车请求
type UpdateCartRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	AddressID     string `json:"address_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
	ReferralCode  string `json:"referral_code"`
}

// 使用 models 包中的结构体定义，删除重复定义

// GetProductsHandler 获取商品列表处理器
func GetProductsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 20
		}

		// 1. 从数据库获取商品列表
		products, total, err := GetProductsList(page, limit)
		if err != nil {
			InternalServerErrorResponse(c, "获取商品列表失败", err)
			return
		}

		// 2. 计算分页信息
		totalPages := (total + limit - 1) / limit

		SuccessResponse(c, "获取商品列表成功", gin.H{
			"products": products,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		})
	}
}

// GetProductHandler 获取单个商品详情处理器
func GetProductHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("product_id")

		// 1. 根据product_id查询商品
		product, err := GetProductByID(productID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				NotFoundResponse(c, "商品不存在", err)
				return
			}
			InternalServerErrorResponse(c, "获取商品详情失败", err)
			return
		}

		SuccessResponse(c, "获取商品详情成功", product)
	}
}

// AddToCartHandler 添加到购物车处理器
func AddToCartHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		userID := c.Param("user_id")
		var req AddToCartRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 1. 验证商品是否存在
		product, err := GetProductByID(req.ProductID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "商品不存在",
				})
				return
			}
			if middlewares.HandleError(err, "验证商品失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "验证商品失败",
					"error":   err.Error(),
				})
				return
			}
		}

		// 2. 检查库存
		if product.Stock < req.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "库存不足",
				"data": gin.H{
					"available_stock": product.Stock,
					"requested":       req.Quantity,
				},
			})
			return
		}

		// 3. 添加到购物车或更新数量
		cart, err := AddProductToCart(userID, req.ProductID, req.Quantity, product.Name, product.Price)
		if err != nil {
			if middlewares.HandleError(err, "添加到购物车失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "添加到购物车失败",
					"error":   err.Error(),
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "添加到购物车成功",
			"data":    cart,
		})
	}
}

// UpdateCartItemHandler 更新购物车商品数量处理器
func UpdateCartItemHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		userID := c.Param("user_id")
		productID := c.Param("product_id")
		var req UpdateCartRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 1. 验证商品库存
		product, err := GetProductByID(productID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "商品不存在",
				})
				return
			}
			if middlewares.HandleError(err, "验证商品失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "验证商品失败",
					"error":   err.Error(),
				})
				return
			}
		}

		// 2. 检查库存
		if product.Stock < req.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "库存不足",
				"data": gin.H{
					"available_stock": product.Stock,
					"requested":       req.Quantity,
				},
			})
			return
		}

		// 3. 更新购物车商品数量
		cart, err := UpdateCartItemQuantity(userID, productID, req.Quantity)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "购物车中没有此商品",
				})
				return
			}
			if middlewares.HandleError(err, "更新购物车失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "更新购物车失败",
					"error":   err.Error(),
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "更新购物车成功",
			"data":    cart,
		})
	}
}

// DeleteCartItemHandler 删除购物车商品处理器
func DeleteCartItemHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		userID := c.Param("user_id")
		productID := c.Param("product_id")

		// 删除购物车商品
		cart, err := RemoveProductFromCart(userID, productID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "购物车中没有此商品",
				})
				return
			}
			if middlewares.HandleError(err, "删除商品失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "删除商品失败",
					"error":   err.Error(),
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "删除商品成功",
			"data":    cart,
		})
	}
}

// CalculateOrderTotal 计算订单总价和折扣
func CalculateOrderTotal(items []models.OrderItem, referralCode string) (subtotal, discountAmount, discountRate, total float64, referrerOpenID string, err error) {
	// 1. 计算原价小计
	for _, item := range items {
		subtotal += item.Price * float64(item.Quantity)
	}

	// 2. 如果没有推荐码，返回原价
	if referralCode == "" {
		return subtotal, 0, 0, subtotal, "", nil
	}

	// 3. 验证推荐码并获取推荐人信息
	referrer, err := GetUserByReferralCode(referralCode)
	if err != nil {
		return subtotal, 0, 0, subtotal, "", err
	}

	// 4. 根据推荐人的代理等级计算折扣率
	discountRate = calculateDiscountRate(referrer.AgentLevel)
	discountAmount = subtotal * discountRate
	total = subtotal - discountAmount

	return subtotal, discountAmount, discountRate, total, referrer.OpenID, nil
}

// CreateOrderHandler 创建订单处理器
func CreateOrderHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		userID := c.Param("user_id")
		var req CreateOrderRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 1. 验证用户是否存在
		_, err := GetUserByOpenID(userID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "用户不存在",
				})
				return
			}
			if middlewares.HandleError(err, "获取用户信息失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "获取用户信息失败",
					"error":   err.Error(),
				})
				return
			}
		}

		// 2. 验证推荐码（如果提供）
		var discountInfo gin.H
		if req.ReferralCode != "" {
			// 验证推荐码是否有效
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

			// 获取推荐人信息用于显示
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

			discountInfo = gin.H{
				"referrer_name":   referrer.UserName,
				"referrer_school": referrer.School,
				"agent_level":     referrer.AgentLevel,
			}
		}

		// 3. 从数据库获取用户的购物车商品
		cart, err := GetCartByUserOpenID(userID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "购物车为空，无法创建订单",
				})
				return
			}
			if middlewares.HandleError(err, "获取购物车失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "获取购物车失败",
					"error":   err.Error(),
				})
				return
			}
		}

		if len(cart.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "购物车为空，无法创建订单",
			})
			return
		}

		// 转换购物车商品为订单项
		orderItems := make([]models.OrderItem, len(cart.Items))
		for i, item := range cart.Items {
			orderItems[i] = models.OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     item.Price,
			}
		}

		// 4. 计算订单价格和折扣
		subtotal, discountAmount, discountRate, total, referrerOpenID, err := CalculateOrderTotal(orderItems, req.ReferralCode)
		if err != nil {
			if middlewares.HandleError(err, "计算订单价格失败", false) {
				InternalServerErrorResponse(c, "计算订单价格失败", err)
				return
			}
		}

		// 5. 创建订单对象并保存到数据库
		order := models.Order{
			UserOpenID:     userID,
			Items:          orderItems,
			SubtotalAmount: subtotal,
			DiscountAmount: discountAmount,
			DiscountRate:   discountRate,
			TotalAmount:    total,
			Status:         "pending_payment",
			AddressID:      req.AddressID,
			PaymentMethod:  req.PaymentMethod,
			ReferralCode:   req.ReferralCode,
			ReferrerOpenID: referrerOpenID,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// 保存订单到数据库
		orderID, err := CreateOrder(&order)
		if err != nil {
			if middlewares.HandleError(err, "创建订单失败", false) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "创建订单失败",
					"error":   err.Error(),
				})
				return
			}
		}

		// 6. 准备响应数据
		responseItems := make([]gin.H, len(cart.Items))
		for i, item := range cart.Items {
			responseItems[i] = gin.H{
				"product_id": item.ProductID,
				"name":       item.Name,
				"price":      item.Price,
				"quantity":   item.Quantity,
				"subtotal":   item.Subtotal,
			}
		}

		responseData := gin.H{
			"_id":          orderID,
			"user_id":      userID,
			"order_status": order.Status,
			"items":        responseItems,
			"shipping_address": gin.H{
				"address_id": req.AddressID,
				// TODO: 这里应该根据AddressID查询具体地址信息
			},
			"price_breakdown": gin.H{
				"subtotal":        subtotal,
				"discount_rate":   discountRate,
				"discount_amount": discountAmount,
				"final_total":     total,
			},
			"payment_method": req.PaymentMethod,
			"referral_code":  req.ReferralCode,
			"created_at":     order.CreatedAt.Format(time.RFC3339),
		}

		// 7. 如果有推荐码，添加推荐人信息
		if req.ReferralCode != "" && discountInfo != nil {
			responseData["referrer_info"] = discountInfo
		}

		c.JSON(http.StatusCreated, gin.H{
			"code":    201,
			"message": "订单创建成功",
			"data":    responseData,
		})
	}
}

// GetOrdersHandler 获取订单历史处理器
func GetOrdersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		userID := c.Param("user_id")
		status := c.Query("status")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		// 1. 验证用户权限（从JWT token中获取当前用户OpenID）
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未授权访问",
			})
			return
		}

		claims, ok := user.(*middlewares.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证信息",
			})
			return
		}

		// 验证用户只能查看自己的订单（比较OpenID）
		if claims.UserId != userID {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足，只能查看自己的订单",
			})
			return
		}

		// 2. 根据用户OpenID查询订单
		orders, err := GetUserOrdersByOpenID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取订单失败",
				"error":   err.Error(),
			})
			return
		}

		// 3. 根据状态筛选（如果提供）
		var filteredOrders []models.Order
		for _, order := range orders {
			if status == "" || order.Status == status {
				filteredOrders = append(filteredOrders, order)
			}
		}

		// 4. 实现分页
		total := len(filteredOrders)
		start := (page - 1) * limit
		end := start + limit
		if end > total {
			end = total
		}
		if start > total {
			start = total
		}

		pagedOrders := filteredOrders[start:end]
		totalPages := (total + limit - 1) / limit

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "获取订单历史成功",
			"data": gin.H{
				"orders": pagedOrders,
				"pagination": gin.H{
					"page":        page,
					"limit":       limit,
					"total":       total,
					"total_pages": totalPages,
				},
			},
		})
	}
}

// GetUserOrdersByOpenID 根据用户OpenID获取订单列表
func GetUserOrdersByOpenID(openID string) ([]models.Order, error) {
	collection := GetCollection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 直接使用OpenID查询订单
	cursor, err := collection.Find(ctx, bson.M{"user_openid": openID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	err = cursor.All(ctx, &orders)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// ProcessOrderCompletion 处理订单完成（支付成功后调用）
func ProcessOrderCompletion(orderID string) error {
	collection := GetCollection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. 获取订单信息
	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return err
	}

	var order models.Order
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&order)
	if err != nil {
		return err
	}

	// 2. 如果订单使用了推荐码，处理推荐奖励
	if order.ReferralCode != "" && order.ReferrerOpenID != "" {
		// 处理推荐奖励（使用购买用户的OpenID）
		err = ProcessReferralReward(order.UserOpenID, orderID, order.TotalAmount)
		if err != nil {
			middlewares.HandleError(err, "处理推荐奖励失败", false)
			// 不返回错误，继续订单处理流程
		}
	}

	// 3. 更新订单状态为已完成
	update := bson.M{
		"$set": bson.M{
			"status":     "completed",
			"updated_at": time.Now(),
		},
	}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

// ===== 商品相关数据库操作函数 =====

// GetProductsList 获取商品列表（分页）
func GetProductsList(page, limit int) ([]models.Product, int, error) {
	collection := GetCollection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 计算跳过的文档数
	skip := (page - 1) * limit

	// 获取总数
	total, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	// 查询商品列表
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.M{"created_at": -1}) // 按创建时间倒序

	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var products []models.Product
	err = cursor.All(ctx, &products)
	if err != nil {
		return nil, 0, err
	}

	return products, int(total), nil
}

// GetProductByID 根据产品ID获取商品详情
func GetProductByID(productID string) (*models.Product, error) {
	collection := GetCollection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var product models.Product
	err := collection.FindOne(ctx, bson.M{"product_id": productID}).Decode(&product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

// ===== 购物车相关数据库操作函数 =====

// GetCartByUserOpenID 根据用户OpenID获取购物车
func GetCartByUserOpenID(userOpenID string) (*models.Cart, error) {
	collection := GetCollection("carts")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cart models.Cart
	err := collection.FindOne(ctx, bson.M{"user_openid": userOpenID}).Decode(&cart)
	if err != nil {
		return nil, err
	}

	return &cart, nil
}

// CreateCartForUser 为用户创建新购物车
func CreateCartForUser(userOpenID string) (*models.Cart, error) {
	collection := GetCollection("carts")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cart := models.Cart{
		CartID:      "CART_" + userOpenID + "_" + strconv.FormatInt(time.Now().Unix(), 10),
		UserOpenID:  userOpenID,
		Items:       []models.CartItem{},
		TotalAmount: 0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := collection.InsertOne(ctx, cart)
	if err != nil {
		return nil, err
	}

	cart.ID = result.InsertedID.(primitive.ObjectID)
	return &cart, nil
}

// AddProductToCart 添加商品到购物车
func AddProductToCart(userOpenID, productID string, quantity int, productName string, price float64) (*models.Cart, error) {
	// 获取或创建购物车
	cart, err := GetCartByUserOpenID(userOpenID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 如果没有购物车，创建新的
			cart, err = CreateCartForUser(userOpenID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// 检查商品是否已在购物车中
	found := false
	for i, item := range cart.Items {
		if item.ProductID == productID {
			// 更新数量
			cart.Items[i].Quantity += quantity
			cart.Items[i].Subtotal = cart.Items[i].Price * float64(cart.Items[i].Quantity)
			found = true
			break
		}
	}

	// 如果商品不在购物车中，添加新项
	if !found {
		newItem := models.CartItem{
			ProductID: productID,
			Name:      productName,
			Price:     price,
			Quantity:  quantity,
			Subtotal:  price * float64(quantity),
		}
		cart.Items = append(cart.Items, newItem)
	}

	// 重新计算总价
	cart.TotalAmount = 0
	for _, item := range cart.Items {
		cart.TotalAmount += item.Subtotal
	}
	cart.UpdatedAt = time.Now()

	// 更新数据库
	err = UpdateCart(cart)
	if err != nil {
		return nil, err
	}

	return cart, nil
}

// UpdateCartItemQuantity 更新购物车商品数量
func UpdateCartItemQuantity(userOpenID, productID string, quantity int) (*models.Cart, error) {
	cart, err := GetCartByUserOpenID(userOpenID)
	if err != nil {
		return nil, err
	}

	// 查找并更新商品
	found := false
	for i, item := range cart.Items {
		if item.ProductID == productID {
			cart.Items[i].Quantity = quantity
			cart.Items[i].Subtotal = cart.Items[i].Price * float64(quantity)
			found = true
			break
		}
	}

	if !found {
		return nil, mongo.ErrNoDocuments
	}

	// 重新计算总价
	cart.TotalAmount = 0
	for _, item := range cart.Items {
		cart.TotalAmount += item.Subtotal
	}
	cart.UpdatedAt = time.Now()

	// 更新数据库
	err = UpdateCart(cart)
	if err != nil {
		return nil, err
	}

	return cart, nil
}

// RemoveProductFromCart 从购物车删除商品
func RemoveProductFromCart(userOpenID, productID string) (*models.Cart, error) {
	cart, err := GetCartByUserOpenID(userOpenID)
	if err != nil {
		return nil, err
	}

	// 查找并删除商品
	found := false
	newItems := []models.CartItem{}
	for _, item := range cart.Items {
		if item.ProductID != productID {
			newItems = append(newItems, item)
		} else {
			found = true
		}
	}

	if !found {
		return nil, mongo.ErrNoDocuments
	}

	cart.Items = newItems

	// 重新计算总价
	cart.TotalAmount = 0
	for _, item := range cart.Items {
		cart.TotalAmount += item.Subtotal
	}
	cart.UpdatedAt = time.Now()

	// 更新数据库
	err = UpdateCart(cart)
	if err != nil {
		return nil, err
	}

	return cart, nil
}

// UpdateCart 更新购物车到数据库
func UpdateCart(cart *models.Cart) error {
	collection := GetCollection("carts")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"items":        cart.Items,
			"total_amount": cart.TotalAmount,
			"updated_at":   time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"user_openid": cart.UserOpenID}, update)
	return err
}

// ===== 订单相关数据库操作函数 =====

// CreateOrder 创建订单
func CreateOrder(order *models.Order) (string, error) {
	collection := GetCollection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, order)
	if err != nil {
		return "", err
	}

	orderID := result.InsertedID.(primitive.ObjectID)
	order.ID = orderID
	return orderID.Hex(), nil
}
