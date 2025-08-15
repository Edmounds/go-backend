package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"miniprogram/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	OpenID       string `json:"openID" binding:"required"`
	UserName     string `json:"user_name,omitempty"`
	UserPassword string `json:"user_password,omitempty"`
	Class        string `json:"class,omitempty"`
	Age          int    `json:"age,omitempty"`
	School       string `json:"school,omitempty"`
	Phone        string `json:"phone,omitempty"`
	City         string `json:"city,omitempty"`
	ReferredBy   string `json:"referred_by,omitempty"`
}

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	UserName     string `json:"user_name,omitempty"`
	UserPassword string `json:"user_password,omitempty"`
	Class        string `json:"class,omitempty"`
	Age          int    `json:"age,omitempty"`
	School       string `json:"school,omitempty"`
	Phone        string `json:"phone,omitempty"`
	City         string `json:"city,omitempty"`
	AgentLevel   int    `json:"agent_level,omitempty"`
}

// AddressRequest 地址请求
type AddressRequest struct {
	RecipientName string `json:"recipient_name" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
	Province      string `json:"province" binding:"required"`
	City          string `json:"city" binding:"required"`
	District      string `json:"district"`
	Street        string `json:"street" binding:"required"`
	PostalCode    string `json:"postal_code"`
	IsDefault     bool   `json:"is_default"`
}

// ===== 用户信息处理函数 =====

// ===== 数据库查询函数 =====

// GetUserByOpenID 根据openid获取用户信息
func GetUserByOpenID(openID string) (*models.User, error) {
	collection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var user models.User
	err := collection.FindOne(ctx, bson.M{"openID": openID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser 创建用户
func CreateUser(user *models.User) error {
	collection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 生成推荐码
	referralCode, err := GenerateReferralCode()
	if err != nil {
		return err
	}

	user.ReferralCode = referralCode
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	// 将插入后的ID设置回用户对象
	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// ValidateReferralCode 验证推荐码是否存在
func ValidateReferralCode(code string) (bool, error) {
	if code == "" {
		return true, nil // 推荐码为空是允许的
	}

	collection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := collection.CountDocuments(ctx, bson.M{"referral_code": code})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GenerateReferralCode 生成推荐码
func GenerateReferralCode() (string, error) {
	bytes := make([]byte, 4) // 8位十六进制字符
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPassword 验证密码
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// CreateUserHandler 创建或更新用户处理器
func CreateUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 1. 检查用户是否已存在
		existingUser, err := GetUserByOpenID(req.OpenID)
		if err != nil && err != mongo.ErrNoDocuments {
			InternalServerErrorResponse(c, "查询用户失败", err)
			return
		}

		// 2. 如果用户已存在，返回现有用户信息
		if existingUser != nil {
			SuccessResponse(c, "用户已存在", existingUser)
			return
		}

		// 3. 验证推荐码（如果提供）
		if req.ReferredBy != "" {
			valid, err := ValidateReferralCode(req.ReferredBy)
			if err != nil {
				InternalServerErrorResponse(c, "验证推荐码失败", err)
				return
			}
			if !valid {
				BadRequestResponse(c, "推荐码不存在", nil)
				return
			}
		}

		// 4. 创建用户对象
		user := &models.User{
			OpenID:     req.OpenID,
			UserName:   req.UserName,
			Class:      req.Class,
			Age:        req.Age,
			School:     req.School,
			Phone:      req.Phone,
			City:       req.City,
			ReferredBy: req.ReferredBy,
		}

		// 5. 如果提供了密码，进行加密
		if req.UserPassword != "" {
			hashedPassword, err := HashPassword(req.UserPassword)
			if err != nil {
				InternalServerErrorResponse(c, "密码加密失败", err)
				return
			}
			user.UserPassword = hashedPassword
		}

		// 6. 保存用户信息
		if err := CreateUser(user); err != nil {
			InternalServerErrorResponse(c, "创建用户失败", err)
			return
		}

		CreatedResponse(c, "用户创建成功", user)
	}
}

// UpdateUserHandler 更新用户信息处理器
func UpdateUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		openID := c.Param("user_id")
		var req UpdateUserRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 1. 验证用户是否存在
		_, err := GetUserByOpenID(openID)
		if err != nil {
			NotFoundResponse(c, "用户不存在", err)
			return
		}

		// 2. 构建更新数据
		collection := GetCollection("users")
		ctx, cancel := CreateDBContext()
		defer cancel()

		updateData := bson.M{"updated_at": time.Now()}
		if req.UserName != "" {
			updateData["user_name"] = req.UserName
		}
		if req.Class != "" {
			updateData["class"] = req.Class
		}
		if req.Age > 0 {
			updateData["age"] = req.Age
		}
		if req.School != "" {
			updateData["school"] = req.School
		}
		if req.Phone != "" {
			updateData["phone"] = req.Phone
		}
		if req.City != "" {
			updateData["city"] = req.City
		}
		if req.AgentLevel > 0 {
			updateData["agent_level"] = req.AgentLevel
			updateData["is_agent"] = true
		}

		if req.UserPassword != "" {
			hashedPassword, err := HashPassword(req.UserPassword)
			if err != nil {
				InternalServerErrorResponse(c, "密码加密失败", err)
				return
			}
			updateData["user_password"] = hashedPassword
		}

		update := bson.M{"$set": updateData}
		filter := bson.M{"openID": openID}
		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			InternalServerErrorResponse(c, "更新用户信息失败", err)
			return
		}

		SuccessResponse(c, "用户信息更新成功", nil)
	}
}

// GetUserHandler 获取用户信息处理器
func GetUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		openID := c.Param("user_id")

		user, err := GetUserByOpenID(openID)
		if err != nil {
			NotFoundResponse(c, "用户不存在", err)
			return
		}

		SuccessResponse(c, "获取用户信息成功", user)
	}
}
