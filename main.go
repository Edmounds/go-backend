package main

import (
	"fmt"
	"miniprogram/controllers"
	"miniprogram/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
	// "miniprogram/database"
)

// 处理Token相关操作，简化错误处理逻辑
func handleTokenOperations(user controllers.User) {
	// 生成token
	token, err := controllers.GenerateToken(user)
	if utils.HandleError(err, "生成Token失败", true) {
		return
	}
	fmt.Println("生成的Token:", token)

	// 验证token
	tokenClaims, err := controllers.ValidateToken(token)
	if utils.HandleError(err, "验证Token失败", true) {
		return
	}
	fmt.Println("Token信息:", tokenClaims)

	// 刷新token
	token, err = controllers.RefreshToken(token)
	if utils.HandleError(err, "刷新Token失败", true) {
		return
	}
	fmt.Println("刷新后的Token:", token)
}

func main() {
	// database.CreateBook()
	// database.CreateWord()
	// database.CreateUser()
	controllers.InitMongoDB()
	defer controllers.CloseMongoDB()

	user, err := controllers.GetUser("cqc")
	if utils.HandleError(err, "获取用户信息失败", true) {
		return
	}

	fmt.Println("用户信息:", user)
	fmt.Println("用户年龄:", user["age"])

	user_ := controllers.User{
		UserName:     user["user_name"].(string),
		UserId:       user["_id"].(primitive.ObjectID).Hex(), // 使用 Hex() 方法转换为字符串
		UserPassword: user["user_password"].(string),
	}

	// 处理Token相关操作
	handleTokenOperations(user_)

	// 打印一条消息，表示程序已成功运行
	fmt.Println("程序已成功运行！")
}
