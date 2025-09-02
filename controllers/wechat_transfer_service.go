package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"miniprogram/config"
	"miniprogram/models"
	"miniprogram/utils"
	"net/http"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	wechatutils "github.com/wechatpay-apiv3/wechatpay-go/utils"
)

// WechatTransferService 微信转账服务
type WechatTransferService struct {
	config     *config.Config
	client     *core.Client
	httpClient *http.Client
}

// NewWechatTransferService 创建微信转账服务实例
func NewWechatTransferService() (*WechatTransferService, error) {
	cfg := config.GetConfig()

	// 加载商户私钥
	mchPrivateKey, err := wechatutils.LoadPrivateKeyWithPath(cfg.WechatMchPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("加载商户私钥失败: %w", err)
	}

	ctx := context.Background()

	// 检查是否使用微信支付公钥
	var opts []core.ClientOption

	// 设置服务器URL
	opts = append(opts, option.WithHTTPClient(
		&http.Client{
			Timeout: 30 * time.Second,
		},
	))

	if cfg.WechatPayPublicKeyID != "" && cfg.WechatPayPublicKeyPath != "" {
		// 使用微信支付公钥进行身份验证
		wechatPayPublicKey, err := wechatutils.LoadPublicKeyWithPath(cfg.WechatPayPublicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("加载微信支付公钥失败: %w", err)
		}

		opts = append(opts, option.WithWechatPayPublicKeyAuthCipher(
			cfg.WechatMchID,
			cfg.WechatMchCertificateSerialNumber,
			mchPrivateKey,
			cfg.WechatPayPublicKeyID,
			wechatPayPublicKey,
		))
	} else {
		// 使用自动证书获取方式
		opts = append(opts, option.WithWechatPayAutoAuthCipher(
			cfg.WechatMchID,
			cfg.WechatMchCertificateSerialNumber,
			mchPrivateKey,
			cfg.WechatMchAPIv3Key,
		))
	}

	// 创建微信支付客户端
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("创建微信支付客户端失败: %w", err)
	}

	return &WechatTransferService{
		config: cfg,
		client: client,
		httpClient: func() *http.Client {
			cfg := config.GetConfig()
			timeout, err := time.ParseDuration(cfg.HTTPClientTimeout)
			if err != nil {
				timeout = 30 * time.Second // 默认超时时间
			}
			return &http.Client{Timeout: timeout}
		}(),
	}, nil
}

// TransferToUser 发起转账到用户（使用SDK客户端发送HTTP请求）
func (s *WechatTransferService) TransferToUser(request *models.TransferToUserRequest) (*models.TransferToUserResponse, error) {
	// 使用SDK客户端的Post方法发送请求
	ctx := context.Background()

	// 转账API完整URL
	cfg := s.config
	apiURL := cfg.WechatMchAPIURL + "/v3/fund-app/mch-transfer/transfer-bills"

	// 发送POST请求
	result, err := s.client.Post(ctx, apiURL, request)
	if err != nil {
		return nil, fmt.Errorf("发送转账请求失败: %w", err)
	}

	// 检查响应状态
	if result.Response.StatusCode != 200 {
		return nil, fmt.Errorf("转账API返回错误: 状态码=%d", result.Response.StatusCode)
	}

	// 解析响应体 - 直接读取并解析响应
	body, err := io.ReadAll(result.Response.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}
	defer result.Response.Body.Close()

	var transferResponse models.TransferToUserResponse
	if err := json.Unmarshal(body, &transferResponse); err != nil {
		return nil, fmt.Errorf("解析转账响应失败: %w", err)
	}

	return &transferResponse, nil
}

// 已删除手动HTTP实现方法，现在使用SDK的标准方式

// BuildTransferRequest 构建转账请求
func (s *WechatTransferService) BuildTransferRequest(userOpenID string, amount float64, userName string) *models.TransferToUserRequest {
	cfg := s.config

	// 转换金额为分
	transferAmount := utils.ConvertToWechatPayCents(amount)

	// 构建转账场景报备信息 - 佣金报酬场景必需
	reportInfos := []models.TransferSceneReportInfo{
		{
			InfoType:    "岗位类型",
			InfoContent: "代理推荐员",
		},
		{
			InfoType:    "报酬说明",
			InfoContent: "代理推荐佣金结算",
		},
	}

	request := &models.TransferToUserRequest{
		Appid:                    cfg.WechatAppID,
		OutBillNo:                "",     // 由调用方设置商户单号
		TransferSceneId:          "1005", // 佣金报酬场景ID
		Openid:                   userOpenID,
		TransferAmount:           transferAmount,
		TransferRemark:           "代理佣金提取",
		UserRecvPerception:       "劳务报酬",
		TransferSceneReportInfos: reportInfos,
	}

	// 根据转账金额决定是否传递用户姓名
	// 转账金额>=2000元时，该笔明细必须填写用户姓名
	if amount >= 2.0 && userName != "" {
		// 这里应该使用微信支付公钥加密用户姓名
		// 暂时使用原始姓名，实际项目中需要加密
		request.UserName = userName
		log.Printf("转账金额: %.2f元，>=2元，传入用户姓名: %s", amount, userName)
	} else {
		log.Printf("转账金额: %.2f元，<2元，不传入用户姓名", amount)
	}

	// 设置回调URL（可选）
	if notifyURL, err := cfg.GetValidatedNotifyURL(); err == nil {
		request.NotifyUrl = notifyURL + "/transfer"
		log.Printf("设置转账回调URL: %s", request.NotifyUrl)
	}

	return request
}
