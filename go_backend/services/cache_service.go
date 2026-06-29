package services

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/xiaohu/pingjiao/config"
)

// CacheItem 缓存项
type CacheItem struct {
	Data      []byte    `json:"data"`
	ExpiresAt time.Time `json:"expires_at"`
}

// IsExpired 检查缓存项是否过期
func (item *CacheItem) IsExpired() bool {
	return time.Now().After(item.ExpiresAt)
}

// CacheService 缓存服务
type CacheService struct {
	// 内存缓存（二级缓存）
	memoryCache map[string]*CacheItem
	mutex       sync.RWMutex

	// Redis缓存（一级缓存）
	redisService *RedisService

	// 配置
	enabled   bool
	cacheType string
	ttl       time.Duration
}

// NewCacheService 创建新的缓存服务
func NewCacheService() *CacheService {
	enabled := config.GetCacheEnabled()
	cacheType := config.GetCacheType()
	ttl := time.Duration(config.GetCacheTTL()) * time.Second

	service := &CacheService{
		memoryCache:  make(map[string]*CacheItem),
		redisService: GetRedisService(),
		enabled:      enabled,
		cacheType:    cacheType,
		ttl:          ttl,
	}

	// 启动清理过期缓存的goroutine
	if enabled {
		go service.cleanupExpiredItems()
	}

	return service
}

// Get 获取缓存数据（两级缓存逻辑）
func (cs *CacheService) Get(key string) ([]byte, bool) {
	if !cs.enabled {
		return nil, false
	}

	// 1. 先尝试从内存缓存获取（二级缓存）
	if data, found := cs.getFromMemory(key); found {
		return data, true
	}

	// 2. 如果内存缓存未命中，尝试从Redis获取（一级缓存）
	if cs.cacheType == "redis" || cs.cacheType == "hybrid" {
		if data, found := cs.getFromRedis(key); found {
			// 将Redis中的数据同步到内存缓存
			cs.setToMemory(key, data, cs.ttl)
			return data, true
		}
	}

	return nil, false
}

// getFromMemory 从内存缓存获取数据
func (cs *CacheService) getFromMemory(key string) ([]byte, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	item, exists := cs.memoryCache[key]
	if !exists {
		return nil, false
	}

	if item.IsExpired() {
		// 异步删除过期项
		go cs.deleteFromMemory(key)
		return nil, false
	}

	return item.Data, true
}

// getFromRedis 从Redis缓存获取数据
func (cs *CacheService) getFromRedis(key string) ([]byte, bool) {
	if cs.redisService == nil || !cs.redisService.IsEnabled() {
		return nil, false
	}

	return cs.redisService.Get(key)
}

// Set 设置缓存数据（两级缓存逻辑）
func (cs *CacheService) Set(key string, data []byte, ttl ...time.Duration) {
	if !cs.enabled {
		return
	}

	// 使用自定义TTL或默认TTL
	expireDuration := cs.ttl
	if len(ttl) > 0 {
		expireDuration = ttl[0]
	}

	// 1. 设置到Redis缓存（一级缓存）
	if cs.cacheType == "redis" || cs.cacheType == "hybrid" {
		cs.setToRedis(key, data, expireDuration)
	}

	// 2. 设置到内存缓存（二级缓存）
	if cs.cacheType == "memory" || cs.cacheType == "hybrid" {
		cs.setToMemory(key, data, expireDuration)
	}
}

// setToMemory 设置数据到内存缓存
func (cs *CacheService) setToMemory(key string, data []byte, ttl time.Duration) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.memoryCache[key] = &CacheItem{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// setToRedis 设置数据到Redis缓存
func (cs *CacheService) setToRedis(key string, data []byte, ttl time.Duration) {
	if cs.redisService == nil || !cs.redisService.IsEnabled() {
		return
	}

	if err := cs.redisService.Set(key, data, ttl); err != nil {
		log.Printf("Redis设置缓存失败: %v", err)
	}
}

// Delete 删除缓存数据（两级缓存逻辑）
func (cs *CacheService) Delete(key string) {
	if !cs.enabled {
		return
	}

	// 1. 从Redis删除（一级缓存）
	if cs.cacheType == "redis" || cs.cacheType == "hybrid" {
		cs.deleteFromRedis(key)
	}

	// 2. 从内存删除（二级缓存）
	if cs.cacheType == "memory" || cs.cacheType == "hybrid" {
		cs.deleteFromMemory(key)
	}
}

// deleteFromMemory 从内存缓存删除数据
func (cs *CacheService) deleteFromMemory(key string) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	delete(cs.memoryCache, key)
}

// deleteFromRedis 从Redis缓存删除数据
func (cs *CacheService) deleteFromRedis(key string) {
	if cs.redisService == nil || !cs.redisService.IsEnabled() {
		return
	}

	if err := cs.redisService.Delete(key); err != nil {
		log.Printf("Redis删除缓存失败: %v", err)
	}
}

// DeletePattern 根据模式删除缓存数据
func (cs *CacheService) DeletePattern(pattern string) {
	if !cs.enabled {
		return
	}

	// 1. 从Redis删除匹配的键（一级缓存）
	if cs.cacheType == "redis" || cs.cacheType == "hybrid" {
		cs.deletePatternFromRedis(pattern)
	}

	// 2. 从内存删除匹配的键（二级缓存）
	if cs.cacheType == "memory" || cs.cacheType == "hybrid" {
		cs.deletePatternFromMemory(pattern)
	}
}

// deletePatternFromMemory 从内存缓存删除匹配模式的数据
func (cs *CacheService) deletePatternFromMemory(pattern string) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// 简单的前缀匹配实现
	prefix := pattern
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix = pattern[:len(pattern)-1]
	}

	keysToDelete := make([]string, 0)
	for key := range cs.memoryCache {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(cs.memoryCache, key)
	}
}

// deletePatternFromRedis 从Redis缓存删除匹配模式的数据
func (cs *CacheService) deletePatternFromRedis(pattern string) {
	if cs.redisService == nil || !cs.redisService.IsEnabled() {
		return
	}

	if err := cs.redisService.DeletePattern(pattern); err != nil {
		log.Printf("Redis模式删除缓存失败: %v", err)
	}
}

// Clear 清空所有缓存（两级缓存逻辑）
func (cs *CacheService) Clear() {
	if !cs.enabled {
		return
	}

	// 1. 清空Redis缓存（一级缓存）
	if cs.cacheType == "redis" || cs.cacheType == "hybrid" {
		if cs.redisService != nil && cs.redisService.IsEnabled() {
			if err := cs.redisService.Clear(); err != nil {
				log.Printf("Redis清空缓存失败: %v", err)
			}
		}
	}

	// 2. 清空内存缓存（二级缓存）
	if cs.cacheType == "memory" || cs.cacheType == "hybrid" {
		cs.mutex.Lock()
		defer cs.mutex.Unlock()
		cs.memoryCache = make(map[string]*CacheItem)
	}
}

// GetStats 获取缓存统计信息
func (cs *CacheService) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled":     cs.enabled,
		"cache_type":  cs.cacheType,
		"ttl_seconds": int(cs.ttl.Seconds()),
	}

	// 内存缓存统计
	cs.mutex.RLock()
	totalMemoryItems := len(cs.memoryCache)
	expiredMemoryItems := 0
	for _, item := range cs.memoryCache {
		if item.IsExpired() {
			expiredMemoryItems++
		}
	}
	cs.mutex.RUnlock()

	stats["memory_cache"] = map[string]interface{}{
		"total_items":   totalMemoryItems,
		"expired_items": expiredMemoryItems,
		"active_items":  totalMemoryItems - expiredMemoryItems,
	}

	// Redis缓存统计
	if cs.redisService != nil && cs.redisService.IsEnabled() {
		stats["redis_cache"] = cs.redisService.GetStats()
	} else {
		stats["redis_cache"] = map[string]interface{}{
			"enabled": false,
		}
	}

	return stats
}

// GetCacheType 获取缓存类型
func (cs *CacheService) GetCacheType() string {
	return cs.cacheType
}

// cleanupExpiredItems 定期清理过期的缓存项（仅清理内存缓存）
func (cs *CacheService) cleanupExpiredItems() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		// 只清理内存缓存，Redis有自己的过期机制
		if cs.cacheType == "memory" || cs.cacheType == "hybrid" {
			cs.mutex.Lock()
			for key, item := range cs.memoryCache {
				if item.IsExpired() {
					delete(cs.memoryCache, key)
				}
			}
			cs.mutex.Unlock()
		}
	}
}

// GenerateKey 生成缓存键
func (cs *CacheService) GenerateKey(prefix, userID string, params map[string]string) string {
	key := fmt.Sprintf("%s:user:%s", prefix, userID)

	if len(params) > 0 {
		// 将参数按字母顺序排序并添加到键中
		for k, v := range params {
			key += fmt.Sprintf(":%s:%s", k, v)
		}
	}

	return key
}

// SetJSON 设置JSON格式的缓存数据
func (cs *CacheService) SetJSON(key string, data interface{}, ttl ...time.Duration) error {
	if !cs.enabled {
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	cs.Set(key, jsonData, ttl...)
	return nil
}

// GetJSON 获取JSON格式的缓存数据
func (cs *CacheService) GetJSON(key string, target interface{}) (bool, error) {
	if !cs.enabled {
		return false, nil
	}

	data, exists := cs.Get(key)
	if !exists {
		return false, nil
	}

	err := json.Unmarshal(data, target)
	if err != nil {
		return false, fmt.Errorf("反序列化数据失败: %w", err)
	}

	return true, nil
}

// 全局缓存服务实例
var globalCacheService *CacheService

// InitCacheService 初始化全局缓存服务
func InitCacheService() {
	globalCacheService = NewCacheService()
}

// GetCacheService 获取全局缓存服务实例
func GetCacheService() *CacheService {
	if globalCacheService == nil {
		InitCacheService()
	}
	return globalCacheService
}
