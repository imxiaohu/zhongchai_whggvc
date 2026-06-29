package services

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// ScoreCheckService 成绩检查服务
type ScoreCheckService struct {
	enabled          bool
	schoolAPIService *SchoolAPIService
}

var scoreCheckService *ScoreCheckService

// InitScoreCheckService 初始化成绩检查服务
func InitScoreCheckService() {
	scoreCheckService = NewScoreCheckService()
}

// GetScoreCheckService 获取成绩检查服务实例
func GetScoreCheckService() *ScoreCheckService {
	return scoreCheckService
}

// NewScoreCheckService 创建新的成绩检查服务
func NewScoreCheckService() *ScoreCheckService {
	service := &ScoreCheckService{
		enabled:          true,
		schoolAPIService: NewSchoolAPIService(),
	}

	log.Printf("成绩检查服务初始化完成")
	return service
}

// IsEnabled 检查成绩检查服务是否启用
func (s *ScoreCheckService) IsEnabled() bool {
	return s.enabled
}

// CheckUserScores 检查用户成绩更新（修复版：先 diff 再更新，解决版本控制时序 bug）
func (s *ScoreCheckService) CheckUserScores(userID uint) ([]ScoreUpdate, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("成绩检查服务未启用")
	}

	user, err := models.FindUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	if user.Username == "" || user.Password == "" {
		return nil, fmt.Errorf("用户未绑定学校账号")
	}

	channel, err := models.GetNotificationChannelByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户通知配置失败: %w", err)
	}

	log.Printf("[ScoreCheck] 开始为用户%d检查成绩更新，检测学期配置: %s", userID, channel.ScoreCheckSemester)

	semestersToCheck, err := s.getSemestersToCheck(user, channel.ScoreCheckSemester)
	if err != nil {
		return nil, fmt.Errorf("获取检测学期失败: %w", err)
	}

	var allScoreUpdates []ScoreUpdate

	for _, semester := range semestersToCheck {
		// 检查该学期是否曾有过快照（用于区分首次同步 vs 成绩更新）
		hasPrevious, err := models.GetPreviousScoreSnapshotsByUserIDAndSemester(userID, semester)
		if err != nil {
			log.Printf("[ScoreCheck] 检查用户%d学期%s历史快照失败: %v", userID, semester, err)
		}

		// 从学校服务器获取最新成绩
		currentScores, err := s.fetchCurrentScoresBySemesterWithCache(user, semester)
		if err != nil {
			log.Printf("[ScoreCheck] 获取学期%s成绩失败: %v", semester, err)
			continue
		}

		// 获取上一版本的快照（用于 diff）
		previousSnapshots, err := models.GetScoreSnapshotsByUserIDAndSemesterWithVersion(userID, semester, models.ScoreVersionPrevious)
		if err != nil {
			previousSnapshots = []models.ScoreSnapshot{}
		}

		log.Printf("[ScoreCheck] 学期%s: 当前%d条, 历史快照%d条, 首次同步=%v",
			semester, len(currentScores), len(previousSnapshots), !hasPrevious)

		// 核心修复：先 diff，再决定是否更新快照
		scoreUpdates := s.compareScoresWithSemester(currentScores, previousSnapshots, semester, !hasPrevious)

		if len(scoreUpdates) > 0 {
			// 过滤掉"无效更新"：首次同步时所有课程都被标记为新增，但我们不希望发通知
			// 仅当 isFirstSync=false（即之前有过快照）时才加入通知列表
			var realUpdates []ScoreUpdate
			for _, u := range scoreUpdates {
				if !u.IsFirstSync {
					realUpdates = append(realUpdates, u)
				}
			}
			if len(realUpdates) > 0 {
				log.Printf("[ScoreCheck] 学期%s 检测到%d条真实成绩更新（已过滤首次同步）", semester, len(realUpdates))
				allScoreUpdates = append(allScoreUpdates, realUpdates...)
			} else {
				log.Printf("[ScoreCheck] 学期%s 检测到%d条变化但均为首次同步，不发送通知", semester, len(scoreUpdates))
			}
		}

		// 修复后的版本控制：先 diff 确定变化，再原地更新 current 快照
		if err := s.updateSnapshotsInPlace(userID, semester, currentScores); err != nil {
			log.Printf("[ScoreCheck] 更新学期%s成绩快照失败: %v", semester, err)
		}
	}

	log.Printf("[ScoreCheck] 用户%d成绩检查完成，共检测到%d条真实变化", userID, len(allScoreUpdates))
	return allScoreUpdates, nil
}

// fetchCurrentScoresBySemesterWithCache 获取学期成绩数据
func (s *ScoreCheckService) fetchCurrentScoresBySemesterWithCache(user *models.User, semester string) ([]models.ScoreSnapshot, error) {
	log.Printf("[ScoreCheck] 获取学期%s的成绩数据", semester)
	return s.fetchCurrentScoresBySemester(user, semester)
}

// fetchScoreDataFromSchool 从学校服务器获取成绩数据
func (s *ScoreCheckService) fetchScoreDataFromSchool(user *models.User, semester string) (interface{}, error) {
	log.Printf("[ScoreCheck] 从学校服务器获取学期%s的成绩数据", semester)

	// 直接调用学校服务器API，获取原始响应数据
	proxyClient := utils.NewProxyClient()

	// 构建请求参数
	params := url.Values{}
	params.Set("currentSemester", semester) // 使用正确的参数名
	params.Set("current", "1")
	params.Set("size", "100") // 增加单页数量以覆盖更多课程

	// 调用学校服务器API（使用正确的API路径）
	body, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET",
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
		message, _ := response["message"].(string)
		if message == "" {
			message = "No message available"
		}
		return nil, fmt.Errorf("获取成绩失败: %s", message)
	}

	log.Printf("[ScoreCheck] 成功获取学期%s的成绩数据", semester)
	return response, nil
}

// parseScoreDataToSnapshots 解析成绩数据为ScoreSnapshot格式
func (s *ScoreCheckService) parseScoreDataToSnapshots(data interface{}, userID uint, semester string) ([]models.ScoreSnapshot, error) {
	var snapshots []models.ScoreSnapshot

	// 尝试解析数据结构
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("数据格式错误，不是map类型")
	}

	log.Printf("[ScoreCheck] 开始解析成绩数据，数据结构: %+v", dataMap)

	// 尝试多种数据格式
	records, err := s.extractRecordsFromData(dataMap)
	if err != nil {
		return nil, fmt.Errorf("提取records失败: %w", err)
	}

	log.Printf("[ScoreCheck] 成功提取到%d条成绩记录", len(records))

	// 转换每条记录为ScoreSnapshot
	for _, record := range records {
		recordMap, ok := record.(map[string]interface{})
		if !ok {
			continue
		}

		// 获取课程基本信息
		courseName := s.getStringValue(recordMap, "courseName")
		courseCode := s.generateCourseCode(recordMap) // 生成课程代码
		credit := s.getFloatValue(recordMap, "credit")
		gpa := s.getFloatValue(recordMap, "getPoint") // 使用getPoint作为GPA

		// 为每种成绩类型创建快照
		scoreTypes := s.extractScoreTypes(recordMap)
		for scoreType, score := range scoreTypes {
			if score != "" && score != "0" { // 只记录有效成绩
				snapshot := models.ScoreSnapshot{
					UserID:     userID,
					Semester:   semester,
					CourseCode: courseCode,
					CourseName: courseName,
					ScoreType:  scoreType,
					Score:      score,
					Credit:     credit,
					GPA:        gpa,
					Rank:       "", // API中没有排名信息
				}

				// 计算校验和
				snapshot.CheckSum = s.calculateCheckSum(&snapshot)
				snapshots = append(snapshots, snapshot)
			}
		}
	}

	return snapshots, nil
}

// extractRecordsFromData 从数据中提取records数组，支持多种格式
func (s *ScoreCheckService) extractRecordsFromData(dataMap map[string]interface{}) ([]interface{}, error) {
	// 格式1: 标准格式 data.result.records
	if result, ok := dataMap["result"].(map[string]interface{}); ok {
		if records, ok := result["records"].([]interface{}); ok {
			log.Printf("[ScoreCheck] 使用标准格式: data.result.records")
			return records, nil
		}

		// 格式2: 嵌套格式 data.result.result.records
		if nestedResult, ok := result["result"].(map[string]interface{}); ok {
			if records, ok := nestedResult["records"].([]interface{}); ok {
				log.Printf("[ScoreCheck] 使用嵌套格式: data.result.result.records")
				return records, nil
			}
		}
	}

	// 格式3: 直接在data下的records
	if records, ok := dataMap["records"].([]interface{}); ok {
		log.Printf("[ScoreCheck] 使用直接格式: data.records")
		return records, nil
	}

	// 格式4: data本身就是数组
	if records, ok := dataMap["data"].([]interface{}); ok {
		log.Printf("[ScoreCheck] 使用data数组格式: data.data")
		return records, nil
	}

	// 格式5: 检查是否有其他可能的字段名
	for key, value := range dataMap {
		if records, ok := value.([]interface{}); ok && len(records) > 0 {
			// 检查第一个元素是否像成绩记录
			if len(records) > 0 {
				if record, ok := records[0].(map[string]interface{}); ok {
					if _, hasCourseName := record["courseName"]; hasCourseName {
						log.Printf("[ScoreCheck] 使用发现的格式: data.%s", key)
						return records, nil
					}
				}
			}
		}
	}

	// 打印数据结构以便调试
	log.Printf("[ScoreCheck] 无法识别的数据格式，数据结构: %+v", dataMap)

	return nil, fmt.Errorf("无法从数据中提取records数组，支持的格式: result.records, result.result.records, records, data")
}

// generateCourseCode 生成课程代码（如果API中没有提供）
func (s *ScoreCheckService) generateCourseCode(recordMap map[string]interface{}) string {
	// 优先使用courseNumber
	if courseNumber := s.getStringValue(recordMap, "courseNumber"); courseNumber != "" && courseNumber != "<nil>" {
		return courseNumber
	}

	// 使用id作为课程代码，处理科学计数法格式
	if idValue, ok := recordMap["id"]; ok && idValue != nil {
		switch v := idValue.(type) {
		case float64:
			return fmt.Sprintf("COURSE_%.0f", v)
		case int:
			return fmt.Sprintf("COURSE_%d", v)
		case string:
			if v != "" && v != "<nil>" {
				return fmt.Sprintf("COURSE_%s", v)
			}
		}
	}

	// 使用课程名称的哈希作为代码
	courseName := s.getStringValue(recordMap, "courseName")
	if courseName != "" {
		hash := fmt.Sprintf("%x", md5.Sum([]byte(courseName)))
		return fmt.Sprintf("COURSE_%s", hash[:8])
	}

	return "UNKNOWN"
}

// extractScoreTypes 从记录中提取各种成绩类型
func (s *ScoreCheckService) extractScoreTypes(recordMap map[string]interface{}) map[string]string {
	scoreTypes := make(map[string]string)

	// 平时分
	if dailyScore := s.getStringValue(recordMap, "dailyScore"); dailyScore != "" && dailyScore != "0" {
		scoreTypes["daily"] = dailyScore
	}

	// 考试分
	if courseScore := s.getStringValue(recordMap, "courseScore"); courseScore != "" && courseScore != "0" {
		scoreTypes["exam"] = courseScore
	}

	// 实践分
	if practicalScore := s.getStringValue(recordMap, "practicalScore"); practicalScore != "" && practicalScore != "0" {
		scoreTypes["practical"] = practicalScore
	}

	// 最终分
	if finalScore := s.getStringValue(recordMap, "finalScore"); finalScore != "" && finalScore != "0" {
		scoreTypes["final"] = finalScore
	}

	// 补考分
	if supplementaryScore := s.getStringValue(recordMap, "supplementaryScore"); supplementaryScore != "" && supplementaryScore != "0" {
		scoreTypes["supplementary"] = supplementaryScore
	}

	return scoreTypes
}

// getStringValue 安全获取字符串值
func (s *ScoreCheckService) getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", value)
	}
	return ""
}

// getFloatValue 安全获取浮点数值
func (s *ScoreCheckService) getFloatValue(data map[string]interface{}, key string) float64 {
	if value, ok := data[key]; ok {
		switch v := value.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return 0.0
}

// compareScoresWithSemester 比较成绩变化（按课程聚合通知）
// isFirstSync: 是否为首次同步（之前没有任何快照记录），首次同步的变化不发送通知
func (s *ScoreCheckService) compareScoresWithSemester(currentScores []models.ScoreSnapshot, snapshots []models.ScoreSnapshot, semester string, isFirstSync bool) []ScoreUpdate {
	var updates []ScoreUpdate

	// 创建快照映射，便于查找
	snapshotMap := make(map[string]models.ScoreSnapshot)
	for _, snapshot := range snapshots {
		key := fmt.Sprintf("%s_%s_%s", snapshot.Semester, snapshot.CourseCode, snapshot.ScoreType)
		snapshotMap[key] = snapshot
	}

	// 按课程分组当前成绩
	courseScoreMap := make(map[string][]models.ScoreSnapshot)
	for _, current := range currentScores {
		courseKey := fmt.Sprintf("%s_%s", current.Semester, current.CourseCode)
		courseScoreMap[courseKey] = append(courseScoreMap[courseKey], current)
	}

	processedCourses := make(map[string]bool)

	for _, current := range currentScores {
		courseKey := fmt.Sprintf("%s_%s", current.Semester, current.CourseCode)

		if processedCourses[courseKey] {
			continue
		}
		processedCourses[courseKey] = true

		courseScores := courseScoreMap[courseKey]
		detailedScoreInfo := s.buildDetailedScoreInfo(courseScores, current)

		hasChanges := false
		changeType := "new"
		changedScoreTypesSet := make(map[string]bool)

		for _, score := range courseScores {
			key := fmt.Sprintf("%s_%s_%s", score.Semester, score.CourseCode, score.ScoreType)

			if snapshot, exists := snapshotMap[key]; exists {
				if score.CheckSum != snapshot.CheckSum {
					hasChanges = true
					changeType = "update"
					changedScoreTypesSet[s.getScoreTypeText(score.ScoreType)] = true
					log.Printf("[ScoreCheck] 检测到成绩变化: %s (%s): %s -> %s",
						score.CourseName, s.getScoreTypeText(score.ScoreType), snapshot.Score, score.Score)
				}
			} else {
				// 新成绩（仅在非首次同步时标记为新增）
				hasChanges = true
				changedScoreTypesSet[s.getScoreTypeText(score.ScoreType)] = true
				log.Printf("[ScoreCheck] 检测到新成绩: %s (%s): %s",
					score.CourseName, s.getScoreTypeText(score.ScoreType), score.Score)
			}
		}

		var changedScoreTypes []string
		for scoreType := range changedScoreTypesSet {
			changedScoreTypes = append(changedScoreTypes, scoreType)
		}

		if hasChanges {
			primaryScore := s.getPrimaryScore(courseScores)
			if primaryScore != nil {
				update := s.createCourseUpdate(primaryScore, semester, detailedScoreInfo, changeType, changedScoreTypes, isFirstSync)
				updates = append(updates, update)
			}
		}
	}

	return updates
}

// buildDetailedScoreInfo 构建详细成绩信息
func (s *ScoreCheckService) buildDetailedScoreInfo(courseScores []models.ScoreSnapshot, currentScore models.ScoreSnapshot) map[string]string {
	scoreInfo := map[string]string{
		"daily":         "0",
		"exam":          "0",
		"final":         "0",
		"practical":     "0",
		"supplementary": "0",
	}

	// 从当前课程的所有成绩中提取各种分数
	for _, score := range courseScores {
		switch score.ScoreType {
		case "daily":
			scoreInfo["daily"] = score.Score
		case "exam":
			scoreInfo["exam"] = score.Score
		case "final":
			scoreInfo["final"] = score.Score
		case "practical":
			scoreInfo["practical"] = score.Score
		case "supplementary":
			scoreInfo["supplementary"] = score.Score
		}
	}

	return scoreInfo
}

// getPrimaryScore 获取课程的主要成绩（优先最终分，其次考试分）
func (s *ScoreCheckService) getPrimaryScore(courseScores []models.ScoreSnapshot) *models.ScoreSnapshot {
	// 优先级：final > exam > daily > practical > supplementary
	priorities := map[string]int{
		"final":         1,
		"exam":          2,
		"daily":         3,
		"practical":     4,
		"supplementary": 5,
	}

	var primaryScore *models.ScoreSnapshot
	highestPriority := 999

	for i, score := range courseScores {
		if priority, exists := priorities[score.ScoreType]; exists {
			if priority < highestPriority {
				highestPriority = priority
				primaryScore = &courseScores[i]
			}
		}
	}

	// 如果没有找到优先级成绩，返回第一个
	if primaryScore == nil && len(courseScores) > 0 {
		primaryScore = &courseScores[0]
	}

	return primaryScore
}

// createCourseUpdate 创建课程级别的更新记录（聚合多种成绩类型）
func (s *ScoreCheckService) createCourseUpdate(primaryScore *models.ScoreSnapshot, semester string, detailedScores map[string]string, changeType string, changedScoreTypes []string, isFirstSync bool) ScoreUpdate {
	originalData, err := s.getOriginalScoreData(primaryScore.UserID, primaryScore.CourseCode, semester)
	if err != nil {
		log.Printf("[ScoreCheck] 获取原始成绩数据失败: %v", err)
	}

	scoreTypeText := "综合成绩"
	if len(changedScoreTypes) == 1 {
		scoreTypeText = changedScoreTypes[0]
	} else if len(changedScoreTypes) > 1 {
		scoreTypeText = fmt.Sprintf("多项成绩(%s)", strings.Join(changedScoreTypes, "、"))
	}

	update := ScoreUpdate{
		CourseName: fmt.Sprintf("[%s] %s", semester, primaryScore.CourseName),
		CourseCode: primaryScore.CourseCode,
		Semester:   semester,
		ScoreType:  scoreTypeText,
		OldScore:   "",
		NewScore:   detailedScores["final"],
		UpdateTime: time.Now().Format("2006-01-02 15:04:05"),

		// 详细成绩信息
		DailyScore:         detailedScores["daily"],
		ExamScore:          detailedScores["exam"],
		FinalScore:         detailedScores["final"],
		PracticalScore:     detailedScores["practical"],
		SupplementaryScore: detailedScores["supplementary"],

		// 课程信息
		Credit:     primaryScore.Credit,
		GPA:        primaryScore.GPA,
		ChangeType: changeType,

		// 标记是否为首次同步（首次同步的变化不发送通知）
		IsFirstSync: isFirstSync,
	}

	if originalData != nil {
		update.CourseProperty = originalData.CourseProperty
		update.TeacherNames = originalData.TeacherNames
		update.TestNote = originalData.TestNote
	}

	return update
}
