package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"miniprogram/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 统一的API响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// 成功响应
func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

// 创建成功响应
func CreatedResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Code:    201,
		Message: message,
		Data:    data,
	})
}

// 错误响应
func ErrorResponse(c *gin.Context, httpStatus int, code int, message string, err error) {
	response := APIResponse{
		Code:    code,
		Message: message,
	}

	if err != nil {
		response.Error = err.Error()
	}

	c.JSON(httpStatus, response)
}

// 常见错误响应快捷方法
func BadRequestResponse(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusBadRequest, 400, message, err)
}

func NotFoundResponse(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusNotFound, 404, message, err)
}

func InternalServerErrorResponse(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusInternalServerError, 500, message, err)
}

func UnauthorizedResponse(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusUnauthorized, 401, message, err)
}

// 统一的数据库上下文创建
func CreateDBContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

// 数据库连接管理
var mongoClient *mongo.Client

// InitMongoDB 初始化MongoDB连接
func InitMongoDB() {
	// 创建一个带有超时的上下文
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 从配置文件获取MongoDB连接字符串
	cfg := config.GetConfig()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDBURL))
	if err != nil {
		log.Fatal(err)
	}

	// 检查连接
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	mongoClient = client
	log.Println("成功连接到MongoDB")
}

// GetCollection 获取指定名称的集合
func GetCollection(collectionName string) *mongo.Collection {
	if mongoClient == nil {
		log.Fatal("MongoDB客户端未初始化")
	}
	return mongoClient.Database("miniprogram_db").Collection(collectionName)
}

// CloseMongoDB 关闭MongoDB连接
func CloseMongoDB() {
	if mongoClient != nil {
		ctx, cancel := CreateDBContext()
		defer cancel()

		err := mongoClient.Disconnect(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("MongoDB连接已关闭")
	}
}

// HealthCheckHandler 健康检查处理器
func HealthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		SuccessResponse(c, "服务运行正常", gin.H{
			"status": "ok",
		})
	}
}

// ===== 公共工具函数 =====

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CalculateCommissionRate 根据代理等级计算佣金率
func CalculateCommissionRate(agentLevel int) float64 {
	switch agentLevel {
	case 0: // 普通用户
		return 0.01 // 1%佣金
	case 1: // 校代理
		return 0.03 // 3%佣金
	case 2: // 区域代理
		return 0.05 // 5%佣金
	default:
		return 0.01 // 默认1%佣金
	}
}

// GenerateWithdrawID 生成提现ID
func GenerateWithdrawID() string {
	return "WD" + time.Now().Format("20060102150405") + GenerateRandomString(4)
}

// GenerateCommissionID 生成佣金ID
func GenerateCommissionID() string {
	return "COMM" + time.Now().Format("20060102150405") + GenerateRandomString(4)
}
