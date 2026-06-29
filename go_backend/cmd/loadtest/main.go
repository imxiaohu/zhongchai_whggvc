package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// TestConfig 测试配置
type TestConfig struct {
	BaseURL         string        `json:"base_url"`
	MaxConcurrency  int           `json:"max_concurrency"`
	RampUpDuration  time.Duration `json:"ramp_up_duration"`
	TestDuration    time.Duration `json:"test_duration"`
	RequestInterval time.Duration `json:"request_interval"`
	Timeout         time.Duration `json:"timeout"`
	EnableAuth      bool          `json:"enable_auth"`
	TestUsername    string        `json:"test_username"`
	TestPassword    string        `json:"test_password"`
	OutputFile      string        `json:"output_file"`
	EnableTLS       bool          `json:"enable_tls"`
	SkipTLSVerify   bool          `json:"skip_tls_verify"`
}

// TestEndpoint 测试端点配置
type TestEndpoint struct {
	Name        string            `json:"name"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	RequireAuth bool              `json:"require_auth"`
	Weight      int               `json:"weight"` // 权重，用于控制请求分布
}

// TestResult 单次请求结果
type TestResult struct {
	Endpoint     string
	StatusCode   int
	ResponseTime time.Duration
	Error        error
	Timestamp    time.Time
	Success      bool
}

// TestMetrics 测试指标
type TestMetrics struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalResponseTime  time.Duration
	MinResponseTime    time.Duration
	MaxResponseTime    time.Duration
	ResponseTimes      []time.Duration
	ErrorCounts        map[string]int64
	StatusCodes        map[int]int64
	StartTime          time.Time
	EndTime            time.Time
	mutex              sync.RWMutex
}

// LoadTester 负载测试器
type LoadTester struct {
	config    TestConfig
	endpoints []TestEndpoint
	client    *http.Client
	metrics   *TestMetrics
	authToken string
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewLoadTester 创建新的负载测试器
func NewLoadTester(config TestConfig) *LoadTester {
	ctx, cancel := context.WithCancel(context.Background())

	// 配置HTTP客户端
	transport := &http.Transport{
		MaxIdleConns:        config.MaxConcurrency * 2,
		MaxIdleConnsPerHost: config.MaxConcurrency,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
	}

	if config.EnableTLS && config.SkipTLSVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

	metrics := &TestMetrics{
		ResponseTimes: make([]time.Duration, 0),
		ErrorCounts:   make(map[string]int64),
		StatusCodes:   make(map[int]int64),
	}

	return &LoadTester{
		config:  config,
		client:  client,
		metrics: metrics,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// 默认测试端点配置
func getDefaultEndpoints() []TestEndpoint {
	return []TestEndpoint{
		{
			Name:        "Health Check",
			Method:      "GET",
			Path:        "/api/health/school-server",
			Headers:     map[string]string{"Content-Type": "application/json"},
			RequireAuth: false,
			Weight:      30,
		},
		{
			Name:        "Init",
			Method:      "GET",
			Path:        "/scloud/init",
			Headers:     map[string]string{"Content-Type": "application/json"},
			RequireAuth: false,
			Weight:      20,
		},
		{
			Name:        "User Info",
			Method:      "GET",
			Path:        "/api/user/info",
			Headers:     map[string]string{"Content-Type": "application/json"},
			RequireAuth: true,
			Weight:      25,
		},
		{
			Name:        "Clubs List",
			Method:      "GET",
			Path:        "/api/clubs",
			Headers:     map[string]string{"Content-Type": "application/json"},
			RequireAuth: true,
			Weight:      25,
		},
	}
}

// authenticate 获取认证令牌
func (lt *LoadTester) authenticate() error {
	if !lt.config.EnableAuth {
		return nil
	}

	loginData := map[string]string{
		"username": lt.config.TestUsername,
		"password": lt.config.TestPassword,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return fmt.Errorf("failed to marshal login data: %w", err)
	}

	req, err := http.NewRequestWithContext(lt.ctx, "POST", lt.config.BaseURL+"/scloud/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := lt.client.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read login response: %w", err)
	}

	var loginResp map[string]interface{}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	if token, ok := loginResp["token"].(string); ok {
		lt.authToken = token
		log.Printf("Authentication successful, token obtained")
		return nil
	}

	return fmt.Errorf("no token found in login response")
}

// makeRequest 执行单次请求
func (lt *LoadTester) makeRequest(endpoint TestEndpoint) TestResult {
	start := time.Now()
	result := TestResult{
		Endpoint:  endpoint.Name,
		Timestamp: start,
	}

	// 创建请求
	var body io.Reader
	if endpoint.Body != "" {
		body = bytes.NewBufferString(endpoint.Body)
	}

	req, err := http.NewRequestWithContext(lt.ctx, endpoint.Method, lt.config.BaseURL+endpoint.Path, body)
	if err != nil {
		result.Error = err
		result.ResponseTime = time.Since(start)
		return result
	}

	// 设置请求头
	for key, value := range endpoint.Headers {
		req.Header.Set(key, value)
	}

	// 如果需要认证，添加认证头
	if endpoint.RequireAuth && lt.authToken != "" {
		req.Header.Set("X-Access-Token", lt.authToken)
	}

	// 执行请求
	resp, err := lt.client.Do(req)
	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Error = err
		return result
	//nolint:errcheck
	}
	//nolint:errcheck
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 400

	// 读取响应体（避免连接泄漏）
	_, _ = io.Copy(io.Discard, resp.Body)

	return result
}

// recordResult 记录测试结果
func (lt *LoadTester) recordResult(result TestResult) {
	lt.metrics.mutex.Lock()
	defer lt.metrics.mutex.Unlock()

	atomic.AddInt64(&lt.metrics.TotalRequests, 1)

	if result.Success {
		atomic.AddInt64(&lt.metrics.SuccessfulRequests, 1)
	} else {
		atomic.AddInt64(&lt.metrics.FailedRequests, 1)
	}

	// 记录响应时间
	lt.metrics.ResponseTimes = append(lt.metrics.ResponseTimes, result.ResponseTime)
	lt.metrics.TotalResponseTime += result.ResponseTime

	if lt.metrics.MinResponseTime == 0 || result.ResponseTime < lt.metrics.MinResponseTime {
		lt.metrics.MinResponseTime = result.ResponseTime
	}
	if result.ResponseTime > lt.metrics.MaxResponseTime {
		lt.metrics.MaxResponseTime = result.ResponseTime
	}

	// 记录状态码
	lt.metrics.StatusCodes[result.StatusCode]++

	// 记录错误
	if result.Error != nil {
		errorKey := result.Error.Error()
		lt.metrics.ErrorCounts[errorKey]++
	}
}

// selectEndpoint 根据权重选择端点
func (lt *LoadTester) selectEndpoint() TestEndpoint {
	if len(lt.endpoints) == 0 {
		return TestEndpoint{}
	}

	// 简单轮询选择
	index := int(atomic.LoadInt64(&lt.metrics.TotalRequests)) % len(lt.endpoints)
	return lt.endpoints[index]
}

// worker 工作协程
func (lt *LoadTester) worker(workerID int, requestChan <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-lt.ctx.Done():
			return
		case _, ok := <-requestChan:
			if !ok {
				return
			}

			endpoint := lt.selectEndpoint()
			if endpoint.Name == "" {
				continue
			}

			result := lt.makeRequest(endpoint)
			lt.recordResult(result)

			// 如果配置了请求间隔，等待
			if lt.config.RequestInterval > 0 {
				time.Sleep(lt.config.RequestInterval)
			}
		}
	}
}

// Run 运行负载测试
func (lt *LoadTester) Run() error {
	log.Printf("Starting load test with %d max concurrent users", lt.config.MaxConcurrency)
	log.Printf("Ramp-up duration: %v, Test duration: %v", lt.config.RampUpDuration, lt.config.TestDuration)

	// 设置默认端点
	lt.endpoints = getDefaultEndpoints()

	// 如果启用认证，先进行认证
	if lt.config.EnableAuth {
		log.Println("Authenticating...")
		if err := lt.authenticate(); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	// 过滤端点（如果没有认证，跳过需要认证的端点）
	if !lt.config.EnableAuth {
		var filteredEndpoints []TestEndpoint
		for _, endpoint := range lt.endpoints {
			if !endpoint.RequireAuth {
				filteredEndpoints = append(filteredEndpoints, endpoint)
			}
		}
		lt.endpoints = filteredEndpoints
	}

	log.Printf("Testing %d endpoints", len(lt.endpoints))
	for _, endpoint := range lt.endpoints {
		log.Printf("- %s %s (weight: %d, auth: %v)", endpoint.Method, endpoint.Path, endpoint.Weight, endpoint.RequireAuth)
	}

	lt.metrics.StartTime = time.Now()

	// 创建请求通道
	requestChan := make(chan struct{}, lt.config.MaxConcurrency*2)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < lt.config.MaxConcurrency; i++ {
		wg.Add(1)
		go lt.worker(i, requestChan, &wg)
	}

	// 启动请求生成器
	go lt.requestGenerator(requestChan)

	// 启动实时监控
	go lt.realTimeMonitor()

	// 等待测试完成
	testTimer := time.NewTimer(lt.config.TestDuration)
	<-testTimer.C

	log.Println("Test duration completed, stopping...")
	lt.cancel()

	// 关闭请求通道并等待所有工作协程完成
	close(requestChan)
	wg.Wait()

	lt.metrics.EndTime = time.Now()

	// 生成报告
	return lt.generateReport()
}

// requestGenerator 请求生成器
func (lt *LoadTester) requestGenerator(requestChan chan<- struct{}) {
	rampUpInterval := lt.config.RampUpDuration / time.Duration(lt.config.MaxConcurrency)
	currentConcurrency := 0

	log.Printf("Starting ramp-up with interval: %v", rampUpInterval)

	rampUpTicker := time.NewTicker(rampUpInterval)
	defer rampUpTicker.Stop()

	requestTicker := time.NewTicker(10 * time.Millisecond) // 高频率发送请求
	defer requestTicker.Stop()

	for {
		select {
		case <-lt.ctx.Done():
			return
		case <-rampUpTicker.C:
			if currentConcurrency < lt.config.MaxConcurrency {
				currentConcurrency++
				if currentConcurrency%100 == 0 || currentConcurrency <= 10 {
					log.Printf("Ramping up to %d concurrent users", currentConcurrency)
				}
			}
		case <-requestTicker.C:
			// 发送请求信号，但不超过当前并发数
			for i := 0; i < currentConcurrency && len(requestChan) < cap(requestChan)-1; i++ {
				select {
				case requestChan <- struct{}{}:
				default:
					// 通道满了，跳过
				}
			}
		}
	}
}

// realTimeMonitor 实时监控
func (lt *LoadTester) realTimeMonitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-lt.ctx.Done():
			return
		case <-ticker.C:
			lt.printRealTimeStats()
		}
	}
}

// printRealTimeStats 打印实时统计
func (lt *LoadTester) printRealTimeStats() {
	lt.metrics.mutex.RLock()
	defer lt.metrics.mutex.RUnlock()

	elapsed := time.Since(lt.metrics.StartTime)
	totalReqs := atomic.LoadInt64(&lt.metrics.TotalRequests)
	successReqs := atomic.LoadInt64(&lt.metrics.SuccessfulRequests)
	failedReqs := atomic.LoadInt64(&lt.metrics.FailedRequests)

	rps := float64(totalReqs) / elapsed.Seconds()
	successRate := float64(successReqs) / float64(totalReqs) * 100

	avgResponseTime := time.Duration(0)
	if totalReqs > 0 {
		avgResponseTime = lt.metrics.TotalResponseTime / time.Duration(totalReqs)
	}

	log.Printf("=== Real-time Stats (Elapsed: %v) ===", elapsed.Round(time.Second))
	log.Printf("Total Requests: %d | Success: %d | Failed: %d", totalReqs, successReqs, failedReqs)
	log.Printf("RPS: %.2f | Success Rate: %.2f%% | Avg Response Time: %v", rps, successRate, avgResponseTime)
	log.Printf("Min Response Time: %v | Max Response Time: %v", lt.metrics.MinResponseTime, lt.metrics.MaxResponseTime)
}

// calculatePercentile 计算百分位数
func (lt *LoadTester) calculatePercentile(percentile float64) time.Duration {
	if len(lt.metrics.ResponseTimes) == 0 {
		return 0
	}

	// 复制并排序响应时间
	times := make([]time.Duration, len(lt.metrics.ResponseTimes))
	copy(times, lt.metrics.ResponseTimes)
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})

	index := int(math.Ceil(float64(len(times))*percentile/100)) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(times) {
		index = len(times) - 1
	}

	return times[index]
}

// generateReport 生成测试报告
func (lt *LoadTester) generateReport() error {
	lt.metrics.mutex.RLock()
	defer lt.metrics.mutex.RUnlock()

	totalDuration := lt.metrics.EndTime.Sub(lt.metrics.StartTime)
	totalReqs := atomic.LoadInt64(&lt.metrics.TotalRequests)
	successReqs := atomic.LoadInt64(&lt.metrics.SuccessfulRequests)
	failedReqs := atomic.LoadInt64(&lt.metrics.FailedRequests)

	rps := float64(totalReqs) / totalDuration.Seconds()
	successRate := float64(successReqs) / float64(totalReqs) * 100

	avgResponseTime := time.Duration(0)
	if totalReqs > 0 {
		avgResponseTime = lt.metrics.TotalResponseTime / time.Duration(totalReqs)
	}

	// 计算百分位数
	p50 := lt.calculatePercentile(50)
	p90 := lt.calculatePercentile(90)
	p95 := lt.calculatePercentile(95)
	p99 := lt.calculatePercentile(99)

	// 生成报告
	report := fmt.Sprintf(`
================================================================================
                           LOAD TEST REPORT
================================================================================
Test Configuration:
  Target URL:           %s
  Max Concurrency:      %d
  Ramp-up Duration:     %v
  Test Duration:        %v
  Request Interval:     %v
  Timeout:              %v
  Authentication:       %v

Test Results:
  Total Duration:       %v
  Total Requests:       %d
  Successful Requests:  %d
  Failed Requests:      %d
  Success Rate:         %.2f%%
  Requests per Second:  %.2f

Response Time Statistics:
  Average:              %v
  Minimum:              %v
  Maximum:              %v
  50th Percentile:      %v
  90th Percentile:      %v
  95th Percentile:      %v
  99th Percentile:      %v

Status Code Distribution:`,
		lt.config.BaseURL,
		lt.config.MaxConcurrency,
		lt.config.RampUpDuration,
		lt.config.TestDuration,
		lt.config.RequestInterval,
		lt.config.Timeout,
		lt.config.EnableAuth,
		totalDuration,
		totalReqs,
		successReqs,
		failedReqs,
		successRate,
		rps,
		avgResponseTime,
		lt.metrics.MinResponseTime,
		lt.metrics.MaxResponseTime,
		p50,
		p90,
		p95,
		p99,
	)

	// 添加状态码分布
	for statusCode, count := range lt.metrics.StatusCodes {
		percentage := float64(count) / float64(totalReqs) * 100
		report += fmt.Sprintf("\n  %d: %d (%.2f%%)", statusCode, count, percentage)
	}

	// 添加错误分布
	if len(lt.metrics.ErrorCounts) > 0 {
		report += "\n\nError Distribution:"
		for errorMsg, count := range lt.metrics.ErrorCounts {
			percentage := float64(count) / float64(totalReqs) * 100
			report += fmt.Sprintf("\n  %s: %d (%.2f%%)", errorMsg, count, percentage)
		}
	}

	report += "\n\nRecommendations:"

	// 性能建议
	if successRate < 95 {
		report += "\n  ⚠️  Success rate is below 95%. Consider reducing load or checking server capacity."
	}
	if avgResponseTime > 1*time.Second {
		report += "\n  ⚠️  Average response time is high. Consider optimizing server performance."
	}
	if p95 > 2*time.Second {
		report += "\n  ⚠️  95th percentile response time is high. Some requests are experiencing significant delays."
	}
	if rps < 100 {
		report += "\n  ℹ️  RPS is relatively low. Server may be able to handle higher loads."
	}

	report += "\n\n================================================================================\n"

	// 打印报告
	fmt.Println(report)

	// 保存到文件
	if lt.config.OutputFile != "" {
		if err := os.WriteFile(lt.config.OutputFile, []byte(report), 0644); err != nil {
			log.Printf("Failed to save report to file: %v", err)
		} else {
			log.Printf("Report saved to: %s", lt.config.OutputFile)
		}
	}

	return nil
}

// main 主函数
func main() {
	// 默认配置
	config := TestConfig{
		BaseURL:         "http://go.server.zhongchai.imxiaohu.cn",
		MaxConcurrency:  10000,
		RampUpDuration:  2 * time.Minute,  // 2分钟逐步增加到10000并发
		TestDuration:    5 * time.Minute,  // 测试持续5分钟
		RequestInterval: 0,                // 无间隔，最大压力
		Timeout:         30 * time.Second, // 30秒超时
		EnableAuth:      false,            // 默认不启用认证
		TestUsername:    "",
		TestPassword:    "",
		OutputFile:      "load_test_report.txt",
		EnableTLS:       true,
		SkipTLSVerify:   true, // 跳过TLS验证，适用于测试环境
	}

	// 解析命令行参数
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "quick":
			// 快速测试模式
			config.MaxConcurrency = 100
			config.RampUpDuration = 30 * time.Second
			config.TestDuration = 2 * time.Minute
			log.Println("Running in quick test mode (100 concurrent users, 2 minutes)")
		case "medium":
			// 中等测试模式
			config.MaxConcurrency = 1000
			config.RampUpDuration = 1 * time.Minute
			config.TestDuration = 3 * time.Minute
			log.Println("Running in medium test mode (1000 concurrent users, 3 minutes)")
		case "full":
			// 完整测试模式（默认）
			log.Println("Running in full test mode (10000 concurrent users, 5 minutes)")
		case "auth":
			// 认证测试模式
			config.EnableAuth = true
			config.TestUsername = "test_user"
			config.TestPassword = "test_password"
			config.MaxConcurrency = 500 // 认证测试使用较少并发
			config.RampUpDuration = 1 * time.Minute
			config.TestDuration = 3 * time.Minute
			log.Println("Running in authentication test mode (500 concurrent users, 3 minutes)")
		case "help":
			fmt.Print(`
Load Test Tool for Go Backend Server

Usage: go run main.go [mode]

Available modes:
  quick   - Quick test (100 concurrent users, 2 minutes)
  medium  - Medium test (1000 concurrent users, 3 minutes)
  full    - Full test (10000 concurrent users, 5 minutes) [default]
  auth    - Authentication test (500 concurrent users, 3 minutes)
  help    - Show this help message

Examples:
  go run main.go quick
  go run main.go medium
  go run main.go full
  go run main.go auth

Safety Notes:
  - Start with 'quick' mode to test server response
  - Monitor server resources during testing
  - Use 'full' mode only when confident about server capacity
  - The tool includes gradual ramp-up to avoid overwhelming the server
`)
			return
		default:
			log.Printf("Unknown mode: %s. Using default full test mode.", os.Args[1])
		}
	}

	// 创建负载测试器
	tester := NewLoadTester(config)

	// 显示测试配置
	log.Printf("=== Load Test Configuration ===")
	log.Printf("Target URL: %s", config.BaseURL)
	log.Printf("Max Concurrency: %d", config.MaxConcurrency)
	log.Printf("Ramp-up Duration: %v", config.RampUpDuration)
	log.Printf("Test Duration: %v", config.TestDuration)
	log.Printf("Timeout: %v", config.Timeout)
	log.Printf("Authentication: %v", config.EnableAuth)
	log.Printf("Output File: %s", config.OutputFile)
	log.Printf("===============================")

	// 安全提示
	if config.MaxConcurrency >= 5000 {
		log.Printf("⚠️  WARNING: High concurrency test (%d users). Make sure the target server can handle this load!", config.MaxConcurrency)
		log.Printf("⚠️  Consider starting with a smaller test first.")
		time.Sleep(3 * time.Second)
	}

	// 运行测试
	log.Println("Starting load test...")
	if err := tester.Run(); err != nil {
		log.Fatalf("Load test failed: %v", err)
	}

	log.Println("Load test completed successfully!")
}
