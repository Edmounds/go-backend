package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	OpenID       string `json:"openID"`
	UserName     string `json:"user_name" binding:"required"`
	UserPassword string `json:"user_password" binding:"required"`
	Class        string `json:"class"`
	Age          int    `json:"age"`
	School       string `json:"school"`
	Phone        string `json:"phone"`
	City         string `json:"city"`
	ReferredBy   string `json:"referred_by"`
}

// UpdateUserRequest 更新基础用户信息请求
type UpdateUserRequest struct {
	UserName string `json:"user_name"`
	Class    string `json:"class"`
	Age      int    `json:"age"`
	School   string `json:"school"`
	Phone    string `json:"phone"`
	City     string `json:"city"`
}

// UpdateAgentRequest 更新代理信息请求
type UpdateAgentRequest struct {
	AgentLevel     int      `json:"agent_level"`
	AgentType      string   `json:"agent_type"`
	ManagedSchools []string `json:"managed_schools"`
	ManagedRegions []string `json:"managed_regions"`
}

// UserBasicResponse 基础用户信息响应（用于注册、更新等简单操作）
type UserBasicResponse struct {
	ID           string `json:"_id"`
	UserName     string `json:"user_name"`
	Age          int    `json:"age"`
	School       string `json:"school"`
	Phone        string `json:"phone"`
	Class        string `json:"class"`
	City         string `json:"city"`
	ReferralCode string `json:"referral_code"`
}

// UserProfileResponse 完整用户档案响应（用于个人中心、登录等需要完整信息的场景）
type UserProfileResponse struct {
	ID             string   `json:"_id"`
	OpenID         string   `json:"openID,omitempty"`
	UserName       string   `json:"user_name"`
	Class          string   `json:"class"`
	Age            int      `json:"age"`
	School         string   `json:"school"`
	Phone          string   `json:"phone"`
	City           string   `json:"city"`
	AgentLevel     int      `json:"agent_level"`
	ReferralCode   string   `json:"referral_code"`
	ReferredBy     string   `json:"referred_by"`
	CollectedCards []string `json:"collected_cards"`
	IsAgent        bool     `json:"is_agent"`
	AgentType      string   `json:"agent_type,omitempty"`
	ManagedSchools []string `json:"managed_schools,omitempty"`
	ManagedRegions []string `json:"managed_regions,omitempty"`
	Progress       Progress `json:"progress"`
	AddressCount   int      `json:"address_count"`
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

// Progress 学习进度结构体
type Progress struct {
	CurrentUnit     string   `bson:"current_unit" json:"current_unit"`
	CurrentSentence string   `bson:"current_sentence" json:"current_sentence"`
	LearnedWords    []string `bson:"learned_words" json:"learned_words"`
}

// Address 地址结构体
type Address struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	RecipientName string             `bson:"recipient_name" json:"recipient_name"`
	Phone         string             `bson:"phone" json:"phone"`
	Province      string             `bson:"province" json:"province"`
	City          string             `bson:"city" json:"city"`
	District      string             `bson:"district" json:"district"`
	Street        string             `bson:"street" json:"street"`
	PostalCode    string             `bson:"postal_code" json:"postal_code"`
	IsDefault     bool               `bson:"is_default" json:"is_default"`
}

// User 用户结构体
type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	OpenID         string             `bson:"openID" json:"openID"`
	UserName       string             `bson:"user_name" json:"user_name"`
	UserPassword   string             `bson:"user_password" json:"-"` // 不在JSON中显示密码
	Class          string             `bson:"class" json:"class"`
	Age            int                `bson:"age" json:"age"`
	School         string             `bson:"school" json:"school"`
	Phone          string             `bson:"phone" json:"phone"`
	City           string             `bson:"city" json:"city"`
	AgentLevel     int                `bson:"agent_level" json:"agent_level"`
	ReferralCode   string             `bson:"referral_code" json:"referral_code"`
	ReferredBy     string             `bson:"referred_by" json:"referred_by"`
	CollectedCards []string           `bson:"collected_cards" json:"collected_cards"`
	Addresses      []Address          `bson:"addresses" json:"addresses"`
	Progress       Progress           `bson:"progress" json:"progress"`
	IsAgent        bool               `bson:"is_agent" json:"is_agent"`
	AgentType      string             `bson:"agent_type" json:"agent_type"`
	ManagedSchools []string           `bson:"managed_schools" json:"managed_schools"`
	ManagedRegions []string           `bson:"managed_regions" json:"managed_regions"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

// ===== DTO转换函数 =====

// ToUserBasicResponse 将User转换为基础响应
func (u *User) ToUserBasicResponse() UserBasicResponse {
	return UserBasicResponse{
		ID:           u.ID.Hex(),
		UserName:     u.UserName,
		Age:          u.Age,
		School:       u.School,
		Phone:        u.Phone,
		Class:        u.Class,
		City:         u.City,
		ReferralCode: u.ReferralCode,
	}
}

// ToUserProfileResponse 将User转换为完整档案响应
func (u *User) ToUserProfileResponse() UserProfileResponse {
	return UserProfileResponse{
		ID:             u.ID.Hex(),
		OpenID:         u.OpenID,
		UserName:       u.UserName,
		Class:          u.Class,
		Age:            u.Age,
		School:         u.School,
		Phone:          u.Phone,
		City:           u.City,
		AgentLevel:     u.AgentLevel,
		ReferralCode:   u.ReferralCode,
		ReferredBy:     u.ReferredBy,
		CollectedCards: u.CollectedCards,
		IsAgent:        u.IsAgent,
		AgentType:      u.AgentType,
		ManagedSchools: u.ManagedSchools,
		ManagedRegions: u.ManagedRegions,
		Progress:       u.Progress,
		AddressCount:   len(u.Addresses),
	}
}

// ===== 数据库查询函数 =====

// GetUserByID 根据用户ID获取用户信息
func GetUserByID(userID string) (*User, error) {
	collection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 尝试将字符串转换为ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID格式: %v", err)
	}

	var user User
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户信息
func GetUserByUsername(username string) (*User, error) {
	collection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user User
	err := collection.FindOne(ctx, bson.M{"user_name": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByPhone 根据手机号获取用户信息
func GetUserByPhone(phone string) (*User, error) {
	collection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var user User
	err := collection.FindOne(ctx, bson.M{"phone": phone}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByOpenID 根据openid获取用户信息
func GetUserByOpenID(openID string) (*User, error) {
	collection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var user User
	err := collection.FindOne(ctx, bson.M{"openID": openID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser 创建新用户
func CreateUser(user *User) error {
	collection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

// CreateSimpleUser 创建简单用户（仅需要openid）
func CreateSimpleUser(user *User) error {
	collection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 为简单用户生成推荐码
	referralCode, err := GenerateReferralCode()
	if err != nil {
		return err
	}

	user.ReferralCode = referralCode
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// 初始化基本字段
	user.CollectedCards = make([]string, 0)
	user.Addresses = make([]Address, 0)
	user.ManagedSchools = make([]string, 0)
	user.ManagedRegions = make([]string, 0)

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

// CreateUserHandler 创建用户处理器
func CreateUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
				"error":   err.Error(),
			})
			return
		}

		// 1. 验证用户名是否已存在
		existingUser, err := GetUserByUsername(req.UserName)
		if err != nil && err != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询用户失败",
				"error":   err.Error(),
			})
			return
		}
		if existingUser != nil {
			c.JSON(http.StatusConflict, gin.H{
				"code":    409,
				"message": "用户名已存在",
			})
			return
		}

		// 验证手机号是否已存在
		existingPhone, err := GetUserByPhone(req.Phone)
		if err != nil && err != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询手机号失败",
				"error":   err.Error(),
			})
			return
		}
		if existingPhone != nil {
			c.JSON(http.StatusConflict, gin.H{
				"code":    409,
				"message": "手机号已被注册",
			})
			return
		}

		// 2. 验证推荐码（如果提供）
		if req.ReferredBy != "" {
			valid, err := ValidateReferralCode(req.ReferredBy)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "验证推荐码失败",
					"error":   err.Error(),
				})
				return
			}
			if !valid {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "推荐码不存在",
				})
				return
			}
		}

		// 3. 加密密码
		hashedPassword, err := HashPassword(req.UserPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "密码加密失败",
				"error":   err.Error(),
			})
			return
		}

		// 4. 生成推荐码
		referralCode, err := GenerateReferralCode()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "生成推荐码失败",
				"error":   err.Error(),
			})
			return
		}

		// 5. 创建用户对象
		user := &User{
			OpenID:         req.OpenID,
			UserName:       req.UserName,
			UserPassword:   hashedPassword,
			Class:          req.Class,
			Age:            req.Age,
			School:         req.School,
			Phone:          req.Phone,
			City:           req.City,
			AgentLevel:     0,
			ReferralCode:   referralCode,
			ReferredBy:     req.ReferredBy,
			CollectedCards: []string{},
			Addresses:      []Address{},
			Progress:       Progress{},
			IsAgent:        false,
			AgentType:      "",
			ManagedSchools: []string{},
			ManagedRegions: []string{},
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// 6. 保存用户信息
		if err := CreateUser(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "创建用户失败",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"code":    201,
			"message": "用户创建成功",
			"data":    user.ToUserBasicResponse(),
		})
	}
}

// UpdateUserHandler 更新用户信息处理器
func UpdateUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		var req UpdateUserRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
				"error":   err.Error(),
			})
			return
		}

		// 1. 验证用户是否存在
		_, err := GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在",
				"error":   err.Error(),
			})
			return
		}

		// 2. 验证权限（从JWT token中获取当前用户ID）
		tokenUserID, exists := c.Get("user_id")
		if !exists || tokenUserID.(string) != userID {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足，只能更新自己的信息",
			})
			return
		}

		// 3. 更新用户信息
		collection := GetCollection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		update := bson.M{
			"$set": bson.M{
				"user_name":  req.UserName,
				"class":      req.Class,
				"age":        req.Age,
				"school":     req.School,
				"phone":      req.Phone,
				"city":       req.City,
				"updated_at": time.Now(),
			},
		}

		objectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的用户ID格式",
				"error":   err.Error(),
			})
			return
		}

		filter := bson.M{"_id": objectID}
		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新用户信息失败",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "用户信息更新成功",
			"data": gin.H{
				"_id":       userID,
				"user_name": req.UserName,
				"class":     req.Class,
				"age":       req.Age,
				"school":    req.School,
				"phone":     req.Phone,
				"city":      req.City,
			},
		})
	}
}

// CreateAddressHandler 创建收货地址处理器
func CreateAddressHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		var req AddressRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
				"error":   err.Error(),
			})
			return
		}

		// TODO: 实现地址创建逻辑
		// 1. 验证用户是否存在
		// 2. 如果设为默认地址，取消其他默认地址
		// 3. 保存地址信息

		c.JSON(http.StatusCreated, gin.H{
			"code":    201,
			"message": "地址创建成功",
			"data": gin.H{
				"_id":            "507f1f77bcf86cd799439013",
				"user_id":        userID,
				"recipient_name": req.RecipientName,
				"phone":          req.Phone,
				"province":       req.Province,
				"city":           req.City,
				"district":       req.District,
				"street":         req.Street,
				"postal_code":    req.PostalCode,
				"is_default":     req.IsDefault,
			},
		})
	}
}

// DeleteAddressHandler 删除收货地址处理器
func DeleteAddressHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		addressID := c.Param("address_id")

		// TODO: 实现地址删除逻辑
		// 1. 验证用户是否存在
		// 2. 验证地址是否属于该用户
		// 3. 删除地址

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "地址删除成功",
		})

		// 避免未使用变量的警告
		_ = userID
		_ = addressID
	}
}

// UpdateAgentHandler 更新代理信息处理器
func UpdateAgentHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		var req UpdateAgentRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
				"error":   err.Error(),
			})
			return
		}

		// 验证用户是否存在
		_, err := GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在",
				"error":   err.Error(),
			})
			return
		}

		// 只更新代理相关字段
		collection := GetCollection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的用户ID格式",
				"error":   err.Error(),
			})
			return
		}

		update := bson.M{
			"$set": bson.M{
				"agent_level":     req.AgentLevel,
				"agent_type":      req.AgentType,
				"managed_schools": req.ManagedSchools,
				"managed_regions": req.ManagedRegions,
				"is_agent":        req.AgentLevel > 0,
				"updated_at":      time.Now(),
			},
		}

		filter := bson.M{"_id": objectID}
		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新代理信息失败",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "代理信息更新成功",
			"data": gin.H{
				"_id":             userID,
				"agent_level":     req.AgentLevel,
				"agent_type":      req.AgentType,
				"managed_schools": req.ManagedSchools,
				"managed_regions": req.ManagedRegions,
				"is_agent":        req.AgentLevel > 0,
			},
		})
	}
}
