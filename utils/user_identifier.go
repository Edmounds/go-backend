package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

// UserIdentifierManager 用户标识符管理器
type UserIdentifierManager struct {
	gcm cipher.AEAD
}

// NewUserIdentifierManager 创建用户标识符管理器实例
func NewUserIdentifierManager() (*UserIdentifierManager, error) {
	// 从环境变量获取密钥，如果没有则使用默认密钥（生产环境必须设置）
	secretKey := os.Getenv("USER_ID_SECRET_KEY")
	if secretKey == "" {
		// 默认密钥，生产环境必须更换
		secretKey = "miniprogram_default_secret_key_2024"
	}

	// 使用SHA256生成32字节密钥
	hash := sha256.Sum256([]byte(secretKey))
	key := hash[:]

	// 创建AES加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建AES加密器失败: %v", err)
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建GCM模式失败: %v", err)
	}

	return &UserIdentifierManager{gcm: gcm}, nil
}

// EncodeOpenID 将openID编码为安全的用户标识符
func (m *UserIdentifierManager) EncodeOpenID(openID string) (string, error) {
	if openID == "" {
		return "", errors.New("openID不能为空")
	}

	// 生成随机nonce
	nonce := make([]byte, m.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成nonce失败: %v", err)
	}

	// 加密openID
	ciphertext := m.gcm.Seal(nonce, nonce, []byte(openID), nil)

	// Base64编码
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	// 添加前缀标识
	return "uid_" + encoded, nil
}

// DecodeUserID 将用户标识符解码为openID
func (m *UserIdentifierManager) DecodeUserID(userID string) (string, error) {
	if userID == "" {
		return "", errors.New("userID不能为空")
	}

	// 检查前缀
	if len(userID) < 4 || userID[:4] != "uid_" {
		return "", errors.New("无效的用户标识符格式")
	}

	// 去掉前缀
	encoded := userID[4:]

	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %v", err)
	}

	// 检查密文长度
	nonceSize := m.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("密文长度不足")
	}

	// 提取nonce和密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := m.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %v", err)
	}

	return string(plaintext), nil
}

// GetSafeUserID 获取安全的用户标识符（用于API响应）
func (m *UserIdentifierManager) GetSafeUserID(openID string) string {
	if openID == "" {
		return ""
	}

	encoded, err := m.EncodeOpenID(openID)
	if err != nil {
		// 如果编码失败，返回一个基于openID的hash值作为fallback
		hash := sha256.Sum256([]byte(openID))
		return "uid_" + base64.StdEncoding.EncodeToString(hash[:16])
	}

	return encoded
}

// ValidateUserID 验证用户标识符格式是否正确
func (m *UserIdentifierManager) ValidateUserID(userID string) bool {
	if userID == "" {
		return false
	}

	if len(userID) < 4 || userID[:4] != "uid_" {
		return false
	}

	// 尝试解码验证
	_, err := m.DecodeUserID(userID)
	return err == nil
}

// 全局实例（单例模式）
var globalUserIDManager *UserIdentifierManager

// GetUserIDManager 获取全局用户标识符管理器实例
func GetUserIDManager() *UserIdentifierManager {
	if globalUserIDManager == nil {
		var err error
		globalUserIDManager, err = NewUserIdentifierManager()
		if err != nil {
			// 如果初始化失败，创建一个fallback实例
			panic(fmt.Sprintf("初始化用户标识符管理器失败: %v", err))
		}
	}
	return globalUserIDManager
}

// 便捷函数

// EncodeOpenIDToSafeID 将openID编码为安全ID（便捷函数）
func EncodeOpenIDToSafeID(openID string) string {
	return GetUserIDManager().GetSafeUserID(openID)
}

// DecodeSafeIDToOpenID 将安全ID解码为openID（便捷函数）
func DecodeSafeIDToOpenID(safeID string) (string, error) {
	return GetUserIDManager().DecodeUserID(safeID)
}

// ValidateSafeUserID 验证安全用户ID格式（便捷函数）
func ValidateSafeUserID(safeID string) bool {
	return GetUserIDManager().ValidateUserID(safeID)
}
