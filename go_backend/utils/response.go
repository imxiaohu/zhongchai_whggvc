package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 通用响应结构
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Code      int         `json:"code"`
	Result    interface{} `json:"result,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// StandardResponse 保持与 Response 一致，用于渐进式替换
type StandardResponse = Response

// NewSuccessResponse 创建成功响应 (Code: 200 为前端兼容)
func NewSuccessResponse(data interface{}) Response {
	return Response{
		Success:   true,
		Message:   "操作成功",
		Code:      200,
		Result:    data,
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(message string, code int) Response {
	return Response{
		Success:   false,
		Message:   message,
		Code:      code,
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewStandardSuccessResponse 创建标准成功响应 (别名)
func NewStandardSuccessResponse(data interface{}) Response {
	return NewSuccessResponse(data)
}

// NewStandardErrorResponse 创建标准错误响应 (别名)
func NewStandardErrorResponse(message string, code int) Response {
	return NewErrorResponse(message, code)
}

// NewStandardResponse 创建标准响应（带自定义消息）
func NewStandardResponse(message string, data interface{}) Response {
	resp := NewSuccessResponse(data)
	resp.Message = message
	return resp
}

// SuccessResponse 发送 HTTP 200 成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, NewSuccessResponse(data))
}

// ErrorResponse 发送错误响应
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, NewErrorResponse(message, statusCode))
}
