package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"miniprogram/middlewares"
	"miniprogram/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

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
		var req models.CreateUserRequest
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

		// 2. 如果用户已存在，检查是否为信息完善请求
		if existingUser != nil {
			// 检查是否尝试修改已有推荐码
			if existingUser.ReferredBy != "" && req.ReferredBy != "" && existingUser.ReferredBy != req.ReferredBy {
				BadRequestResponse(c, "用户已有推荐码，不允许修改", nil)
				return
			}

			// 用户已存在且没有尝试修改推荐码，允许更新其他信息
			updateData := make(map[string]interface{})
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
			if req.UserPassword != "" {
				hashedPassword, err := HashPassword(req.UserPassword)
				if err != nil {
					InternalServerErrorResponse(c, "密码加密失败", err)
					return
				}
				updateData["user_password"] = hashedPassword
			}
			// 如果用户没有推荐码且提供了推荐码，则设置推荐码
			if existingUser.ReferredBy == "" && req.ReferredBy != "" {
				updateData["referred_by"] = req.ReferredBy
			}

			if len(updateData) > 0 {
				updateData["updated_at"] = time.Now()
				collection := GetCollection("users")
				ctx, cancel := CreateDBContext()
				defer cancel()

				filter := bson.M{"openID": req.OpenID}
				update := bson.M{"$set": updateData}
				_, err = collection.UpdateOne(ctx, filter, update)
				if err != nil {
					InternalServerErrorResponse(c, "更新用户信息失败", err)
					return
				}

				// 如果设置了推荐码，处理推荐关系
				if existingUser.ReferredBy == "" && req.ReferredBy != "" {
					err := ProcessNewUserReferral(existingUser.OpenID, req.ReferredBy)
					if err != nil {
						// 记录错误但不影响用户信息更新
						middlewares.HandleError(err, "处理推荐关系失败", false)
					}
				}
			}

			// 重新获取更新后的用户信息
			updatedUser, err := GetUserByOpenID(req.OpenID)
			if err != nil {
				InternalServerErrorResponse(c, "获取更新后的用户信息失败", err)
				return
			}

			SuccessResponse(c, "用户信息更新成功", updatedUser)
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
			OpenID:         req.OpenID,
			UserName:       req.UserName,
			Class:          req.Class,
			Age:            req.Age,
			School:         req.School,
			Phone:          req.Phone,
			City:           req.City,
			ReferredBy:     req.ReferredBy,
			CollectedCards: []string{},         // 初始化收藏单词卡为空数组
			Addresses:      []models.Address{}, // 初始化地址为空数组
			Progress: models.Progress{
				CurrentUnit:  "",
				LearnedWords: []string{}, // 初始化已学习单词为空数组
			},
			ManagedSchools: []string{}, // 初始化管理学校为空数组
			ManagedRegions: []string{}, // 初始化管理区域为空数组
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
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

// ===== 地址管理处理函数 =====

// CreateAddressHandler 创建用户地址处理器
func CreateAddressHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数（这里的 user_id 实际上是微信的 openID）
		openID := c.Param("user_id")

		// 解析请求体
		var req models.AddressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 检查用户是否存在
		_, err := GetUserByOpenID(openID)
		if err != nil {
			NotFoundResponse(c, "用户不存在", err)
			return
		}

		// 创建地址对象
		address := models.Address{
			ID:            primitive.NewObjectID(),
			RecipientName: req.RecipientName,
			Phone:         req.Phone,
			Province:      req.Province,
			City:          req.City,
			District:      req.District,
			Street:        req.Street,
			PostalCode:    req.PostalCode,
			IsDefault:     req.IsDefault,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// 暂时注释掉默认地址逻辑，先让基本功能工作
		// if req.IsDefault {
		// 	err := SetDefaultAddress(openID, address.ID)
		// 	if err != nil {
		// 		InternalServerErrorResponse(c, "设置默认地址失败", err)
		// 		return
		// 	}
		// }

		// 将地址添加到用户的地址列表中
		collection := GetCollection("users")
		ctx, cancel := CreateDBContext()
		defer cancel()

		// 直接添加地址到用户的地址数组（新用户都有正确的空数组初始化）
		filter := bson.M{"openID": openID}
		update := bson.M{
			"$push": bson.M{"addresses": address},
			"$set":  bson.M{"updated_at": time.Now()},
		}

		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			InternalServerErrorResponse(c, "添加地址失败", err)
			return
		}

		CreatedResponse(c, "地址创建成功", address)
	}
}

// GetUserAddressesHandler 获取用户地址列表处理器
func GetUserAddressesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数（这里的 user_id 实际上是微信的 openID）
		openID := c.Param("user_id")

		// 获取用户信息
		user, err := GetUserByOpenID(openID)
		if err != nil {
			NotFoundResponse(c, "用户不存在", err)
			return
		}

		SuccessResponse(c, "获取地址列表成功", user.Addresses)
	}
}

// UpdateAddressHandler 更新用户地址处理器
func UpdateAddressHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取参数
		openID := c.Param("user_id")
		addressIDStr := c.Param("address_id")

		// 转换地址ID
		addressID, err := primitive.ObjectIDFromHex(addressIDStr)
		if err != nil {
			BadRequestResponse(c, "地址ID格式错误", err)
			return
		}

		// 解析请求体
		var req models.AddressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 检查用户是否存在
		_, err = GetUserByOpenID(openID)
		if err != nil {
			NotFoundResponse(c, "用户不存在", err)
			return
		}

		// 如果设置为默认地址，需要先将其他地址设为非默认
		if req.IsDefault {
			err := SetDefaultAddress(openID, addressID)
			if err != nil {
				InternalServerErrorResponse(c, "设置默认地址失败", err)
				return
			}
		}

		// 更新地址信息
		collection := GetCollection("users")
		ctx, cancel := CreateDBContext()
		defer cancel()

		filter := bson.M{
			"openID":        openID,
			"addresses._id": addressID,
		}
		update := bson.M{
			"$set": bson.M{
				"addresses.$.recipient_name": req.RecipientName,
				"addresses.$.phone":          req.Phone,
				"addresses.$.province":       req.Province,
				"addresses.$.city":           req.City,
				"addresses.$.district":       req.District,
				"addresses.$.street":         req.Street,
				"addresses.$.postal_code":    req.PostalCode,
				"addresses.$.is_default":     req.IsDefault,
				"addresses.$.updated_at":     time.Now(),
				"updated_at":                 time.Now(),
			},
		}

		result, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			InternalServerErrorResponse(c, "更新地址失败", err)
			return
		}

		if result.MatchedCount == 0 {
			NotFoundResponse(c, "地址不存在", nil)
			return
		}

		SuccessResponse(c, "地址更新成功", nil)
	}
}

// DeleteAddressHandler 删除用户地址处理器
func DeleteAddressHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取参数
		openID := c.Param("user_id")
		addressIDStr := c.Param("address_id")

		// 转换地址ID
		addressID, err := primitive.ObjectIDFromHex(addressIDStr)
		if err != nil {
			BadRequestResponse(c, "地址ID格式错误", err)
			return
		}

		// 检查用户是否存在
		_, err = GetUserByOpenID(openID)
		if err != nil {
			NotFoundResponse(c, "用户不存在", err)
			return
		}

		// 删除地址
		collection := GetCollection("users")
		ctx, cancel := CreateDBContext()
		defer cancel()

		filter := bson.M{"openID": openID}
		update := bson.M{
			"$pull": bson.M{"addresses": bson.M{"_id": addressID}},
			"$set":  bson.M{"updated_at": time.Now()},
		}

		result, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			InternalServerErrorResponse(c, "删除地址失败", err)
			return
		}

		if result.MatchedCount == 0 {
			NotFoundResponse(c, "用户不存在", nil)
			return
		}

		SuccessResponse(c, "地址删除成功", nil)
	}
}

// ===== 地址管理辅助函数 =====

// SetDefaultAddress 设置默认地址（将其他地址设为非默认）
func SetDefaultAddress(openID string, defaultAddressID primitive.ObjectID) error {
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 确保用户有addresses数组，如果没有则初始化
	filter := bson.M{"openID": openID, "addresses": bson.M{"$exists": false}}
	update := bson.M{
		"$set": bson.M{
			"addresses":  []models.Address{},
			"updated_at": time.Now(),
		},
	}
	collection.UpdateOne(ctx, filter, update)

	// 如果addresses字段为null，也进行初始化
	filter = bson.M{"openID": openID, "addresses": nil}
	update = bson.M{
		"$set": bson.M{
			"addresses":  []models.Address{},
			"updated_at": time.Now(),
		},
	}
	collection.UpdateOne(ctx, filter, update)

	// 先将所有地址设为非默认
	filter = bson.M{"openID": openID, "addresses": bson.M{"$ne": nil}}
	update = bson.M{
		"$set": bson.M{
			"addresses.$[].is_default": false,
			"updated_at":               time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// 再将指定地址设为默认
	filter = bson.M{
		"openID":        openID,
		"addresses._id": defaultAddressID,
	}
	update = bson.M{
		"$set": bson.M{
			"addresses.$.is_default": true,
			"updated_at":             time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	return err
}
