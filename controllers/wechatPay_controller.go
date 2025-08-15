package controllers

import (
	"crypto/rsa"
	"io"
	"log"
	"miniprogram/models"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
)

// WechatPayClient 微信支付客户端单例
var wechatPayClient *core.Client

// 微信支付相关全局变量
var (
	merchantPrivateKey *rsa.PrivateKey
)

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
		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("读取回调请求体失败: %v", err)
			c.Status(500)
			return
		}

		// 获取请求头
		headers := make(map[string]string)
		headers["Wechatpay-Timestamp"] = c.GetHeader("Wechatpay-Timestamp")
		headers["Wechatpay-Nonce"] = c.GetHeader("Wechatpay-Nonce")
		headers["Wechatpay-Signature"] = c.GetHeader("Wechatpay-Signature")
		headers["Wechatpay-Serial"] = c.GetHeader("Wechatpay-Serial")

		// 初始化支付服务
		paymentService := GetPaymentService()

		// 处理支付回调
		err = paymentService.ProcessPaymentNotification(body, headers)
		if err != nil {
			log.Printf("处理支付回调失败: %v", err)
			c.Status(500)
			return
		}

		// 返回成功响应
		c.JSON(200, gin.H{
			"code":    "SUCCESS",
			"message": "",
		})
	}
}
