package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WithdrawRequest 提取佣金请求
type WithdrawRequest struct {
	Amount         float64 `json:"amount" binding:"required,min=0.01"`
	WithdrawMethod string  `json:"withdraw_method" binding:"required"`
	AccountInfo    struct {
		AccountName   string `json:"account_name"`
		AccountNumber string `json:"account_number"`
	} `json:"account_info"`
}

// GetAgentUsersHandler 获取代理管理的用户列表处理器
func GetAgentUsersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		school := c.Query("school")
		region := c.Query("region")

		// TODO: 实现代理用户列表查询逻辑
		// 1. 验证用户是否为代理
		// 2. 根据代理等级查询管理的用户
		// 3. 根据学校或地区筛选
		// 4. 返回用户列表和统计信息

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "获取管理用户列表成功",
			"data": gin.H{
				"users": []gin.H{
					{
						"_id":           "507f1f77bcf86cd799439012",
						"user_name":     "alice_wang",
						"school":        "北京大学",
						"age":           20,
						"phone":         "13800138001",
						"agent_level":   0,
						"referral_code": "ALICE456",
						"created_at":    "2024-01-10T10:00:00Z",
						"last_login":    "2024-01-15T09:00:00Z",
						"total_orders":  3,
						"total_spent":   149.7,
					},
					{
						"_id":           "507f1f77bcf86cd799439013",
						"user_name":     "bob_li",
						"school":        "清华大学",
						"age":           21,
						"phone":         "13800138002",
						"agent_level":   0,
						"referral_code": "BOB789",
						"created_at":    "2024-01-12T14:30:00Z",
						"last_login":    "2024-01-15T08:30:00Z",
						"total_orders":  1,
						"total_spent":   59.9,
					},
				},
				"total_users": 2,
				"agent_info": gin.H{
					"agent_level":     1,
					"agent_type":      "school",
					"managed_schools": []string{"北京大学", "清华大学"},
					"managed_regions": []string{},
				},
				"statistics": gin.H{
					"active_users":    2,
					"new_users_month": 2,
					"total_revenue":   209.6,
					"total_orders":    4,
				},
			},
		})

		// 避免未使用变量的警告
		_ = userID
		_ = school
		_ = region
	}
}

// GetAgentSalesHandler 获取代理销售数据处理器
func GetAgentSalesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")

		// TODO: 实现代理销售数据查询逻辑
		// 1. 验证用户是否为代理
		// 2. 根据时间范围查询销售数据
		// 3. 计算佣金统计
		// 4. 生成销售趋势数据

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "获取销售数据成功",
			"data": gin.H{
				"sales_summary": gin.H{
					"total_sales":      2096.5,
					"total_orders":     25,
					"total_commission": 104.8, // 5%佣金率
					"period": gin.H{
						"start_date": "2024-01-01",
						"end_date":   "2024-01-31",
					},
				},
				"sales_by_product": []gin.H{
					{
						"product_id":    "PROD001",
						"product_name":  "高级英语词汇书",
						"quantity_sold": 20,
						"total_revenue": 998.0,
						"commission":    49.9,
					},
					{
						"product_id":    "PROD002",
						"product_name":  "英语语法精讲",
						"quantity_sold": 15,
						"total_revenue": 598.5,
						"commission":    29.9,
					},
				},
				"monthly_trend": []gin.H{
					{
						"month":      "2024-01",
						"sales":      2096.5,
						"orders":     25,
						"commission": 104.8,
					},
					{
						"month":      "2023-12",
						"sales":      1850.0,
						"orders":     22,
						"commission": 92.5,
					},
				},
				"top_performers": []gin.H{
					{
						"user_id":      "507f1f77bcf86cd799439012",
						"user_name":    "alice_wang",
						"total_spent":  299.4,
						"orders_count": 6,
					},
					{
						"user_id":      "507f1f77bcf86cd799439013",
						"user_name":    "bob_li",
						"total_spent":  179.7,
						"orders_count": 3,
					},
				},
			},
		})

		// 避免未使用变量的警告
		_ = userID
		_ = startDate
		_ = endDate
	}
}

// WithdrawCommissionHandler 提取佣金处理器
func WithdrawCommissionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		var req WithdrawRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
				"error":   err.Error(),
			})
			return
		}

		// TODO: 实现佣金提取逻辑
		// 1. 验证用户是否为代理
		// 2. 检查可提取佣金余额
		// 3. 验证提取金额
		// 4. 创建提取记录
		// 5. 更新佣金余额
		// 6. 发起提取流程

		// 验证提取方式
		validMethods := map[string]bool{
			"wechat":        true,
			"alipay":        true,
			"bank_transfer": true,
		}

		if !validMethods[req.WithdrawMethod] {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
				"error":   "不支持的提取方式",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "提取申请提交成功",
			"data": gin.H{
				"withdraw_id":       "WD20240115001",
				"user_id":           userID,
				"amount":            req.Amount,
				"withdraw_method":   req.WithdrawMethod,
				"account_info":      req.AccountInfo,
				"status":            "pending",
				"estimated_arrival": "2024-01-17T10:00:00Z", // 预计2个工作日到账
				"created_at":        "2024-01-15T10:00:00Z",
				"processing_fee":    req.Amount * 0.01, // 1%手续费
				"actual_amount":     req.Amount * 0.99, // 实际到账金额
			},
		})
	}
}
