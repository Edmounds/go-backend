package controllers

import (
	"io"
	"log"
	"miniprogram/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WechatPayClient 微信支付客户端单例
var wechatPayClient *core.Client

// 微信支付回调处理器
var wechatPayNotifyHandler *notify.Handler

// InitWechatPayClient 初始化微信支付客户端
func InitWechatPayClient() error {
	return InitializePaymentService()
}

// ===== HTTP 处理器 =====

// CreateWechatPayOrderHandler 创建微信支付订单处理器
func CreateWechatPayOrderHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userOpenID := c.Param("user_id")

		var req models.CreateWechatPayOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 初始化支付服务
		paymentService := GetPaymentService()

		// 创建微信支付订单
		wechatPayOrder, err := paymentService.CreateWechatPayOrder(req.OrderID, userOpenID)
		if err != nil {
			InternalServerErrorResponse(c, "创建微信支付订单失败", err)
			return
		}

		SuccessResponse(c, "微信支付订单创建成功", wechatPayOrder)
	}
}

// WechatPayNotifyHandler 微信支付回调处理器
func WechatPayNotifyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录回调开始日志
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		log.Printf("[微信支付回调] 开始处理回调通知 - IP: %s, UserAgent: %s", clientIP, userAgent)

		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("[微信支付回调] 读取请求体失败: %v", err)
			InternalServerErrorResponse(c, "读取请求体失败", err)
			return
		}

		// 记录请求体大小
		log.Printf("[微信支付回调] 请求体大小: %d bytes", len(body))

		// 获取请求头
		headers := make(map[string]string)
		headers["Wechatpay-Timestamp"] = c.GetHeader("Wechatpay-Timestamp")
		headers["Wechatpay-Nonce"] = c.GetHeader("Wechatpay-Nonce")
		headers["Wechatpay-Signature"] = c.GetHeader("Wechatpay-Signature")
		headers["Wechatpay-Serial"] = c.GetHeader("Wechatpay-Serial")

		// 记录关键请求头
		log.Printf("[微信支付回调] 关键请求头 - Timestamp: %s, Nonce: %s, Serial: %s",
			headers["Wechatpay-Timestamp"],
			headers["Wechatpay-Nonce"],
			headers["Wechatpay-Serial"])

		// 验证必要的请求头
		if headers["Wechatpay-Timestamp"] == "" || headers["Wechatpay-Nonce"] == "" ||
			headers["Wechatpay-Signature"] == "" || headers["Wechatpay-Serial"] == "" {
			log.Printf("[微信支付回调] 缺少必要的请求头")
			BadRequestResponse(c, "缺少必要的微信支付签名头部", nil)
			return
		}

		// 初始化支付服务
		paymentService := GetPaymentService()

		// 处理支付回调
		err = paymentService.ProcessPaymentNotification(body, headers)
		if err != nil {
			log.Printf("[微信支付回调] 处理回调失败: %v", err)
			InternalServerErrorResponse(c, "处理支付回调失败", err)
			return
		}

		log.Printf("[微信支付回调] 回调处理成功")

		// 返回成功响应（微信要求的格式）
		c.JSON(200, gin.H{
			"code":    "SUCCESS",
			"message": "",
		})
	}
}

// TestUpdateOrderStatusHandler 测试订单状态更新（调试用）
func TestUpdateOrderStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		type TestRequest struct {
			OrderID       string `json:"order_id" binding:"required"`
			TransactionID string `json:"transaction_id" binding:"required"`
		}

		var req TestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		log.Printf("[测试] 手动测试订单状态更新 - 订单ID: %s, 交易ID: %s", req.OrderID, req.TransactionID)

		// 转换订单ID
		orderID, err := primitive.ObjectIDFromHex(req.OrderID)
		if err != nil {
			BadRequestResponse(c, "订单ID格式错误", err)
			return
		}

		// 更新订单状态
		orderService := GetOrderService()
		err = orderService.UpdateOrderPayment(orderID, req.TransactionID)
		if err != nil {
			InternalServerErrorResponse(c, "更新订单状态失败", err)
			return
		}

		SuccessResponse(c, "订单状态更新成功", gin.H{
			"order_id":       req.OrderID,
			"transaction_id": req.TransactionID,
			"status":         "paid",
		})
	}
}

// ===== 退款相关处理器 =====

// CreateRefundHandler 创建退款申请处理器
func CreateRefundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userOpenID := c.Param("user_id")

		var req models.RefundRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		log.Printf("[退款控制器] 收到退款申请 - 用户: %s, 退款金额: %d分", userOpenID, req.RefundAmount)

		// 初始化支付服务
		paymentService := GetPaymentService()

		// 创建退款申请
		refundResponse, err := paymentService.CreateRefund(userOpenID, &req)
		if err != nil {
			log.Printf("[退款控制器] 退款申请失败: %v", err)
			InternalServerErrorResponse(c, "退款申请失败", err)
			return
		}

		log.Printf("[退款控制器] 退款申请成功 - 退款单号: %s", refundResponse.RefundID)
		SuccessResponse(c, "退款申请提交成功", refundResponse)
	}
}

// GetRefundRecordsHandler 获取用户退款记录处理器
func GetRefundRecordsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userOpenID := c.Param("user_id")

		// 获取分页参数
		page := 1
		limit := 20
		if pageStr := c.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}
		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
				limit = l
			}
		}

		log.Printf("[退款控制器] 获取退款记录 - 用户: %s, 页码: %d, 每页: %d", userOpenID, page, limit)

		// 初始化支付服务
		paymentService := GetPaymentService()

		// 获取退款记录
		records, total, err := paymentService.GetRefundRecords(userOpenID, page, limit)
		if err != nil {
			log.Printf("[退款控制器] 获取退款记录失败: %v", err)
			InternalServerErrorResponse(c, "获取退款记录失败", err)
			return
		}

		// 计算分页信息
		totalPages := (total + int64(limit) - 1) / int64(limit)

		response := gin.H{
			"records":     records,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		}

		log.Printf("[退款控制器] 退款记录获取成功 - 总数: %d, 当前页: %d", total, page)
		SuccessResponse(c, "获取退款记录成功", response)
	}
}

// GetRefundHandler 获取单个退款记录处理器
func GetRefundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userOpenID := c.Param("user_id")
		refundID := c.Param("refund_id")

		log.Printf("[退款控制器] 获取退款详情 - 用户: %s, 退款单号: %s", userOpenID, refundID)

		// 初始化支付服务
		paymentService := GetPaymentService()

		// 获取退款记录
		record, err := paymentService.GetRefundByID(refundID)
		if err != nil {
			log.Printf("[退款控制器] 获取退款记录失败: %v", err)
			NotFoundResponse(c, "退款记录不存在", err)
			return
		}

		// 验证用户权限
		if record.UserOpenID != userOpenID {
			log.Printf("[退款控制器] 用户无权限访问退款记录 - 用户: %s, 记录用户: %s", userOpenID, record.UserOpenID)
			ForbiddenResponse(c, "无权限访问此退款记录", nil)
			return
		}

		log.Printf("[退款控制器] 退款详情获取成功 - 退款单号: %s", refundID)
		SuccessResponse(c, "获取退款详情成功", record)
	}
}
