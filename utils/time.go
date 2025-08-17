package utils

import (
	"log"
	"miniprogram/config"
	"time"
)

// 应用时区设置
var AppTimeZone *time.Location

// InitTimeZone 初始化应用时区
func InitTimeZone() {
	cfg := config.GetConfig()

	loc, err := time.LoadLocation(cfg.AppTimeZone)
	if err != nil {
		log.Printf("无法加载时区 %s，使用UTC: %v", cfg.AppTimeZone, err)
		AppTimeZone = time.UTC
	} else {
		AppTimeZone = loc
		log.Printf("应用时区设置为: %s", cfg.AppTimeZone)
	}
}

// GetCurrentUTCTime 获取当前UTC时间（推荐用于数据库存储）
func GetCurrentUTCTime() time.Time {
	return time.Now().UTC()
}

// GetCurrentAppTime 获取当前应用时区时间（用于显示）
func GetCurrentAppTime() time.Time {
	if AppTimeZone == nil {
		InitTimeZone()
	}
	return time.Now().In(AppTimeZone)
}

// ConvertToAppTimeZone 将UTC时间转换为应用时区（用于显示）
func ConvertToAppTimeZone(utcTime time.Time) time.Time {
	if AppTimeZone == nil {
		InitTimeZone()
	}
	return utcTime.In(AppTimeZone)
}

// FormatTimeForResponse 格式化时间用于API响应
func FormatTimeForResponse(t time.Time) string {
	if AppTimeZone == nil {
		InitTimeZone()
	}
	return t.In(AppTimeZone).Format(time.RFC3339)
}
