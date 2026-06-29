package services

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"log"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// WechatPayService 微信支付服务
type WechatPayService struct {
	client    *core.Client
	enabled   bool
	appID     string
	mchID     string
	notifyURL string
}

// PaymentRequest 支付请求
type PaymentRequest struct {
	OrderID     string
	Amount      int64 // 金额（分）
	Description string
	UserID      uint
}

// PaymentResponse 支付响应
type PaymentResponse struct {
	CodeURL string `json:"codeUrl"` // 二维码链接
	OrderID string `json:"orderId"`
}

var wechatPayService *WechatPayService

// InitWechatPayService 初始化微信支付服务
func InitWechatPayService() {
	wechatPayService = NewWechatPayService()
}

// GetWechatPayService 获取微信支付服务实例
func GetWechatPayService() *WechatPayService {
	return wechatPayService
}

// NewWechatPayService 创建新的微信支付服务
func NewWechatPayService() *WechatPayService {
	service := &WechatPayService{
		enabled:   config.GetEnvBool("WECHAT_PAY_ENABLED", true),
		appID:     config.GetEnv("WECHAT_PAY_APP_ID", ""),
		mchID:     config.GetEnv("WECHAT_PAY_MCH_ID", ""),
		notifyURL: config.GetEnv("WECHAT_PAY_NOTIFY_URL", ""),
	}

	if service.enabled {
		log.Printf("微信支付服务启用，开始初始化客户端...")
		if err := service.initClient(); err != nil {
			log.Printf("初始化微信支付客户端失败: %v", err)
			service.enabled = false
		} else {
			log.Printf("微信支付客户端初始化成功")
		}
	} else {
		log.Printf("微信支付服务未启用")
	}

	return service
}

// initClient 初始化微信支付客户端
func (w *WechatPayService) initClient() error {
	if w.appID == "" || w.mchID == "" {
		return fmt.Errorf("微信支付配置不完整")
	}

	// 读取商户私钥
	privateKeyPath := config.GetEnv("WECHAT_PAY_PRIVATE_KEY_PATH", "./certs/apiclient_key.pem")
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("读取商户私钥失败: %w", err)
	}

	// 解析私钥
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return fmt.Errorf("解析私钥PEM格式失败")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("解析私钥失败: %w", err)
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("私钥不是RSA格式")
	}

	// 读取商户证书序列号
	certPath := config.GetEnv("WECHAT_PAY_CERT_PATH", "./certs/apiclient_cert.pem")
	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("读取商户证书失败: %w", err)
	}

	// 获取证书序列号
	serialNumber, serialErr := utils.GetCertificateSerialNumber(certBytes)
	if serialErr != nil {
		return fmt.Errorf("获取证书序列号失败: %w", serialErr)
	}

	// 获取API v3密钥
	apiV3Key := config.GetEnv("WECHAT_PAY_API_V3_KEY", "")
	if apiV3Key == "" {
		return fmt.Errorf("微信支付API v3密钥未配置")
	}

	// 创建微信支付客户端
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(w.mchID, serialNumber, rsaPrivateKey, apiV3Key),
	}

	client, err := core.NewClient(context.Background(), opts...)
	if err != nil {
		return fmt.Errorf("创建微信支付客户端失败: %w", err)
	}

	w.client = client
	return nil
}

// IsEnabled 检查微信支付服务是否启用
func (w *WechatPayService) IsEnabled() bool {
	return w.enabled && w.client != nil
}

// CreateNativePayment 创建Native支付订单
func (w *WechatPayService) CreateNativePayment(req PaymentRequest) (*PaymentResponse, error) {
	if !w.IsEnabled() {
		return nil, fmt.Errorf("微信支付服务未启用")
	}

	// 创建Native支付服务
	svc := native.NativeApiService{Client: w.client}

	// 构建支付请求
	request := native.PrepayRequest{
		Appid:       core.String(w.appID),
		Mchid:       core.String(w.mchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OrderID),
		NotifyUrl:   core.String(w.notifyURL),
		Amount: &native.Amount{
			Total:    core.Int64(req.Amount),
			Currency: core.String("CNY"),
		},
	}

	// 发起支付请求
	resp, result, err := svc.Prepay(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("创建支付订单失败: %w", err)
	}

	if result.Response.StatusCode != 200 {
		return nil, fmt.Errorf("微信支付API返回错误: %d", result.Response.StatusCode)
	}

	// 创建交易记录
	transaction := &models.SMSTransaction{
		UserID:        req.UserID,
		Type:          models.TransactionTypeRecharge,
		Amount:        int(req.Amount),
		Description:   req.Description,
		OrderID:       req.OrderID,
		PaymentMethod: "wechat",
		Status:        models.TransactionStatusPending,
	}

	if err := models.CreateSMSTransaction(transaction); err != nil {
		log.Printf("创建交易记录失败: %v", err)
	}

	return &PaymentResponse{
		CodeURL: *resp.CodeUrl,
		OrderID: req.OrderID,
	}, nil
}

// HandlePaymentNotify 处理支付回调
func (w *WechatPayService) HandlePaymentNotify(notifyData []byte) error {
	if !w.IsEnabled() {
		return fmt.Errorf("微信支付服务未启用")
	}

	// 验证签名
	// 注意：这里需要根据微信支付的回调格式进行解析和验证
	// 具体实现需要根据微信支付API v3的回调格式来处理

	// 解析回调数据（简化示例）
	// 实际实现需要解密回调数据并验证签名

	log.Printf("收到微信支付回调: %s", string(notifyData))

	// TODO: 实现具体的回调处理逻辑
	// 1. 验证签名
	// 2. 解密回调数据
	// 3. 更新订单状态
	// 4. 更新用户余额

	return nil
}

// ProcessPaymentSuccess 处理支付成功
func (w *WechatPayService) ProcessPaymentSuccess(orderID string, transactionID string) error {
	log.Printf("开始处理支付成功: 订单%s, 交易ID: %s", orderID, transactionID)

	// 获取交易记录
	transaction, err := models.GetSMSTransactionByOrderID(orderID)
	if err != nil {
		log.Printf("获取交易记录失败: 订单%s, 错误: %v", orderID, err)
		return fmt.Errorf("获取交易记录失败: %w", err)
	}

	log.Printf("获取到交易记录: ID=%d, 用户ID=%d, 金额=%d分, 状态=%s",
		transaction.ID, transaction.UserID, transaction.Amount, transaction.Status)

	if transaction.Status == models.TransactionStatusSuccess {
		log.Printf("交易已处理过，跳过: 订单%s", orderID)
		return nil // 已经处理过
	}

	// 更新交易状态
	transaction.Status = models.TransactionStatusSuccess
	transaction.ExtraData = fmt.Sprintf(`{"transaction_id": "%s"}`, transactionID)

	log.Printf("更新交易状态为成功: 订单%s", orderID)
	if err := models.UpdateSMSTransaction(transaction); err != nil {
		log.Printf("更新交易记录失败: 订单%s, 错误: %v", orderID, err)
		return fmt.Errorf("更新交易记录失败: %w", err)
	}

	// 更新用户余额
	log.Printf("开始更新用户余额: 用户ID=%d, 金额=%d分, 类型=%s",
		transaction.UserID, transaction.Amount, models.TransactionTypeRecharge)

	if err := models.UpdateSMSBalance(transaction.UserID, transaction.Amount,
		models.TransactionTypeRecharge, transaction.Description); err != nil {
		log.Printf("更新用户余额失败: 用户ID=%d, 金额=%d分, 错误: %v",
			transaction.UserID, transaction.Amount, err)
		return fmt.Errorf("更新用户余额失败: %w", err)
	}

	log.Printf("处理支付成功完成: 订单%s, 用户%d, 金额%d分", orderID, transaction.UserID, transaction.Amount)
	return nil
}

// ProcessPaymentFailed 处理支付失败
func (w *WechatPayService) ProcessPaymentFailed(orderID string, reason string) error {
	// 获取交易记录
	transaction, err := models.GetSMSTransactionByOrderID(orderID)
	if err != nil {
		return fmt.Errorf("获取交易记录失败: %w", err)
	}

	// 更新交易状态
	transaction.Status = models.TransactionStatusFailed
	transaction.ExtraData = fmt.Sprintf(`{"fail_reason": "%s"}`, reason)

	if err := models.UpdateSMSTransaction(transaction); err != nil {
		return fmt.Errorf("更新交易记录失败: %w", err)
	}

	log.Printf("处理支付失败: 订单%s, 原因: %s", orderID, reason)
	return nil
}

// QueryPaymentStatus 查询支付状态
func (w *WechatPayService) QueryPaymentStatus(orderID string) (string, error) {
	if !w.IsEnabled() {
		return "", fmt.Errorf("微信支付服务未启用")
	}

	// 创建Native支付服务
	svc := native.NativeApiService{Client: w.client}

	// 查询订单
	request := native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(orderID),
		Mchid:      core.String(w.mchID),
	}

	resp, result, err := svc.QueryOrderByOutTradeNo(context.Background(), request)
	if err != nil {
		return "", fmt.Errorf("查询订单失败: %w", err)
	}

	if result.Response.StatusCode != 200 {
		return "", fmt.Errorf("微信支付API返回错误: %d", result.Response.StatusCode)
	}

	// 返回交易状态
	if resp.TradeState != nil {
		return *resp.TradeState, nil
	}

	return "UNKNOWN", nil
}

// GenerateOrderID 生成订单ID
func (w *WechatPayService) GenerateOrderID(userID uint) string {
	return fmt.Sprintf("SMS_%d_%d", userID, time.Now().Unix())
}

// GetRechargePackages 获取充值套餐
func (w *WechatPayService) GetRechargePackages() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":          1,
			"name":        "10元套餐",
			"amount":      1000, // 分
			"smsCount":    100,  // 可发送短信数量
			"description": "10元充值，可发送100条短信",
		},
		{
			"id":          2,
			"name":        "20元套餐",
			"amount":      2000,
			"smsCount":    200,
			"description": "20元充值，可发送200条短信",
		},
		{
			"id":          3,
			"name":        "50元套餐",
			"amount":      5000,
			"smsCount":    500,
			"description": "50元充值，可发送500条短信",
		},
		{
			"id":          4,
			"name":        "100元套餐",
			"amount":      10000,
			"smsCount":    1000,
			"description": "100元充值，可发送1000条短信",
		},
	}
}
