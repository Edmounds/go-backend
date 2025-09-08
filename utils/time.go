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

// GetTodayStartUTC 获取今日开始时间（UTC）
func GetTodayStartUTC() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

// GetTodayEndUTC 获取今日结束时间（UTC）
func GetTodayEndUTC() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC)
}

// GetCurrentMonthStartUTC 获取当前月开始时间（UTC）
func GetCurrentMonthStartUTC() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// GetCurrentMonthEndUTC 获取当前月结束时间（UTC）
func GetCurrentMonthEndUTC() time.Time {
	now := time.Now().UTC()
	nextMonth := now.AddDate(0, 1, 0)
	return time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// ParseDateString 解析日期字符串（YYYY-MM-DD格式）
func ParseDateString(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
