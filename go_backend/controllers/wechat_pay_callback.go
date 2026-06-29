package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// WechatPayNotifyData 微信支付回调数据结构
type WechatPayNotifyData struct {
	ID           string                `json:"id"`
	CreateTime   string                `json:"create_time"`
	EventType    string                `json:"event_type"`
	ResourceType string                `json:"resource_type"`
	Resource     WechatPayResourceData `json:"resource"`
	Summary      string                `json:"summary"`
}

// WechatPayResourceData 微信支付资源数据
type WechatPayResourceData struct {
	Algorithm      string `json:"algorithm"`
	Ciphertext     string `json:"ciphertext"`
	AssociatedData string `json:"associated_data"`
	OriginalType   string `json:"original_type"`
	Nonce          string `json:"nonce"`
}

// WechatPayDecryptedData 解密后的支付数据
type WechatPayDecryptedData struct {
	AppID          string              `json:"appid"`
	MchID          string              `json:"mchid"`
	OutTradeNo     string              `json:"out_trade_no"`
	TransactionID  string              `json:"transaction_id"`
	TradeType      string              `json:"trade_type"`
	TradeState     string              `json:"trade_state"`
	TradeStateDesc string              `json:"trade_state_desc"`
	BankType       string              `json:"bank_type"`
	Attach         string              `json:"attach"`
	SuccessTime    string              `json:"success_time"`
	Payer          WechatPayPayerData  `json:"payer"`
	Amount         WechatPayAmountData `json:"amount"`
}

// WechatPayPayerData 支付者数据
type WechatPayPayerData struct {
	OpenID string `json:"openid"`
}

// WechatPayAmountData 金额数据
type WechatPayAmountData struct {
	Total         int    `json:"total"`
	PayerTotal    int    `json:"payer_total"`
	Currency      string `json:"currency"`
	PayerCurrency string `json:"payer_currency"`
}

// WechatPayCallback 微信支付回调处理
func WechatPayCallback(c *gin.Context) {
	log.Printf("收到微信支付回调请求")

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("读取回调请求体失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "FAIL",
			"message": "读取请求体失败",
		})
		return
	}

	log.Printf("微信支付回调数据: %s", string(body))

	// 解析回调数据
	var notifyData WechatPayNotifyData
	if err := json.Unmarshal(body, &notifyData); err != nil {
		log.Printf("解析回调数据失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "FAIL",
			"message": "解析回调数据失败",
		})
		return
	}

	// 验证事件类型
	if notifyData.EventType != "TRANSACTION.SUCCESS" {
		log.Printf("忽略非支付成功事件: %s", notifyData.EventType)
		c.JSON(http.StatusOK, gin.H{
			"code":    "SUCCESS",
			"message": "OK",
		})
		return
	}

	// 处理支付成功回调
	if err := handlePaymentSuccess(&notifyData); err != nil {
		log.Printf("处理支付成功回调失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAIL",
			"message": "处理回调失败",
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "OK",
	})
}

// handlePaymentSuccess 处理支付成功
func handlePaymentSuccess(notifyData *WechatPayNotifyData) error {
	// 这里需要解密回调数据
	// 由于微信支付API v3使用了加密，需要使用平台证书进行解密
	// 为了简化示例，这里假设已经解密了数据

	// TODO: 实现数据解密逻辑
	// decryptedData, err := decryptNotifyData(notifyData.Resource)
	// if err != nil {
	//     return fmt.Errorf("解密回调数据失败: %v", err)
	// }

	// 模拟解密后的数据
	decryptedData := WechatPayDecryptedData{
		OutTradeNo:    "SMS_1_1703123456", // 从实际解密数据中获取
		TransactionID: "4200001234567890", // 从实际解密数据中获取
		TradeState:    "SUCCESS",
		Amount: WechatPayAmountData{
			Total: 1000, // 从实际解密数据中获取
		},
	}

	// 处理支付成功
	wechatPayService := services.GetWechatPayService()
	if wechatPayService == nil {
		return fmt.Errorf("微信支付服务未初始化")
	}

	if decryptedData.TradeState == "SUCCESS" {
		return wechatPayService.ProcessPaymentSuccess(decryptedData.OutTradeNo, decryptedData.TransactionID)
	} else {
		return wechatPayService.ProcessPaymentFailed(decryptedData.OutTradeNo, decryptedData.TradeStateDesc)
	}
}

// QueryPaymentStatus 查询支付状态
func QueryPaymentStatus(c *gin.Context) {
	// 验证用户认证
	_, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Response{
			Success: false,
			Message: config.MsgUnauthorized,
			Code:    config.CodeUnauthorized,
		})
		return
	}

	// 获取订单ID
	orderID := c.Param("orderId")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "订单ID不能为空",
			Code:    400,
		})
		return
	}

	// 查询支付状态
	wechatPayService := services.GetWechatPayService()
	if wechatPayService == nil || !wechatPayService.IsEnabled() {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "微信支付服务未启用",
			Code:    config.CodeServerError,
		})
		return
	}

	status, err := wechatPayService.QueryPaymentStatus(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "查询支付状态失败: " + err.Error(),
			Code:    config.CodeServerError,
		})
		return
	}

	// 如果支付成功，自动处理余额更新
	if status == "SUCCESS" {
		log.Printf("支付状态为SUCCESS，开始检查交易记录: 订单%s", orderID)

		// 检查是否已经处理过
		transaction, err := models.GetSMSTransactionByOrderID(orderID)
		if err != nil {
			log.Printf("获取交易记录失败: 订单%s, 错误: %v", orderID, err)
		} else {
			log.Printf("找到交易记录: 订单%s, 状态: %s, 用户ID: %d, 金额: %d分",
				orderID, transaction.Status, transaction.UserID, transaction.Amount)

			if transaction.Status == models.TransactionStatusPending {
				log.Printf("交易状态为待处理，开始处理支付成功: 订单%s", orderID)
				// 处理支付成功，更新余额
				err = wechatPayService.ProcessPaymentSuccess(orderID, "AUTO_DETECTED")
				if err != nil {
					log.Printf("自动处理支付成功失败: 订单%s, 错误: %v", orderID, err)
				} else {
					log.Printf("自动处理支付成功完成: 订单%s", orderID)
				}
			} else {
				log.Printf("交易已处理过，跳过: 订单%s, 当前状态: %s", orderID, transaction.Status)
			}
		}
	} else {
		log.Printf("支付状态不是SUCCESS: 订单%s, 状态: %s", orderID, status)
	}

	result := map[string]interface{}{
		"orderId": orderID,
		"status":  status,
		"paid":    status == "SUCCESS",
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "查询支付状态成功",
		Code:    config.CodeSuccess,
		Result:  result,
	})
}

// ManualProcessPayment 手动处理支付（用于测试）
func ManualProcessPayment(c *gin.Context) {
	// 检查管理员权限
	userType, exists := c.Get("userType")
	if !exists || userType != "admin" {
		c.JSON(http.StatusForbidden, utils.Response{
			Success: false,
			Message: "权限不足",
			Code:    config.CodeForbidden,
		})
		return
	}

	// 解析请求参数
	var req struct {
		OrderID       string `json:"orderId" binding:"required"`
		TransactionID string `json:"transactionId" binding:"required"`
		Success       bool   `json:"success"`
		Reason        string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "参数错误: " + err.Error(),
			Code:    400,
		})
		return
	}

	// 处理支付结果
	wechatPayService := services.GetWechatPayService()
	if wechatPayService == nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "微信支付服务未初始化",
			Code:    config.CodeServerError,
		})
		return
	}

	var err error
	if req.Success {
		err = wechatPayService.ProcessPaymentSuccess(req.OrderID, req.TransactionID)
	} else {
		reason := req.Reason
		if reason == "" {
			reason = "手动标记为失败"
		}
		err = wechatPayService.ProcessPaymentFailed(req.OrderID, reason)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "处理支付结果失败: " + err.Error(),
			Code:    config.CodeServerError,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "处理支付结果成功",
		Code:    config.CodeSuccess,
	})
}

// GetPaymentCertificates 获取微信支付平台证书（用于验证回调签名）
func GetPaymentCertificates(c *gin.Context) {
	// 检查管理员权限
	userType, exists := c.Get("userType")
	if !exists || userType != "admin" {
		c.JSON(http.StatusForbidden, utils.Response{
			Success: false,
			Message: "权限不足",
			Code:    config.CodeForbidden,
		})
		return
	}

	// TODO: 实现获取微信支付平台证书的逻辑
	// 这个功能用于定期更新平台证书，确保回调验签正常

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取平台证书功能待实现",
		Code:    config.CodeSuccess,
	})
}

// decryptNotifyData 解密微信支付回调数据
// func decryptNotifyData(resource WechatPayResourceData) (*WechatPayDecryptedData, error) {
//     // TODO: 实现微信支付回调数据解密
//     // 1. 使用平台证书验证签名
//     // 2. 使用APIv3密钥解密数据
//     // 3. 返回解密后的数据
//     return nil, fmt.Errorf("解密功能待实现")
// }
