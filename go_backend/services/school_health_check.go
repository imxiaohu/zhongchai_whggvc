package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/xiaohu/pingjiao/models"
)

// SchoolServerStatus 学校服务器状态
type SchoolServerStatus struct {
	IsAlive      bool      `json:"isAlive"`
	LastCheck    time.Time `json:"lastCheck"`
	LastAlive    time.Time `json:"lastAlive"`
	ErrorCount   int       `json:"errorCount"`
	ErrorMsg     string    `json:"errorMsg"`
	ResponseTime int64     `json:"responseTime"` // 响应时间(毫秒)
}

// SchoolHealthCheckService 学校服务器健康检查服务
type SchoolHealthCheckService struct {
	status        *SchoolServerStatus
	mutex         sync.RWMutex
	ticker        *time.Ticker
	stopChan      chan bool
	checkURL      string
	checkInterval time.Duration
	timeout       time.Duration
}

// 全局健康检查服务实例
var globalSchoolHealthCheckService *SchoolHealthCheckService
var schoolHealthOnce sync.Once

// GetSchoolHealthCheckService 获取学校健康检查服务实例
func GetSchoolHealthCheckService() *SchoolHealthCheckService {
	schoolHealthOnce.Do(func() {
		globalSchoolHealthCheckService = NewSchoolHealthCheckService()
	})
	return globalSchoolHealthCheckService
}

// NewSchoolHealthCheckService 创建新的学校健康检查服务
func NewSchoolHealthCheckService() *SchoolHealthCheckService {
	service := &SchoolHealthCheckService{
		status: &SchoolServerStatus{
			IsAlive:    true, // 默认认为服务器是活跃的
			LastCheck:  time.Now(),
			LastAlive:  time.Now(),
			ErrorCount: 0,
		},
		checkURL:      "http://scs.whggvc.net/scloudoa/sys/common/ping", // 健康检查端点
		checkInterval: 5 * time.Minute,                                  // 5分钟检查一次
		timeout:       10 * time.Second,                                 // 10秒超时
		stopChan:      make(chan bool),
	}

	// 启动时从缓存加载状态
	service.loadStatusFromCache()

	return service
}

// Start 启动健康检查服务
func (s *SchoolHealthCheckService) Start() {
	log.Println("启动学校服务器健康检查服务...")

	// 立即执行一次检查
	go s.checkHealth()

	// 启动定时检查
	s.ticker = time.NewTicker(s.checkInterval)
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.checkHealth()
			case <-s.stopChan:
				return
			}
		}
	}()
}

// Stop 停止健康检查服务
func (s *SchoolHealthCheckService) Stop() {
	log.Println("停止学校服务器健康检查服务...")
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopChan)
}

// checkHealth 执行健康检查
func (s *SchoolHealthCheckService) checkHealth() {
	log.Println("执行学校服务器健康检查...")

	startTime := time.Now()

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: s.timeout,
	}

	// 发送健康检查请求
	resp, err := client.Get(s.checkURL)
	responseTime := time.Since(startTime).Milliseconds()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.status.LastCheck = time.Now()
	s.status.ResponseTime = responseTime

	if err != nil {
		// 请求失败
		s.status.IsAlive = false
		s.status.ErrorCount++
		s.status.ErrorMsg = err.Error()
		log.Printf("学校服务器健康检查失败: %v", err)
	} else {
		//nolint:errcheck
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// 服务器正常
			s.status.IsAlive = true
			s.status.LastAlive = time.Now()
			s.status.ErrorCount = 0
			s.status.ErrorMsg = ""
			log.Printf("学校服务器健康检查成功，响应时间: %dms", responseTime)
		} else {
			// 服务器返回错误状态码
			s.status.IsAlive = false
			s.status.ErrorCount++
			s.status.ErrorMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
			log.Printf("学校服务器健康检查失败，状态码: %d", resp.StatusCode)
		}
	}

	// 保存状态到缓存
	s.saveStatusToCache()
}

// GetStatus 获取当前服务器状态
func (s *SchoolHealthCheckService) GetStatus() SchoolServerStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 返回状态副本
	return *s.status
}

// IsServerAlive 检查服务器是否存活
func (s *SchoolHealthCheckService) IsServerAlive() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.status.IsAlive
}

// GetLastAliveTime 获取最后存活时间
func (s *SchoolHealthCheckService) GetLastAliveTime() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.status.LastAlive
}

// ForceCheck 强制执行一次健康检查
func (s *SchoolHealthCheckService) ForceCheck() {
	go s.checkHealth()
}

// saveStatusToCache 保存状态到全局缓存
func (s *SchoolHealthCheckService) saveStatusToCache() {
	statusJSON, err := json.Marshal(s.status)
	if err != nil {
		log.Printf("序列化服务器状态失败: %v", err)
		return
	}

	// 保存到全局缓存，TTL设置为10分钟（比检查间隔长一些）
	err = models.SetGlobalCache("school_server_status", string(statusJSON), 10*time.Minute)
	if err != nil {
		log.Printf("保存服务器状态到缓存失败: %v", err)
	}
}

// loadStatusFromCache 从全局缓存加载状态
func (s *SchoolHealthCheckService) loadStatusFromCache() {
	cache, err := models.GetGlobalCache("school_server_status")
	if err != nil {
		log.Printf("从缓存加载服务器状态失败: %v", err)
		return
	}

	var status SchoolServerStatus
	err = json.Unmarshal([]byte(cache.CacheData), &status)
	if err != nil {
		log.Printf("反序列化服务器状态失败: %v", err)
		return
	}

	s.mutex.Lock()
	s.status = &status
	s.mutex.Unlock()

	log.Printf("从缓存加载服务器状态: 存活=%v, 最后检查=%v", status.IsAlive, status.LastCheck)
}

// GetMaintenanceMessage 获取维护提示信息
func (s *SchoolHealthCheckService) GetMaintenanceMessage() map[string]interface{} {
	status := s.GetStatus()

	return map[string]interface{}{
		"isServerAlive":  status.IsAlive,
		"lastCheck":      status.LastCheck.Format("2006-01-02 15:04:05"),
		"lastAlive":      status.LastAlive.Format("2006-01-02 15:04:05"),
		"errorCount":     status.ErrorCount,
		"errorMsg":       status.ErrorMsg,
		"responseTime":   status.ResponseTime,
		"maintenanceMsg": "学校服务器正在维护，请稍后再试",
	}
}
