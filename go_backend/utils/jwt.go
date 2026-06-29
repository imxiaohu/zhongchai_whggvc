package utils

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/xiaohu/pingjiao/config"
)

// GenerateToken 生成JWT令牌
func GenerateToken(userId, username string, userType string) (string, error) {
	// 创建JWT声明
	claims := jwt.MapClaims{
		"userId":   userId,
		"username": username,
		"userType": userType,
		"exp":      time.Now().Add(time.Second * time.Duration(config.GetJWTExpiration())).Unix(),
		"iat":      time.Now().Unix(),
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(config.GetJWTSecret()))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证JWT令牌
func ValidateToken(tokenString string) (*jwt.Token, error) {
	// 解析令牌
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetJWTSecret()), nil
	})

	return token, err
}

// GetUserIDFromContext 从Gin上下文中获取用户ID并转换为uint
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userIDInterface, exists := c.Get("userId")
	if !exists {
		return 0, false
	}

	// JWT中的userId是字符串类型
	userIDStr, ok := userIDInterface.(string)
	if !ok {
		return 0, false
	}

	// 转换为uint
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, false
	}

	return uint(userID), true
}

// GetCertificateSerialNumber 获取证书序列号
func GetCertificateSerialNumber(certData []byte) (string, error) {
	// 解析PEM格式的证书
	block, _ := pem.Decode(certData)
	if block == nil {
		return "", fmt.Errorf("无法解析PEM格式的证书")
	}

	// 解析X.509证书
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("解析X.509证书失败: %w", err)
	}

	// 返回序列号的十六进制字符串
	return fmt.Sprintf("%X", cert.SerialNumber), nil
}
