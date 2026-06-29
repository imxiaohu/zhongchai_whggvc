package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xiaohu/pingjiao/config"
)

// RedisService Redis服务
type RedisService struct {
	client  *redis.Client
	enabled bool
	ctx     context.Context
}

// GetClient 获取Redis客户端
func (rs *RedisService) GetClient() *redis.Client {
	if rs.IsEnabled() {
		return rs.client
	}
	return nil
}

// NewRedisService 创建新的Redis服务
func NewRedisService() *RedisService {
	// Redis启用条件：
	// 1. Redis功能启用 (REDIS_ENABLED=true)
	// 2. 缓存功能启用 (CACHE_ENABLED=true)
	// 3. 缓存类型为redis或hybrid (CACHE_TYPE=redis|hybrid)
	enabled := config.GetRedisEnabled() &&
		config.GetCacheEnabled() &&
		(config.GetCacheType() == "redis" || config.GetCacheType() == "hybrid")

	service := &RedisService{
		enabled: enabled,
		ctx:     context.Background(),
	}

	if enabled {
		log.Printf("Redis服务启用，开始初始化连接...")
		service.initRedisClient()
	} else {
		log.Printf("Redis服务未启用 - Redis启用:%v, 缓存启用:%v, 缓存类型:%s",
			config.GetRedisEnabled(), config.GetCacheEnabled(), config.GetCacheType())
	}

	return service
}

// initRedisClient 初始化Redis客户端
func (rs *RedisService) initRedisClient() {
	// 创建Redis客户端配置
	options := &redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.GetRedisHost(), config.GetRedisPort()),
		Password:     config.GetRedisPassword(),
		DB:           config.GetRedisDB(),
		PoolSize:     config.GetRedisPoolSize(),
		ReadTimeout:  time.Duration(config.GetRedisTimeout()) * time.Second,
		WriteTimeout: time.Duration(config.GetRedisTimeout()) * time.Second,
		DialTimeout:  time.Duration(config.GetRedisTimeout()) * time.Second,
	}

	// 创建Redis客户端
	rs.client = redis.NewClient(options)

	// 测试连接
	if err := rs.testConnection(); err != nil {
		log.Printf("Redis连接失败，将禁用Redis缓存: %v", err)
		rs.enabled = false
		rs.client = nil
	} else {
		log.Printf("Redis连接成功: %s:%s DB:%d", config.GetRedisHost(), config.GetRedisPort(), config.GetRedisDB())
	}
}

// testConnection 测试Redis连接
func (rs *RedisService) testConnection() error {
	if rs.client == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}

	ctx, cancel := context.WithTimeout(rs.ctx, 5*time.Second)
	defer cancel()

	_, err := rs.client.Ping(ctx).Result()
	return err
}

// IsEnabled 检查Redis服务是否启用
func (rs *RedisService) IsEnabled() bool {
	return rs.enabled && rs.client != nil
}

// Get 获取缓存数据
func (rs *RedisService) Get(key string) ([]byte, bool) {
	if !rs.IsEnabled() {
		return nil, false
	}

	ctx, cancel := context.WithTimeout(rs.ctx, 3*time.Second)
	defer cancel()

	result, err := rs.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 键不存在
			return nil, false
		}
		log.Printf("Redis GET错误: %v", err)
		return nil, false
	}

	return result, true
}

// Set 设置缓存数据
func (rs *RedisService) Set(key string, data []byte, ttl time.Duration) error {
	if !rs.IsEnabled() {
		return nil
	}

	ctx, cancel := context.WithTimeout(rs.ctx, 3*time.Second)
	defer cancel()

	err := rs.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		log.Printf("Redis SET错误: %v", err)
		return err
	}

	return nil
}

// Delete 删除缓存数据
func (rs *RedisService) Delete(key string) error {
	if !rs.IsEnabled() {
		return nil
	}

	ctx, cancel := context.WithTimeout(rs.ctx, 3*time.Second)
	defer cancel()

	err := rs.client.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Redis DELETE错误: %v", err)
		return err
	}

	return nil
}

// DeletePattern 根据模式删除缓存数据
func (rs *RedisService) DeletePattern(pattern string) error {
	if !rs.IsEnabled() {
		return nil
	}

	ctx, cancel := context.WithTimeout(rs.ctx, 10*time.Second)
	defer cancel()

	// 使用SCAN命令查找匹配的键
	var cursor uint64
	var keys []string

	for {
		var scanKeys []string
		var err error
		scanKeys, cursor, err = rs.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			log.Printf("Redis SCAN错误: %v", err)
			return err
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	// 批量删除找到的键
	if len(keys) > 0 {
		err := rs.client.Del(ctx, keys...).Err()
		if err != nil {
			log.Printf("Redis批量删除错误: %v", err)
			return err
		}
		log.Printf("Redis删除了%d个匹配模式'%s'的键", len(keys), pattern)
	}

	return nil
}

// Clear 清空所有缓存（谨慎使用）
func (rs *RedisService) Clear() error {
	if !rs.IsEnabled() {
		return nil
	}

	ctx, cancel := context.WithTimeout(rs.ctx, 10*time.Second)
	defer cancel()

	err := rs.client.FlushDB(ctx).Err()
	if err != nil {
		log.Printf("Redis FLUSHDB错误: %v", err)
		return err
	}

	return nil
}

// Exists 检查键是否存在
func (rs *RedisService) Exists(key string) (bool, error) {
	if !rs.IsEnabled() {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(rs.ctx, 3*time.Second)
	defer cancel()

	result, err := rs.client.Exists(ctx, key).Result()
	if err != nil {
		log.Printf("Redis EXISTS错误: %v", err)
		return false, err
	}

	return result > 0, nil
}

// TTL 获取键的剩余过期时间
func (rs *RedisService) TTL(key string) (time.Duration, error) {
	if !rs.IsEnabled() {
		return 0, nil
	}

	ctx, cancel := context.WithTimeout(rs.ctx, 3*time.Second)
	defer cancel()

	result, err := rs.client.TTL(ctx, key).Result()
	if err != nil {
		log.Printf("Redis TTL错误: %v", err)
		return 0, err
	}

	return result, nil
}

// GetStats 获取Redis统计信息
func (rs *RedisService) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled":   rs.enabled,
		"connected": false,
	}

	if !rs.IsEnabled() {
		return stats
	}

	// 测试连接状态
	if err := rs.testConnection(); err == nil {
		stats["connected"] = true

		// 获取Redis信息
		ctx, cancel := context.WithTimeout(rs.ctx, 3*time.Second)
		defer cancel()

		if info, err := rs.client.Info(ctx, "memory", "stats").Result(); err == nil {
			stats["info"] = info
		}

		// 获取数据库大小
		if dbSize, err := rs.client.DBSize(ctx).Result(); err == nil {
			stats["db_size"] = dbSize
		}
	}

	return stats
}

// Close 关闭Redis连接
func (rs *RedisService) Close() error {
	if rs.client != nil {
		return rs.client.Close()
	}
	return nil
}

// 全局Redis服务实例
var globalRedisService *RedisService

// InitRedisService 初始化全局Redis服务
func InitRedisService() {
	globalRedisService = NewRedisService()
}

// GetRedisService 获取全局Redis服务实例
func GetRedisService() *RedisService {
	if globalRedisService == nil {
		InitRedisService()
	}
	return globalRedisService
}
