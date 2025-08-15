package controllers

import (
	"miniprogram/models"

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
			"openID":      openID,
			"agent_level": req.AgentLevel,
			"is_agent":    req.AgentLevel > 0,
		})
	}
}
