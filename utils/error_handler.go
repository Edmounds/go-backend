package utils

import (
	"fmt"
	"os"
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

// HandleErrorWithResult 处理错误并返回默认结果
// 如果 err 不为 nil，打印错误信息并返回提供的默认结果
// 适用于需要返回值的函数中进行错误处理
func HandleErrorWithResult[T any](err error, message string, defaultResult T) (T, bool) {
	if err != nil {
		fmt.Printf("%s: %v\n", message, err)
		return defaultResult, true
	}
	return defaultResult, false
}
