// Package controllers handles HTTP request handlers for proxy-related endpoints.
// This file contains response standardization utilities and helper functions.
package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// DataSourceType 数据源类型
type DataSourceType string

const (
	DataSourceSchool   DataSourceType = "school"   // 学校服务器数据
	DataSourceDatabase DataSourceType = "database" // 用户缓存数据库数据
)

// StandardResponse 标准化响应结构
type StandardResponse struct {
	Success        bool           `json:"success"`
	Code           int            `json:"code"`
	Message        string         `json:"message"`
	Result         interface{}    `json:"result"`
	DataSourceType DataSourceType `json:"dataSourceType"`
	CacheUpdatedAt string         `json:"cacheUpdatedAt"`
	Timestamp      string         `json:"timestamp"`
}

// standardizeAPIResponse 统一API响应格式
func standardizeAPIResponse(rawResponse []byte, dataSourceType DataSourceType, cacheUpdatedAt string) (*StandardResponse, error) {
	var response map[string]interface{}
	if err := json.Unmarshal(rawResponse, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查是否需要格式化
	if needsStandardization(response) {
		response = standardizeResponse(response)
	}

	// 构建标准化响应
	standardResp := &StandardResponse{
		Success:        getSuccessFromResponse(response),
		Code:           getCodeFromResponse(response),
		Message:        getMessageFromResponse(response),
		Result:         getResultFromResponse(response),
		DataSourceType: dataSourceType,
		CacheUpdatedAt: cacheUpdatedAt,
		Timestamp:      time.Now().Format("2006-01-02 15:04:05"),
	}

	return standardResp, nil
}

// getSuccessFromResponse 从响应中提取success字段
func getSuccessFromResponse(response map[string]interface{}) bool {
	if success, ok := response["success"].(bool); ok {
		return success
	}
	// 如果没有success字段，根据code判断
	if code, ok := response["code"].(float64); ok {
		return code == 200 || code == 0
	}
	return true // 默认为成功
}

// getCodeFromResponse 从响应中提取code字段
func getCodeFromResponse(response map[string]interface{}) int {
	if code, ok := response["code"].(float64); ok {
		return int(code)
	}
	return 200 // 默认为200
}

// getResultFromResponse 从响应中提取result字段
func getResultFromResponse(response map[string]interface{}) interface{} {
	if result, ok := response["result"]; ok {
		return result
	}
	return nil
}

// needsStandardization 检查响应是否需要标准化
func needsStandardization(response map[string]interface{}) bool {
	fmt.Printf("检查是否需要标准化，响应结构: %+v\n", response)

	// 检查是否有多层嵌套的result结构或重复的元数据
	if result, ok := response["result"].(map[string]interface{}); ok {
		// 检查是否有重复的元数据（表明需要标准化）
		if _, hasCacheType := result["cacheType"]; hasCacheType {
			fmt.Printf("检测到重复的元数据，需要标准化\n")
			return true
		}

		// 检查是否有嵌套的result结构
		if innerResult, ok := result["result"].(map[string]interface{}); ok {
			fmt.Printf("检测到二层嵌套，需要标准化\n")
			// 检查三层嵌套
			if _, hasThirdLevel := innerResult["result"]; hasThirdLevel {
				fmt.Printf("检测到三层嵌套，需要标准化\n")
			}
			return true // 任何嵌套的result结构都需要标准化
		}
	}

	// 检查是否是已经标准化的格式（有success、code、message等标准字段）
	hasStandardFields := false
	if _, hasSuccess := response["success"]; hasSuccess {
		hasStandardFields = true
	}
	if _, hasCode := response["code"]; hasCode {
		hasStandardFields = true
	}

	// 如果没有标准字段且有result字段，可能需要标准化
	if !hasStandardFields {
		if _, hasResult := response["result"]; hasResult {
			fmt.Printf("检测到非标准格式，需要标准化\n")
			return true
		}
	}

	fmt.Printf("无需标准化\n")
	return false
}

// standardizeResponse 标准化响应格式
func standardizeResponse(response map[string]interface{}) map[string]interface{} {
	fmt.Printf("标准化前的响应结构: %+v\n", response)

	// 提取最深层的实际数据和最外层的元数据
	var finalData interface{}
	var finalMessage string
	var finalSuccess interface{}
	var finalCode interface{}

	// 从最外层获取元数据（这些是最准确的）
	outerSuccess := response["success"]
	outerCode := response["code"]
	outerFromCache := response["fromCache"]
	outerCacheType := response["cacheType"]
	outerCacheUpdatedAt := response["cacheUpdatedAt"]
	outerServerDown := response["serverDown"]
	outerTimestamp := response["timestamp"]

	// 递归查找最深层的实际数据
	if result, ok := response["result"].(map[string]interface{}); ok {
		if innerResult, ok := result["result"].(map[string]interface{}); ok {
			if thirdResult, ok := innerResult["result"].(map[string]interface{}); ok {
				// 三层嵌套：使用最内层的数据
				finalData = thirdResult
				finalMessage = getStringFromMap(innerResult, "message", "获取成功")
				finalSuccess = outerSuccess // 使用外层的success状态
				finalCode = outerCode
				fmt.Printf("检测到三层嵌套，使用最内层数据\n")
			} else {
				// 二层嵌套：使用第二层数据
				finalData = innerResult
				finalMessage = getStringFromMap(result, "message", "获取成功")
				finalSuccess = outerSuccess
				finalCode = outerCode
				fmt.Printf("检测到二层嵌套，使用第二层数据\n")
			}
		} else {
			// 一层嵌套：使用第一层数据
			finalData = result
			finalMessage = getStringFromMap(response, "message", "获取成功")
			finalSuccess = outerSuccess
			finalCode = outerCode
			fmt.Printf("检测到一层嵌套，使用第一层数据\n")
		}
	} else {
		// 没有嵌套：直接使用result
		finalData = response["result"]
		finalMessage = getStringFromMap(response, "message", "获取成功")
		finalSuccess = outerSuccess
		finalCode = outerCode
		fmt.Printf("没有嵌套，直接使用result\n")
	}

	// 确保必要字段有默认值
	if finalSuccess == nil {
		finalSuccess = true
	}
	if finalCode == nil {
		finalCode = 200
	}

	// 构建标准化的响应（单层结构）
	standardizedResponse := map[string]interface{}{
		"success":        finalSuccess,
		"code":           finalCode,
		"message":        finalMessage,
		"result":         finalData,
		"fromCache":      outerFromCache,
		"cacheType":      outerCacheType,
		"cacheUpdatedAt": outerCacheUpdatedAt,
		"serverDown":     outerServerDown,
		"timestamp":      outerTimestamp,
	}

	fmt.Printf("标准化后的响应结构: %+v\n", standardizedResponse)
	return standardizedResponse
}

// getStringFromMap 从map中安全获取字符串值，带默认值
func getStringFromMap(data map[string]interface{}, key string, defaultValue string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok && str != "" {
			return str
		}
	}
	return defaultValue
}

// getMessageFromResponse 从响应中提取消息（保留用于兼容性）
func getMessageFromResponse(response map[string]interface{}) string {
	if message, ok := response["message"].(string); ok && message != "" {
		return message
	}

	// 尝试从嵌套结构中获取消息
	if result, ok := response["result"].(map[string]interface{}); ok {
		if message, ok := result["message"].(string); ok && message != "" {
			return message
		}
		if innerResult, ok := result["result"].(map[string]interface{}); ok {
			if message, ok := innerResult["message"].(string); ok && message != "" {
				return message
			}
		}
	}

	return "获取成功"
}

// 辅助函数：从map中安全获取字符串值
func getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// 辅助函数：从map中安全获取整数值
func getIntValue(data map[string]interface{}, key string) int {
	if value, ok := data[key]; ok {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if intVal, err := strconv.Atoi(v); err == nil {
				return intVal
			}
		}
	}
	return 0
}

// 辅助函数：将星期字符串转换为数字
func getWeekdayFromString(weekStr string) int {
	switch weekStr {
	case "一":
		return 1
	case "二":
		return 2
	case "三":
		return 3
	case "四":
		return 4
	case "五":
		return 5
	case "六":
		return 6
	case "日", "七":
		return 7
	default:
		return 1
	}
}
