package controllers

import (
	"miniprogram/models"
	"miniprogram/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ===== HTTP 处理器 =====

// UpdateAgentLevelHandler 管理员更新用户代理等级处理器
func UpdateAgentLevelHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 注意：这里的 user_id 实际上是微信的 openID
		openID := c.Param("user_id")

		var req models.UpdateAgentLevelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 初始化管理员服务
		adminService := GetAdminService()

		// 更新代理等级
		err := adminService.UpdateUserAgentLevel(openID, req.AgentLevel)
		if err != nil {
			InternalServerErrorResponse(c, "更新代理等级失败", err)
			return
		}

		SuccessResponse(c, "代理等级更新成功", gin.H{
			"user_id":     utils.EncodeOpenIDToSafeID(openID), // 使用安全的用户标识符
			"agent_level": req.AgentLevel,
			"is_agent":    req.AgentLevel > 0,
		})
	}
}

// ===== 用户管理处理器 =====

// GetAllUsersHandler 获取所有用户列表处理器
func GetAllUsersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.AdminUserListRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		adminService := GetAdminService()
		response, err := adminService.GetAllUsers(req)
		if err != nil {
			InternalServerErrorResponse(c, "获取用户列表失败", err)
			return
		}

		SuccessResponse(c, "获取用户列表成功", response)
	}
}

// GetUserDetailHandler 获取用户详细信息处理器
func GetUserDetailHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.Param("user_id")

		adminService := GetAdminService()
		user, err := adminService.GetUserDetail(openID)
		if err != nil {
			if err.Error() == "用户不存在" {
				NotFoundResponse(c, "用户不存在", err)
				return
			}
			InternalServerErrorResponse(c, "获取用户详情失败", err)
			return
		}

		SuccessResponse(c, "获取用户详情成功", user)
	}
}

// UpdateUserAdminStatusHandler 设置/取消用户管理员权限处理器
func UpdateUserAdminStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.Param("user_id")

		var req models.UpdateUserAdminRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		adminService := GetAdminService()
		err := adminService.UpdateUserAdminStatus(openID, req.IsAdmin)
		if err != nil {
			InternalServerErrorResponse(c, "更新管理员权限失败", err)
			return
		}

		SuccessResponse(c, "管理员权限更新成功", gin.H{
			"user_id":  utils.EncodeOpenIDToSafeID(openID),
			"is_admin": req.IsAdmin,
		})
	}
}

// GetUserOrdersHandler 获取指定用户订单列表处理器
func GetUserOrdersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.Param("user_id")

		// 解析分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

		adminService := GetAdminService()
		orders, pagination, err := adminService.GetUserOrders(openID, page, limit)
		if err != nil {
			InternalServerErrorResponse(c, "获取用户订单失败", err)
			return
		}

		SuccessResponse(c, "获取用户订单成功", gin.H{
			"orders":     orders,
			"pagination": pagination,
		})
	}
}

// ===== 代理管理处理器扩展 =====

// UpdateAgentSchoolsHandler 设置校代理管理学校处理器
func UpdateAgentSchoolsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.Param("user_id")

		var req models.UpdateAgentSchoolsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		adminService := GetAdminService()
		err := adminService.UpdateAgentSchools(openID, req.Schools)
		if err != nil {
			InternalServerErrorResponse(c, "设置校代理学校失败", err)
			return
		}

		SuccessResponse(c, "校代理学校设置成功", gin.H{
			"user_id": utils.EncodeOpenIDToSafeID(openID),
			"schools": req.Schools,
		})
	}
}

// UpdateAgentRegionsHandler 设置区代理管理区域处理器
func UpdateAgentRegionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.Param("user_id")

		var req models.UpdateAgentRegionsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		adminService := GetAdminService()
		err := adminService.UpdateAgentRegions(openID, req.Regions)
		if err != nil {
			InternalServerErrorResponse(c, "设置区代理区域失败", err)
			return
		}

		SuccessResponse(c, "区代理区域设置成功", gin.H{
			"user_id": utils.EncodeOpenIDToSafeID(openID),
			"regions": req.Regions,
		})
	}
}

// GetAgentStatsHandler 获取代理统计信息处理器
func GetAgentStatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.Param("user_id")

		adminService := GetAdminService()
		stats, err := adminService.GetAgentStats(openID)
		if err != nil {
			if err.Error() == "用户不是代理" {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "权限不足: 用户不是代理",
					"error":   "forbidden",
				})
				return
			}
			InternalServerErrorResponse(c, "获取代理统计失败", err)
			return
		}

		SuccessResponse(c, "获取代理统计成功", stats)
	}
}

// ===== 商品管理处理器 =====

// CreateProductHandler 创建商品处理器
func CreateProductHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		adminService := GetAdminService()
		product, err := adminService.CreateProduct(req)
		if err != nil {
			InternalServerErrorResponse(c, "创建商品失败", err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"code":    201,
			"message": "商品创建成功",
			"data":    product,
		})
	}
}

// UpdateProductHandler 更新商品处理器
func UpdateProductHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("product_id")

		var req models.UpdateProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		adminService := GetAdminService()
		product, err := adminService.UpdateProduct(productID, req)
		if err != nil {
			InternalServerErrorResponse(c, "更新商品失败", err)
			return
		}

		SuccessResponse(c, "商品更新成功", product)
	}
}

// DeleteProductHandler 删除商品处理器
func DeleteProductHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("product_id")

		adminService := GetAdminService()
		err := adminService.DeleteProduct(productID)
		if err != nil {
			if err.Error() == "商品不存在" {
				NotFoundResponse(c, "商品不存在", err)
				return
			}
			InternalServerErrorResponse(c, "删除商品失败", err)
			return
		}

		SuccessResponse(c, "商品删除成功", nil)
	}
}

// UpdateProductStatusHandler 更新商品状态处理器
func UpdateProductStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("product_id")

		var req models.UpdateProductStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		adminService := GetAdminService()
		err := adminService.UpdateProductStatus(productID, req.Status)
		if err != nil {
			if err.Error() == "商品不存在" {
				NotFoundResponse(c, "商品不存在", err)
				return
			}
			if err.Error() == "无效的状态值" {
				BadRequestResponse(c, "无效的状态值", err)
				return
			}
			InternalServerErrorResponse(c, "更新商品状态失败", err)
			return
		}

		SuccessResponse(c, "商品状态更新成功", gin.H{
			"product_id": productID,
			"status":     req.Status,
		})
	}
}

// ===== 仪表盘API处理器 =====

// GetDashboardStatsHandler 获取仪表盘统计数据处理器
func GetDashboardStatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		adminService := GetAdminService()
		stats, err := adminService.GetDashboardStats()
		if err != nil {
			InternalServerErrorResponse(c, "获取仪表盘统计失败", err)
			return
		}

		SuccessResponse(c, "获取仪表盘统计成功", stats)
	}
}

// GetDashboardRecentOrdersHandler 获取最近订单列表处理器
func GetDashboardRecentOrdersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		adminService := GetAdminService()
		orders, err := adminService.GetRecentOrders(limit)
		if err != nil {
			InternalServerErrorResponse(c, "获取最近订单失败", err)
			return
		}

		SuccessResponse(c, "获取最近订单成功", gin.H{
			"orders": orders,
		})
	}
}

// GetSalesTrendHandler 获取销售趋势数据处理器
func GetSalesTrendHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

		adminService := GetAdminService()
		trend, err := adminService.GetSalesTrend(days)
		if err != nil {
			InternalServerErrorResponse(c, "获取销售趋势失败", err)
			return
		}

		SuccessResponse(c, "获取销售趋势成功", gin.H{
			"trend": trend,
		})
	}
}

// GetUserGrowthHandler 获取用户增长趋势数据处理器
func GetUserGrowthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

		adminService := GetAdminService()
		growth, err := adminService.GetUserGrowth(days)
		if err != nil {
			InternalServerErrorResponse(c, "获取用户增长趋势失败", err)
			return
		}

		SuccessResponse(c, "获取用户增长趋势成功", gin.H{
			"growth": growth,
		})
	}
}

// GetAllOrdersHandler 管理员获取所有订单列表处理器
func GetAllOrdersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.AdminOrderListRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		adminService := GetAdminService()
		response, err := adminService.GetAllOrders(req)
		if err != nil {
			InternalServerErrorResponse(c, "获取订单列表失败", err)
			return
		}

		SuccessResponse(c, "获取订单列表成功", response)
	}
}
