package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"miniprogram/config"
	"miniprogram/models"
	"miniprogram/utils"
	"net/http"
	"strings"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	wechatutils "github.com/wechatpay-apiv3/wechatpay-go/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ===== 支付服务层 =====

// testVerifier 测试用的验证器，暂时跳过签名验证
// 注意：这仅用于调试，生产环境必须使用正确的证书验证器
type testVerifier struct{}

func (v *testVerifier) Verify(ctx context.Context, serial, message, signature string) error {
	// 临时跳过验证，仅用于调试
	log.Printf("[测试验证器] 跳过签名验证 - Serial: %s", serial)
	return nil
}

func (v *testVerifier) GetSerial(ctx context.Context) (string, error) {
	// 返回一个测试序列号
	return "TEST_SERIAL", nil
}

// PaymentService 支付服务
type PaymentService struct {
	client        *core.Client
	notifyHandler *notify.Handler
}

// GetPaymentService 获取支付服务实例
func GetPaymentService() *PaymentService {
	return &PaymentService{
		client:        wechatPayClient,
		notifyHandler: wechatPayNotifyHandler,
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

	// 4. 获取验证过的回调URL
	notifyURL, err := cfg.GetValidatedNotifyURL()
	if err != nil {
		return nil, fmt.Errorf("回调URL验证失败: %w", err)
	}

	// 5. 构建微信支付请求
	req := jsapi.PrepayRequest{
		Appid:       core.String(cfg.WechatAppID),
		Mchid:       core.String(cfg.WechatMchID),
		Description: core.String("商品购买"),
		OutTradeNo:  core.String(orderID),
		TimeExpire:  core.Time(utils.GetCurrentUTCTime().Add(30 * time.Minute)), // 30分钟过期
		NotifyUrl:   core.String(notifyURL),                                     // 微信支付回调URL（已验证）
		Amount: &jsapi.Amount{
			Total:    core.Int64(utils.ConvertToWechatPayCents(order.TotalAmount)), // 转为分，确保符合微信支付要求
			Currency: core.String("CNY"),
		},
		Payer: &jsapi.Payer{
			Openid: core.String(userOpenID),
		},
	}

	// 6. 调用微信支付API
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
	log.Printf("[支付回调处理] 开始处理支付回调通知")

	// 1. 构建 HTTP 请求对象进行验证
	log.Printf("[支付回调处理] 开始验证签名和解密数据")

	// 创建 HTTP 请求对象
	req, err := http.NewRequest("POST", "", bytes.NewReader(body))
	if err != nil {
		log.Printf("[支付回调处理] 创建请求对象失败: %v", err)
		return fmt.Errorf("创建请求对象失败: %w", err)
	}

	// 设置必要的头部信息
	req.Header.Set("Wechatpay-Timestamp", headers["Wechatpay-Timestamp"])
	req.Header.Set("Wechatpay-Nonce", headers["Wechatpay-Nonce"])
	req.Header.Set("Wechatpay-Serial", headers["Wechatpay-Serial"])
	req.Header.Set("Wechatpay-Signature", headers["Wechatpay-Signature"])

	// 使用 notify.Handler 进行验证和解密
	var transaction map[string]interface{} // 定义接收解密数据的结构体
	_, err = s.notifyHandler.ParseNotifyRequest(context.Background(), req, &transaction)
	if err != nil {
		log.Printf("[支付回调处理] SDK 验证签名或解密失败: %v", err)
		return fmt.Errorf("签名验证失败: %w", err)
	}
	log.Printf("[支付回调处理] SDK 验证签名和解密成功")

	// 2. 检查事件类型和交易数据
	outTradeNo, outTradeNoOk := transaction["out_trade_no"].(string)
	transactionId, transactionIdOk := transaction["transaction_id"].(string)
	tradeState, tradeStateOk := transaction["trade_state"].(string)

	if !outTradeNoOk || !transactionIdOk || !tradeStateOk {
		log.Printf("[支付回调处理] 回调数据缺少必要字段")
		return fmt.Errorf("回调数据格式错误")
	}

	log.Printf("[支付回调处理] 订单数据解析成功 - 订单号: %s, 交易状态: %s, 微信交易号: %s",
		outTradeNo, tradeState, transactionId)

	// 3. 处理支付成功
	if tradeState == "SUCCESS" {
		log.Printf("[支付回调处理] 订单支付成功，开始处理支付成功逻辑")
		return s.ProcessPaymentSuccess(outTradeNo, transactionId)
	} else {
		log.Printf("[支付回调处理] 订单支付状态非成功: %s，跳过处理", tradeState)
	}

	log.Printf("[支付回调处理] 支付回调处理完成")
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

	// 4. 根据订单来源决定是否清空购物车
	if order.OrderSource == "cart" {
		if len(order.SelectedCartItems) > 0 {
			err = s.ClearSelectedCartItems(order.UserOpenID, order.SelectedCartItems)
			if err != nil {
				log.Printf("清空选中的购物车商品失败: %v", err)
			} else {
				log.Printf("购物车订单支付成功，已清空选中的商品: %v", order.SelectedCartItems)
			}
		} else {
			// 兼容旧订单，清空整个购物车
			err = s.ClearUserCart(order.UserOpenID)
			if err != nil {
				log.Printf("清空购物车失败: %v", err)
			} else {
				log.Printf("购物车订单支付成功，已清空购物车（兼容模式）")
			}
		}
	} else {
		log.Printf("直接购买订单，跳过清空购物车操作")
	}

	// 5. 处理推荐奖励
	if order.ReferrerOpenID != "" {
		referralService := NewReferralRewardService()
		err := referralService.ProcessReferralReward(order.UserOpenID, order.ID.Hex(), order.TotalAmount)
		if err != nil {
			log.Printf("处理推荐奖励失败: %v", err)
		}
	}

	// 6. 标记用户已使用推荐优惠（如果此订单享受了优惠）
	if order.DiscountAmount > 0 && order.ReferrerOpenID != "" {
		err = s.markUserUsedReferralDiscount(order.UserOpenID)
		if err != nil {
			log.Printf("标记用户已使用推荐优惠失败: %v", err)
		}
	}

	// 7. 处理代理分级提成（基于原始销售金额，不包含优惠）
	agentCommissionService := NewAgentTieredCommissionService()
	err = agentCommissionService.ProcessAgentCommission(order.UserOpenID, order.SubtotalAmount, order.ID.Hex())
	if err != nil {
		log.Printf("处理代理分级提成失败: %v", err)
	}

	// 8. 处理书籍权限解锁
	err = orderService.ProcessOrderUnlockBooks(orderID)
	if err != nil {
		log.Printf("处理书籍权限解锁失败: %v", err)
		// 权限解锁失败不影响支付流程
	} else {
		log.Printf("订单 %s 书籍权限解锁成功", orderIDHex)
	}

	log.Printf("订单 %s 支付成功处理完成", orderIDHex)
	return nil
}

// markUserUsedReferralDiscount 标记用户已使用推荐优惠
func (s *PaymentService) markUserUsedReferralDiscount(userOpenID string) error {
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	filter := bson.M{"openID": userOpenID}
	update := bson.M{
		"$set": bson.M{
			"has_used_referral_discount": true,
			"updated_at":                 utils.GetCurrentUTCTime(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
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
			"updated_at":   utils.GetCurrentUTCTime(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// ClearSelectedCartItems 清空用户购物车中选中的商品
func (s *PaymentService) ClearSelectedCartItems(userOpenID string, selectedItemIDs []string) error {
	collection := GetCollection("carts")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 先获取当前购物车
	var cart models.Cart
	err := collection.FindOne(ctx, bson.M{"user_openid": userOpenID}).Decode(&cart)
	if err != nil {
		return err
	}

	// 创建商品ID的映射表，便于快速查找
	selectedMap := make(map[string]bool)
	for _, id := range selectedItemIDs {
		selectedMap[id] = true
	}

	// 过滤掉选中的商品，保留未选中的商品
	var remainingItems []models.CartItem
	totalAmount := 0.0
	for _, item := range cart.Items {
		if !selectedMap[item.ProductID] {
			remainingItems = append(remainingItems, item)
			if item.Selected {
				totalAmount += item.Subtotal
			}
		}
	}

	// 更新购物车
	filter := bson.M{"user_openid": userOpenID}
	update := bson.M{
		"$set": bson.M{
			"items":        remainingItems,
			"total_amount": totalAmount,
			"updated_at":   utils.GetCurrentUTCTime(),
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	return err
}

// InitializePaymentService 初始化支付服务
func InitializePaymentService() error {
	cfg := config.GetConfig()

	// 验证回调URL配置
	log.Printf("[支付服务初始化] 验证回调URL配置")
	_, err := cfg.GetValidatedNotifyURL()
	if err != nil {
		log.Printf("[支付服务初始化] 回调URL配置验证失败: %v", err)
		return fmt.Errorf("回调URL配置验证失败: %w", err)
	}
	log.Printf("[支付服务初始化] 回调URL配置验证成功")

	var (
		mchID                      string = cfg.WechatMchID
		mchCertificateSerialNumber string = cfg.WechatMchCertificateSerialNumber
		mchAPIv3Key                string = cfg.WechatMchAPIv3Key
	)

	// 从配置文件中加载商户私钥
	mchPrivateKey, err := wechatutils.LoadPrivateKeyWithPath(cfg.WechatMchPrivateKeyPath)
	if err != nil {
		log.Printf("[支付服务初始化] 从配置路径加载私钥失败: %s, 错误: %v", cfg.WechatMchPrivateKeyPath, err)
		return fmt.Errorf("无法从配置路径加载商户私钥: %w", err)
	}
	log.Printf("[支付服务初始化] 成功从路径加载私钥: %s", cfg.WechatMchPrivateKeyPath)

	// 创建客户端，使用自动证书管理
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, mchCertificateSerialNumber, mchPrivateKey, mchAPIv3Key),
	}

	client, err := core.NewClient(context.Background(), opts...)
	if err != nil {
		return fmt.Errorf("创建微信支付客户端失败: %w", err)
	}

	// 创建 notify.Handler，暂时使用简单的验证器
	// 注意：在生产环境中应该使用正确的证书验证器
	log.Printf("[支付服务初始化] 创建回调处理器，使用默认验证器")

	// 使用空验证器进行测试，这将跳过签名验证
	// 在生产环境中，应该配置正确的证书验证器
	notifyHandler := notify.NewNotifyHandler(mchAPIv3Key, &testVerifier{})

	// 设置全局变量
	wechatPayClient = client
	wechatPayNotifyHandler = notifyHandler

	log.Println("微信支付客户端和回调处理器初始化成功")
	return nil
}

// ProcessAgentWithdraw 处理代理提取申请（使用新版商家转账API）
func (s *PaymentService) ProcessAgentWithdraw(withdrawalID, userOpenID string, amount float64) error {
	log.Printf("开始处理代理提取申请，提取ID: %s, 用户: %s, 金额: %.2f", withdrawalID, userOpenID, amount)

	// 1. 获取用户信息以获取收款人姓名
	ctx, cancel := CreateDBContext()
	defer cancel()

	collection := GetCollection("users")
	var user models.User
	err := collection.FindOne(ctx, bson.M{"openID": userOpenID}).Decode(&user)
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
		s.updateWithdrawStatus(withdrawalID, "failed", "获取用户信息失败")
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 2. 创建微信转账服务
	transferService, err := NewWechatTransferService()
	if err != nil {
		log.Printf("创建微信转账服务失败: %v", err)
		s.updateWithdrawStatus(withdrawalID, "failed", "转账服务初始化失败")
		return fmt.Errorf("创建微信转账服务失败: %w", err)
	}

	// 3. 构建转账请求
	transferRequest := transferService.BuildTransferRequest(userOpenID, amount, user.UserName)

	// 使用withdrawalID作为商户单号，确保唯一性
	transferRequest.OutBillNo = withdrawalID

	log.Printf("构建转账请求完成: 商户单号=%s, 金额=%d分, 收款人=%s",
		transferRequest.OutBillNo, transferRequest.TransferAmount, transferRequest.Openid)

	// 4. 发起转账
	transferResponse, err := transferService.TransferToUser(transferRequest)
	if err != nil {
		log.Printf("调用新版商家转账API失败: %v", err)

		// 检查是否为资金不足错误
		errorStr := err.Error()
		if strings.Contains(errorStr, "NOT_ENOUGH") || strings.Contains(errorStr, "资金不足") {
			log.Printf("检测到资金不足错误，保留提现记录状态为pending: %v", err)
			// 不更新状态为failed，保持pending状态，保留商户单号
			return fmt.Errorf("商户运营账户资金不足，请联系管理员")
		}

		// 其他错误，更新提取记录状态为失败
		s.updateWithdrawStatus(withdrawalID, "failed", err.Error())
		return fmt.Errorf("商家转账失败: %w", err)
	}

	log.Printf("商家转账API调用成功: 商户单号=%s, 微信转账单号=%s, 状态=%v",
		transferResponse.OutBillNo, transferResponse.TransferBillNo, transferResponse.State)

	// 5. 根据转账状态更新提取记录
	var status string
	var failureReason string

	if transferResponse.State != nil {
		switch *transferResponse.State {
		case models.TRANSFERBILLSTATUS_ACCEPTED, models.TRANSFERBILLSTATUS_PROCESSING:
			status = "processing"
			log.Printf("转账已受理，状态: %s", *transferResponse.State)
		case models.TRANSFERBILLSTATUS_WAIT_USER_CONFIRM, models.TRANSFERBILLSTATUS_TRANSFERING:
			status = "processing"
			log.Printf("待用户确认或转账中，状态: %s", *transferResponse.State)
		case models.TRANSFERBILLSTATUS_SUCCESS:
			status = "completed"
			log.Printf("转账成功，状态: %s", *transferResponse.State)
		case models.TRANSFERBILLSTATUS_FAIL:
			status = "failed"
			failureReason = "微信支付转账失败"
			log.Printf("转账失败，状态: %s", *transferResponse.State)
		case models.TRANSFERBILLSTATUS_CANCELLED, models.TRANSFERBILLSTATUS_CANCELING:
			status = "failed"
			failureReason = "转账已取消"
			log.Printf("转账已取消，状态: %s", *transferResponse.State)
		default:
			status = "processing"
			log.Printf("未知转账状态: %s，设为处理中", *transferResponse.State)
		}
	} else {
		status = "processing" // 默认状态
		log.Printf("转账响应无状态信息，设为处理中")
	}

	// 6. 更新提取记录状态
	err = s.updateWithdrawStatus(withdrawalID, status, failureReason)
	if err != nil {
		log.Printf("更新提取记录状态失败: %v", err)
		// 数据库操作失败，返回错误确保数据一致性
		return fmt.Errorf("更新提取记录状态失败: %w", err)
	}

	// 7. 保存新版转账信息到数据库
	err = s.saveNewTransferInfo(withdrawalID, transferResponse)
	if err != nil {
		log.Printf("保存转账信息失败: %v", err)
		// 数据库操作失败，返回错误确保数据一致性
		return fmt.Errorf("保存转账信息失败: %w", err)
	}

	log.Printf("代理提取处理完成，提取ID: %s，转账状态: %s", withdrawalID, status)

	// 如果转账失败，返回错误
	if status == "failed" {
		return fmt.Errorf("转账失败: %s", failureReason)
	}

	return nil
}

// updateWithdrawStatus 更新提取记录状态
func (s *PaymentService) updateWithdrawStatus(withdrawalID, status, failureReason string) error {
	collection := GetCollection("withdrawals")
	ctx, cancel := CreateDBContext()
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": utils.GetCurrentUTCTime(),
		},
	}

	if failureReason != "" {
		update["$set"].(bson.M)["failure_reason"] = failureReason
	}

	if status == "completed" {
		update["$set"].(bson.M)["completed_at"] = utils.GetCurrentUTCTime()
	}

	_, err := collection.UpdateOne(ctx, bson.M{"withdraw_id": withdrawalID}, update)
	return err
}

// saveNewTransferInfo 保存新版转账信息
func (s *PaymentService) saveNewTransferInfo(withdrawalID string, transferResponse *models.TransferToUserResponse) error {
	collection := GetCollection("withdrawals")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 构建更新字段
	updateFields := bson.M{
		"updated_at": utils.GetCurrentUTCTime(),
	}

	// 保存新版转账API返回的信息
	if transferResponse.OutBillNo != "" {
		updateFields["out_bill_no"] = transferResponse.OutBillNo
	}
	if transferResponse.TransferBillNo != "" {
		updateFields["transfer_bill_no"] = transferResponse.TransferBillNo
	}
	if transferResponse.CreateTime != "" {
		updateFields["transfer_create_time"] = transferResponse.CreateTime
	}
	if transferResponse.State != nil {
		updateFields["transfer_state"] = string(*transferResponse.State)
	}
	if transferResponse.PackageInfo != "" {
		updateFields["package_info"] = transferResponse.PackageInfo
	}

	update := bson.M{"$set": updateFields}

	_, err := collection.UpdateOne(ctx, bson.M{"withdraw_id": withdrawalID}, update)
	if err != nil {
		return fmt.Errorf("更新转账信息失败: %w", err)
	}

	log.Printf("保存转账信息成功: withdrawalID=%s, transferBillNo=%s", withdrawalID, transferResponse.TransferBillNo)
	return nil
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

// ===== 退款服务 =====

// CreateRefund 创建退款申请
func (s *PaymentService) CreateRefund(userOpenID string, req *models.RefundRequest) (*models.RefundResponse, error) {
	log.Printf("[退款服务] 开始处理退款申请 - 用户: %s, 退款金额: %d分", userOpenID, req.RefundAmount)

	// 1. 验证退款参数
	if err := s.validateRefundRequest(req); err != nil {
		return nil, fmt.Errorf("退款参数验证失败: %w", err)
	}

	// 2. 获取订单信息进行验证
	var orderID string
	if req.OutTradeNo != nil {
		orderID = *req.OutTradeNo
	}

	order, err := s.GetOrderByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("获取订单失败: %w", err)
	}

	// 验证订单用户
	if order.UserOpenID != userOpenID {
		return nil, fmt.Errorf("无权限退款此订单")
	}

	// 验证订单状态
	if order.Status != "paid" {
		return nil, fmt.Errorf("订单状态不支持退款")
	}

	// 3. 生成商户退款单号
	outRefundNo := s.generateRefundNo()

	// 4. 构建微信退款请求JSON
	refundReq := map[string]interface{}{
		"out_refund_no": outRefundNo,
		"amount": map[string]interface{}{
			"refund":   req.RefundAmount,
			"total":    req.TotalAmount,
			"currency": "CNY",
		},
	}

	// 设置订单号（优先使用微信支付订单号）
	if req.TransactionID != nil && *req.TransactionID != "" {
		refundReq["transaction_id"] = *req.TransactionID
	} else if req.OutTradeNo != nil && *req.OutTradeNo != "" {
		refundReq["out_trade_no"] = *req.OutTradeNo
	}

	// 设置退款原因
	if req.Reason != "" {
		refundReq["reason"] = req.Reason
	}

	// 设置回调URL
	if req.NotifyUrl != "" {
		refundReq["notify_url"] = req.NotifyUrl
	}

	log.Printf("[退款服务] 构建微信退款请求 - 商户退款单号: %s", outRefundNo)

	// 5. 将请求转为JSON
	requestBody, err := json.Marshal(refundReq)
	if err != nil {
		return nil, fmt.Errorf("构建退款请求失败: %w", err)
	}

	// 6. 使用HTTP客户端调用微信退款API
	apiResult, err := s.client.Post(context.Background(), "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds",
		bytes.NewReader(requestBody))
	if err != nil {
		log.Printf("[退款服务] 调用微信退款API失败: %v", err)
		return nil, fmt.Errorf("退款申请失败: %w", err)
	}

	// 7. 读取响应体
	responseBody, err := io.ReadAll(apiResult.Response.Body)
	if err != nil {
		return nil, fmt.Errorf("读取退款响应失败: %w", err)
	}
	defer apiResult.Response.Body.Close()

	// 8. 解析微信返回的响应
	var wechatResp map[string]interface{}
	if err := json.Unmarshal(responseBody, &wechatResp); err != nil {
		return nil, fmt.Errorf("解析退款响应失败: %w", err)
	}

	// 获取响应字段的辅助函数
	getStringField := func(key string) string {
		if val, ok := wechatResp[key].(string); ok {
			return val
		}
		return ""
	}

	refundID := getStringField("refund_id")
	status := getStringField("status")

	log.Printf("[退款服务] 微信退款API调用成功 - 退款单号: %s, 状态: %s", refundID, status)

	// 9. 保存退款记录到数据库
	refundRecord := &models.RefundRecord{
		RefundID:      refundID,
		OutRefundNo:   getStringField("out_refund_no"),
		TransactionID: getStringField("transaction_id"),
		OutTradeNo:    getStringField("out_trade_no"),
		UserOpenID:    userOpenID,
		RefundAmount:  req.RefundAmount,
		TotalAmount:   req.TotalAmount,
		Status:        status,
		Reason:        req.Reason,
		CreateTime:    getStringField("create_time"),
		CreatedAt:     utils.GetCurrentUTCTime(),
		UpdatedAt:     utils.GetCurrentUTCTime(),
	}

	// 设置可选字段
	refundRecord.Channel = getStringField("channel")
	refundRecord.UserReceivedAccount = getStringField("user_received_account")
	refundRecord.SuccessTime = getStringField("success_time")
	refundRecord.FundsAccount = getStringField("funds_account")

	err = s.saveRefundRecord(refundRecord)
	if err != nil {
		log.Printf("[退款服务] 保存退款记录失败: %v", err)
		// 不返回错误，因为退款可能已经成功
	}

	// 10. 构建响应
	response := &models.RefundResponse{
		RefundID:      refundID,
		OutRefundNo:   getStringField("out_refund_no"),
		TransactionID: getStringField("transaction_id"),
		OutTradeNo:    getStringField("out_trade_no"),
		Status:        status,
		CreateTime:    getStringField("create_time"),
	}

	// 设置可选字段
	response.Channel = getStringField("channel")
	response.UserReceivedAccount = getStringField("user_received_account")
	response.SuccessTime = getStringField("success_time")
	response.FundsAccount = getStringField("funds_account")

	// 转换金额信息
	if amountData, ok := wechatResp["amount"].(map[string]interface{}); ok {
		getInt64Field := func(key string) int64 {
			if val, ok := amountData[key].(float64); ok {
				return int64(val)
			}
			return 0
		}

		response.Amount = &models.RefundAmount{
			Total:       getInt64Field("total"),
			Refund:      getInt64Field("refund"),
			PayerTotal:  getInt64Field("payer_total"),
			PayerRefund: getInt64Field("payer_refund"),
			Currency:    getStringField("currency"),
		}

		response.Amount.SettlementRefund = getInt64Field("settlement_refund")
		response.Amount.SettlementTotal = getInt64Field("settlement_total")
		response.Amount.DiscountRefund = getInt64Field("discount_refund")

		// 转换退款出资信息
		if fromData, ok := amountData["from"].([]interface{}); ok {
			response.Amount.From = make([]models.RefundFundsFrom, len(fromData))
			for i, fromItem := range fromData {
				if fromMap, ok := fromItem.(map[string]interface{}); ok {
					response.Amount.From[i] = models.RefundFundsFrom{
						Account: getStringFromMap(fromMap, "account"),
						Amount:  getInt64FromMap(fromMap, "amount"),
					}
				}
			}
		}
	}

	log.Printf("[退款服务] 退款申请处理完成 - 退款单号: %s", response.RefundID)
	return response, nil
}

// getStringFromMap 从map中获取字符串值的辅助函数
func getStringFromMap(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

// getInt64FromMap 从map中获取int64值的辅助函数
func getInt64FromMap(data map[string]interface{}, key string) int64 {
	if val, ok := data[key].(float64); ok {
		return int64(val)
	}
	return 0
}

// validateRefundRequest 验证退款请求参数
func (s *PaymentService) validateRefundRequest(req *models.RefundRequest) error {
	// 验证订单号
	if (req.TransactionID == nil || *req.TransactionID == "") &&
		(req.OutTradeNo == nil || *req.OutTradeNo == "") {
		return fmt.Errorf("微信支付订单号和商户订单号必须提供其中之一")
	}

	// 验证退款金额
	if req.RefundAmount <= 0 {
		return fmt.Errorf("退款金额必须大于0")
	}

	if req.TotalAmount <= 0 {
		return fmt.Errorf("订单总金额必须大于0")
	}

	if req.RefundAmount > req.TotalAmount {
		return fmt.Errorf("退款金额不能大于订单总金额")
	}

	return nil
}

// generateRefundNo 生成商户退款单号
func (s *PaymentService) generateRefundNo() string {
	return fmt.Sprintf("REFUND_%d", time.Now().UnixNano())
}

// saveRefundRecord 保存退款记录到数据库
func (s *PaymentService) saveRefundRecord(record *models.RefundRecord) error {
	collection := GetCollection("refunds")
	ctx, cancel := CreateDBContext()
	defer cancel()

	_, err := collection.InsertOne(ctx, record)
	if err != nil {
		return fmt.Errorf("保存退款记录失败: %w", err)
	}

	log.Printf("退款记录保存成功: %s", record.RefundID)
	return nil
}

// GetRefundRecords 获取用户退款记录
func (s *PaymentService) GetRefundRecords(userOpenID string, page, limit int) ([]models.RefundRecord, int64, error) {
	collection := GetCollection("refunds")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 构建查询条件
	filter := bson.M{"user_openid": userOpenID}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("获取退款记录总数失败: %w", err)
	}

	// 分页查询
	skip := int64((page - 1) * limit)
	findOptions := options.Find().
		SetSort(bson.M{"created_at": -1}). // 按创建时间倒序
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("查询退款记录失败: %w", err)
	}
	defer cursor.Close(ctx)

	var records []models.RefundRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, 0, fmt.Errorf("解析退款记录失败: %w", err)
	}

	return records, total, nil
}

// GetRefundByID 根据退款单号获取退款记录
func (s *PaymentService) GetRefundByID(refundID string) (*models.RefundRecord, error) {
	collection := GetCollection("refunds")
	ctx, cancel := CreateDBContext()
	defer cancel()

	var record models.RefundRecord
	err := collection.FindOne(ctx, bson.M{"refund_id": refundID}).Decode(&record)
	if err != nil {
		return nil, fmt.Errorf("获取退款记录失败: %w", err)
	}

	return &record, nil
}

// ===== 向后兼容函数 =====

// CreateRefund 创建退款 (向后兼容)
func CreateRefund(userOpenID string, req *models.RefundRequest) (*models.RefundResponse, error) {
	service := GetPaymentService()
	return service.CreateRefund(userOpenID, req)
}

// GetRefundRecords 获取退款记录 (向后兼容)
func GetRefundRecords(userOpenID string, page, limit int) ([]models.RefundRecord, int64, error) {
	service := GetPaymentService()
	return service.GetRefundRecords(userOpenID, page, limit)
}

// GetRefundByID 获取退款记录 (向后兼容)
func GetRefundByID(refundID string) (*models.RefundRecord, error) {
	service := GetPaymentService()
	return service.GetRefundByID(refundID)
}
