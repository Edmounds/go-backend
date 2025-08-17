package controllers

import (
	"log"
	"miniprogram/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ===== HTTP 处理器 =====

// GetAgentUsersHandler 获取代理管理的用户列表处理器
func GetAgentUsersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		openID := c.Param("user_id")
		school := c.Query("school")
		region := c.Query("region")

		// 初始化服务
		userService := NewAgentUserService()

		// 1. 验证用户是否为代理
		agent, err := userService.IsValidAgent(openID)
		if err != nil {
			ErrorResponse(c, http.StatusForbidden, 403, "用户不是代理或权限不足", err)
			return
		}

		// 2. 根据代理等级查询管理的用户
		managedUsers, err := userService.GetManagedUsers(agent, school, region)
		if err != nil {
			InternalServerErrorResponse(c, "获取管理用户列表失败", err)
			return
		}

		// 3. 为每个用户添加订单统计信息和推荐码使用详情
		referralService := NewReferralCodeService()
		var usersData []gin.H
		for _, user := range managedUsers {
			totalOrders, totalSpent, err := userService.GetUserOrderStats(user.OpenID)
			if err != nil {
				totalOrders = 0
				totalSpent = 0.0
			}

			// 获取用户推荐码使用情况
			var referralUsageCount int
			var referralUsageDetails []gin.H
			if user.ReferralCode != "" {
				referral, err := referralService.GetReferralByCode(user.ReferralCode)
				if err == nil {
					referralUsageCount = len(referral.UsedBy)
					for _, usage := range referral.UsedBy {
						referralUsageDetails = append(referralUsageDetails, gin.H{
							"user_name": usage.UserName,
							"used_at":   usage.UsedAt.Format(time.RFC3339),
						})
					}
				}
			}

			usersData = append(usersData, gin.H{
				"_id":         user.ID,
				"openID":      user.OpenID,
				"user_name":   user.UserName,
				"school":      user.School,
				"city":        user.City,
				"age":         user.Age,
				"phone":       user.Phone,
				"agent_level": user.AgentLevel,

				"referral_code":          user.ReferralCode,
				"referral_usage_count":   referralUsageCount,
				"referral_usage_details": referralUsageDetails,
				"created_at":             user.CreatedAt.Format(time.RFC3339),
				"updated_at":             user.UpdatedAt.Format(time.RFC3339),
				"total_orders":           totalOrders,
				"total_spent":            totalSpent,
			})
		}

		// 4. 获取代理统计信息
		statistics, err := userService.GetAgentStatistics(agent)
		if err != nil {
			// 如果获取统计失败，使用默认值
			statistics = map[string]interface{}{
				"active_users":    0,
				"new_users_month": 0,
				"total_revenue":   0.0,
				"total_orders":    0,
			}
		}

		// 5. 构建代理信息
		SuccessResponse(c, "获取管理用户列表成功", gin.H{
			"users":       usersData,
			"total_users": len(managedUsers),
			"agent_info": gin.H{
				"agent_level": agent.AgentLevel,

				"managed_schools": agent.ManagedSchools,
				"managed_regions": agent.ManagedRegions,
			},
			"statistics": statistics,
		})
	}
}

// GetAgentSalesHandler 获取代理销售数据处理器
func GetAgentSalesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		openID := c.Param("user_id")
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")

		// 初始化服务
		userService := NewAgentUserService()
		salesService := NewAgentSalesService()

		// 1. 验证用户是否为代理
		agent, err := userService.IsValidAgent(openID)
		if err != nil {
			ErrorResponse(c, http.StatusForbidden, 403, "用户不是代理或权限不足", err)
			return
		}

		// 2. 根据时间范围查询销售数据
		salesData, err := salesService.GetSalesData(agent, startDate, endDate)
		if err != nil {
			InternalServerErrorResponse(c, "获取销售数据失败", err)
			return
		}

		SuccessResponse(c, "获取销售数据成功", salesData)
	}
}

// GetAgentCommissionDashboardHandler 获取代理佣金仪表板处理器
func GetAgentCommissionDashboardHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID
		openID := c.Param("user_id")

		// 初始化服务
		userService := NewAgentUserService()
		commissionService := NewAgentCommissionService()

		// 1. 验证用户是否为代理
		_, err := userService.IsValidAgent(openID)
		if err != nil {
			ErrorResponse(c, http.StatusForbidden, 403, "用户不是代理或权限不足", err)
			return
		}

		// 2. 获取代理佣金仪表板数据
		dashboardData, err := commissionService.GetCommissionDashboard(openID)
		if err != nil {
			InternalServerErrorResponse(c, "获取佣金仪表板数据失败", err)
			return
		}

		SuccessResponse(c, "获取佣金仪表板数据成功", dashboardData)
	}
}

// GetAgentCommissionDetailsHandler 获取代理佣金明细处理器
func GetAgentCommissionDetailsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID
		openID := c.Param("user_id")

		// 获取查询参数
		monthsParam := c.DefaultQuery("months", "6") // 默认查询6个月
		months := 6
		if m, err := strconv.Atoi(monthsParam); err == nil {
			months = m
		}
		// 限制查询范围
		if months < 1 {
			months = 1
		}
		if months > 12 {
			months = 12
		}

		// 初始化服务
		userService := NewAgentUserService()
		commissionService := NewAgentCommissionService()

		// 1. 验证用户是否为代理
		_, err := userService.IsValidAgent(openID)
		if err != nil {
			ErrorResponse(c, http.StatusForbidden, 403, "用户不是代理或权限不足", err)
			return
		}

		// 2. 获取代理佣金明细数据
		detailsData, err := commissionService.GetCommissionDetails(openID, months)
		if err != nil {
			InternalServerErrorResponse(c, "获取佣金明细数据失败", err)
			return
		}

		SuccessResponse(c, "获取佣金明细数据成功", detailsData)
	}
}

// WithdrawCommissionHandler 提取佣金处理器
func WithdrawCommissionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		openID := c.Param("user_id")
		var req models.WithdrawRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 初始化服务
		userService := NewAgentUserService()
		withdrawService := NewAgentWithdrawService()

		// 1. 验证用户是否为代理
		agent, err := userService.IsValidAgent(openID)
		if err != nil {
			ErrorResponse(c, http.StatusForbidden, 403, "用户不是代理或权限不足", err)
			return
		}

		// 验证提取方式
		validMethods := map[string]bool{
			"wechat":        true,
			"alipay":        true,
			"bank_transfer": true,
		}

		if !validMethods[req.WithdrawMethod] {
			BadRequestResponse(c, "不支持的提取方式", nil)
			return
		}

		// 2. 检查可提取佣金余额
		availableAmount, err := withdrawService.GetAvailableCommission(agent.OpenID)
		if err != nil {
			InternalServerErrorResponse(c, "获取可提取佣金失败", err)
			return
		}

		// 3. 验证提取金额
		if req.Amount > availableAmount {
			ErrorResponse(c, http.StatusBadRequest, 400, "提取金额不能超过可提取佣金余额", nil)
			return
		}

		// 最小提取金额检查
		if req.Amount < 10.0 {
			BadRequestResponse(c, "最小提取金额为10元", nil)
			return
		}

		// 4. 创建提取记录
		withdrawRecord, err := withdrawService.CreateWithdrawRecord(agent.OpenID, req.Amount, req.WithdrawMethod, req.AccountInfo)
		if err != nil {
			InternalServerErrorResponse(c, "创建提取记录失败", err)
			return
		}

		// 5. 调用微信支付企业转账处理提取
		err = ProcessAgentWithdraw(withdrawRecord.WithdrawID, agent.OpenID, req.Amount)
		if err != nil {
			// 转账失败，更新记录状态但不影响响应
			log.Printf("微信支付企业转账失败: %v", err)
		}

		SuccessResponse(c, "提取申请提交成功", gin.H{
			"withdraw_id":       withdrawRecord.WithdrawID,
			"openID":            openID,
			"amount":            withdrawRecord.Amount,
			"withdraw_method":   withdrawRecord.WithdrawMethod,
			"account_info":      withdrawRecord.AccountInfo,
			"status":            withdrawRecord.Status,
			"estimated_arrival": withdrawRecord.EstimatedArrival.Format(time.RFC3339),
			"created_at":        withdrawRecord.CreatedAt.Format(time.RFC3339),
			"processing_fee":    withdrawRecord.ProcessingFee,
			"actual_amount":     withdrawRecord.ActualAmount,
		})
	}
}
