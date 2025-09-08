package middlewares

import (
	"context"
	"errors"
	"net/http"
	"time"

	"miniprogram/config"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AdminAuthMiddleware 管理员权限验证中间件
// 必须在JWT认证中间件之后使用
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取JWT声明
		claims, exists := GetUserFromGinContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未授权: 未找到用户信息",
				"error":   "unauthorized",
			})
			c.Abort()
			return
		}

		// 从数据库获取完整的用户信息以检查管理员权限
		user, err := getUserByOpenID(claims.UserId) // UserId 实际存储的是 OpenID
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "服务器内部错误: 无法获取用户信息",
				"error":   "internal_server_error",
			})
			c.Abort()
			return
		}

		// 检查用户是否为管理员
		isAdmin, ok := user["is_admin"].(bool)
		if !ok || !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足: 需要管理员权限",
				"error":   "insufficient_privileges",
			})
			c.Abort()
			return
		}

		// 将完整的用户信息添加到上下文中，供后续处理器使用
		c.Set("admin_user", user)

		// 继续处理请求
		c.Next()
	}
}

// getUserByOpenID 根据OpenID从数据库获取用户信息
func getUserByOpenID(openID string) (bson.M, error) {
	if GetCollectionFunc == nil {
		return nil, ErrDatabaseAccessFuncNotSet
	}

	collection := GetCollectionFunc("users")
	cfg := config.GetConfig()
	timeout, err := time.ParseDuration(cfg.MongoDBTimeout)
	if err != nil {
		timeout = 10 * time.Second // 默认超时时间
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var user bson.M
	err = collection.FindOne(ctx, bson.M{"openID": openID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetAdminUserFromGinContext 从Gin上下文中获取管理员用户信息
func GetAdminUserFromGinContext(c *gin.Context) (bson.M, bool) {
	user, exists := c.Get("admin_user")
	if !exists {
		return nil, false
	}
	adminUser, ok := user.(bson.M)
	return adminUser, ok
}

// 错误定义
var (
	ErrDatabaseAccessFuncNotSet = errors.New("数据库访问函数未设置")
	ErrUserNotFound             = errors.New("用户不存在")
)
