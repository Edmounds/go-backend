package utils

import "math"

/**
 * 金额处理工具函数
 * 用于确保微信支付金额符合要求（最小单位为分，即0.01元）
 */

// FormatMoneyForWechatPay 格式化金额用于微信支付
// 将金额舍去到分（0.01元），并确保最小支付金额为0.01元
func FormatMoneyForWechatPay(amount float64) float64 {
	// 舍去到分（保留两位小数）
	rounded := math.Floor(amount*100) / 100

	// 确保最小支付金额为0.01元
	if rounded < 0.01 {
		return 0.01
	}

	return rounded
}

// ConvertToWechatPayCents 将金额转换为微信支付要求的分
// 返回整数分，用于微信支付API调用
func ConvertToWechatPayCents(amount float64) int64 {
	// 先格式化金额，再转换为分
	formattedAmount := FormatMoneyForWechatPay(amount)
	return int64(formattedAmount * 100)
}
