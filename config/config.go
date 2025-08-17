package config

import (
	"os"
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

	// 数据库配置
	MongoDBURL string
	OBSURL     string

	// 时区配置
	AppTimeZone string

	// JWT配置
	JWTSecret string
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

		// 数据库配置
		MongoDBURL: getEnv("MONGODB_URL", "mongodb://localhost:27017"),
		OBSURL:     getEnv("OBS_URL", ""),

		// 时区配置
		AppTimeZone: getEnv("APP_TIMEZONE", "UTC"),

		// JWT配置
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
