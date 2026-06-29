/**
 * Schedule Display Utilities
 * Week/day formatters, course card styling, time slot definitions, and view mode helpers
 */

// Week Day Names
export const WEEK_DAYS = ['一', '二', '三', '四', '五', '六', '日'];

export const WEEK_DAYS_FULL = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];

export const WEEK_DAYS_EN = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];

// Time Slot Definitions (Default)
export const DEFAULT_TIME_SLOTS = [
	{ name: '1', startTime: '08:30', endTime: '09:15' },
	{ name: '2', startTime: '09:20', endTime: '10:05' },
	{ name: '3', startTime: '10:20', endTime: '11:05' },
	{ name: '4', startTime: '11:10', endTime: '11:55' },
	{ name: '5', startTime: '13:30', endTime: '14:15' },
	{ name: '6', startTime: '14:20', endTime: '15:05' },
	{ name: '7', startTime: '15:20', endTime: '16:05' },
	{ name: '8', startTime: '16:10', endTime: '16:55' },
	{ name: '9', startTime: '18:00', endTime: '18:45' },
	{ name: '10', startTime: '18:50', endTime: '19:35' },
	{ name: '11', startTime: '19:40', endTime: '20:25' },
	{ name: '12', startTime: '20:30', endTime: '21:15' }
];

// Week Formatters
export function formatWeekLabel(weekNum, totalWeeks = 20) {
	if (weekNum < 1 || weekNum > totalWeeks) {
		return `第${weekNum}周`;
	}
	return `第${weekNum}周`;
}

export function formatWeekRange(startWeek, endWeek) {
	if (startWeek === endWeek) {
		return `第${startWeek}周`;
	}
	return `第${startWeek}-${endWeek}周`;
}

export function formatWeekdayShort(weekday) {
	const index = WEEK_DAYS.indexOf(weekday);
	if (index !== -1) {
		return WEEK_DAYS_FULL[index];
	}
	return weekday;
}

export function getWeekdayIndex(weekday) {
	return WEEK_DAYS.indexOf(weekday);
}

// Day Messages
export const EMPTY_DAY_MESSAGES = {
	default: '今日无课程',
	weekend: '周末愉快，无课程安排',
	holiday: '节假日，无课程安排'
};

export const EMPTY_WEEK_MESSAGE = '本周暂无课程安排';

export function getEmptyDayMessage(weekday) {
	if (!weekday) return EMPTY_DAY_MESSAGES.default;
	const index = WEEK_DAYS.indexOf(weekday);
	if (index >= 5) {
		return EMPTY_DAY_MESSAGES.weekend;
	}
	return EMPTY_DAY_MESSAGES.default;
}

// Time Slot Helpers
export function formatTimeSlot(slot) {
	if (!slot) return '';
	return `${slot.startTime}-${slot.endTime}`;
}

export function getTimeSlotByIndex(timePeriods, index) {
	if (!timePeriods || !Array.isArray(timePeriods) || index < 0) {
		return DEFAULT_TIME_SLOTS[index] || null;
	}
	return timePeriods[index] || null;
}

export function calculateSlotHeight(lessonScopeLength, baseHeight = 80) {
	return baseHeight * lessonScopeLength;
}

// Course Card Styling
const COURSE_COLORS = [
	{ bg: '#6366f1', text: '#ffffff' }, // Indigo
	{ bg: '#8b5cf6', text: '#ffffff' }, // Violet
	{ bg: '#ec4899', text: '#ffffff' }, // Pink
	{ bg: '#f43f5e', text: '#ffffff' }, // Rose
	{ bg: '#f97316', text: '#ffffff' }, // Orange
	{ bg: '#eab308', text: '#000000' }, // Yellow
	{ bg: '#22c55e', text: '#ffffff' }, // Green
	{ bg: '#14b8a6', text: '#ffffff' }, // Teal
	{ bg: '#06b6d4', text: '#ffffff' }, // Cyan
	{ bg: '#3b82f6', text: '#ffffff' }  // Blue
];

export function getCourseColor(courseName) {
	if (!courseName) return COURSE_COLORS[0];
	let hash = 0;
	for (let i = 0; i < courseName.length; i++) {
		hash = courseName.charCodeAt(i) + ((hash << 5) - hash);
	}
	const index = Math.abs(hash) % COURSE_COLORS.length;
	return COURSE_COLORS[index];
}

export function getCourseCardStyle(course, timePeriods) {
	const color = getCourseColor(course.courseName);
	const style = {
		backgroundColor: color.bg,
		color: color.text
	};

	if (course.isPreview) {
		style.opacity = 0.85;
		style.border = '2rpx dashed rgba(255,255,255,0.5)';
	}

	return style;
}

export function getCourseCardClass(course) {
	const classes = ['course-card'];
	if (course.isPreview) {
		classes.push('course-card--preview');
	}
	if (course.isCurrent) {
		classes.push('course-card--current');
	}
	return classes.join(' ');
}

// Schedule View Mode Helpers
export const VIEW_MODES = {
	WEEK: 'week',
	DAY: 'day',
	LIST: 'list'
};

export function isValidViewMode(mode) {
	return Object.values(VIEW_MODES).includes(mode);
}

export function getDefaultViewMode() {
	return VIEW_MODES.WEEK;
}

// Semester Info Helpers
export function formatSemesterInfo(semester) {
	if (!semester) return '';
	return semester.replace('学年', '\n学年');
}

export function createSemesterInfo(currentSemester, nowweek, weekCount) {
	return {
		currentSemester: currentSemester || '',
		nowweek: nowweek || 1,
		weekCount: weekCount || 20,
		startDate: ''
	};
}

// Week Array Generator
export function generateWeekArray(totalWeeks) {
	return Array.from({ length: totalWeeks }, (_, i) => `${i + 1}`);
}

// Course Data Helpers
export function parseLessonScope(scopeStr) {
	if (!scopeStr) return [];
	const matches = scopeStr.match(/#(\d+)#/g);
	if (!matches) return [];
	return matches.map(m => parseInt(m.replace(/#/g, '')));
}

export function formatCourseInfo(course) {
	if (!course) return '';
	const parts = [course.courseName];
	if (course.teacherName) parts.push(course.teacherName);
	if (course.classroomName) parts.push(course.classroomName);
	return parts.join(' | ');
}

// Preview Data Generator
export function generatePreviewTimePeriods() {
	return [
		{ name: 1, startTime: '8:30' },
		{ name: 2, startTime: '9:20' },
		{ name: 3, startTime: '10:20' },
		{ name: 4, startTime: '11:10' },
		{ name: 5, startTime: '13:30' },
		{ name: 6, startTime: '14:20' },
		{ name: 7, startTime: '15:20' },
		{ name: 8, startTime: '16:10' },
		{ name: 9, startTime: '18:00' },
		{ name: 10, startTime: '18:50' },
		{ name: 11, startTime: '19:40' },
		{ name: 12, startTime: '20:30' }
	];
}

export function generatePreviewCourseData() {
	return [
		{
			id: 'preview-1',
			courseName: '高等数学',
			teacherNames: '张教授',
			classroomName: 'A101',
			week: '一',
			lessonScope: '#1#,#2#',
			lessonScopeLenght: 2,
			weeks: '1-18',
			isPreview: true
		},
		{
			id: 'preview-2',
			courseName: '大学英语',
			teacherNames: '李老师',
			classroomName: 'B203',
			week: '一',
			lessonScope: '#3#,#4#',
			lessonScopeLenght: 2,
			weeks: '1-18',
			isPreview: true
		},
		{
			id: 'preview-3',
			courseName: '计算机基础',
			teacherNames: '王老师',
			classroomName: '机房1',
			week: '二',
			lessonScope: '#1#,#2#',
			lessonScopeLenght: 2,
			weeks: '1-18',
			isPreview: true
		},
		{
			id: 'preview-4',
			courseName: '思想政治理论',
			teacherNames: '陈老师',
			classroomName: 'C301',
			week: '三',
			lessonScope: '#5#,#6#',
			lessonScopeLenght: 2,
			weeks: '1-18',
			isPreview: true
		},
		{
			id: 'preview-5',
			courseName: '体育',
			teacherNames: '刘教练',
			classroomName: '体育馆',
			week: '四',
			lessonScope: '#7#,#8#',
			lessonScopeLenght: 2,
			weeks: '1-18',
			isPreview: true
		},
		{
			id: 'preview-6',
			courseName: '专业课程',
			teacherNames: '赵老师',
			classroomName: 'D402',
			week: '五',
			lessonScope: '#1#,#2#,#3#',
			lessonScopeLenght: 3,
			weeks: '1-18',
			isPreview: true
		}
	];
}

// Server Maintenance Detection
export const MAINTENANCE_KEYWORDS = [
	'服务器关闭',
	'服务器维护',
	'连接超时',
	'网络错误',
	'ECONNREFUSED',
	'ETIMEDOUT',
	'Network Error',
	'timeout'
];

export function isServerMaintenanceError(error) {
	const errorMessage = error.message || '';
	return MAINTENANCE_KEYWORDS.some(keyword =>
		errorMessage.toLowerCase().includes(keyword.toLowerCase())
	);
}
