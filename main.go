package main

import (
	"log"
	"miniprogram/config"
	"miniprogram/controllers"
	"miniprogram/middlewares"
	"miniprogram/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载.env文件
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: 未找到.env文件，将使用默认配置或系统环境变量")
	}

	// 获取配置
	cfg := config.GetConfig()

	// 初始化MongoDB连接
	controllers.InitMongoDB()
	defer controllers.CloseMongoDB()

	// 设置middleware中的数据库访问函数
	middlewares.SetGetCollectionFunc(controllers.GetCollection)

	// 创建Gin路由器
	r := gin.Default()

	// 添加全局中间件
	r.Use(middlewares.ErrorHandlerMiddleware())
	r.Use(middlewares.CORSMiddleware())

	// 设置路由
	routes.SetupRoutes(r)

	// 启动服务器
	log.Printf("服务器启动在端口 %s", cfg.ServerPort)
	log.Printf("基础API地址: %s", cfg.BaseAPIURL)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}
