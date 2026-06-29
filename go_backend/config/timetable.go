package config

// TimeSlotToLesson 课表时间段到节次的映射
var TimeSlotToLesson = map[string]int{
	"08:00": 1, "08:45": 1,
	"08:55": 2, "09:40": 2,
	"10:00": 3, "10:45": 3,
	"10:55": 4, "11:40": 4,
	"14:00": 5, "14:45": 5,
	"14:55": 6, "15:40": 6,
	"16:00": 7, "16:45": 7,
	"16:55": 8, "17:40": 8,
	"19:00": 9, "19:45": 9,
	"19:55": 10, "20:40": 10,
	"20:50": 11, "21:35": 11,
}

// GetLessonFromTime 根据时间获取节次
func GetLessonFromTime(timeStr string) int {
	if lesson, ok := TimeSlotToLesson[timeStr]; ok {
		return lesson
	}
	return 1 // 默认返回第一节
}
