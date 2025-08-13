package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// 定义用户模型结构体 (简化版，用于JWT token)
type User struct {
	UserName     string
	UserId       string
	UserPassword string
	OpenID       string
}

// 定义JWT声明结构体
type Claims struct {
	UserName string
	UserId   string
	jwt.RegisteredClaims
}

// JWT密钥
var jwtKey = []byte("Chenqichen666")

// 生成JWT令牌
func GenerateToken(user User) (string, error) {
	// 设置过期时间 - 此处设置为24小时
	expirationTime := time.Now().Add(24 * time.Hour)

	// 创建JWT声明
	claims := &Claims{
		UserName: user.UserName,
		UserId:   user.UserId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "miniprogram",
			Subject:   user.UserName,
		},
	}

	// 使用指定的签名方法创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并获取完整的编码后的字符串令牌
	tokenString, err := token.SignedString(jwtKey)
	if HandleError(err, "生成Token失败", false) {
		return "", err
	}
	return tokenString, nil
}

// 验证JWT令牌
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	// 解析JWT令牌e
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if HandleError(err, "验证Token失败", false) {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("令牌无效")
	}

	return claims, nil
}

// 从请求头中提取Bearer token
func ExtractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("未提供Authorization头")
	}

	// 检查是否是Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("Authorization头格式错误，应为'Bearer {token}'")
	}

	return parts[1], nil
}

// JWT认证中间件 - Gin版本
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求中提取token
		tokenString, err := ExtractBearerTokenFromGin(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权: " + err.Error()})
			c.Abort()
			return
		}

		// 验证token
		claims, err := ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权: " + err.Error()})
			c.Abort()
			return
		}

		// 检查令牌是否过期
		if time.Now().Unix() > claims.ExpiresAt.Unix() {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权: 令牌已过期"})
			c.Abort()
			return
		}

		// 将用户信息添加到gin上下文
		c.Set("user", claims)

		// 继续处理请求
		c.Next()
	}
}

// 从Gin上下文中提取Bearer token
func ExtractBearerTokenFromGin(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("未提供Authorization头")
	}

	// 检查是否是Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("Authorization头格式错误，应为'Bearer {token}'")
	}

	return parts[1], nil
}

// 刷新JWT令牌
func RefreshToken(tokenString string, getUserFunc func(string) (bson.M, error)) (string, error) {
	// 验证旧令牌
	claims, err := ValidateToken(tokenString)
	if HandleError(err, "验证Token失败", false) {
		return "", err
	}

	// 获取用户信息，使用传入的函数而不是直接依赖controllers包
	user, err := getUserFunc(claims.UserName)
	if HandleError(err, "获取用户信息失败", false) {
		return "", err
	}

	openID := ""
	if openIDVal, exists := user["openID"]; exists && openIDVal != nil {
		openID = openIDVal.(string)
	}

	new_user := User{
		UserName:     user["user_name"].(string),
		UserId:       user["_id"].(primitive.ObjectID).Hex(),
		UserPassword: user["user_password"].(string),
		OpenID:       openID,
	}

	// 生成新令牌
	newToken, err := GenerateToken(new_user)
	if HandleError(err, "刷新Token失败", false) {
		return "", err
	}
	return newToken, nil
}

// 从上下文中获取用户信息
func GetUserFromContext(r *http.Request) (*Claims, bool) {
	claims, ok := r.Context().Value("user").(*Claims)
	return claims, ok
}

// 从Gin上下文中获取用户信息
func GetUserFromGinContext(c *gin.Context) (*Claims, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	claims, ok := user.(*Claims)
	return claims, ok
}

// 将用户信息添加到上下文
func AddUserToContext(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, "user", claims)
}

// 数据库访问函数类型定义
var GetCollectionFunc func(string) *mongo.Collection

// 设置数据库访问函数
func SetGetCollectionFunc(fn func(string) *mongo.Collection) {
	GetCollectionFunc = fn
}

// 从数据库获取用户信息
func getUserByUsername(username string) (bson.M, error) {
	if GetCollectionFunc == nil {
		return nil, errors.New("数据库访问函数未设置")
	}

	collection := GetCollectionFunc("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user bson.M
	err := collection.FindOne(ctx, bson.M{"user_name": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// RefreshTokenMiddleware Token刷新中间件
func RefreshTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求中提取token
		tokenString, err := ExtractBearerTokenFromGin(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未授权: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 刷新token
		newTokenString, err := RefreshToken(tokenString, getUserByUsername)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token刷新失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 返回新的token
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Token刷新成功",
			"data": gin.H{
				"token": newTokenString,
			},
		})
	}
}
