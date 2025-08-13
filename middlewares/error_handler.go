package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// HandleError 处理错误并决定是否退出程序
// 如果 err 不为 nil，打印错误信息并根据 shouldExit 决定是否退出程序
// message 参数是描述错误上下文的信息
// 如果不需要退出程序，则返回 true 表示发生了错误
func HandleError(err error, message string, shouldExit bool) bool {
	if err != nil {
		fmt.Printf("%s: %v\n", message, err)
		if shouldExit {
			os.Exit(1)
		}
		return true
	}
	return false
}

// MustSucceed 处理错误并在出错时退出程序
// 如果 err 不为 nil，打印错误信息并退出程序
// 适用于程序无法继续执行的关键错误
func MustSucceed(err error, message string) {
	if err != nil {
		fmt.Printf("%s: %v\n", message, err)
		os.Exit(1)
	}
}

// ErrorHandlerMiddleware Gin错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "服务器内部错误",
					"error":   fmt.Sprintf("%v", err),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
