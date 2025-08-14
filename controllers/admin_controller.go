package controllers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// UpdateAgentLevelRequest 更新代理等级请求
type UpdateAgentLevelRequest struct {
	AgentLevel int `json:"agent_level" binding:"required,min=0"`
}

// UpdateAgentLevelHandler 管理员更新用户代理等级（基于 openID）
func UpdateAgentLevelHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 注意：这里的 user_id 实际上是微信的 openID
		openID := c.Param("user_id")

		var req UpdateAgentLevelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 验证用户是否存在
		_, err := GetUserByOpenID(openID)
		if err != nil {
			NotFoundResponse(c, "用户不存在", err)
			return
		}

		// 执行更新：agent_level 与 is_agent（level>0 则为代理）
		collection := GetCollection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		update := bson.M{
			"$set": bson.M{
				"agent_level": req.AgentLevel,
				"is_agent":    req.AgentLevel > 0,
				"updated_at":  time.Now(),
			},
		}
		filter := bson.M{"openID": openID}

		if _, err := collection.UpdateOne(ctx, filter, update); err != nil {
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
