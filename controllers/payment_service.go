package controllers

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"miniprogram/config"
	"miniprogram/models"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/transferbatch"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ===== 支付服务层 =====

// PaymentService 支付服务
type PaymentService struct {
	client             *core.Client
	merchantPrivateKey *rsa.PrivateKey
}

// GetPaymentService 获取支付服务实例
func GetPaymentService() *PaymentService {
	return &PaymentService{
		client:             wechatPayClient,
		merchantPrivateKey: merchantPrivateKey,
	}
}

// CreateWechatPayOrder 创建微信支付订单
func (s *PaymentService) CreateWechatPayOrder(orderID string, userOpenID string) (*models.WechatPayOrder, error) {
	// 1. 获取订单信息
	order, err := s.GetOrderByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("获取订单失败: %w", err)
	}

	// 2. 检查订单状态
	if order.Status != "pending" {
		return nil, fmt.Errorf("订单状态不正确")
	}

	// 3. 检查订单用户
	if order.UserOpenID != userOpenID {
		return nil, fmt.Errorf("订单用户不匹配")
	}

	cfg := config.GetConfig()

	// 4. 构建微信支付请求
	req := jsapi.PrepayRequest{
		Appid:       core.String(cfg.WechatAppID),
		Mchid:       core.String(cfg.WechatMchID),
		Description: core.String("商品购买"),
		OutTradeNo:  core.String(orderID),
		TimeExpire:  core.Time(time.Now().Add(30 * time.Minute)),            // 30分钟过期
		NotifyUrl:   core.String(cfg.BaseAPIURL + "/api/wechat/pay/notify"), // 微信支付回调URL
		Amount: &jsapi.Amount{
			Total:    core.Int64(int64(order.TotalAmount * 100)), // 转为分
			Currency: core.String("CNY"),
		},
		Payer: &jsapi.Payer{
			Openid: core.String(userOpenID),
		},
	}

	// 5. 调用微信支付API
	svc := jsapi.JsapiApiService{Client: s.client}
	resp, _, err := svc.PrepayWithRequestPayment(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("调用微信支付API失败: %w", err)
	}

	return &models.WechatPayOrder{
		PrepayId:  *resp.PrepayId,
		TimeStamp: *resp.TimeStamp,
		NonceStr:  *resp.NonceStr,
		Package:   *resp.Package,
		SignType:  *resp.SignType,
		PaySign:   *resp.PaySign,
	}, nil
}

// ProcessPaymentNotification 处理支付回调通知
func (s *PaymentService) ProcessPaymentNotification(body []byte, headers map[string]string) error {
	// 1. 验证签名
	if !s.verifyNotificationSignature(body, headers) {
		return fmt.Errorf("签名验证失败")
	}

	// 2. 解析回调数据
	var notification struct {
		Resource struct {
			Ciphertext     string `json:"ciphertext"`
			Nonce          string `json:"nonce"`
			AssociatedData string `json:"associated_data"`
		} `json:"resource"`
	}

	err := json.Unmarshal(body, &notification)
	if err != nil {
		return fmt.Errorf("解析回调数据失败: %w", err)
	}

	// 3. 解密回调数据
	cfg := config.GetConfig()
	plaintext, err := utils.DecryptAES256GCM(
		cfg.WechatMchAPIv3Key,
		notification.Resource.AssociatedData,
		notification.Resource.Nonce,
		notification.Resource.Ciphertext,
	)
	if err != nil {
		return fmt.Errorf("解密回调数据失败: %w", err)
	}

	// 4. 解析订单数据
	var orderData struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionId string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
	}

	err = json.Unmarshal([]byte(plaintext), &orderData)
	if err != nil {
		return fmt.Errorf("解析订单数据失败: %w", err)
	}

	// 5. 处理支付成功
	if orderData.TradeState == "SUCCESS" {
		return s.ProcessPaymentSuccess(orderData.OutTradeNo, orderData.TransactionId)
	}

	return nil
}

// ProcessPaymentSuccess 处理支付成功
func (s *PaymentService) ProcessPaymentSuccess(orderIDHex, transactionID string) error {
	log.Printf("开始处理订单 %s 的支付成功逻辑", orderIDHex)

	// 1. 验证订单ID格式
	orderID, err := primitive.ObjectIDFromHex(orderIDHex)
	if err != nil {
		return fmt.Errorf("订单ID格式错误: %w", err)
	}

	// 2. 更新订单状态
	orderService := GetOrderService()
	err = orderService.UpdateOrderPayment(orderID, transactionID)
	if err != nil {
		return fmt.Errorf("更新订单支付状态失败: %w", err)
	}

	// 3. 获取订单信息
	order, err := s.GetOrderByID(orderIDHex)
	if err != nil {
		log.Printf("获取订单信息失败: %v", err)
		return nil // 不阻止支付流程
	}

	// 4. 清空购物车
	err = s.ClearUserCart(order.UserOpenID)
	if err != nil {
		log.Printf("清空购物车失败: %v", err)
	}

	// 5. 处理推荐奖励
	if order.ReferrerOpenID != "" {
		referralService := NewReferralRewardService()
		err := referralService.ProcessReferralReward(order.UserOpenID, order.ID.Hex(), order.TotalAmount)
		if err != nil {
			log.Printf("处理推荐奖励失败: %v", err)
		}
	}

	log.Printf("订单 %s 支付成功处理完成", orderIDHex)
	return nil
}

// GetOrderByID 根据ID获取订单
func (s *PaymentService) GetOrderByID(orderID string) (*models.Order, error) {
	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, err
	}

	collection := GetCollection("orders")
	ctx, cancel := CreateDBContext()
	defer cancel()

	var order models.Order
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

// ClearUserCart 清空用户购物车
func (s *PaymentService) ClearUserCart(userOpenID string) error {
	collection := GetCollection("carts")
	ctx, cancel := CreateDBContext()
	defer cancel()

	filter := bson.M{"user_openid": userOpenID}
	update := bson.M{
		"$set": bson.M{
			"items":        []models.CartItem{},
			"total_amount": 0,
			"updated_at":   time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// verifyNotificationSignature 验证回调签名
func (s *PaymentService) verifyNotificationSignature(body []byte, headers map[string]string) bool {
	// 获取签名相关头部
	timestamp := headers["Wechatpay-Timestamp"]
	nonce := headers["Wechatpay-Nonce"]
	signature := headers["Wechatpay-Signature"]
	serial := headers["Wechatpay-Serial"]

	if timestamp == "" || nonce == "" || signature == "" || serial == "" {
		log.Printf("缺少必要的签名头部")
		return false
	}

	// 构建签名字符串
	signStr := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, string(body))

	// 验证签名
	err := s.verifySignature(signStr, signature)
	if err != nil {
		log.Printf("签名验证失败: %v", err)
		return false
	}

	return true
}

// verifySignature 验证签名
func (s *PaymentService) verifySignature(message, signature string) error {
	decodedSignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	h := sha256.New()
	h.Write([]byte(message))
	digest := h.Sum(nil)

	return rsa.VerifyPKCS1v15(&s.merchantPrivateKey.PublicKey, crypto.SHA256, digest, decodedSignature)
}

// InitializePaymentService 初始化支付服务
func InitializePaymentService() error {
	cfg := config.GetConfig()

	var (
		mchID                      string = cfg.WechatMchID
		mchCertificateSerialNumber string = cfg.WechatMchCertificateSerialNumber
		mchAPIv3Key                string = cfg.WechatMchAPIv3Key
	)

	// 加载商户私钥
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath("apiclient_key.pem")
	if err != nil {
		return fmt.Errorf("加载商户私钥失败: %w", err)
	}

	// 创建客户端
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, mchCertificateSerialNumber, mchPrivateKey, mchAPIv3Key),
	}

	client, err := core.NewClient(context.Background(), opts...)
	if err != nil {
		return fmt.Errorf("创建微信支付客户端失败: %w", err)
	}

	// 设置全局变量
	wechatPayClient = client
	merchantPrivateKey = mchPrivateKey

	log.Println("微信支付客户端初始化成功")
	return nil
}

// ProcessAgentWithdraw 处理代理提取申请
func (s *PaymentService) ProcessAgentWithdraw(withdrawalID, userOpenID string, amount float64) error {
	log.Printf("开始处理代理提取申请，提取ID: %s, 用户: %s, 金额: %.2f", withdrawalID, userOpenID, amount)

	cfg := config.GetConfig()

	// 1. 构建企业转账请求
	outBatchNo := fmt.Sprintf("WITHDRAW_%s_%d", withdrawalID, time.Now().Unix())
	outDetailNo := fmt.Sprintf("DETAIL_%s_%d", withdrawalID, time.Now().Unix())

	transferAmount := int64(amount * 100) // 转换为分

	req := transferbatch.InitiateBatchTransferRequest{
		Appid:       core.String(cfg.WechatAppID),
		OutBatchNo:  core.String(outBatchNo),
		BatchName:   core.String("代理佣金提取"),
		BatchRemark: core.String("代理佣金提取转账"),
		TotalAmount: core.Int64(transferAmount),
		TotalNum:    core.Int64(1),
		TransferDetailList: []transferbatch.TransferDetailInput{
			{
				OutDetailNo:    core.String(outDetailNo),
				TransferAmount: core.Int64(transferAmount),
				TransferRemark: core.String("代理佣金提取"),
				Openid:         core.String(userOpenID),
			},
		},
		TransferSceneId: core.String("1000"), // 商家转账场景ID
	}

	// 2. 调用微信支付企业转账API
	svc := transferbatch.TransferBatchApiService{Client: s.client}
	resp, result, err := svc.InitiateBatchTransfer(context.Background(), req)
	if err != nil {
		log.Printf("调用微信企业转账API失败: %v", err)
		// 更新提取记录状态为失败
		s.updateWithdrawStatus(withdrawalID, "failed", err.Error())
		return fmt.Errorf("企业转账失败: %w", err)
	}

	log.Printf("企业转账API调用成功, 状态码: %d, 批次号: %s", result.Response.StatusCode, *resp.BatchId)

	// 3. 更新提取记录状态
	err = s.updateWithdrawStatus(withdrawalID, "processing", "")
	if err != nil {
		log.Printf("更新提取记录状态失败: %v", err)
		// 不返回错误，因为转账已经成功
	}

	// 4. 保存转账批次信息到数据库
	err = s.saveTransferBatchInfo(withdrawalID, *resp.BatchId, outBatchNo, outDetailNo)
	if err != nil {
		log.Printf("保存转账批次信息失败: %v", err)
		// 不返回错误，因为转账已经成功
	}

	log.Printf("代理提取处理完成，提取ID: %s", withdrawalID)
	return nil
}

// updateWithdrawStatus 更新提取记录状态
func (s *PaymentService) updateWithdrawStatus(withdrawalID, status, failureReason string) error {
	objectID, err := primitive.ObjectIDFromHex(withdrawalID)
	if err != nil {
		return fmt.Errorf("无效的提取记录ID: %w", err)
	}

	collection := GetCollection("withdrawals")
	ctx, cancel := CreateDBContext()
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	if failureReason != "" {
		update["$set"].(bson.M)["failure_reason"] = failureReason
	}

	if status == "completed" {
		update["$set"].(bson.M)["completed_at"] = time.Now()
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

// saveTransferBatchInfo 保存转账批次信息
func (s *PaymentService) saveTransferBatchInfo(withdrawalID, batchID, outBatchNo, outDetailNo string) error {
	objectID, err := primitive.ObjectIDFromHex(withdrawalID)
	if err != nil {
		return fmt.Errorf("无效的提取记录ID: %w", err)
	}

	collection := GetCollection("withdrawals")
	ctx, cancel := CreateDBContext()
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"wechat_batch_id": batchID,
			"out_batch_no":    outBatchNo,
			"out_detail_no":   outDetailNo,
			"updated_at":      time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

// ===== 向后兼容函数 =====

// GetOrderByID 根据ID获取订单 (向后兼容)
func GetOrderByID(orderID string) (*models.Order, error) {
	service := GetPaymentService()
	return service.GetOrderByID(orderID)
}

// ClearUserCart 清空用户购物车 (向后兼容)
func ClearUserCart(userOpenID string) error {
	service := GetPaymentService()
	return service.ClearUserCart(userOpenID)
}

// ProcessAgentWithdraw 处理代理提取 (向后兼容)
func ProcessAgentWithdraw(withdrawalID, userOpenID string, amount float64) error {
	service := GetPaymentService()
	return service.ProcessAgentWithdraw(withdrawalID, userOpenID, amount)
}
