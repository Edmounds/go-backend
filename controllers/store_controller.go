package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// Order 订单结构体
type Order struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	Items         []OrderItem        `bson:"items" json:"items"`
	TotalAmount   float64            `bson:"total_amount" json:"total_amount"`
	Status        string             `bson:"status" json:"status"`
	AddressID     string             `bson:"address_id" json:"address_id"`
	PaymentMethod string             `bson:"payment_method" json:"payment_method"`
	ReferralCode  string             `bson:"referral_code" json:"referral_code"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// OrderItem 订单项结构体
type OrderItem struct {
	ProductID string  `bson:"product_id" json:"product_id"`
	Quantity  int     `bson:"quantity" json:"quantity"`
	Price     float64 `bson:"price" json:"price"`
}

// GetProductsHandler 获取商品列表处理器
func GetProductsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

		// TODO: 实现商品列表查询逻辑
		// 1. 从数据库获取商品列表
		// 2. 实现分页
		// 3. 返回商品信息

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "获取商品列表成功",
			"data": gin.H{
				"products": []gin.H{
					{
						"_id":         "PROD001",
						"product_id":  "PROD001",
						"name":        "高级英语词汇书",
						"price":       49.9,
						"description": "包含2000个高频英语词汇",
						"created_at":  "2024-01-01T00:00:00Z",
						"updated_at":  "2024-01-01T00:00:00Z",
					},
				},
				"pagination": gin.H{
					"page":        page,
					"limit":       limit,
					"total":       1,
					"total_pages": 1,
				},
			},
		})
	}
}

// GetProductHandler 获取单个商品详情处理器
func GetProductHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("product_id")

		// TODO: 实现商品详情查询逻辑
		// 1. 根据product_id查询商品
		// 2. 返回商品详细信息

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "获取商品详情成功",
			"data": gin.H{
				"_id":         productID,
				"product_id":  productID,
				"name":        "高级英语词汇书",
				"price":       49.9,
				"description": "包含2000个高频英语词汇，适合大学生和职场人士",
				"stock":       100,
				"images":      []string{"https://example.com/product1.jpg"},
				"created_at":  "2024-01-01T00:00:00Z",
				"updated_at":  "2024-01-01T00:00:00Z",
			},
		})
	}
}

// AddToCartHandler 添加到购物车处理器
func AddToCartHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		var req AddToCartRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
				"error":   err.Error(),
			})
			return
		}

		// TODO: 实现添加到购物车逻辑
		// 1. 验证商品是否存在
		// 2. 检查库存
		// 3. 添加到购物车或更新数量
		// 4. 计算购物车总价

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "添加到购物车成功",
			"data": gin.H{
				"_id":     "507f1f77bcf86cd799439014",
				"user_id": userID,
				"items": []gin.H{
					{
						"product_id": req.ProductID,
						"name":       "高级英语词汇书",
						"price":      49.9,
						"quantity":   req.Quantity,
						"subtotal":   49.9 * float64(req.Quantity),
					},
				},
				"total_amount": 49.9 * float64(req.Quantity),
				"created_at":   "2024-01-15T10:30:00Z",
				"updated_at":   "2024-01-15T10:35:00Z",
			},
		})
	}
}

// UpdateCartItemHandler 更新购物车商品数量处理器
func UpdateCartItemHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		productID := c.Param("product_id")
		var req UpdateCartRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
				"error":   err.Error(),
			})
			return
		}

		// TODO: 实现更新购物车逻辑
		// 1. 验证购物车项是否存在
		// 2. 检查库存
		// 3. 更新数量
		// 4. 重新计算总价

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "更新购物车成功",
			"data": gin.H{
				"_id":     "507f1f77bcf86cd799439014",
				"user_id": userID,
				"items": []gin.H{
					{
						"product_id": productID,
						"name":       "高级英语词汇书",
						"price":      49.9,
						"quantity":   req.Quantity,
						"subtotal":   49.9 * float64(req.Quantity),
					},
				},
				"total_amount": 49.9 * float64(req.Quantity),
			},
		})
	}
}

// DeleteCartItemHandler 删除购物车商品处理器
func DeleteCartItemHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		productID := c.Param("product_id")

		// TODO: 实现删除购物车商品逻辑
		// 1. 验证购物车项是否存在
		// 2. 删除商品
		// 3. 重新计算总价

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "删除商品成功",
			"data": gin.H{
				"_id":          "507f1f77bcf86cd799439014",
				"user_id":      userID,
				"items":        []gin.H{},
				"total_amount": 0,
			},
		})

		// 避免未使用变量的警告
		_ = productID
	}
}

// CreateOrderHandler 创建订单处理器
func CreateOrderHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		var req CreateOrderRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
				"error":   err.Error(),
			})
			return
		}

		// TODO: 实现创建订单逻辑
		// 1. 验证购物车不为空
		// 2. 验证收货地址
		// 3. 计算价格（包括折扣）
		// 4. 创建订单
		// 5. 清空购物车

		c.JSON(http.StatusCreated, gin.H{
			"code":    201,
			"message": "订单创建成功",
			"data": gin.H{
				"_id":          "ORD20240115001",
				"user_id":      userID,
				"order_status": "pending_payment",
				"items": []gin.H{
					{
						"product_id": "PROD001",
						"name":       "高级英语词汇书",
						"price":      49.9,
						"quantity":   1,
						"subtotal":   49.9,
					},
				},
				"shipping_address": gin.H{
					"recipient_name": "张三",
					"phone":          "13800138001",
					"address":        "北京市海淀区中关村大街1号",
				},
				"price_breakdown": gin.H{
					"subtotal":    49.9,
					"shipping":    10.0,
					"discount":    0.0,
					"final_total": 59.9,
				},
				"payment_method": req.PaymentMethod,
				"created_at":     "2024-01-15T10:40:00Z",
			},
		})
	}
}

// GetOrdersHandler 获取订单历史处理器
func GetOrdersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		status := c.Query("status")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		// 1. 验证用户权限（从JWT token中获取当前用户ID）
		tokenUserID, exists := c.Get("user_id")
		if !exists || tokenUserID.(string) != userID {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足，只能查看自己的订单",
			})
			return
		}

		// 2. 根据用户ID查询订单
		orders, err := GetUserOrders(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取订单失败",
				"error":   err.Error(),
			})
			return
		}

		// 3. 根据状态筛选（如果提供）
		var filteredOrders []Order
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

// GetUserOrders 根据用户ID获取订单列表
func GetUserOrders(userID string) ([]Order, error) {
	collection := GetCollection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(ctx, bson.M{"user_id": userObjectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []Order
	err = cursor.All(ctx, &orders)
	if err != nil {
		return nil, err
	}

	return orders, nil
}
