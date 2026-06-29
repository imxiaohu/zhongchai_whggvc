package models

import (
	"time"

	"gorm.io/gorm"
)

// News 新闻模型
type News struct {
	gorm.Model
	Title      string    `gorm:"size:200;not null" json:"title"`
	Content    string    `gorm:"type:text" json:"content"`
	Summary    string    `gorm:"size:500" json:"summary"`
	Cover      string    `gorm:"size:255" json:"cover"`
	Author     string    `gorm:"size:50" json:"author"`
	Source     string    `gorm:"size:100" json:"source"`
	TypeID     int       `json:"typeId"` // 1: 学校新闻, 2: 通知公告, 3: 学术活动
	TypeName   string    `gorm:"size:50" json:"typeName"`
	PublishAt  time.Time `json:"publishAt"`
	IsTop      int       `gorm:"default:0" json:"isTop"` // 0: 不置顶, 1: 置顶
	ViewCount  int       `gorm:"default:0" json:"viewCount"`
	SchoolID   uint      `json:"schoolId"`
	School     School    `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	Status     int       `gorm:"default:1" json:"status"` // 0: 草稿, 1: 已发布
}

// GetNewsByTypeID 根据类型ID获取新闻列表
func GetNewsByTypeID(typeID int, page, pageSize int) ([]News, int64, error) {
	var news []News
	var total int64

	// 计算总数
	query := DB.Model(&News{}).Where("type_id = ? AND status = 1", typeID)
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	result := query.Order("is_top DESC, publish_at DESC").Offset(offset).Limit(pageSize).Find(&news)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return news, total, nil
}

// GetNewsDetail 获取新闻详情
func GetNewsDetail(id uint) (*News, error) {
	var news News
	result := DB.First(&news, id)
	if result.Error != nil {
		return nil, result.Error
	}

	// 更新浏览量
	DB.Model(&news).Update("view_count", news.ViewCount+1)

	return &news, nil
}

// initDefaultNews 初始化默认新闻数据
func initDefaultNews() {
	// 检查是否已存在新闻数据
	var count int64
	DB.Model(&News{}).Count(&count)
	if count > 0 {
		return
	}

	// 获取默认学校ID
	var school School
	DB.First(&school)

	// 创建默认新闻
	news := []News{
		{
			Title:     "关于2023-2024学年第二学期期末考试安排的通知",
			Content:   "<p>各学院、各部门：</p><p>2023-2024学年第二学期期末考试定于2024年7月1日至7月5日进行，现将有关事项通知如下：</p><p>一、考试安排</p><p>1. 考试时间：7月1日至7月5日</p><p>2. 考试科目：详见教务系统公告</p><p>二、注意事项</p><p>1. 学生须携带学生证和身份证参加考试</p><p>2. 考试期间严禁携带手机等通讯设备</p><p>3. 严格遵守考场纪律，违纪者按校规处理</p>",
			Summary:   "2023-2024学年第二学期期末考试定于2024年7月1日至7月5日进行，请各位同学做好准备。",
			Author:    "教务处",
			Source:    "学校官网",
			TypeID:    2,
			TypeName:  "通知公告",
			PublishAt: time.Now().AddDate(0, 0, -2),
			IsTop:     1,
			SchoolID:  school.ID,
			Status:    1,
		},
		{
			Title:     "我校在全国职业院校技能大赛中获佳绩",
			Content:   "<p>近日，第十六届全国职业院校技能大赛落下帷幕，我校学生在比赛中表现优异，共获得金牌2枚、银牌3枚、铜牌5枚，创历史最好成绩。</p><p>本次大赛共有来自全国各地的500余所院校参赛，我校派出了由30名学生组成的代表队参加了15个赛项的角逐。在信息技术类、电子技术类等多个赛项中，我校选手展现出了扎实的专业功底和出色的实践能力。</p><p>学校将继续加强实践教学，提升学生职业技能，为培养更多高素质技术技能人才而努力。</p>",
			Summary:   "我校在第十六届全国职业院校技能大赛中获得金牌2枚、银牌3枚、铜牌5枚，创历史最好成绩。",
			Cover:     "https://example.com/news/skills_competition.jpg",
			Author:    "宣传部",
			Source:    "学校官网",
			TypeID:    1,
			TypeName:  "学校新闻",
			PublishAt: time.Now().AddDate(0, 0, -5),
			IsTop:     0,
			SchoolID:  school.ID,
			Status:    1,
		},
		{
			Title:     "2024届毕业生就业双选会成功举办",
			Content:   "<p>6月15日，我校2024届毕业生就业双选会在体育馆成功举办。本次双选会共有来自全国各地的200余家企业参会，提供就业岗位5000余个，涵盖IT、电子、机械、财会、管理等多个专业领域。</p><p>双选会现场人头攒动，毕业生们积极投递简历，与企业代表深入交流。据统计，现场达成就业意向的毕业生超过1000人，签约率达到40%以上。</p><p>学校就业指导中心表示，将持续关注毕业生就业情况，为毕业生和用人单位搭建更多交流平台，促进毕业生高质量就业。</p>",
			Summary:   "6月15日，我校举办2024届毕业生就业双选会，200余家企业提供5000余个岗位，现场签约率超40%。",
			Author:    "就业指导中心",
			Source:    "学校官网",
			TypeID:    1,
			TypeName:  "学校新闻",
			PublishAt: time.Now().AddDate(0, 0, -10),
			IsTop:     0,
			SchoolID:  school.ID,
			Status:    1,
		},
	}

	for _, n := range news {
		DB.Create(&n)
	}
}