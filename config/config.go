package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config 应用配置结构
type Config struct {
	// 服务器配置
	ServerPort  string
	BaseAPIURL  string
	Environment string

	// 微信小程序配置
	WechatMchID                      string
	WechatMchCertificateSerialNumber string
	WechatMchAPIv3Key                string
	WechatAppID                      string
	WechatAppSecret                  string
	WechatAPIURL                     string
	WechatMchAPIURL                  string // 微信商户API地址

	// 微信支付证书配置
	WechatMchPrivateKeyPath string // 商户API证书私钥文件路径
	WechatPayPublicKeyPath  string // 微信支付公钥文件路径
	WechatPayPublicKeyID    string // 微信支付公钥ID

	// 数据库配置
	MongoDBURL          string
	MongoDBDatabaseName string // 数据库名称
	MongoDBTimeout      string // 数据库连接超时时间
	OBSURL              string

	// 安全配置
	UserIDSecretKey string // 用户标识符加密密钥

	// 时区配置
	AppTimeZone string

	// JWT配置
	JWTSecret string

	// 网络配置
	HTTPClientTimeout string // HTTP客户端超时时间

	// 开发配置
	EnableDevLogin bool   // 是否启用开发登录接口
	LogLevel       string // 日志级别

	// 小程序码配置
	QRCodeEnvVersion string // 小程序码环境版本 (release/trial/develop)
	QRCodeWidth      int    // 小程序码宽度
	QRCodePage       string // 小程序码跳转页面

	// 折扣率配置
	DiscountRateNormalUser    float64 // 普通用户折扣率
	DiscountRateSchoolAgent   float64 // 校代理折扣率
	DiscountRateRegionalAgent float64 // 区域代理折扣率
	DiscountRateDefault       float64 // 默认折扣率

	// 佣金率配置
	CommissionRateNormalUser    float64 // 普通用户佣金率
	CommissionRateSchoolAgent   float64 // 校代理佣金率
	CommissionRateRegionalAgent float64 // 区域代理佣金率
	CommissionRateDefault       float64 // 默认佣金率
}

// GetConfig 获取应用配置
func GetConfig() *Config {
	return &Config{
		// 服务器配置
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		BaseAPIURL:  getEnv("BASE_API_URL", "https://backend.edmounds.top"),
		Environment: getEnv("ENVIRONMENT", "development"),

		// 微信小程序配置
		WechatMchID:                      getEnv("WECHAT_MCH_ID", ""),
		WechatMchCertificateSerialNumber: getEnv("WECHAT_MCH_CERTIFICATE_SERIAL_NUMBER", ""),
		WechatMchAPIv3Key:                getEnv("WECHAT_MCH_API_V3_KEY", ""),
		WechatAppID:                      getEnv("WECHAT_APP_ID", ""),
		WechatAppSecret:                  getEnv("WECHAT_APP_SECRET", ""),
		WechatAPIURL:                     getEnv("WECHAT_API_URL", "https://api.weixin.qq.com"),
		WechatMchAPIURL:                  getEnv("WECHAT_MCH_API_URL", "https://api.mch.weixin.qq.com"),

		// 微信支付证书配置
		WechatMchPrivateKeyPath: getEnv("WECHAT_MCH_PRIVATE_KEY_PATH", "cert/apiclient_key.pem"),
		WechatPayPublicKeyPath:  getEnv("WECHAT_PAY_PUBLIC_KEY_PATH", "cert/pub_key.pem"),
		WechatPayPublicKeyID:    getEnv("WECHAT_PAY_PUBLIC_KEY_ID", "PUB_KEY_ID_0116812471822025081300112175001600"),

		// 数据库配置
		MongoDBURL:          getEnv("MONGODB_URL", "mongodb://miniprogram_db:Chenqichen666@113.45.220.0/miniprogram_db?authSource=miniprogram_db"),
		MongoDBDatabaseName: getEnv("MONGODB_DATABASE_NAME", "miniprogram_db"),
		MongoDBTimeout:      getEnv("MONGODB_TIMEOUT", "10s"),
		OBSURL:              getEnv("OBS_URL", "https://mini-app-89d7.obs.cn-south-1.myhuaweicloud.com"),

		// 安全配置
		UserIDSecretKey: getEnv("USER_ID_SECRET_KEY", "Chenqichen666"),

		// 时区配置
		AppTimeZone: getEnv("APP_TIMEZONE", "UTC"),

		// JWT配置
		JWTSecret: getEnv("JWT_SECRET", "chenqichen666"),

		// 网络配置
		HTTPClientTimeout: getEnv("HTTP_CLIENT_TIMEOUT", "30s"),

		// 开发配置
		EnableDevLogin: getEnv("ENABLE_DEV_LOGIN", "true") == "true",
		LogLevel:       getEnv("LOG_LEVEL", "info"),

		// 小程序码配置
		QRCodeEnvVersion: getEnv("QRCODE_ENV_VERSION", "develop"),
		QRCodeWidth:      getEnvInt("QRCODE_WIDTH", 280),
		QRCodePage:       getEnv("QRCODE_PAGE", "pages/index/index"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数类型的环境变量，如果不存在或无法解析则返回默认值
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// ValidateWechatPayNotifyURL 验证微信支付回调URL是否符合要求
func ValidateWechatPayNotifyURL(baseURL string) error {
	// 1. 检查URL是否为空
	if baseURL == "" {
		return fmt.Errorf("BaseAPIURL不能为空")
	}

	// 2. 解析URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("BaseAPIURL格式错误: %w", err)
	}

	// 3. 检查协议
	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
		return fmt.Errorf("BaseAPIURL必须以https://或http://开头，当前协议: %s", parsedURL.Scheme)
	}

	// 4. 检查域名/IP
	host := parsedURL.Hostname()
	if host == "" {
		return fmt.Errorf("BaseAPIURL缺少域名或IP")
	}

	// 5. 检查是否为本地或内网IP
	localIPs := []string{"localhost", "127.0.0.1", "0.0.0.0"}
	for _, localIP := range localIPs {
		if host == localIP {
			return fmt.Errorf("BaseAPIURL不能使用本地IP: %s", host)
		}
	}

	// 6. 检查内网IP段
	if strings.HasPrefix(host, "192.168.") || strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "172.16.") || strings.HasPrefix(host, "172.17.") ||
		strings.HasPrefix(host, "172.18.") || strings.HasPrefix(host, "172.19.") ||
		strings.HasPrefix(host, "172.20.") || strings.HasPrefix(host, "172.21.") ||
		strings.HasPrefix(host, "172.22.") || strings.HasPrefix(host, "172.23.") ||
		strings.HasPrefix(host, "172.24.") || strings.HasPrefix(host, "172.25.") ||
		strings.HasPrefix(host, "172.26.") || strings.HasPrefix(host, "172.27.") ||
		strings.HasPrefix(host, "172.28.") || strings.HasPrefix(host, "172.29.") ||
		strings.HasPrefix(host, "172.30.") || strings.HasPrefix(host, "172.31.") {
		return fmt.Errorf("BaseAPIURL不能使用内网IP: %s", host)
	}

	// 7. 检查RawQuery（不能携带参数）
	if parsedURL.RawQuery != "" {
		return fmt.Errorf("BaseAPIURL不能携带参数")
	}

	return nil
}

// GetValidatedNotifyURL 获取验证过的微信支付回调URL
func (c *Config) GetValidatedNotifyURL() (string, error) {
	// 验证BaseAPIURL
	if err := ValidateWechatPayNotifyURL(c.BaseAPIURL); err != nil {
		return "", err
	}

	// 构建完整的回调URL
	notifyURL := c.BaseAPIURL + "/api/wechat/pay/notify"

	// 再次验证完整URL
	if err := ValidateWechatPayNotifyURL(notifyURL); err != nil {
		return "", fmt.Errorf("回调URL验证失败: %w", err)
	}

	return notifyURL, nil
}
