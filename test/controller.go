package test

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateReferralCode() (string, error) {
	bytes := make([]byte, 4) // 8位十六进制字符
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
