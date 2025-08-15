package controllers

import (
	"time"
)

// ===== 管理员服务层 =====

// AdminService 管理员服务
type AdminService struct{}

// GetAdminService 获取管理员服务实例
func GetAdminService() *AdminService {
	return &AdminService{}
}

// UpdateUserAgentLevel 更新用户代理等级
func (s *AdminService) UpdateUserAgentLevel(openID string, agentLevel int) error {
	// 验证用户是否存在
	userService := GetUserService()
	_, err := userService.FindUserByOpenID(openID)
	if err != nil {
		return err
	}

	// 更新代理等级
	updates := map[string]interface{}{
		"agent_level": agentLevel,
		"is_agent":    agentLevel > 0,
		"updated_at":  time.Now(),
	}

	_, err = userService.UpdateUser(openID, updates)
	return err
}
