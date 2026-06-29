package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
	"github.com/xiaohu/pingjiao/config"
)

// RabbitMQService RabbitMQ服务
type RabbitMQService struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	enabled    bool
}

// NotificationMessage 通知消息结构
type NotificationMessage struct {
	Type       string                 `json:"type"`       // 通知类型
	UserID     uint                   `json:"userId"`     // 接收者ID
	FromUserID uint                   `json:"fromUserId"` // 发送者ID
	Title      string                 `json:"title"`      // 通知标题
	Content    string                 `json:"content"`    // 通知内容
	RelatedID  uint                   `json:"relatedId"`  // 相关对象ID
	RelatedType string                `json:"relatedType"` // 相关对象类型
	Data       map[string]interface{} `json:"data"`       // 额外数据
	CreatedAt  time.Time              `json:"createdAt"`  // 创建时间
}

// ModerationMessage 内容审核消息结构
type ModerationMessage struct {
	Type       string                 `json:"type"`       // 审核类型: auto, manual
	TargetType string                 `json:"targetType"` // 目标类型: post, comment
	TargetID   uint                   `json:"targetId"`   // 目标ID
	Content    string                 `json:"content"`    // 内容
	UserID     uint                   `json:"userId"`     // 用户ID
	Data       map[string]interface{} `json:"data"`       // 额外数据
	CreatedAt  time.Time              `json:"createdAt"`  // 创建时间
}

var rabbitMQService *RabbitMQService

// InitRabbitMQService 初始化RabbitMQ服务
func InitRabbitMQService() {
	rabbitMQService = NewRabbitMQService()
}

// GetRabbitMQService 获取RabbitMQ服务实例
func GetRabbitMQService() *RabbitMQService {
	return rabbitMQService
}

// NewRabbitMQService 创建新的RabbitMQ服务
func NewRabbitMQService() *RabbitMQService {
	service := &RabbitMQService{
		enabled: config.GetEnvBool("RABBITMQ_ENABLED", true),
	}

	if service.enabled {
		log.Printf("RabbitMQ服务启用，开始初始化连接...")
		if err := service.connect(); err != nil {
			log.Printf("RabbitMQ连接失败，将禁用消息队列: %v", err)
			service.enabled = false
		} else {
			log.Printf("RabbitMQ连接成功")
			// 初始化队列
			service.initQueues()
		}
	} else {
		log.Printf("RabbitMQ服务未启用")
	}

	return service
}

// connect 连接到RabbitMQ服务器
func (r *RabbitMQService) connect() error {
	// 构建连接字符串
	host := config.GetEnv("RABBITMQ_HOST", "")
	port := config.GetEnv("RABBITMQ_PORT", "5672")
	username := config.GetEnv("RABBITMQ_USERNAME", "")
	password := config.GetEnv("RABBITMQ_PASSWORD", "")
	vhost := config.GetEnv("RABBITMQ_VHOST", "/")

	connStr := fmt.Sprintf("amqp://%s:%s@%s:%s%s", username, password, host, port, vhost)

	var err error
	r.connection, err = amqp.Dial(connStr)
	if err != nil {
		return fmt.Errorf("连接RabbitMQ失败: %w", err)
	}

	r.channel, err = r.connection.Channel()
	if err != nil {
		return fmt.Errorf("创建RabbitMQ通道失败: %w", err)
	}

	return nil
}

// initQueues 初始化队列
func (r *RabbitMQService) initQueues() {
	if !r.enabled || r.channel == nil {
		return
	}

	// 声明通知队列
	_, err := r.channel.QueueDeclare(
		"notifications", // 队列名称
		true,           // 持久化
		false,          // 自动删除
		false,          // 排他性
		false,          // 不等待
		nil,            // 参数
	)
	if err != nil {
		log.Printf("声明通知队列失败: %v", err)
	}

	// 声明内容审核队列
	_, err = r.channel.QueueDeclare(
		"moderation", // 队列名称
		true,        // 持久化
		false,       // 自动删除
		false,       // 排他性
		false,       // 不等待
		nil,         // 参数
	)
	if err != nil {
		log.Printf("声明内容审核队列失败: %v", err)
	}
}

// IsEnabled 检查RabbitMQ服务是否启用
func (r *RabbitMQService) IsEnabled() bool {
	return r.enabled && r.connection != nil && r.channel != nil
}

// PublishNotification 发布通知消息
func (r *RabbitMQService) PublishNotification(msg NotificationMessage) error {
	if !r.IsEnabled() {
		log.Printf("RabbitMQ未启用，跳过通知发送")
		return nil
	}

	msg.CreatedAt = time.Now()
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化通知消息失败: %w", err)
	}

	err = r.channel.Publish(
		"",             // 交换机
		"notifications", // 路由键
		false,          // 强制
		false,          // 立即
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // 持久化消息
		},
	)

	if err != nil {
		return fmt.Errorf("发布通知消息失败: %w", err)
	}

	log.Printf("通知消息已发布: %s -> 用户%d", msg.Type, msg.UserID)
	return nil
}

// PublishModerationMessage 发布内容审核消息
func (r *RabbitMQService) PublishModerationMessage(msg ModerationMessage) error {
	if !r.IsEnabled() {
		log.Printf("RabbitMQ未启用，跳过内容审核消息发送")
		return nil
	}

	msg.CreatedAt = time.Now()
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化内容审核消息失败: %w", err)
	}

	err = r.channel.Publish(
		"",          // 交换机
		"moderation", // 路由键
		false,       // 强制
		false,       // 立即
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // 持久化消息
		},
	)

	if err != nil {
		return fmt.Errorf("发布内容审核消息失败: %w", err)
	}

	log.Printf("内容审核消息已发布: %s -> %s:%d", msg.Type, msg.TargetType, msg.TargetID)
	return nil
}

// ConsumeNotifications 消费通知消息
func (r *RabbitMQService) ConsumeNotifications(handler func(NotificationMessage) error) error {
	if !r.IsEnabled() {
		return fmt.Errorf("RabbitMQ未启用")
	}

	msgs, err := r.channel.Consume(
		"notifications", // 队列名称
		"",             // 消费者标签
		false,          // 自动确认
		false,          // 排他性
		false,          // 不等待
		false,          // 参数
		nil,
	)
	if err != nil {
		return fmt.Errorf("开始消费通知消息失败: %w", err)
	}

	go func() {
		for msg := range msgs {
			var notification NotificationMessage
		if err := json.Unmarshal(msg.Body, &notification); err != nil {
			log.Printf("解析通知消息失败: %v", err)
			//nolint:errcheck
			msg.Nack(false, false) // 拒绝消息，不重新入队
			continue
		}

		if err := handler(notification); err != nil {
			log.Printf("处理通知消息失败: %v", err)
			//nolint:errcheck
			msg.Nack(false, true) // 拒绝消息，重新入队
		} else {
			//nolint:errcheck
			msg.Ack(false) // 确认消息
		}
		}
	}()

	log.Printf("开始消费通知消息...")
	return nil
}

// ConsumeModerationMessages 消费内容审核消息
func (r *RabbitMQService) ConsumeModerationMessages(handler func(ModerationMessage) error) error {
	if !r.IsEnabled() {
		return fmt.Errorf("RabbitMQ未启用")
	}

	msgs, err := r.channel.Consume(
		"moderation", // 队列名称
		"",          // 消费者标签
		false,       // 自动确认
		false,       // 排他性
		false,       // 不等待
		false,       // 参数
		nil,
	)
	if err != nil {
		return fmt.Errorf("开始消费内容审核消息失败: %w", err)
	}

	go func() {
		for msg := range msgs {
			var moderation ModerationMessage
			if err := json.Unmarshal(msg.Body, &moderation); err != nil {
				log.Printf("解析内容审核消息失败: %v", err)
				//nolint:errcheck
				msg.Nack(false, false) // 拒绝消息，不重新入队
				continue
			}

			if err := handler(moderation); err != nil {
				log.Printf("处理内容审核消息失败: %v", err)
				//nolint:errcheck
				msg.Nack(false, true) // 拒绝消息，重新入队
			} else {
				//nolint:errcheck
				msg.Ack(false) // 确认消息
			}
		}
	}()

	log.Printf("开始消费内容审核消息...")
	return nil
}

// Close 关闭RabbitMQ连接
func (r *RabbitMQService) Close() error {
	if r.channel != nil {
		//nolint:errcheck
		r.channel.Close()
	}
	if r.connection != nil {
		return r.connection.Close()
	}
	return nil
}
