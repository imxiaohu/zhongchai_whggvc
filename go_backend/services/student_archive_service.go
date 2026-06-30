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

// StudentArchive 学生完整档案数据（来自银行卡接口 + JSON API 拼接）
type StudentArchive struct {
	// 基础资料
	StudentNo       string `json:"studentNo"`
	ExamNo          string `json:"examNo"`
	Realname        string `json:"realname"`
	NameUsedBefore  string `json:"nameUsedBefore"`
	Sex             string `json:"sex"`
	EntranceScore   string `json:"entranceScore"`
	Birthday        string `json:"birthday"`
	GraduateType    string `json:"graduateType"`
	NativePlace     string `json:"nativePlace"`
	GraduateSchool  string `json:"graduateSchool"`
	IDCardNo        string `json:"idCardNo"`
	GraduateForm    string `json:"graduateForm"`
	IsMarried       string `json:"isMarried"`
	SourceProvince  string `json:"sourceProvince"`
	Nation          string `json:"nation"`
	IsPoorArea      string `json:"isPoorArea"`
	PoliticsStatus  string `json:"politicsStatus"`
	HealthCondition string `json:"healthCondition"`
	Avatar          string `json:"avatar"`

	// 学业信息
	Campus           string `json:"campus"`
	StudentStatus    string `json:"studentStatus"`
	FacultyName      string `json:"facultyName"`
	EnrollmentDate   string `json:"enrollmentDate"`
	MajorName        string `json:"majorName"`
	ExpectedGradDate string `json:"expectedGradDate"`
	AdminClass       string `json:"adminClass"`
	ClassName        string `json:"className"`
	Counselor        string `json:"counselor"`
	StudyForm        string `json:"studyForm"`
	Grade            string `json:"grade"`
	EducationYears   string `json:"educationYears"`
	EnrollmentType   string `json:"enrollmentType"`

	// 联系方式
	Phone           string `json:"phone"`
	PersonalAddress string `json:"personalAddress"`
	Email           string `json:"email"`
	QQ              string `json:"qq"`

	// 银行卡信息
	BankCardNumber   string `json:"bankCardNumber"`
	BankName         string `json:"bankName"`
	BankProvinceCity string `json:"bankProvinceCity"`
	BankSubBranch    string `json:"bankSubBranch"`
	BankCardType     string `json:"bankCardType"`
	CardHolder       string `json:"cardHolder"`

	// 家庭资料
	FamilyAddress string         `json:"familyAddress"`
	FamilyPhone   string         `json:"familyPhone"`
	FamilyPost    string         `json:"familyPost"`
	FamilyMembers []FamilyMember `json:"familyMembers"`

	// 学校经历
	SchoolExperiences []SchoolExperience `json:"schoolExperiences"`

	// 学业调整
	AcademicChanges []AcademicChange `json:"academicChanges"`
}

// StudentInfo 学校 JSON API /scloud/student/base/getStudentInfo 返回的数据
type StudentInfo struct {
	StudyForm          string `json:"studyForm"`
	Memo               string `json:"memo"`
	ClassName          string `json:"className"`
	StudySystem        string `json:"studySystem"`
	EnrollmentStatusID int    `json:"enrollmentStatusId"`
	BranchCourts       string `json:"branchCourts"`
	EnrollmentStatus   string `json:"enrollmentStatus"`
	ProfessionName     string `json:"professionName"`
	ExpectedGradDate   string `json:"expectedGraduateDate"`
	Grade              string `json:"grade"`
	Name               string `json:"name"`
	ID                 int    `json:"ID"`
	AdminClass         string `json:"adminClass"`
	FacultyStation     string `json:"facultyStation"`
	StudyNumber        string `json:"studyNumber"`
	EntranceDate       string `json:"entranceDate"`
}

type FamilyMember struct {
	Name            string `json:"name"`
	Relationship    string `json:"relationship"`
	IsGuardian      string `json:"isGuardian"`
	CredentialsType string `json:"credentialsType"`
	CredentialsNo   string `json:"credentialsNo"`
	Workplace       string `json:"workplace"`
	Mobile          string `json:"mobile"`
	PoliticsStatus  string `json:"politicsStatus"`
}

type SchoolExperience struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	School    string `json:"school"`
	Job       string `json:"job"`
	ProveMan  string `json:"proveMan"`
}

type AcademicChange struct {
	ChangeType   string `json:"changeType"`
	ChangeDetail string `json:"changeDetail"`
	Operator     string `json:"operator"`
	ChangeDate   string `json:"changeDate"`
	OperateDate  string `json:"operateDate"`
}

// ===== 内存缓存 =====

type archiveCacheEntry struct {
	archive  *StudentArchive
	expireAt time.Time
}

var (
	archiveCache sync.Map // map[uint]*archiveCacheEntry
)

const archiveCacheTTL = 10 * time.Minute

func setUserArchiveCache(userID uint, archive *StudentArchive) {
	archiveCache.Store(userID, &archiveCacheEntry{
		archive:  archive,
		expireAt: time.Now().Add(archiveCacheTTL),
	})
	log.Printf("[ARCHIVE-CACHE] Set cache for userID=%d", userID)
}

func getUserArchiveCache(userID uint) *StudentArchive {
	val, ok := archiveCache.Load(userID)
	if !ok {
		return nil
	}
	entry := val.(*archiveCacheEntry)
	if time.Now().After(entry.expireAt) {
		archiveCache.Delete(userID)
		return nil
	}
	return entry.archive
}

// ===== 数据获取 =====

// FetchStudentInfoFromAPI 调用学校 JSON API 获取学生信息
func FetchStudentInfoFromAPI(user *models.User) (*StudentInfo, error) {
	proxyClient := utils.NewProxyClient()

	if err := proxyClient.AutoLogin(user, false); err != nil {
		return nil, fmt.Errorf("自动登录失败: %w", err)
	}

	if _, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloud/init", nil); err != nil {
		log.Printf("[STUDENT-ARCHIVE-API] init 调用失败，继续尝试: %v", err)
	}

	body, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloud/student/base/getStudentInfo", nil)
	if err != nil {
		return nil, fmt.Errorf("请求学生信息接口失败: %w", err)
	}

	if strings.TrimSpace(string(body)) == "" || strings.HasPrefix(strings.TrimSpace(string(body)), "<") {
		log.Printf("[STUDENT-ARCHIVE-API] 首次请求返回空/HTML，强制重新登录后重试")
		if err := proxyClient.AutoLogin(user, true); err != nil {
			return nil, fmt.Errorf("强制重新登录失败: %w", err)
		}
		if _, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloud/init", nil); err != nil {
			log.Printf("[STUDENT-ARCHIVE-API] 重试 init 失败: %v", err)
		}
		body, err = proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloud/student/base/getStudentInfo", nil)
		if err != nil {
			return nil, fmt.Errorf("重试请求学生信息接口失败: %w", err)
		}
		if strings.TrimSpace(string(body)) == "" || strings.HasPrefix(strings.TrimSpace(string(body)), "<") {
			return nil, fmt.Errorf("重试后仍返回空或 HTML 响应")
		}
	}

	var info StudentInfo
	if err := json.Unmarshal(body, &info); err != nil {
		log.Printf("[STUDENT-ARCHIVE-API] JSON 解析失败: %v, raw: %s", err, string(body[:min(200, len(body))]))
		return nil, fmt.Errorf("解析学生信息失败: %w", err)
	}

	log.Printf("[STUDENT-ARCHIVE-API] name=%s studyNumber=%s faculty=%s profession=%s class=%s",
		info.Name, info.StudyNumber, info.FacultyStation, info.ProfessionName, info.AdminClass)

	return &info, nil
}

// FetchBankCardInfo 获取银行卡信息（含学生档案字段）
func FetchBankCardInfo(user *models.User, clientID, overrideCookie string) (map[string]interface{}, error) {
	proxyClient := utils.NewProxyClient()
	params := make(url.Values)
	if clientID != "" {
		params.Set("clientId", clientID)
	}

	body, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloudoa/studentBank/getStudentBank", params)
	if err != nil {
		return nil, fmt.Errorf("请求银行卡接口失败: %w", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析银行卡响应失败: %w", err)
	}

	if result, ok := resp["result"].(map[string]interface{}); ok && result != nil {
		log.Printf("[BANK-CARD] got bank info: bankName=%v cardNumber=%v",
			result["bankname"], result["bankcardnumber"])
		return result, nil
	}

	return nil, fmt.Errorf("银行卡接口返回结果为空")
}

// BuildMergedArchive 构建合并后的学生档案
// data priority: bank card info (highest) > JSON API > user DB
func BuildMergedArchive(user *models.User, bankInfo map[string]interface{}, apiInfo *StudentInfo) *StudentArchive {
	archive := &StudentArchive{}

	// 1. 银行卡信息（最高优先级，包含姓名/学号/性别/民族/政治面貌/学业等）
	if bankInfo != nil {
		archive.StudentNo = getString(bankInfo, "studentnumber", "studynumber")
		archive.Realname = getString(bankInfo, "name")
		archive.Sex = getString(bankInfo, "gender")
		archive.Birthday = getString(bankInfo, "birthday")
		archive.IDCardNo = getString(bankInfo, "identitycard")
		archive.Nation = getString(bankInfo, "nation")
		archive.PoliticsStatus = getString(bankInfo, "politicsstatus")
		archive.FacultyName = getString(bankInfo, "facultystation")
		archive.MajorName = getString(bankInfo, "professionname")
		archive.AdminClass = getString(bankInfo, "classnumber")
		archive.ClassName = getString(bankInfo, "classname")
		archive.Campus = getString(bankInfo, "branchcourts")
		archive.Grade = getString(bankInfo, "grade")
		archive.StudyForm = getString(bankInfo, "studyform")
		archive.EducationYears = getString(bankInfo, "studysystem")
		archive.EnrollmentDate = getString(bankInfo, "entrancedate")
		archive.ExpectedGradDate = getString(bankInfo, "expectedgraduatedate")
		archive.Phone = getString(bankInfo, "phoneNumber")
		archive.Email = getString(bankInfo, "email")
		archive.BankCardNumber = getString(bankInfo, "bankcardnumber")
		archive.BankName = getString(bankInfo, "bankname")
		archive.BankProvinceCity = getString(bankInfo, "bankprovincecity")
		archive.BankSubBranch = getString(bankInfo, "banksubbranch")
		archive.BankCardType = getString(bankInfo, "bankcardtype")
		archive.CardHolder = getString(bankInfo, "cardholder")
	}

	// 2. JSON API 补充（覆盖空字段）
	if apiInfo != nil {
		if archive.StudentNo == "" {
			archive.StudentNo = apiInfo.StudyNumber
		}
		if archive.Realname == "" {
			archive.Realname = apiInfo.Name
		}
		if archive.AdminClass == "" {
			archive.AdminClass = apiInfo.AdminClass
		}
		if archive.ClassName == "" {
			archive.ClassName = apiInfo.ClassName
		}
		if archive.MajorName == "" {
			archive.MajorName = apiInfo.ProfessionName
		}
		if archive.FacultyName == "" {
			archive.FacultyName = apiInfo.FacultyStation
		}
		if archive.Grade == "" {
			archive.Grade = apiInfo.Grade
		}
		if archive.ExpectedGradDate == "" {
			archive.ExpectedGradDate = apiInfo.ExpectedGradDate
		}
		if archive.EnrollmentDate == "" {
			archive.EnrollmentDate = apiInfo.EntranceDate
		}
		if archive.StudyForm == "" {
			archive.StudyForm = apiInfo.StudyForm
		}
		if archive.Campus == "" {
			archive.Campus = apiInfo.BranchCourts
		}
		if archive.StudentStatus == "" {
			archive.StudentStatus = apiInfo.EnrollmentStatus
		}
		if archive.EducationYears == "" {
			archive.EducationYears = apiInfo.StudySystem
		}
		if archive.Counselor == "" {
			archive.Counselor = apiInfo.Memo
		}
	}

	// 3. 用户数据库兜底（覆盖空字段）
	if archive.Realname == "" {
		archive.Realname = user.Realname
	}
	if archive.Phone == "" {
		archive.Phone = user.Phone
	}
	if archive.Email == "" {
		archive.Email = user.Email
	}
	if archive.Birthday == "" {
		archive.Birthday = user.Birthday
	}
	if archive.IDCardNo == "" {
		archive.IDCardNo = user.IdentityCard
	}
	if archive.ClassName == "" {
		archive.ClassName = user.ClassName
	}
	if archive.Sex == "" && user.Sex != 0 {
		archive.Sex = fmt.Sprintf("%d", user.Sex)
	}

	return archive
}

// getString 尝试从 map 中获取字符串值（支持多个 key）
func getString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			s := strings.TrimSpace(fmt.Sprintf("%v", v))
			if s != "" && s != "<nil>" && s != "0" {
				return s
			}
		}
	}
	return ""
}

// ===== 缓存封装 =====

// FetchAndCacheArchive 获取并缓存学生档案
// 优先从内存缓存返回，缓存未命中则从银行接口+JSON API 拉取并缓存
func FetchAndCacheArchive(user *models.User, clientID, overrideCookie string) (*StudentArchive, error) {
	// 1. 尝试从缓存获取
	if cached := getUserArchiveCache(user.ID); cached != nil {
		log.Printf("[ARCHIVE-CACHE] Hit cache for userID=%d", user.ID)
		return cached, nil
	}

	// 2. 检查是否有学校账号
	hasSchoolAccount := user.Username != "" &&
		user.Password != "" &&
		!strings.HasPrefix(user.Username, "wx_") &&
		user.Username != user.WechatOpenID

	if !hasSchoolAccount || user.SchoolToken == "" {
		// 无学校账号，仅返回用户本地数据
		archive := BuildMergedArchive(user, nil, nil)
		setUserArchiveCache(user.ID, archive)
		return archive, nil
	}

	// 3. 并发拉取银行卡 + JSON API
	type fetchResult struct {
		bankInfo map[string]interface{}
	}
	resultCh := make(chan fetchResult, 1)

	go func() {
		bankInfo, _ := FetchBankCardInfo(user, clientID, overrideCookie)
		resultCh <- fetchResult{bankInfo: bankInfo}
	}()

	apiInfo, _ := FetchStudentInfoFromAPI(user)

	fetched := <-resultCh

	// 4. 构建并缓存
	archive := BuildMergedArchive(user, fetched.bankInfo, apiInfo)
	setUserArchiveCache(user.ID, archive)

	return archive, nil
}

// CacheArchiveOnLogin 登录时异步预缓存档案（不阻塞登录响应）
func CacheArchiveOnLogin(userID uint, username, password, schoolToken string) {
	go func() {
		time.Sleep(2 * time.Second) // 延迟2秒，等登录完成

		// 已有缓存就直接返回：避免登录后并发请求（包括 onShow 触发）重复拉取
		if cached := getUserArchiveCache(userID); cached != nil {
			log.Printf("[ARCHIVE-LOGIN] cache already warm, skip: userID=%d", userID)
			return
		}

		user, err := models.FindUserByID(userID)
		if err != nil {
			log.Printf("[ARCHIVE-LOGIN] user not found, skip cache: userID=%d", userID)
			return
		}
		if user.SchoolToken == "" {
			log.Printf("[ARCHIVE-LOGIN] no school token, skip cache: userID=%d", userID)
			return
		}

		archive, err := FetchAndCacheArchive(user, "", "")
		if err != nil {
			log.Printf("[ARCHIVE-LOGIN] fetch failed: userID=%d err=%v", userID, err)
			return
		}
		log.Printf("[ARCHIVE-LOGIN] cached on login: userID=%d realname=%s studentNo=%s",
			userID, archive.Realname, archive.StudentNo)
	}()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
