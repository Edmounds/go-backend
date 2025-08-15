package controllers

import (
	"miniprogram/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ===== HTTP 处理器 =====

// CreateUserHandler 创建或更新用户处理器
func CreateUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 初始化用户服务
		userService := GetUserService()

		// 创建或更新用户资料
		user, isNewUser, err := userService.CreateOrUpdateUserProfile(req)
		if err != nil {
			InternalServerErrorResponse(c, "用户操作失败", err)
			return
		}

		// 处理推荐关系（仅对新用户）
		if isNewUser && req.ReferredBy != "" {
			referralService := NewReferralRewardService()
			err := referralService.ProcessNewUserReferral(user.OpenID, req.ReferredBy)
			if err != nil {
				// 推荐关系处理失败不影响用户创建，记录日志即可
			}
		}

		// 构建响应
		responseMessage := "用户信息更新成功"
		if isNewUser {
			responseMessage = "用户创建成功"
		}

		SuccessResponse(c, responseMessage, gin.H{
			"user":        user,
			"is_new_user": isNewUser,
		})
	}
}

// GetUserHandler 获取用户信息处理器
func GetUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.Param("user_id")

		// 初始化用户服务
		userService := GetUserService()

		// 获取用户信息
		user, err := userService.FindUserByOpenID(openID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				NotFoundResponse(c, "用户不存在", err)
			} else {
				InternalServerErrorResponse(c, "获取用户信息失败", err)
			}
			return
		}

		SuccessResponse(c, "获取用户信息成功", user)
	}
}

// CreateAddressHandler 创建地址处理器
func CreateAddressHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.Param("user_id")

		var req models.AddressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 初始化地址服务
		addressService := GetAddressService()

		// 创建地址
		address, err := addressService.CreateAddress(openID, req)
		if err != nil {
			InternalServerErrorResponse(c, "创建地址失败", err)
			return
		}

		CreatedResponse(c, "地址创建成功", address)
	}
}

// GetUserAddressesHandler 获取用户地址列表处理器
func GetUserAddressesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.Param("user_id")

		// 初始化地址服务
		addressService := GetAddressService()

		// 获取地址列表
		addresses, err := addressService.GetUserAddresses(openID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				NotFoundResponse(c, "用户不存在", err)
			} else {
				InternalServerErrorResponse(c, "获取地址列表失败", err)
			}
			return
		}

		SuccessResponse(c, "获取地址列表成功", gin.H{
			"addresses": addresses,
			"total":     len(addresses),
		})
	}
}

// UpdateAddressHandler 更新地址处理器
func UpdateAddressHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.Param("user_id")
		addressIDStr := c.Param("address_id")

		// 解析地址ID
		addressID, err := primitive.ObjectIDFromHex(addressIDStr)
		if err != nil {
			BadRequestResponse(c, "地址ID格式错误", err)
			return
		}

		var req models.AddressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 初始化地址服务
		addressService := GetAddressService()

		// 更新地址
		updatedAddress, err := addressService.UpdateAddress(openID, addressID, req)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				NotFoundResponse(c, "地址不存在", err)
			} else {
				InternalServerErrorResponse(c, "更新地址失败", err)
			}
			return
		}

		SuccessResponse(c, "地址更新成功", updatedAddress)
	}
}

// DeleteAddressHandler 删除地址处理器
func DeleteAddressHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.Param("user_id")
		addressIDStr := c.Param("address_id")

		// 解析地址ID
		addressID, err := primitive.ObjectIDFromHex(addressIDStr)
		if err != nil {
			BadRequestResponse(c, "地址ID格式错误", err)
			return
		}

		// 初始化地址服务
		addressService := GetAddressService()

		// 删除地址
		err = addressService.DeleteAddress(openID, addressID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				NotFoundResponse(c, "地址不存在", err)
			} else {
				InternalServerErrorResponse(c, "删除地址失败", err)
			}
			return
		}

		SuccessResponse(c, "地址删除成功", nil)
	}
}

// SetDefaultAddressHandler 设置默认地址处理器
func SetDefaultAddressHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.Param("user_id")
		addressIDStr := c.Query("address_id")

		if addressIDStr == "" {
			BadRequestResponse(c, "缺少地址ID参数", nil)
			return
		}

		// 解析地址ID
		addressID, err := primitive.ObjectIDFromHex(addressIDStr)
		if err != nil {
			BadRequestResponse(c, "地址ID格式错误", err)
			return
		}

		// 初始化地址服务
		addressService := GetAddressService()

		// 设置默认地址
		err = addressService.SetDefaultAddress(openID, addressID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				NotFoundResponse(c, "地址不存在", err)
			} else {
				InternalServerErrorResponse(c, "设置默认地址失败", err)
			}
			return
		}

		SuccessResponse(c, "默认地址设置成功", nil)
	}
}
