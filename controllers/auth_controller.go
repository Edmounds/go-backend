package controllers

import (
	"errors"
	"miniprogram/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 定义用户模型结构体
type User struct {
	UserName     string
	UserId       string
	UserPassword string
}

// 定义JWT声明结构体
type Claims struct {
	UserName string
	UserId   string
	jwt.RegisteredClaims
}

// JWT密钥 - 在实际生产环境中应该从环境变量或配置文件中获取
var jwtKey = []byte("miniprogram_secret_key")

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
	if utils.HandleError(err, "生成Token失败", false) {
		return "", err
	}
	return tokenString, nil
}

// 验证JWT令牌
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	// 解析JWT令牌
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if utils.HandleError(err, "验证Token失败", false) {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("令牌无效")
	}

	return claims, nil
}

func AuthMiddleware(tokenString string) (*Claims, error) {
	claims, err := ValidateToken(tokenString)

	if utils.HandleError(err, "验证Token失败", false) {
		return nil, err
	}
	// 检查令牌是否过期
	if time.Now().Unix() > claims.ExpiresAt.Unix() {
		return nil, errors.New("令牌已过期")
	}

	return claims, nil
}

// // 刷新JWT令牌
func RefreshToken(tokenString string) (string, error) {
	// 验证旧令牌
	claims, err := ValidateToken(tokenString)

	if utils.HandleError(err, "验证Token失败", false) {
		return "", err
	}

	// 获取用户信息
	user, err := GetUser(claims.UserName)
	if utils.HandleError(err, "获取用户信息失败", false) {
		return "", err
	}

	new_user := User{
		UserName:     user["user_name"].(string),
		UserId:       user["_id"].(primitive.ObjectID).Hex(), // 使用 Hex() 方法转换为字符串
		UserPassword: user["user_password"].(string),
	}
	// 生成新令牌
	newToken, err := GenerateToken(new_user)
	if utils.HandleError(err, "刷新Token失败", false) {
		return "", err
	}
	return newToken, nil
}
