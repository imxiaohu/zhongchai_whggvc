package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// SchoolAPIService 学校API服务
type SchoolAPIService struct {
	proxyClient *utils.ProxyClient
}

// NewSchoolAPIService 创建学校API服务实例
func NewSchoolAPIService() *SchoolAPIService {
	return &SchoolAPIService{
		proxyClient: utils.NewProxyClient(),
	}
}

// SemesterInfo 学期信息
type SemesterInfo struct {
	Name      string `json:"name"`
	Code      string `json:"code"`
	IsCurrent bool   `json:"isCurrent"`
}

// ScoreInfo 成绩信息
type ScoreInfo struct {
	Semester    string  `json:"semester"`
	CourseCode  string  `json:"courseCode"`
	CourseName  string  `json:"courseName"`
	ScoreType   string  `json:"scoreType"`
	Score       string  `json:"score"`
	Credit      float64 `json:"credit"`
	GPA         float64 `json:"gpa"`
	Rank        string  `json:"rank"`
	TeacherName string  `json:"teacherName"`
	ExamType    string  `json:"examType"`
	ExamDate    string  `json:"examDate"`
}

// GetCurrentSemester 获取当前学期信息
func (s *SchoolAPIService) GetCurrentSemester(user *models.User) (*SemesterInfo, error) {
	// 首先尝试从学校服务器获取当前时间信息
	currentTimeResp, err := s.proxyClient.ProxyRequestWithAutoLogin(user, "GET",
		"/scloudoa/scs/course/tCourseTimetableDetail/getCurrentTime", nil)
	if err != nil {
		log.Printf("获取当前时间信息失败: %v", err)
		return s.getFallbackCurrentSemester(), nil
	}

	var currentTimeData map[string]interface{}
	if err := json.Unmarshal(currentTimeResp, &currentTimeData); err != nil {
		log.Printf("解析当前时间信息失败: %v", err)
		return s.getFallbackCurrentSemester(), nil
	}

	// 检查响应是否成功
	if success, ok := currentTimeData["success"].(bool); !ok || !success {
		log.Printf("学校服务器返回失败响应: %+v", currentTimeData)
		return s.getFallbackCurrentSemester(), nil
	}

	// 提取结果数据
	result, ok := currentTimeData["result"].(map[string]interface{})
	if !ok {
		log.Printf("学校服务器数据格式错误: %+v", currentTimeData)
		return s.getFallbackCurrentSemester(), nil
	}

	// 提取学期信息
	currentSemester := ""
	if v, ok := result["currentSemester"].(string); ok {
		currentSemester = v
	}
	if currentSemester == "" {
		return s.getFallbackCurrentSemester(), nil
	}

	return &SemesterInfo{
		Name:      currentSemester,
		Code:      s.generateSemesterCode(currentSemester),
		IsCurrent: true,
	}, nil
}

// GetAvailableSemesters 获取可用的学期列表
func (s *SchoolAPIService) GetAvailableSemesters(user *models.User) ([]SemesterInfo, error) {
	// 调用学校服务器获取学期列表
	semesterResp, err := s.proxyClient.ProxyRequestWithAutoLogin(user, "GET",
		"/scloudoa/scs/course/tCourseScore/getSemester", nil)
	if err != nil {
		log.Printf("获取学期列表失败: %v", err)
		return s.getFallbackSemesters(), nil
	}

	var semesterData interface{}
	if err := json.Unmarshal(semesterResp, &semesterData); err != nil {
		log.Printf("解析学期列表失败: %v", err)
		return s.getFallbackSemesters(), nil
	}

	// 处理不同的响应格式
	var semesters []SemesterInfo

	// 如果是标准响应格式
	if respMap, ok := semesterData.(map[string]interface{}); ok {
		if success, ok := respMap["success"].(bool); ok && success {
			if result, ok := respMap["result"].([]interface{}); ok {
				semesters = s.parseSemesterList(result)
			}
		}
	} else if semesterList, ok := semesterData.([]interface{}); ok {
		// 如果直接是数组格式
		semesters = s.parseSemesterList(semesterList)
	}

	if len(semesters) == 0 {
		return s.getFallbackSemesters(), nil
	}

	return semesters, nil
}

// GetScoresBySemester 根据学期获取成绩信息
func (s *SchoolAPIService) GetScoresBySemester(user *models.User, semester string) ([]ScoreInfo, error) {
	// 构建请求参数
	params := url.Values{}
	params.Set("semesterName", semester)

	// 调用学校服务器获取成绩列表
	scoresResp, err := s.proxyClient.ProxyRequestWithAutoLogin(user, "GET",
		"/scloudoa/scs/course/tCourseScore/getScoreList", params)
	if err != nil {
		log.Printf("获取成绩列表失败: %v", err)
		return nil, err
	}

	var scoresData map[string]interface{}
	if err := json.Unmarshal(scoresResp, &scoresData); err != nil {
		log.Printf("解析成绩列表失败: %v", err)
		return nil, err
	}

	// 检查响应是否成功
	if success, ok := scoresData["success"].(bool); !ok || !success {
		msg := ""
		if m, ok := scoresData["message"].(string); ok {
			msg = m
		}
		return nil, fmt.Errorf("获取成绩失败: %s", msg)
	}

	// 提取成绩数据
	result, ok := scoresData["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("成绩数据格式错误")
	}

	var scores []ScoreInfo

	// 处理分页数据
	if records, ok := result["records"].([]interface{}); ok {
		scores = s.parseScoreList(records, semester)
	}

	return scores, nil
}

// parseSemesterList 解析学期列表
func (s *SchoolAPIService) parseSemesterList(semesterList []interface{}) []SemesterInfo {
	var semesters []SemesterInfo
	semesterSet := make(map[string]bool) // 用于去重

	for _, item := range semesterList {
		if semesterData, ok := item.(map[string]interface{}); ok {
			if semesterName, ok := semesterData["currentSemester"].(string); ok && semesterName != "" {
				if !semesterSet[semesterName] { // 去重
					semesters = append(semesters, SemesterInfo{
						Name:      semesterName,
						Code:      s.generateSemesterCode(semesterName),
						IsCurrent: strings.Contains(semesterName, "2024-2025"),
					})
					semesterSet[semesterName] = true
				}
			}
		}
	}

	return semesters
}

// parseScoreList 解析成绩列表
func (s *SchoolAPIService) parseScoreList(scoreList []interface{}, semester string) []ScoreInfo {
	var scores []ScoreInfo

	for _, item := range scoreList {
		if scoreMap, ok := item.(map[string]interface{}); ok {
			score := ScoreInfo{
				Semester:    semester,
				CourseCode:  s.getStringValue(scoreMap, "courseCode"),
				CourseName:  s.getStringValue(scoreMap, "courseName"),
				Score:       s.getStringValue(scoreMap, "score"),
				ScoreType:   s.getStringValue(scoreMap, "scoreType"),
				Credit:      s.getFloatValue(scoreMap, "credit"),
				GPA:         s.getFloatValue(scoreMap, "gpa"),
				Rank:        s.getStringValue(scoreMap, "rank"),
				TeacherName: s.getStringValue(scoreMap, "teacherName"),
				ExamType:    s.getStringValue(scoreMap, "examType"),
				ExamDate:    s.getStringValue(scoreMap, "examDate"),
			}
			scores = append(scores, score)
		}
	}

	return scores
}

// 辅助方法
func (s *SchoolAPIService) getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}

func (s *SchoolAPIService) getFloatValue(data map[string]interface{}, key string) float64 {
	if value, ok := data[key].(float64); ok {
		return value
	}
	return 0.0
}

func (s *SchoolAPIService) generateSemesterCode(semesterName string) string {
	// 从学期名称生成代码，例如 "2024-2025学年第一学期" -> "2024-2025-1"
	if strings.Contains(semesterName, "第一学期") {
		return strings.Replace(semesterName, "学年第一学期", "-1", 1)
	} else if strings.Contains(semesterName, "第二学期") {
		return strings.Replace(semesterName, "学年第二学期", "-2", 1)
	}
	return semesterName
}

func (s *SchoolAPIService) getFallbackCurrentSemester() *SemesterInfo {
	return &SemesterInfo{
		Name:      "2024-2025学年第二学期",
		Code:      "2024-2025-2",
		IsCurrent: true,
	}
}

func (s *SchoolAPIService) getFallbackSemesters() []SemesterInfo {
	return []SemesterInfo{
		{Name: "2024-2025学年第二学期", Code: "2024-2025-2", IsCurrent: true},
		{Name: "2024-2025学年第一学期", Code: "2024-2025-1", IsCurrent: false},
		{Name: "2023-2024学年第二学期", Code: "2023-2024-2", IsCurrent: false},
		{Name: "2023-2024学年第一学期", Code: "2023-2024-1", IsCurrent: false},
	}
}
