//go:build loadtest_cn
// +build loadtest_cn

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

// 压测配置
type 压测配置 struct {
	目标地址    string        `json:"目标地址"`
	最大并发数   int           `json:"最大并发数"`
	预热时长    time.Duration `json:"预热时长"`
	测试时长    time.Duration `json:"测试时长"`
	请求间隔    time.Duration `json:"请求间隔"`
	超时时间    time.Duration `json:"超时时间"`
	启用认证    bool          `json:"启用认证"`
	测试用户名   string        `json:"测试用户名"`
	测试密码    string        `json:"测试密码"`
	输出文件    string        `json:"输出文件"`
	启用TLS   bool          `json:"启用TLS"`
	跳过TLS验证 bool          `json:"跳过TLS验证"`
}

// 测试接口配置
type 测试接口 struct {
	接口名称 string            `json:"接口名称"`
	请求方法 string            `json:"请求方法"`
	接口路径 string            `json:"接口路径"`
	请求头  map[string]string `json:"请求头"`
	请求体  string            `json:"请求体"`
	需要认证 bool              `json:"需要认证"`
	权重   int               `json:"权重"` // 权重，用于控制请求分布
}

// 单次请求结果
type 请求结果 struct {
	接口名称 string
	状态码  int
	响应时间 time.Duration
	错误信息 error
	时间戳  time.Time
	是否成功 bool
}

// 测试指标
type 测试指标 struct {
	总请求数   int64
	成功请求数  int64
	失败请求数  int64
	总响应时间  time.Duration
	最小响应时间 time.Duration
	最大响应时间 time.Duration
	响应时间列表 []time.Duration
	错误统计   map[string]int64
	状态码统计  map[int]int64
	开始时间   time.Time
	结束时间   time.Time
	读写锁    sync.RWMutex
}

// 压力测试器
type 压力测试器 struct {
	配置      压测配置
	接口列表    []测试接口
	HTTP客户端 *http.Client
	测试指标    *测试指标
	认证令牌    string
	上下文     context.Context
	取消函数    context.CancelFunc
}

// 创建新的压力测试器
func 创建压力测试器(配置 压测配置) *压力测试器 {
	上下文, 取消函数 := context.WithCancel(context.Background())

	// 配置HTTP客户端
	传输配置 := &http.Transport{
		MaxIdleConns:        配置.最大并发数 * 2,
		MaxIdleConnsPerHost: 配置.最大并发数,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
	}

	if 配置.启用TLS && 配置.跳过TLS验证 {
		传输配置.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	客户端 := &http.Client{
		Timeout:   配置.超时时间,
		Transport: 传输配置,
	}

	指标 := &测试指标{
		响应时间列表: make([]time.Duration, 0),
		错误统计:   make(map[string]int64),
		状态码统计:  make(map[int]int64),
	}

	return &压力测试器{
		配置:      配置,
		HTTP客户端: 客户端,
		测试指标:    指标,
		上下文:     上下文,
		取消函数:    取消函数,
	}
}

// 获取默认测试接口配置
func 获取默认接口配置() []测试接口 {
	return []测试接口{
		{
			接口名称: "健康检查",
			请求方法: "GET",
			接口路径: "/api/health/school-server",
			请求头:  map[string]string{"Content-Type": "application/json"},
			需要认证: false,
			权重:   30,
		},
		{
			接口名称: "系统初始化",
			请求方法: "GET",
			接口路径: "/scloud/init",
			请求头:  map[string]string{"Content-Type": "application/json"},
			需要认证: false,
			权重:   20,
		},
		{
			接口名称: "用户信息",
			请求方法: "GET",
			接口路径: "/api/user/info",
			请求头:  map[string]string{"Content-Type": "application/json"},
			需要认证: true,
			权重:   25,
		},
		{
			接口名称: "社团列表",
			请求方法: "GET",
			接口路径: "/api/clubs",
			请求头:  map[string]string{"Content-Type": "application/json"},
			需要认证: true,
			权重:   25,
		},
	}
}

// 用户认证
func (测试器 *压力测试器) 用户认证() error {
	if !测试器.配置.启用认证 {
		return nil
	}

	登录数据 := map[string]string{
		"username": 测试器.配置.测试用户名,
		"password": 测试器.配置.测试密码,
	}

	JSON数据, err := json.Marshal(登录数据)
	if err != nil {
		return fmt.Errorf("序列化登录数据失败: %v", err)
	}

	请求, err := http.NewRequestWithContext(测试器.上下文, "POST", 测试器.配置.目标地址+"/scloud/login", bytes.NewBuffer(JSON数据))
	if err != nil {
		return fmt.Errorf("创建登录请求失败: %v", err)
	}

	请求.Header.Set("Content-Type", "application/json")

	响应, err := 测试器.HTTP客户端.Do(请求)
	if err != nil {
		return fmt.Errorf("登录请求失败: %v", err)
	}
	defer 响应.Body.Close()

	if 响应.StatusCode != http.StatusOK {
		return fmt.Errorf("登录失败，状态码: %d", 响应.StatusCode)
	}

	响应体, err := io.ReadAll(响应.Body)
	if err != nil {
		return fmt.Errorf("读取登录响应失败: %v", err)
	}

	var 登录响应 map[string]interface{}
	if err := json.Unmarshal(响应体, &登录响应); err != nil {
		return fmt.Errorf("解析登录响应失败: %v", err)
	}

	if 令牌, ok := 登录响应["token"].(string); ok {
		测试器.认证令牌 = 令牌
		log.Printf("用户认证成功，已获取令牌")
		return nil
	}

	return fmt.Errorf("登录响应中未找到令牌")
}

// 执行单次请求
func (测试器 *压力测试器) 执行请求(接口 测试接口) 请求结果 {
	开始时间 := time.Now()
	结果 := 请求结果{
		接口名称: 接口.接口名称,
		时间戳:  开始时间,
	}

	// 创建请求
	var 请求体 io.Reader
	if 接口.请求体 != "" {
		请求体 = bytes.NewBufferString(接口.请求体)
	}

	请求, err := http.NewRequestWithContext(测试器.上下文, 接口.请求方法, 测试器.配置.目标地址+接口.接口路径, 请求体)
	if err != nil {
		结果.错误信息 = err
		结果.响应时间 = time.Since(开始时间)
		return 结果
	}

	// 设置请求头
	for 键, 值 := range 接口.请求头 {
		请求.Header.Set(键, 值)
	}

	// 如果需要认证，添加认证头
	if 接口.需要认证 && 测试器.认证令牌 != "" {
		请求.Header.Set("X-Access-Token", 测试器.认证令牌)
	}

	// 执行请求
	响应, err := 测试器.HTTP客户端.Do(请求)
	结果.响应时间 = time.Since(开始时间)

	if err != nil {
		结果.错误信息 = err
		return 结果
	}
	defer 响应.Body.Close()

	结果.状态码 = 响应.StatusCode
	结果.是否成功 = 响应.StatusCode >= 200 && 响应.StatusCode < 400

	// 读取响应体（避免连接泄漏）
	_, _ = io.Copy(io.Discard, 响应.Body)

	return 结果
}

// 记录测试结果
func (测试器 *压力测试器) 记录结果(结果 请求结果) {
	测试器.测试指标.读写锁.Lock()
	defer 测试器.测试指标.读写锁.Unlock()

	atomic.AddInt64(&测试器.测试指标.总请求数, 1)

	if 结果.是否成功 {
		atomic.AddInt64(&测试器.测试指标.成功请求数, 1)
	} else {
		atomic.AddInt64(&测试器.测试指标.失败请求数, 1)
	}

	// 记录响应时间
	测试器.测试指标.响应时间列表 = append(测试器.测试指标.响应时间列表, 结果.响应时间)
	测试器.测试指标.总响应时间 += 结果.响应时间

	if 测试器.测试指标.最小响应时间 == 0 || 结果.响应时间 < 测试器.测试指标.最小响应时间 {
		测试器.测试指标.最小响应时间 = 结果.响应时间
	}
	if 结果.响应时间 > 测试器.测试指标.最大响应时间 {
		测试器.测试指标.最大响应时间 = 结果.响应时间
	}

	// 记录状态码
	测试器.测试指标.状态码统计[结果.状态码]++

	// 记录错误
	if 结果.错误信息 != nil {
		错误键 := 结果.错误信息.Error()
		测试器.测试指标.错误统计[错误键]++
	}
}

// 选择测试接口
func (测试器 *压力测试器) 选择接口() 测试接口 {
	if len(测试器.接口列表) == 0 {
		return 测试接口{}
	}

	// 简单轮询选择
	索引 := int(atomic.LoadInt64(&测试器.测试指标.总请求数)) % len(测试器.接口列表)
	return 测试器.接口列表[索引]
}

// 工作协程
func (测试器 *压力测试器) 工作协程(工作ID int, 请求通道 <-chan struct{}, 等待组 *sync.WaitGroup) {
	defer 等待组.Done()

	for {
		select {
		case <-测试器.上下文.Done():
			return
		case _, ok := <-请求通道:
			if !ok {
				return
			}

			接口 := 测试器.选择接口()
			if 接口.接口名称 == "" {
				continue
			}

			结果 := 测试器.执行请求(接口)
			测试器.记录结果(结果)

			// 如果配置了请求间隔，等待
			if 测试器.配置.请求间隔 > 0 {
				time.Sleep(测试器.配置.请求间隔)
			}
		}
	}
}

// 运行压力测试
func (测试器 *压力测试器) 运行测试() error {
	log.Printf("开始压力测试，最大并发用户数: %d", 测试器.配置.最大并发数)
	log.Printf("预热时长: %v，测试时长: %v", 测试器.配置.预热时长, 测试器.配置.测试时长)

	// 设置默认接口
	测试器.接口列表 = 获取默认接口配置()

	// 如果启用认证，先进行认证
	if 测试器.配置.启用认证 {
		log.Println("正在进行用户认证...")
		if err := 测试器.用户认证(); err != nil {
			return fmt.Errorf("用户认证失败: %v", err)
		}
	}

	// 过滤接口（如果没有认证，跳过需要认证的接口）
	if !测试器.配置.启用认证 {
		var 过滤后接口 []测试接口
		for _, 接口 := range 测试器.接口列表 {
			if !接口.需要认证 {
				过滤后接口 = append(过滤后接口, 接口)
			}
		}
		测试器.接口列表 = 过滤后接口
	}

	log.Printf("测试接口数量: %d", len(测试器.接口列表))
	for _, 接口 := range 测试器.接口列表 {
		log.Printf("- %s %s (权重: %d, 需要认证: %v)", 接口.请求方法, 接口.接口路径, 接口.权重, 接口.需要认证)
	}

	测试器.测试指标.开始时间 = time.Now()

	// 创建请求通道
	请求通道 := make(chan struct{}, 测试器.配置.最大并发数*2)

	// 启动工作协程
	var 等待组 sync.WaitGroup
	for i := 0; i < 测试器.配置.最大并发数; i++ {
		等待组.Add(1)
		go 测试器.工作协程(i, 请求通道, &等待组)
	}

	// 启动请求生成器
	go 测试器.请求生成器(请求通道)

	// 启动实时监控
	go 测试器.实时监控()

	// 等待测试完成
	测试定时器 := time.NewTimer(测试器.配置.测试时长)
	<-测试定时器.C

	log.Println("测试时长已完成，正在停止...")
	测试器.取消函数()

	// 关闭请求通道并等待所有工作协程完成
	close(请求通道)
	等待组.Wait()

	测试器.测试指标.结束时间 = time.Now()

	// 生成报告
	return 测试器.生成报告()
}

// 请求生成器
func (测试器 *压力测试器) 请求生成器(请求通道 chan<- struct{}) {
	预热间隔 := 测试器.配置.预热时长 / time.Duration(测试器.配置.最大并发数)
	当前并发数 := 0

	log.Printf("开始预热，预热间隔: %v", 预热间隔)

	预热定时器 := time.NewTicker(预热间隔)
	defer 预热定时器.Stop()

	请求定时器 := time.NewTicker(10 * time.Millisecond) // 高频率发送请求
	defer 请求定时器.Stop()

	for {
		select {
		case <-测试器.上下文.Done():
			return
		case <-预热定时器.C:
			if 当前并发数 < 测试器.配置.最大并发数 {
				当前并发数++
				if 当前并发数%100 == 0 || 当前并发数 <= 10 {
					log.Printf("预热至 %d 并发用户", 当前并发数)
				}
			}
		case <-请求定时器.C:
			// 发送请求信号，但不超过当前并发数
			for i := 0; i < 当前并发数 && len(请求通道) < cap(请求通道)-1; i++ {
				select {
				case 请求通道 <- struct{}{}:
				default:
					// 通道满了，跳过
				}
			}
		}
	}
}

// 实时监控
func (测试器 *压力测试器) 实时监控() {
	定时器 := time.NewTicker(10 * time.Second)
	defer 定时器.Stop()

	for {
		select {
		case <-测试器.上下文.Done():
			return
		case <-定时器.C:
			测试器.打印实时统计()
		}
	}
}

// 打印实时统计
func (测试器 *压力测试器) 打印实时统计() {
	测试器.测试指标.读写锁.RLock()
	defer 测试器.测试指标.读写锁.RUnlock()

	已用时间 := time.Since(测试器.测试指标.开始时间)
	总请求数 := atomic.LoadInt64(&测试器.测试指标.总请求数)
	成功请求数 := atomic.LoadInt64(&测试器.测试指标.成功请求数)
	失败请求数 := atomic.LoadInt64(&测试器.测试指标.失败请求数)

	每秒请求数 := float64(总请求数) / 已用时间.Seconds()
	成功率 := float64(成功请求数) / float64(总请求数) * 100

	平均响应时间 := time.Duration(0)
	if 总请求数 > 0 {
		平均响应时间 = 测试器.测试指标.总响应时间 / time.Duration(总请求数)
	}

	log.Printf("=== 实时统计 (已用时间: %v) ===", 已用时间.Round(time.Second))
	log.Printf("总请求数: %d | 成功: %d | 失败: %d", 总请求数, 成功请求数, 失败请求数)
	log.Printf("每秒请求数: %.2f | 成功率: %.2f%% | 平均响应时间: %v", 每秒请求数, 成功率, 平均响应时间)
	log.Printf("最小响应时间: %v | 最大响应时间: %v", 测试器.测试指标.最小响应时间, 测试器.测试指标.最大响应时间)
}

// 计算百分位数
func (测试器 *压力测试器) 计算百分位数(百分位 float64) time.Duration {
	if len(测试器.测试指标.响应时间列表) == 0 {
		return 0
	}

	// 复制并排序响应时间
	时间列表 := make([]time.Duration, len(测试器.测试指标.响应时间列表))
	copy(时间列表, 测试器.测试指标.响应时间列表)
	sort.Slice(时间列表, func(i, j int) bool {
		return 时间列表[i] < 时间列表[j]
	})

	索引 := int(math.Ceil(float64(len(时间列表))*百分位/100)) - 1
	if 索引 < 0 {
		索引 = 0
	}
	if 索引 >= len(时间列表) {
		索引 = len(时间列表) - 1
	}

	return 时间列表[索引]
}

// 生成测试报告
func (测试器 *压力测试器) 生成报告() error {
	测试器.测试指标.读写锁.RLock()
	defer 测试器.测试指标.读写锁.RUnlock()

	总时长 := 测试器.测试指标.结束时间.Sub(测试器.测试指标.开始时间)
	总请求数 := atomic.LoadInt64(&测试器.测试指标.总请求数)
	成功请求数 := atomic.LoadInt64(&测试器.测试指标.成功请求数)
	失败请求数 := atomic.LoadInt64(&测试器.测试指标.失败请求数)

	每秒请求数 := float64(总请求数) / 总时长.Seconds()
	成功率 := float64(成功请求数) / float64(总请求数) * 100

	平均响应时间 := time.Duration(0)
	if 总请求数 > 0 {
		平均响应时间 = 测试器.测试指标.总响应时间 / time.Duration(总请求数)
	}

	// 计算百分位数
	p50 := 测试器.计算百分位数(50)
	p90 := 测试器.计算百分位数(90)
	p95 := 测试器.计算百分位数(95)
	p99 := 测试器.计算百分位数(99)

	// 生成报告
	报告 := fmt.Sprintf(`
================================================================================
                           压力测试报告
================================================================================
测试配置:
  目标地址:             %s
  最大并发数:           %d
  预热时长:             %v
  测试时长:             %v
  请求间隔:             %v
  超时时间:             %v
  启用认证:             %v

测试结果:
  总测试时长:           %v
  总请求数:             %d
  成功请求数:           %d
  失败请求数:           %d
  成功率:               %.2f%%
  每秒请求数:           %.2f

响应时间统计:
  平均响应时间:         %v
  最小响应时间:         %v
  最大响应时间:         %v
  50%% 百分位:          %v
  90%% 百分位:          %v
  95%% 百分位:          %v
  99%% 百分位:          %v

状态码分布:`,
		测试器.配置.目标地址,
		测试器.配置.最大并发数,
		测试器.配置.预热时长,
		测试器.配置.测试时长,
		测试器.配置.请求间隔,
		测试器.配置.超时时间,
		测试器.配置.启用认证,
		总时长,
		总请求数,
		成功请求数,
		失败请求数,
		成功率,
		每秒请求数,
		平均响应时间,
		测试器.测试指标.最小响应时间,
		测试器.测试指标.最大响应时间,
		p50,
		p90,
		p95,
		p99,
	)

	// 添加状态码分布
	for 状态码, 数量 := range 测试器.测试指标.状态码统计 {
		百分比 := float64(数量) / float64(总请求数) * 100
		报告 += fmt.Sprintf("\n  %d: %d (%.2f%%)", 状态码, 数量, 百分比)
	}

	// 添加错误分布
	if len(测试器.测试指标.错误统计) > 0 {
		报告 += "\n\n错误分布:"
		for 错误信息, 数量 := range 测试器.测试指标.错误统计 {
			百分比 := float64(数量) / float64(总请求数) * 100
			报告 += fmt.Sprintf("\n  %s: %d (%.2f%%)", 错误信息, 数量, 百分比)
		}
	}

	报告 += "\n\n性能建议:"

	// 性能建议
	if 成功率 < 95 {
		报告 += "\n  ⚠️  成功率低于95%，建议减少负载或检查服务器容量。"
	}
	if 平均响应时间 > 1*time.Second {
		报告 += "\n  ⚠️  平均响应时间较高，建议优化服务器性能。"
	}
	if p95 > 2*time.Second {
		报告 += "\n  ⚠️  95%百分位响应时间较高，部分请求存在明显延迟。"
	}
	if 每秒请求数 < 100 {
		报告 += "\n  ℹ️  每秒请求数相对较低，服务器可能能够处理更高的负载。"
	}

	报告 += "\n\n================================================================================\n"

	// 打印报告
	fmt.Println(报告)

	// 保存到文件
	if 测试器.配置.输出文件 != "" {
		if err := os.WriteFile(测试器.配置.输出文件, []byte(报告), 0644); err != nil {
			log.Printf("保存报告到文件失败: %v", err)
		} else {
			log.Printf("报告已保存到: %s", 测试器.配置.输出文件)
		}
	}

	return nil
}

// 获取服务器地址
func 获取服务器地址() string {
	// 优先级：命令行参数 > 环境变量 > 默认值

	// 检查环境变量
	if 地址 := os.Getenv("压测服务器地址"); 地址 != "" {
		return 地址
	}
	if 地址 := os.Getenv("LOAD_TEST_URL"); 地址 != "" {
		return 地址
	}

	// 默认地址
	return "https://go.server.zhongchai.imxiaohu.cn"
}

// 主函数
func main() {
	// 获取服务器地址
	服务器地址 := 获取服务器地址()

	// 检查命令行参数中是否有服务器地址
	for i, arg := range os.Args {
		if arg == "-u" || arg == "--url" {
			if i+1 < len(os.Args) {
				服务器地址 = os.Args[i+1]
				// 从参数列表中移除URL参数
				os.Args = append(os.Args[:i], os.Args[i+2:]...)
				break
			}
		}
	}

	// 默认配置
	配置 := 压测配置{
		目标地址:    服务器地址,
		最大并发数:   10000,
		预热时长:    2 * time.Minute,  // 2分钟逐步增加到10000并发
		测试时长:    5 * time.Minute,  // 测试持续5分钟
		请求间隔:    0,                // 无间隔，最大压力
		超时时间:    30 * time.Second, // 30秒超时
		启用认证:    false,            // 默认不启用认证
		测试用户名:   "",
		测试密码:    "",
		输出文件:    "压测报告.txt",
		启用TLS:   true,
		跳过TLS验证: true, // 跳过TLS验证，适用于测试环境
	}

	// 解析命令行参数
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "快速", "quick":
			// 快速测试模式
			配置.最大并发数 = 100
			配置.预热时长 = 30 * time.Second
			配置.测试时长 = 2 * time.Minute
			log.Println("运行快速测试模式 (100并发用户, 2分钟)")
		case "中等", "medium":
			// 中等测试模式
			配置.最大并发数 = 1000
			配置.预热时长 = 1 * time.Minute
			配置.测试时长 = 3 * time.Minute
			log.Println("运行中等测试模式 (1000并发用户, 3分钟)")
		case "完整", "full":
			// 完整测试模式（默认）
			log.Println("运行完整测试模式 (10000并发用户, 5分钟)")
		case "认证", "auth":
			// 认证测试模式
			配置.启用认证 = true
			配置.测试用户名 = "test_user"
			配置.测试密码 = "test_password"
			配置.最大并发数 = 500 // 认证测试使用较少并发
			配置.预热时长 = 1 * time.Minute
			配置.测试时长 = 3 * time.Minute
			log.Println("运行认证测试模式 (500并发用户, 3分钟)")
		case "帮助", "help":
			fmt.Println(`
Go后端服务器压力测试工具

用法: go run 压测工具.go [选项] [模式]

可用模式:
  快速/quick   - 快速测试 (100并发用户, 2分钟)
  中等/medium  - 中等测试 (1000并发用户, 3分钟)
  完整/full    - 完整测试 (10000并发用户, 5分钟) [默认]
  认证/auth    - 认证测试 (500并发用户, 3分钟)
  帮助/help    - 显示此帮助信息

选项:
  -u, --url URL    指定目标服务器地址

示例:
  go run 压测工具.go 快速
  go run 压测工具.go 中等
  go run 压测工具.go 完整
  go run 压测工具.go 认证
  go run 压测工具.go -u http://localhost:2333 快速
  go run 压测工具.go --url https://my-server.com 中等

环境变量:
  压测服务器地址      设置目标服务器地址
  LOAD_TEST_URL     设置目标服务器地址（英文）

配置优先级: 命令行参数 > 环境变量 > 默认值

安全提示:
  - 首先使用'快速'模式测试服务器响应
  - 测试期间监控服务器资源
  - 只有在确信服务器容量时才使用'完整'模式
  - 工具包含渐进式预热以避免服务器过载
`)
			return
		default:
			log.Printf("未知模式: %s. 使用默认完整测试模式.", os.Args[1])
		}
	}

	// 创建压力测试器
	测试器 := 创建压力测试器(配置)

	// 显示测试配置
	log.Printf("=== 压力测试配置 ===")
	log.Printf("目标地址: %s", 配置.目标地址)
	log.Printf("最大并发数: %d", 配置.最大并发数)
	log.Printf("预热时长: %v", 配置.预热时长)
	log.Printf("测试时长: %v", 配置.测试时长)
	log.Printf("超时时间: %v", 配置.超时时间)
	log.Printf("启用认证: %v", 配置.启用认证)
	log.Printf("输出文件: %s", 配置.输出文件)
	log.Printf("===================")

	// 安全提示
	if 配置.最大并发数 >= 5000 {
		log.Printf("⚠️  警告: 高并发测试 (%d用户)。请确保目标服务器能够处理此负载!", 配置.最大并发数)
		log.Printf("⚠️  建议先进行较小规模的测试。")
		time.Sleep(3 * time.Second)
	}

	// 运行测试
	log.Println("开始压力测试...")
	if err := 测试器.运行测试(); err != nil {
		log.Fatalf("压力测试失败: %v", err)
	}

	log.Println("压力测试成功完成!")
}
