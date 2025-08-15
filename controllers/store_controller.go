package controllers

import (
	"miniprogram/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// 使用 models 包中的结构体定义，删除重复定义

// ===== HTTP 处理器 =====

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

		// 初始化产品服务
		productService := GetProductService()

		// 获取商品列表
		products, total, err := productService.GetProductsList(page, limit)
		if err != nil {
			InternalServerErrorResponse(c, "获取商品列表失败", err)
			return
		}

		// 计算分页信息
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

		// 初始化产品服务
		productService := GetProductService()

		// 获取商品详情
		product, err := productService.GetProductByID(productID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				NotFoundResponse(c, "商品不存在", err)
			} else {
				InternalServerErrorResponse(c, "获取商品详情失败", err)
			}
			return
		}

		SuccessResponse(c, "获取商品详情成功", product)
	}
}

// AddToCartHandler 添加商品到购物车处理器
func AddToCartHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")

		var req models.AddToCartRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 初始化购物车服务
		cartService := GetCartService()

		// 添加商品到购物车
		cart, err := cartService.AddItemToCart(userID, req.ProductID, req.Quantity)
		if err != nil {
			InternalServerErrorResponse(c, "添加到购物车失败", err)
			return
		}

		SuccessResponse(c, "商品已添加到购物车", cart)
	}
}

// UpdateCartItemHandler 更新购物车商品数量处理器
func UpdateCartItemHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		productID := c.Param("product_id")

		var req models.UpdateCartRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 初始化购物车服务
		cartService := GetCartService()

		// 更新购物车商品
		cart, err := cartService.UpdateCartItem(userID, productID, req.Quantity)
		if err != nil {
			InternalServerErrorResponse(c, "更新购物车失败", err)
			return
		}

		SuccessResponse(c, "购物车已更新", cart)
	}
}

// DeleteCartItemHandler 删除购物车商品处理器
func DeleteCartItemHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		productID := c.Param("product_id")

		// 初始化购物车服务
		cartService := GetCartService()

		// 删除购物车商品
		cart, err := cartService.DeleteCartItem(userID, productID)
		if err != nil {
			InternalServerErrorResponse(c, "删除购物车商品失败", err)
			return
		}

		SuccessResponse(c, "商品已从购物车删除", cart)
	}
}

// GetCartHandler 获取购物车处理器
func GetCartHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")

		// 初始化购物车服务
		cartService := GetCartService()

		// 获取购物车
		cart, err := cartService.GetUserCart(userID)
		if err != nil {
			InternalServerErrorResponse(c, "获取购物车失败", err)
			return
		}

		SuccessResponse(c, "获取购物车成功", cart)
	}
}

// CreateOrderHandler 创建订单处理器
func CreateOrderHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")

		var req models.CreateOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 初始化订单服务
		orderService := GetOrderService()

		// 创建订单
		order, err := orderService.CreateOrder(userID, req)
		if err != nil {
			InternalServerErrorResponse(c, "创建订单失败", err)
			return
		}

		// 处理推荐奖励
		if order.ReferrerOpenID != "" {
			referralService := NewReferralRewardService()
			err := referralService.ProcessReferralReward(userID, order.ID.Hex(), order.TotalAmount)
			if err != nil {
				// 推荐奖励处理失败不影响订单创建，记录日志即可
			}
		}

		CreatedResponse(c, "订单创建成功", order)
	}
}

// GetOrdersHandler 获取用户订单列表处理器
func GetOrdersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")

		// 获取分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 50 {
			limit = 10
		}

		// 初始化订单服务
		orderService := GetOrderService()

		// 获取订单列表
		orders, total, err := orderService.GetUserOrders(userID, page, limit)
		if err != nil {
			InternalServerErrorResponse(c, "获取订单列表失败", err)
			return
		}

		// 计算分页信息
		totalPages := (total + limit - 1) / limit

		SuccessResponse(c, "获取订单列表成功", gin.H{
			"orders": orders,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		})
	}
}

// GetOrderHandler 获取单个订单详情处理器
func GetOrderHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		orderIDStr := c.Param("order_id")

		// 解析订单ID
		orderID, err := primitive.ObjectIDFromHex(orderIDStr)
		if err != nil {
			BadRequestResponse(c, "订单ID格式错误", err)
			return
		}

		// 查询订单
		collection := GetCollection("orders")
		ctx, cancel := CreateDBContext()
		defer cancel()

		var order models.Order
		err = collection.FindOne(ctx, map[string]interface{}{
			"_id":         orderID,
			"user_openid": userID,
		}).Decode(&order)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				NotFoundResponse(c, "订单不存在", err)
			} else {
				InternalServerErrorResponse(c, "获取订单详情失败", err)
			}
			return
		}

		SuccessResponse(c, "获取订单详情成功", order)
	}
}
