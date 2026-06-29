package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// BatchScoreCacheService 批量成绩缓存服务
type BatchScoreCacheService struct {
	proxyClient       *utils.ProxyClient
	schoolAPIService  *SchoolAPIService
	smartCacheService *SmartCacheService
}

// BatchCacheResult 批量缓存结果
type BatchCacheResult struct {
	TotalSemesters  int                   `json:"totalSemesters"`
	SuccessCount    int                   `json:"successCount"`
	FailureCount    int                   `json:"failureCount"`
	CachedSemesters []string              `json:"cachedSemesters"`
	FailedSemesters []string              `json:"failedSemesters"`
	CacheDetails    []SemesterCacheDetail `json:"cacheDetails"`
	StartTime       time.Time             `json:"startTime"`
	EndTime         time.Time             `json:"endTime"`
	Duration        string                `json:"duration"`
}

// SemesterCacheDetail 学期缓存详情
type SemesterCacheDetail struct {
	SemesterName string    `json:"semesterName"`
	Success      bool      `json:"success"`
	ScoreCount   int       `json:"scoreCount"`
	CacheKey     string    `json:"cacheKey"`
	Error        string    `json:"error,omitempty"`
	CachedAt     time.Time `json:"cachedAt"`
}

// NewBatchScoreCacheService 创建批量成绩缓存服务实例
func NewBatchScoreCacheService() *BatchScoreCacheService {
	return &BatchScoreCacheService{
		proxyClient:       utils.NewProxyClient(),
		schoolAPIService:  NewSchoolAPIService(),
		smartCacheService: GetSmartCacheService(),
	}
}

// CacheAllSemesterScores 批量缓存所有学期的成绩数据
func (s *BatchScoreCacheService) CacheAllSemesterScores(user *models.User) (*BatchCacheResult, error) {
	startTime := time.Now()
	log.Printf("开始为用户 %d 批量缓存成绩数据", user.ID)

	result := &BatchCacheResult{
		StartTime:       startTime,
		CachedSemesters: make([]string, 0),
		FailedSemesters: make([]string, 0),
		CacheDetails:    make([]SemesterCacheDetail, 0),
	}

	// 1. 获取学期列表
	semesters, err := s.getSemesterList(user)
	if err != nil {
		return nil, fmt.Errorf("获取学期列表失败: %w", err)
	}

	result.TotalSemesters = len(semesters)
	log.Printf("获取到 %d 个学期", len(semesters))

	// 2. 并发缓存各学期成绩数据
	var wg sync.WaitGroup
	var mutex sync.Mutex
	semesterChan := make(chan SemesterInfo, len(semesters))

	// 启动工作协程池（限制并发数为3，避免对学校服务器造成过大压力）
	workerCount := 3
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for semester := range semesterChan {
				detail := s.cacheSemesterScores(user, semester.Name)

				mutex.Lock()
				result.CacheDetails = append(result.CacheDetails, detail)
				if detail.Success {
					result.SuccessCount++
					result.CachedSemesters = append(result.CachedSemesters, semester.Name)
				} else {
					result.FailureCount++
					result.FailedSemesters = append(result.FailedSemesters, semester.Name)
				}
				mutex.Unlock()
			}
		}()
	}

	// 发送学期任务到通道
	for _, semester := range semesters {
		semesterChan <- semester
	}
	close(semesterChan)

	// 等待所有任务完成
	wg.Wait()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	log.Printf("批量缓存完成，成功: %d, 失败: %d, 耗时: %s",
		result.SuccessCount, result.FailureCount, result.Duration)

	// 3. 创建通知日志
	s.createBatchCacheNotificationLog(user.ID, result)

	return result, nil
}

// createBatchCacheNotificationLog 创建批量缓存通知日志
func (s *BatchScoreCacheService) createBatchCacheNotificationLog(userID uint, result *BatchCacheResult) {
	status := models.SendStatusSuccess
	content := fmt.Sprintf("批量缓存完成，总学期数: %d, 成功: %d, 失败: %d, 耗时: %s",
		result.TotalSemesters, result.SuccessCount, result.FailureCount, result.Duration)

	if result.FailureCount > 0 {
		status = models.SendStatusFailed
		// 构建更详细的失败信息
		failedDetails := s.buildFailedDetails(result.CacheDetails)
		content = fmt.Sprintf("批量缓存完成，总学期数: %d, 成功: %d, 失败: %d, 耗时: %s",
			result.TotalSemesters, result.SuccessCount, result.FailureCount, result.Duration)
		content += fmt.Sprintf("\n失败学期: %s", failedDetails)
	}

	notificationLog := &models.NotificationLog{
		UserID:    userID,
		Channel:   models.ChannelSystem,
		Type:      models.NotificationTypeBatchCache,
		Title:     "批量成绩缓存",
		Content:   content,
		Recipient: models.ChannelSystem,
		Status:    status,
		ExtraData: s.buildExtraData(result),
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建批量缓存通知日志失败: %v", err)
	}
}

// buildFailedDetails 构建失败详情
func (s *BatchScoreCacheService) buildFailedDetails(details []SemesterCacheDetail) string {
	var failed []string
	for _, d := range details {
		if !d.Success {
			failed = append(failed, d.SemesterName)
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(failed, " "))
}

// buildExtraData 构建额外数据（JSON格式）
func (s *BatchScoreCacheService) buildExtraData(result *BatchCacheResult) string {
	extra, _ := json.Marshal(map[string]interface{}{
		"totalSemesters":  result.TotalSemesters,
		"successCount":    result.SuccessCount,
		"failureCount":    result.FailureCount,
		"cachedSemesters":  result.CachedSemesters,
		"failedSemesters":  result.FailedSemesters,
		"duration":        result.Duration,
		"cacheDetails":     result.CacheDetails,
	})
	return string(extra)
}

// getSemesterList 获取学期列表
func (s *BatchScoreCacheService) getSemesterList(user *models.User) ([]SemesterInfo, error) {
	// 注意：学期接口实际路径是 /scloudoa/scs/course/tCourseScore/getSemester（不含 /api/m/ 前缀）
	cacheReq := &CacheRequest{
		APIPath:   "/scloudoa/scs/course/tCourseScore/getSemester",
		UserID:    "", // 学期列表是公共数据
		Params:    map[string]interface{}{},
		IPAddress: "",
		UserAgent: "",
	}

	// 定义数据获取器
	dataFetcher := func() (interface{}, error) {
		return s.schoolAPIService.GetAvailableSemesters(user)
	}

	// 获取或设置缓存
	cacheResp, err := s.smartCacheService.GetOrSetCache(cacheReq, dataFetcher)
	if err != nil {
		return nil, err
	}

	// 解析缓存数据
	var semesters []SemesterInfo

	// 尝试将缓存数据转换为JSON字节
	var dataBytes []byte

	if dataStr, ok := cacheResp.Data.(string); ok {
		dataBytes = []byte(dataStr)
	} else {
		dataBytes, err = json.Marshal(cacheResp.Data)
		if err != nil {
			// 如果序列化失败，直接调用API
			return s.schoolAPIService.GetAvailableSemesters(user)
		}
	}

	if err := json.Unmarshal(dataBytes, &semesters); err != nil {
		// 如果解析失败，直接调用API
		return s.schoolAPIService.GetAvailableSemesters(user)
	}

	return semesters, nil
}

// cacheSemesterScores 缓存单个学期的成绩数据
func (s *BatchScoreCacheService) cacheSemesterScores(user *models.User, semesterName string) SemesterCacheDetail {
	detail := SemesterCacheDetail{
		SemesterName: semesterName,
		CachedAt:     time.Now(),
	}

	log.Printf("开始缓存学期 %s 的成绩数据", semesterName)

	// 构建缓存请求
	// 注意：缓存键应与 fetchScoreDataFromSchool 保持一致，使用相同的 API 路径和参数
	params := map[string]interface{}{
		"current":         "1",
		"size":            "100",
		"currentSemester": semesterName,
	}

	cacheReq := &CacheRequest{
		APIPath:   "/scloudoa/scs/course/tCourseScore/getCourseScore",
		UserID:    fmt.Sprintf("%d", user.ID),
		Params:    params,
		IPAddress: "",
		UserAgent: "",
	}

	// 构建缓存键
	requestHash := models.GenerateRequestHash(params)
	keyBuilder := &models.CacheKeyBuilder{
		APIPath:     cacheReq.APIPath,
		UserID:      cacheReq.UserID,
		RequestHash: requestHash,
	}
	detail.CacheKey = keyBuilder.BuildKey()

	// 定义数据获取器
	dataFetcher := func() (interface{}, error) {
		return s.fetchScoreDataFromSchool(user, semesterName)
	}

	// 获取或设置缓存
	cacheResp, err := s.smartCacheService.GetOrSetCache(cacheReq, dataFetcher)
	if err != nil {
		detail.Success = false
		detail.Error = err.Error()
		log.Printf("缓存学期 %s 成绩数据失败: %v", semesterName, err)
		return detail
	}

	// 统计成绩数量
	detail.ScoreCount = s.countScoresFromResponse(cacheResp.Data)
	detail.Success = true

	log.Printf("成功缓存学期 %s 的成绩数据，共 %d 条记录", semesterName, detail.ScoreCount)
	return detail
}

// fetchScoreDataFromSchool 从学校服务器获取成绩数据
func (s *BatchScoreCacheService) fetchScoreDataFromSchool(user *models.User, semesterName string) (interface{}, error) {
	// 构建请求参数
	// 注意：学校成绩接口使用 currentSemester 参数，而不是 semesterName
	params := url.Values{}
	params.Set("current", "1")
	params.Set("size", "100") // 增加数量以覆盖更多课程
	params.Set("currentSemester", semesterName)

	// 调用学校服务器API（使用正确的API路径 getCourseScore）
	body, err := s.proxyClient.ProxyRequestWithAutoLogin(user, "GET",
		"/scloudoa/scs/course/tCourseScore/getCourseScore", params)
	if err != nil {
		return nil, fmt.Errorf("请求学校服务器失败: %w", err)
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应是否成功
	if success, ok := response["success"].(bool); !ok || !success {
		if msg, ok := response["message"].(string); ok {
			return nil, fmt.Errorf("获取成绩失败: %s", msg)
		}
		return nil, fmt.Errorf("获取成绩失败")
	}

	return response, nil
}

// countScoresFromResponse 从响应中统计成绩数量
func (s *BatchScoreCacheService) countScoresFromResponse(data interface{}) int {
	if dataMap, ok := data.(map[string]interface{}); ok {
		if result, ok := dataMap["result"].(map[string]interface{}); ok {
			if records, ok := result["records"].([]interface{}); ok {
				return len(records)
			}
		}
	}
	return 0
}
