package controllers

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"miniprogram/config"
	"miniprogram/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"

	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateWechatPayOrderRequest 创建微信支付订单请求
type CreateWechatPayOrderRequest struct {
	OrderID string `json:"order_id" binding:"required"`
}

// WechatPayClient 微信支付客户端单例
var wechatPayClient *core.Client

// 微信支付相关全局变量
var (
	merchantPrivateKey *rsa.PrivateKey
)

// InitWechatPayClient 初始化微信支付客户端
func InitWechatPayClient() error {
	cfg := config.GetConfig()

	var (
		mchID                      string = cfg.WechatMchID
		mchCertificateSerialNumber string = cfg.WechatMchCertificateSerialNumber
		mchAPIv3Key                string = cfg.WechatMchAPIv3Key
	)

	// 使用 utils 提供的函数从本地文件中加载商户私钥，商户私钥会用来生成请求的签名
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath("apiclient_key.pem")
	if err != nil {
		return fmt.Errorf("load merchant private key error: %v", err)
	}

	// 保存商户私钥到全局变量
	merchantPrivateKey = mchPrivateKey

	ctx := context.Background()
	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, mchCertificateSerialNumber, mchPrivateKey, mchAPIv3Key),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("new wechat pay client err: %v", err)
	}

	wechatPayClient = client

	log.Println("微信支付客户端初始化成功")
	return nil
}

// CreateWechatPayOrderHandler 创建微信支付订单处理器
func CreateWechatPayOrderHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		var req CreateWechatPayOrderRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 1. 验证订单是否存在且属于当前用户
		order, err := GetOrderByIDAndUserOpenID(req.OrderID, userID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				NotFoundResponse(c, "订单不存在或无权访问", err)
				return
			}
			InternalServerErrorResponse(c, "获取订单信息失败", err)
			return
		}

		// 2. 检查订单状态
		if order.Status != "pending_payment" {
			BadRequestResponse(c, "订单状态不正确，无法支付", fmt.Errorf("order status: %s", order.Status))
			return
		}

		// 3. 生成微信支付订单
		payOrder, err := CreateWechatPayOrder(order, userID)
		if err != nil {
			InternalServerErrorResponse(c, "创建支付订单失败", err)
			return
		}

		// 4. 更新订单状态为待支付
		err = UpdateOrderStatus(req.OrderID, "awaiting_payment")
		if err != nil {
			log.Printf("更新订单状态失败: %v", err)
			// 不返回错误，继续支付流程
		}

		SuccessResponse(c, "支付订单创建成功", gin.H{
			"order_id":   req.OrderID,
			"prepay_id":  payOrder.PrepayId,
			"pay_params": payOrder,
		})
	}
}

// WechatPayOrder 微信支付订单信息
type WechatPayOrder struct {
	PrepayId  string `json:"prepayId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

// CreateWechatPayOrder 创建微信支付订单
func CreateWechatPayOrder(order *models.Order, userOpenID string) (*WechatPayOrder, error) {
	if wechatPayClient == nil {
		return nil, fmt.Errorf("微信支付客户端未初始化")
	}

	cfg := config.GetConfig()
	ctx := context.Background()

	// 生成商户订单号（使用MongoDB的ObjectID确保唯一性）
	outTradeNo := fmt.Sprintf("ORDER_%s_%d", order.ID.Hex(), time.Now().Unix())

	// 计算订单金额（单位：分）
	totalFee := int64(order.TotalAmount * 100)

	// 构建订单描述
	description := "单词卡片商城订单"
	if len(order.Items) > 0 {
		description = fmt.Sprintf("单词卡片商城订单-共%d件商品", len(order.Items))
	}

	// 创建JSAPI支付服务
	svc := jsapi.JsapiApiService{Client: wechatPayClient}

	// 发送预支付请求
	resp, _, err := svc.Prepay(ctx, jsapi.PrepayRequest{
		Appid:       core.String(cfg.WechatAppID),
		Mchid:       core.String(cfg.WechatMchID),
		Description: core.String(description),
		OutTradeNo:  core.String(outTradeNo),
		Attach:      core.String(order.ID.Hex()), // 附加数据，用于回调时识别订单
		NotifyUrl:   core.String(cfg.BaseAPIURL + "/api/wechat/pay/notify"),
		Amount: &jsapi.Amount{
			Total: core.Int64(totalFee),
		},
		Payer: &jsapi.Payer{
			Openid: core.String(userOpenID),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("创建预支付订单失败: %v", err)
	}

	// 生成小程序支付参数
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonceStr := generateNonceStr()
	packageStr := fmt.Sprintf("prepay_id=%s", *resp.PrepayId)

	// 使用微信支付 v3 标准生成签名
	paySign, err := generateWechatPayV3Sign(cfg.WechatAppID, timeStamp, nonceStr, packageStr, merchantPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("生成支付签名失败: %v", err)
	}

	return &WechatPayOrder{
		PrepayId:  *resp.PrepayId,
		TimeStamp: timeStamp,
		NonceStr:  nonceStr,
		Package:   packageStr,
		SignType:  "RSA",
		PaySign:   paySign,
	}, nil
}

// WechatPayNotifyHandler 微信支付回调处理器
func WechatPayNotifyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 读取回调数据
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("读取微信支付回调数据失败: %v", err)
			c.String(http.StatusBadRequest, "FAIL")
			return
		}

		// 获取请求头用于日志记录
		wechatSerial := c.GetHeader("Wechatpay-Serial")
		wechatTimestamp := c.GetHeader("Wechatpay-Timestamp")

		log.Printf("收到微信支付回调，序列号: %s, 时间戳: %s", wechatSerial, wechatTimestamp)
		log.Printf("回调内容: %s", string(body))

		// TODO: 这里应该进行签名验证和解密，目前先简化处理
		// 在生产环境中，必须验证签名确保回调来自微信支付

		// 简单解析回调数据结构
		var notifyData struct {
			EventType    string `json:"event_type"`
			ResourceType string `json:"resource_type"`
			Resource     struct {
				Ciphertext string `json:"ciphertext"`
				Nonce      string `json:"nonce"`
			} `json:"resource"`
		}

		if err := json.Unmarshal(body, &notifyData); err != nil {
			log.Printf("解析微信支付回调数据失败: %v", err)
			c.String(http.StatusBadRequest, "FAIL")
			return
		}

		if notifyData.EventType == "TRANSACTION.SUCCESS" {
			log.Printf("收到支付成功回调")
			// TODO: 解密resource.ciphertext获取具体的交易信息
			// 目前先记录日志，实际项目中需要解密并更新订单状态
		}

		// 返回成功响应给微信
		c.String(http.StatusOK, "SUCCESS")
	}
}

// ProcessPaymentSuccess 处理支付成功
func ProcessPaymentSuccess(orderIDHex, transactionID string) error {
	// 1. 更新订单状态为已支付
	err := UpdateOrderStatusWithTransaction(orderIDHex, "paid", transactionID)
	if err != nil {
		return fmt.Errorf("更新订单状态失败: %v", err)
	}

	// 2. 扣减商品库存
	err = DeductProductStock(orderIDHex)
	if err != nil {
		log.Printf("扣减库存失败: %v", err)
		// 不返回错误，避免影响支付流程
	}

	// 3. 清空用户购物车
	order, err := GetOrderByID(orderIDHex)
	if err != nil {
		log.Printf("获取订单信息失败: %v", err)
	} else {
		err = ClearUserCart(order.UserOpenID)
		if err != nil {
			log.Printf("清空购物车失败: %v", err)
		}
	}

	// 4. 处理推荐奖励（如果有）
	err = ProcessOrderCompletion(orderIDHex)
	if err != nil {
		log.Printf("处理推荐奖励失败: %v", err)
		// 不返回错误，避免影响支付流程
	}

	log.Printf("订单 %s 支付成功处理完成", orderIDHex)
	return nil
}

// ===== 辅助函数 =====

// generateNonceStr 生成随机字符串
func generateNonceStr() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// generateWechatPayV3Sign 生成微信支付 v3 标准签名（RSA-SHA256）
func generateWechatPayV3Sign(appID, timeStamp, nonceStr, packageStr string, privateKey *rsa.PrivateKey) (string, error) {
	if privateKey == nil {
		return "", fmt.Errorf("商户私钥未初始化")
	}

	// 构造签名字符串（微信支付 v3 标准格式）
	signStr := fmt.Sprintf("%s\n%s\n%s\n%s\n", appID, timeStamp, nonceStr, packageStr)

	// 计算SHA256哈希
	hash := sha256.Sum256([]byte(signStr))

	// 使用RSA私钥签名
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("RSA签名失败: %v", err)
	}

	// 返回base64编码的签名
	return base64.StdEncoding.EncodeToString(signature), nil
}

// GetOrderByIDAndUserOpenID 根据订单ID和用户OpenID获取订单
func GetOrderByIDAndUserOpenID(orderIDHex, userOpenID string) (*models.Order, error) {
	collection := GetCollection("orders")
	ctx, cancel := CreateDBContext()
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(orderIDHex)
	if err != nil {
		return nil, err
	}

	var order models.Order
	err = collection.FindOne(ctx, bson.M{
		"_id":         objectID,
		"user_openid": userOpenID,
	}).Decode(&order)

	return &order, err
}

// GetOrderByID 根据订单ID获取订单
func GetOrderByID(orderIDHex string) (*models.Order, error) {
	collection := GetCollection("orders")
	ctx, cancel := CreateDBContext()
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(orderIDHex)
	if err != nil {
		return nil, err
	}

	var order models.Order
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&order)
	return &order, err
}

// UpdateOrderStatus 更新订单状态
func UpdateOrderStatus(orderIDHex, status string) error {
	collection := GetCollection("orders")
	ctx, cancel := CreateDBContext()
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(orderIDHex)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

// UpdateOrderStatusWithTransaction 更新订单状态并记录交易ID
func UpdateOrderStatusWithTransaction(orderIDHex, status, transactionID string) error {
	collection := GetCollection("orders")
	ctx, cancel := CreateDBContext()
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(orderIDHex)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":         status,
			"transaction_id": transactionID,
			"paid_at":        time.Now(),
			"updated_at":     time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

// DeductProductStock 扣减商品库存
func DeductProductStock(orderIDHex string) error {
	order, err := GetOrderByID(orderIDHex)
	if err != nil {
		return err
	}

	collection := GetCollection("products")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 批量扣减库存
	for _, item := range order.Items {
		update := bson.M{
			"$inc": bson.M{
				"stock": -item.Quantity,
			},
			"$set": bson.M{
				"updated_at": time.Now(),
			},
		}
		_, err = collection.UpdateOne(ctx, bson.M{"product_id": item.ProductID}, update)
		if err != nil {
			log.Printf("扣减商品 %s 库存失败: %v", item.ProductID, err)
		}
	}

	return nil
}

// ClearUserCart 清空用户购物车
func ClearUserCart(userOpenID string) error {
	collection := GetCollection("carts")
	ctx, cancel := CreateDBContext()
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"items":        []models.CartItem{},
			"total_amount": 0,
			"updated_at":   time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"user_openid": userOpenID}, update)
	return err
}
