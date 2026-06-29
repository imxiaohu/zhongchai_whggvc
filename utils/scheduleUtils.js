/**
 * 课表工具函数
 * 包含课表格式化、课程卡片样式、空状态处理等功能
 */

/**
 * 课程卡片颜色配置
 */
export const CLASS_CARD_COLORS = [
	{ bg: '#dbeafe', border: '#3b82f6', text: '#1d4ed8' }, // 蓝色
	{ bg: '#d1fae5', border: '#10b981', text: '#047857' }, // 绿色
	{ bg: '#fef3c7', border: '#f59e0b', text: '#b45309' }, // 黄色
	{ bg: '#ede9fe', border: '#8b5cf6', text: '#6d28d9' }, // 紫色
	{ bg: '#fce7f3', border: '#ec4899', text: '#be185d' }, // 粉色
	{ bg: '#fed7aa', border: '#f97316', text: '#c2410c' }  // 橙色
];

/**
 * 课程卡片样式映射（按课程名称hash分配颜色）
 */
const CLASS_STYLE_CACHE = new Map();

/**
 * 获取课程卡片的颜色配置
 * @param {string} courseName - 课程名称
 * @returns {Object} 颜色配置
 */
export function getClassCardColor(courseName) {
	if (!courseName) {
		return CLASS_CARD_COLORS[0];
	}
	
	if (CLASS_STYLE_CACHE.has(courseName)) {
		return CLASS_STYLE_CACHE.get(courseName);
	}
	
	// 根据课程名称生成hash来分配颜色
	let hash = 0;
	for (let i = 0; i < courseName.length; i++) {
		hash = ((hash << 5) - hash) + courseName.charCodeAt(i);
		hash = hash & hash;
	}
	const index = Math.abs(hash) % CLASS_CARD_COLORS.length;
	const color = CLASS_CARD_COLORS[index];
	CLASS_STYLE_CACHE.set(courseName, color);
	
	return color;
}

/**
 * 格式化今日课程数据
 * @param {Object} result - API返回的结果
 * @returns {{originalData: Array, formattedData: Array}} 原始数据和格式化后的数据
 */
export function formatTodayClasses(result) {
	let originalData = [];
	let formattedData = [];
	
	if (Array.isArray(result.originalData)) {
		originalData = result.originalData;
	} else if (Array.isArray(result)) {
		originalData = result;
	}
	
	if (Array.isArray(result.formattedData)) {
		formattedData = result.formattedData;
	} else if (Array.isArray(result)) {
		formattedData = result;
	}
	
	return { originalData, formattedData };
}

/**
 * 创建预览课程数据
 * @returns {Array} 预览课程列表
 */
export function createPreviewClasses() {
	return [
		{
			id: 'preview-1',
			courseName: '示例课程',
			teacherName: '示例教师',
			classroom: '示例教室',
			timeSlot: '08:30-10:10',
			isPreview: true
		}
	];
}

/**
 * 获取空课程状态消息
 * @returns {{title: string, subtitle: string}} 空状态消息
 */
export function getEmptyScheduleMessage() {
	return {
		title: '今天没有课程',
		subtitle: '享受难得的休息时光吧~',
		emoji: '🌿'
	};
}

/**
 * 检查课程列表是否为空
 * @param {Array} classes - 课程列表
 * @returns {boolean} 是否为空
 */
export function isEmptySchedule(classes) {
	return !Array.isArray(classes) || classes.length === 0;
}

/**
 * 获取课程教室信息（带默认值）
 * @param {string|Object} classroom - 教室信息
 * @returns {string} 格式化后的教室文字
 */
export function formatClassroom(classroom) {
	if (!classroom) return '待定';
	if (typeof classroom === 'string') return classroom;
	if (classroom.building && classroom.room) {
		return `${classroom.building}-${classroom.room}`;
	}
	return String(classroom);
}

/**
 * 获取课程时间槽信息
 * @param {string|Object} timeSlot - 时间槽信息
 * @returns {string} 格式化后的时间文字
 */
export function formatTimeSlot(timeSlot) {
	if (!timeSlot) return '';
	if (typeof timeSlot === 'string') return timeSlot;
	if (timeSlot.start && timeSlot.end) {
		return `${timeSlot.start}-${timeSlot.end}`;
	}
	return String(timeSlot);
}

/**
 * 获取课程教师信息（带默认值）
 * @param {string|Object} teacher - 教师信息
 * @returns {string} 格式化后的教师文字
 */
export function formatTeacher(teacher) {
	if (!teacher) return '待定';
	if (typeof teacher === 'string') return teacher;
	if (teacher.name) return teacher.name;
	return String(teacher);
}

/**
 * 获取课程卡片样式对象
 * @param {Object} classItem - 课程项
 * @returns {Object} 样式对象
 */
export function getClassCardStyle(classItem) {
	const color = getClassCardColor(classItem.courseName);
	return {
		backgroundColor: color.bg,
		borderColor: color.border,
		color: color.text
	};
}

/**
 * 判断是否需要显示课程详情
 * @param {Object} classItem - 课程项
 * @returns {boolean} 是否需要显示
 */
export function shouldShowClassDetail(classItem) {
	return classItem && !classItem.isPreview;
}

/**
 * 获取课程状态（进行中/未开始/已结束）
 * @param {string} timeSlot - 时间槽
 * @returns {string} 课程状态
 */
export function getClassStatus(timeSlot) {
	if (!timeSlot) return 'unknown';
	
	try {
		const now = new Date();
		const currentTime = now.getHours() * 60 + now.getMinutes();
		const [start, end] = timeSlot.split('-').map(t => {
			const [h, m] = t.split(':').map(Number);
			return h * 60 + m;
		});
		
		if (currentTime < start) return 'not_started';
		if (currentTime > end) return 'finished';
		return 'ongoing';
	} catch (e) {
		return 'unknown';
	}
}

/**
 * 获取课程状态对应的文字
 * @param {string} status - 课程状态
 * @returns {string} 状态文字
 */
export function getClassStatusText(status) {
	const statusMap = {
		ongoing: '进行中',
		not_started: '未开始',
		finished: '已结束',
		unknown: ''
	};
	return statusMap[status] || '';
}

/**
 * 检查是否从设置页面返回
 * @returns {boolean} 是否从设置页面返回
 */
export function checkFromSettingsPage() {
	const fromSettingsPage = uni.getStorageSync('fromSettingsPage');
	const fromSettingsPageTime = uni.getStorageSync('fromSettingsPageTime');
	return fromSettingsPage === 'true' && fromSettingsPageTime &&
		(Date.now() - parseInt(fromSettingsPageTime)) < 10000;
}

/**
 * 清除设置页面返回标记
 */
export function clearSettingsPageMark() {
	uni.removeStorageSync('fromSettingsPage');
	uni.removeStorageSync('fromSettingsPageTime');
}

/**
 * 设置临时错误抑制
 * @param {number} duration - 抑制时长（毫秒）
 */
export function setErrorSuppress(duration = 15000) {
	const suppressUntil = Date.now() + duration;
	uni.setStorageSync('errorSuppressUntil', suppressUntil.toString());
}

/**
 * 清除错误抑制
 */
export function clearErrorSuppress() {
	uni.removeStorageSync('errorSuppressUntil');
}
