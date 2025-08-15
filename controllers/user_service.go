package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"miniprogram/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// ===== 用户服务层 =====

// UserService 用户服务
type UserService struct{}

// GetUserService 获取用户服务实例
func GetUserService() *UserService {
	return &UserService{}
}

// FindUserByOpenID 根据openid查找用户
func (s *UserService) FindUserByOpenID(openID string) (*models.User, error) {
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"openID": openID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	// 确保数组字段初始化
	s.initializeUserArrays(&user)
	return &user, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(user *models.User) error {
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 生成推荐码
	referralCode, err := s.GenerateReferralCode()
	if err != nil {
		return err
	}
	user.ReferralCode = referralCode

	// 设置创建时间
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// 确保数组字段初始化
	s.initializeUserArrays(user)

	// 如果设置了密码，进行加密
	if user.UserPassword != "" {
		hashedPassword, err := s.HashPassword(user.UserPassword)
		if err != nil {
			return err
		}
		user.UserPassword = hashedPassword
	}

	_, err = collection.InsertOne(ctx, user)
	return err
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(openID string, updates map[string]interface{}) (*models.User, error) {
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 添加更新时间
	updates["updated_at"] = time.Now()

	// 如果更新密码，进行加密
	if password, ok := updates["user_password"].(string); ok && password != "" {
		hashedPassword, err := s.HashPassword(password)
		if err != nil {
			return nil, err
		}
		updates["user_password"] = hashedPassword
	}

	filter := bson.M{"openID": openID}
	update := bson.M{"$set": updates}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	// 返回更新后的用户信息
	return s.FindUserByOpenID(openID)
}

// initializeUserArrays 确保用户的数组字段都初始化为空切片
func (s *UserService) initializeUserArrays(user *models.User) {
	if user.CollectedCards == nil {
		user.CollectedCards = []string{}
	}
	if user.Addresses == nil {
		user.Addresses = []models.Address{}
	}
	if user.ManagedSchools == nil {
		user.ManagedSchools = []string{}
	}
	if user.ManagedRegions == nil {
		user.ManagedRegions = []string{}
	}
	if user.Progress.LearnedWords == nil {
		user.Progress.LearnedWords = []string{}
	}
}

// ValidateReferralCodeExists 验证推荐码是否已存在
func (s *UserService) ValidateReferralCodeExists(code string) (bool, error) {
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	count, err := collection.CountDocuments(ctx, bson.M{"referral_code": code})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GenerateReferralCode 生成推荐码
func (s *UserService) GenerateReferralCode() (string, error) {
	bytes := make([]byte, 4) // 8位十六进制字符
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashPassword 加密密码
func (s *UserService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPassword 验证密码
func (s *UserService) CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// CreateOrUpdateUserProfile 创建或更新用户资料
func (s *UserService) CreateOrUpdateUserProfile(req models.CreateUserRequest) (*models.User, bool, error) {
	// 检查用户是否已存在
	existingUser, err := s.FindUserByOpenID(req.OpenID)
	isNewUser := false

	if err != nil && err != mongo.ErrNoDocuments {
		return nil, false, err
	}

	if existingUser == nil {
		// 创建新用户
		newUser := &models.User{
			OpenID:       req.OpenID,
			UserName:     req.UserName,
			UserPassword: req.UserPassword,
			Class:        req.Class,
			Age:          req.Age,
			School:       req.School,
			Phone:        req.Phone,
			City:         req.City,
			ReferredBy:   req.ReferredBy,
		}

		err = s.CreateUser(newUser)
		if err != nil {
			return nil, false, err
		}

		isNewUser = true
		return newUser, isNewUser, nil
	} else {
		// 更新现有用户
		updates := make(map[string]interface{})

		if req.UserName != "" {
			updates["user_name"] = req.UserName
		}
		if req.UserPassword != "" {
			updates["user_password"] = req.UserPassword
		}
		if req.Class != "" {
			updates["class"] = req.Class
		}
		if req.Age > 0 {
			updates["age"] = req.Age
		}
		if req.School != "" {
			updates["school"] = req.School
		}
		if req.Phone != "" {
			updates["phone"] = req.Phone
		}
		if req.City != "" {
			updates["city"] = req.City
		}

		if len(updates) > 0 {
			updatedUser, err := s.UpdateUser(req.OpenID, updates)
			if err != nil {
				return nil, false, err
			}
			return updatedUser, isNewUser, nil
		}

		return existingUser, isNewUser, nil
	}
}

// AddressService 地址服务
type AddressService struct{}

// GetAddressService 获取地址服务实例
func GetAddressService() *AddressService {
	return &AddressService{}
}

// CreateAddress 创建用户地址
func (s *AddressService) CreateAddress(openID string, req models.AddressRequest) (*models.Address, error) {
	userService := GetUserService()
	user, err := userService.FindUserByOpenID(openID)
	if err != nil {
		return nil, err
	}

	// 创建新地址
	newAddress := models.Address{
		ID:            primitive.NewObjectID(),
		UserOpenID:    openID,
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

	// 如果设置为默认地址，先取消其他默认地址
	if req.IsDefault {
		err = s.ClearDefaultAddresses(user)
		if err != nil {
			return nil, err
		}
	}

	// 添加地址到用户记录
	user.Addresses = append(user.Addresses, newAddress)

	// 更新用户记录
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	filter := bson.M{"openID": openID}
	update := bson.M{
		"$set": bson.M{
			"addresses":  user.Addresses,
			"updated_at": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return &newAddress, nil
}

// GetUserAddresses 获取用户地址列表
func (s *AddressService) GetUserAddresses(openID string) ([]models.Address, error) {
	userService := GetUserService()
	user, err := userService.FindUserByOpenID(openID)
	if err != nil {
		return nil, err
	}

	return user.Addresses, nil
}

// UpdateAddress 更新地址
func (s *AddressService) UpdateAddress(openID string, addressID primitive.ObjectID, req models.AddressRequest) (*models.Address, error) {
	userService := GetUserService()
	user, err := userService.FindUserByOpenID(openID)
	if err != nil {
		return nil, err
	}

	// 查找要更新的地址
	var updatedAddress *models.Address
	for i := range user.Addresses {
		if user.Addresses[i].ID == addressID {
			// 更新地址信息
			user.Addresses[i].RecipientName = req.RecipientName
			user.Addresses[i].Phone = req.Phone
			user.Addresses[i].Province = req.Province
			user.Addresses[i].City = req.City
			user.Addresses[i].District = req.District
			user.Addresses[i].Street = req.Street
			user.Addresses[i].PostalCode = req.PostalCode
			user.Addresses[i].UpdatedAt = time.Now()

			// 处理默认地址设置
			if req.IsDefault && !user.Addresses[i].IsDefault {
				err = s.ClearDefaultAddresses(user)
				if err != nil {
					return nil, err
				}
				user.Addresses[i].IsDefault = true
			} else if !req.IsDefault {
				user.Addresses[i].IsDefault = false
			}

			updatedAddress = &user.Addresses[i]
			break
		}
	}

	if updatedAddress == nil {
		return nil, mongo.ErrNoDocuments
	}

	// 更新数据库
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	filter := bson.M{"openID": openID}
	update := bson.M{
		"$set": bson.M{
			"addresses":  user.Addresses,
			"updated_at": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return updatedAddress, nil
}

// DeleteAddress 删除地址
func (s *AddressService) DeleteAddress(openID string, addressID primitive.ObjectID) error {
	userService := GetUserService()
	user, err := userService.FindUserByOpenID(openID)
	if err != nil {
		return err
	}

	// 查找并删除地址
	found := false
	for i, address := range user.Addresses {
		if address.ID == addressID {
			user.Addresses = append(user.Addresses[:i], user.Addresses[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return mongo.ErrNoDocuments
	}

	// 更新数据库
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	filter := bson.M{"openID": openID}
	update := bson.M{
		"$set": bson.M{
			"addresses":  user.Addresses,
			"updated_at": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	return err
}

// ClearDefaultAddresses 清除用户的所有默认地址
func (s *AddressService) ClearDefaultAddresses(user *models.User) error {
	for i := range user.Addresses {
		user.Addresses[i].IsDefault = false
	}
	return nil
}

// SetDefaultAddress 设置默认地址
func (s *AddressService) SetDefaultAddress(openID string, defaultAddressID primitive.ObjectID) error {
	userService := GetUserService()
	user, err := userService.FindUserByOpenID(openID)
	if err != nil {
		return err
	}

	// 先清除所有默认地址
	err = s.ClearDefaultAddresses(user)
	if err != nil {
		return err
	}

	// 设置新的默认地址
	found := false
	for i := range user.Addresses {
		if user.Addresses[i].ID == defaultAddressID {
			user.Addresses[i].IsDefault = true
			found = true
			break
		}
	}

	if !found {
		return mongo.ErrNoDocuments
	}

	// 更新数据库
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	filter := bson.M{"openID": openID}
	update := bson.M{
		"$set": bson.M{
			"addresses":  user.Addresses,
			"updated_at": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	return err
}

// ===== 向后兼容函数 =====

// GetUserByOpenID 根据openid获取用户信息 (向后兼容)
func GetUserByOpenID(openID string) (*models.User, error) {
	service := GetUserService()
	return service.FindUserByOpenID(openID)
}

// CreateUser 创建用户 (向后兼容)
func CreateUser(user *models.User) error {
	service := GetUserService()
	return service.CreateUser(user)
}

// GenerateReferralCode 生成推荐码 (向后兼容)
func GenerateReferralCode() (string, error) {
	service := GetUserService()
	return service.GenerateReferralCode()
}

// HashPassword 加密密码 (向后兼容)
func HashPassword(password string) (string, error) {
	service := GetUserService()
	return service.HashPassword(password)
}

// CheckPassword 验证密码 (向后兼容)
func CheckPassword(hashedPassword, password string) bool {
	service := GetUserService()
	return service.CheckPassword(hashedPassword, password)
}

// SetDefaultAddress 设置默认地址 (向后兼容)
func SetDefaultAddress(openID string, defaultAddressID primitive.ObjectID) error {
	service := GetAddressService()
	return service.SetDefaultAddress(openID, defaultAddressID)
}
