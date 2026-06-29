package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/utils"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		var tokenString string

		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 如果没有token，返回未授权错误
		if tokenString == "" {
			log.Println("JWT Auth: No token provided")
			c.JSON(http.StatusUnauthorized, utils.Response{
				Success: false,
				Message: config.MsgUnauthorized,
				Code:    config.CodeUnauthorized,
			})
			c.Abort()
			return
		}

		// 解析token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.GetJWTSecret()), nil
		})

		// 处理解析错误
		if err != nil {
			log.Printf("JWT Auth: Token parse error: %v", err)
			c.JSON(http.StatusUnauthorized, utils.Response{
				Success: false,
				Message: "无效的令牌: " + err.Error(),
				Code:    config.CodeUnauthorized,
			})
			c.Abort()
			return
		}

		// 验证token有效性
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 检查token是否过期
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					c.JSON(http.StatusUnauthorized, utils.Response{
						Success: false,
						Message: "令牌已过期",
						Code:    config.CodeUnauthorized,
					})
					c.Abort()
					return
				}
			}

			// 将用户信息存储到上下文中
			c.Set("userId", claims["userId"])
			c.Set("username", claims["username"])
			c.Set("userType", claims["userType"])
			c.Next()
		} else {
			log.Printf("JWT Auth: Invalid token claims or token not valid")
			c.JSON(http.StatusUnauthorized, utils.Response{
				Success: false,
				Message: "无效的令牌",
				Code:    config.CodeUnauthorized,
			})
			c.Abort()
			return
		}
	}
}
