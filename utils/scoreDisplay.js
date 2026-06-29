/**
 * 成绩显示相关工具函数
 */

// 常量定义
export const LOADING_SEMESTER = { name: 'Loading...', value: '' };
export const ERROR_SEMESTER = { name: 'Data loading failed', value: '' };
export const EMPTY_SEMESTER = { name: 'No semester data', value: '' };

/**
 * 初始化统计数据
 */
export const getInitialStats = () => ({
	gpa: '0.00',
	averageScore: '0.00',
	creditTotal: '0',
	passRate: '0'
});

/**
 * 格式化统计数据
 * @param {Object} stats 原始统计数据
 * @returns {Object} 格式化后的统计数据
 */
export const formatStatsData = (stats) => ({
	gpa: stats.gpa || stats.averageGpa || '0.00',
	averageScore: stats.averageScore || stats.avgScore || '0.00',
	creditTotal: stats.creditTotal || stats.totalCredit || '0',
	passRate: stats.passRate || stats.passRatio || '0'
});

/**
 * 获取分数等级的样式类名
 * @param {number} score 分数
 * @returns {string} 样式类名
 */
export const getScoreGradeClass = (score) => {
	if (score >= 90) return 'score-excellent';
	if (score >= 80) return 'score-good';
	if (score >= 70) return 'score-medium';
	if (score >= 60) return 'score-pass';
	return 'score-fail';
};

/**
 * 获取分数等级文本
 * @param {number} score 分数
 * @returns {string} 等级文本
 */
export const getScoreGradeText = (score) => {
	if (score >= 90) return '优秀';
	if (score >= 80) return '良好';
	if (score >= 70) return '中等';
	if (score >= 60) return '及格';
	return '不及格';
};

/**
 * 获取颜色代码
 * @param {number} score 分数
 * @returns {string} 十六进制颜色代码
 */
export const getScoreColor = (score) => {
	if (score >= 90) return '#10b981'; // 绿色
	if (score >= 80) return '#3b82f6'; // 蓝色
	if (score >= 70) return '#f59e0b'; // 黄色
	if (score >= 60) return '#6366f1'; // 紫色
	return '#ef4444'; // 红色
};

/**
 * 判断课程是否挂科
 * @param {Object} course 课程成绩
 * @returns {boolean} 是否挂科
 */
export const isCourseFailed = (course) => {
	const getPoint = parseFloat(course.getPoint || course.gpa || 0);
	if (getPoint > 0) return false;

	const finalScore = parseFloat(course.finalScore);
	if (!isNaN(finalScore)) {
		return finalScore > 0 && finalScore < 60;
	}

	const courseScore = parseFloat(course.courseScore);
	if (!isNaN(courseScore)) {
		return courseScore > 0 && courseScore < 60;
	}

	const letterGrade = course.finalScore;
	return letterGrade === 'F' || letterGrade === 'E';
};

/**
 * 格式化教师姓名
 * @param {string} teacherNames 教师姓名
 * @returns {string} 格式化后的教师姓名
 */
export const formatTeacherName = (teacherNames) => {
	if (!teacherNames) return '';
	return teacherNames.replace(/#/g, '').replace(/\(\d+\)/g, '').trim();
};

/**
 * 获取用户当前学期
 * @returns {string} 当前学期
 */
export const getCurrentUserSemester = () => {
	try {
		const userInfoStr = uni.getStorageSync('userInfo');
		if (userInfoStr) {
			const userInfo = JSON.parse(userInfoStr);
			return userInfo.currentSemester || '';
		}
	} catch (error) {
		console.error('获取用户当前学期失败:', error);
	}
	return '';
};

/**
 * 为当前学期添加标识
 * @param {string} semesterName 学期名称
 * @param {string} currentUserSemester 用户当前学期
 * @returns {string} 添加标识后的学期名称
 */
export const addCurrentSemesterLabel = (semesterName, currentUserSemester) => {
	if (!semesterName || !currentUserSemester) {
		return semesterName;
	}

	if (semesterName === currentUserSemester ||
		semesterName.includes(currentUserSemester) ||
		currentUserSemester.includes(semesterName)) {
		return `${semesterName}（本学期）`;
	}

	return semesterName;
};

/**
 * 格式化学期列表
 * @param {Array} semesterData 学期数据
 * @returns {Array} 格式化后的学期列表
 */
export const formatSemesterList = (semesterData) => {
	const currentUserSemester = getCurrentUserSemester();

	const formattedSemesters = semesterData.map(item => {
		if (typeof item === 'string') {
			const displayName = addCurrentSemesterLabel(item, currentUserSemester);
			return { name: displayName, value: item };
		}

		const semesterName = item.currentSemester ||
						   item.semesterName ||
						   item.name ||
						   item.semester ||
						   '未知学期';

		const semesterValue = item.currentSemester ||
							item.semesterName ||
							item.value ||
							item.name ||
							item.semester ||
							'';

		const displayName = addCurrentSemesterLabel(semesterName, currentUserSemester);

		return { name: displayName, value: semesterValue };
	}).filter(item => item.value);

	const allOption = { name: '全部', value: '全部' };
	return [allOption, ...formattedSemesters];
};

/**
 * 解析学期信息，支持多种格式
 * @param {string} semesterStr 学期字符串
 * @returns {Object|null} 解析后的学期信息
 */
export const parseSemesterInfo = (semesterStr) => {
	if (!semesterStr) return null;

	// 格式1: 2023-2024-1
	const pattern1 = /(\d{4})-(\d{4})-?(\d)/;
	const match1 = semesterStr.match(pattern1);
	if (match1) {
		return { year: parseInt(match1[1]), semester: parseInt(match1[3]) };
	}

	// 格式2: 2024-2025学年第一学期、2024-2025学年第二学期
	const pattern2 = /(\d{4})-(\d{4})学年第([一二])学期/;
	const match2 = semesterStr.match(pattern2);
	if (match2) {
		const semesterMap = { '一': 1, '二': 2 };
		return { year: parseInt(match2[1]), semester: semesterMap[match2[3]] || 1 };
	}

	// 格式3: 2024-2025学年第1学期、2024-2025学年第2学期
	const pattern3 = /(\d{4})-(\d{4})学年第(\d)学期/;
	const match3 = semesterStr.match(pattern3);
	if (match3) {
		return { year: parseInt(match3[1]), semester: parseInt(match3[3]) };
	}

	return null;
};

/**
 * 比较两个学期的时间顺序
 * @param {string} semesterA 学期A
 * @param {string} semesterB 学期B
 * @returns {number} 比较结果
 */
export const compareSemesters = (semesterA, semesterB) => {
	const infoA = parseSemesterInfo(semesterA);
	const infoB = parseSemesterInfo(semesterB);

	if (infoA && infoB) {
		if (infoA.year !== infoB.year) {
			return infoA.year - infoB.year;
		}
		return infoA.semester - infoB.semester;
	}

	return semesterA.localeCompare(semesterB);
};

/**
 * 判断两个学期是否相同
 * @param {string} semesterA 学期A
 * @param {string} semesterB 学期B
 * @returns {boolean} 是否相同
 */
export const isSameSemester = (semesterA, semesterB) => {
	if (!semesterA || !semesterB) return false;

	if (semesterA === semesterB) return true;

	const infoA = parseSemesterInfo(semesterA);
	const infoB = parseSemesterInfo(semesterB);

	if (infoA && infoB) {
		return infoA.year === infoB.year && infoA.semester === infoB.semester;
	}

	return semesterA.includes(semesterB) || semesterB.includes(semesterA);
};

/**
 * 从学期列表中查找上学期
 * @param {Array} semesterList 学期列表
 * @returns {string|null} 上学期值
 */
export const findPreviousSemester = (semesterList) => {
	const currentUserSemester = getCurrentUserSemester();
	if (!currentUserSemester) return null;

	const availableSemesters = semesterList
		.filter(semester => semester.value !== '全部')
		.sort((a, b) => compareSemesters(a.value, b.value));

	const currentIndex = availableSemesters.findIndex(semester =>
		isSameSemester(semester.value, currentUserSemester)
	);

	if (currentIndex > 0) {
		return availableSemesters[currentIndex - 1].value;
	}

	return null;
};

/**
 * 从学期列表中查找下学期
 * @param {Array} semesterList 学期列表
 * @returns {string|null} 下学期值
 */
export const findNextSemester = (semesterList) => {
	const currentUserSemester = getCurrentUserSemester();
	if (!currentUserSemester) return null;

	const availableSemesters = semesterList
		.filter(semester => semester.value !== '全部')
		.sort((a, b) => compareSemesters(a.value, b.value));

	const currentIndex = availableSemesters.findIndex(semester =>
		isSameSemester(semester.value, currentUserSemester)
	);

	if (currentIndex !== -1 && currentIndex < availableSemesters.length - 1) {
		return availableSemesters[currentIndex + 1].value;
	}

	return null;
};

/**
 * 生成预览学期列表
 * @returns {Array} 预览学期列表
 */
export const generatePreviewSemesterList = () => [
	{ name: '全部学期（预览版）', value: '全部' },
	{ name: '2024-2025学年第二学期（本学期）', value: '2024-2025学年第二学期' },
	{ name: '2024-2025学年第一学期', value: '2024-2025学年第一学期' },
	{ name: '2023-2024学年第二学期', value: '2023-2024学年第二学期' }
];

/**
 * 生成预览成绩数据
 * @returns {Array} 预览成绩列表
 */
export const generatePreviewScoreData = () => [
	{
		id: 'preview-1',
		courseName: '高等数学A',
		teacherName: '张教授',
		credit: '4.0',
		dailyScore: '85',
		examScore: '88',
		finalScore: '87',
		supplementaryScore: '良好',
		gpa: '3.7',
		semester: '2024-2025学年第二学期',
		isPreview: true
	},
	{
		id: 'preview-2',
		courseName: '大学英语',
		teacherName: '李老师',
		credit: '3.0',
		dailyScore: '92',
		examScore: '89',
		finalScore: '90',
		supplementaryScore: '优秀',
		gpa: '4.0',
		semester: '2024-2025学年第二学期',
		isPreview: true
	},
	{
		id: 'preview-3',
		courseName: '数据结构与算法',
		teacherName: '王教授',
		credit: '4.0',
		dailyScore: '78',
		examScore: '82',
		finalScore: '80',
		supplementaryScore: '良好',
		gpa: '3.0',
		semester: '2024-2025学年第二学期',
		isPreview: true
	},
	{
		id: 'preview-4',
		courseName: '计算机网络',
		teacherName: '陈老师',
		credit: '3.5',
		dailyScore: '88',
		examScore: '85',
		finalScore: '86',
		supplementaryScore: '良好',
		gpa: '3.6',
		semester: '2024-2025学年第二学期',
		isPreview: true
	},
	{
		id: 'preview-5',
		courseName: '操作系统',
		teacherName: '刘教授',
		credit: '3.5',
		dailyScore: '91',
		examScore: '93',
		finalScore: '92',
		supplementaryScore: '优秀',
		gpa: '4.2',
		semester: '2024-2025学年第一学期',
		isPreview: true
	},
	{
		id: 'preview-6',
		courseName: '软件工程',
		teacherName: '赵老师',
		credit: '3.0',
		dailyScore: '84',
		examScore: '87',
		finalScore: '85',
		supplementaryScore: '良好',
		gpa: '3.5',
		semester: '2024-2025学年第一学期',
		isPreview: true
	},
	{
		id: 'preview-7',
		courseName: '高等数学B',
		teacherName: '李教授',
		credit: '4.0',
		dailyScore: '45',
		examScore: '52',
		finalScore: '48',
		supplementaryScore: '不及格',
		getPoint: '0.0',
		gpa: '0.0',
		semester: '2024-2025学年第一学期',
		isPreview: true
	},
	{
		id: 'preview-8',
		courseName: '形势与政策1（必选）',
		teacherName: '刘老师',
		credit: '0.25',
		dailyScore: '0',
		courseScore: '0',
		finalScore: 'D',
		supplementaryScore: '0',
		getPoint: '1.5',
		gpa: '1.5',
		semester: '2024-2025学年第一学期',
		isPreview: true
	}
];

/**
 * 生成预览统计数据
 * @returns {Object} 预览统计数据
 */
export const generatePreviewStats = () => ({
	gpa: '3.18',
	averageScore: '78.29',
	creditTotal: '25.25',
	passRate: '88'
});

/**
 * 判断是否是服务器维护错误
 * @param {Error} error 错误对象
 * @returns {boolean} 是否是服务器维护错误
 */
export const isServerMaintenanceError = (error) => {
	const errorMessage = error.message || '';
	const maintenanceKeywords = [
		'服务器关闭', '服务器维护', '连接超时', '网络错误',
		'ECONNREFUSED', 'ETIMEDOUT', 'Network Error', 'timeout'
	];

	return maintenanceKeywords.some(keyword =>
		errorMessage.toLowerCase().includes(keyword.toLowerCase())
	);
};

/**
 * 获取默认学期列表
 * @returns {Array} 默认学期列表
 */
export const getDefaultSemesterList = () => [{ name: '全部', value: '全部' }];

/**
 * 验证学期索引
 * @param {number} index 索引
 * @param {Array} semesterList 学期列表
 * @returns {boolean} 是否有效
 */
export const isValidSemesterIndex = (index, semesterList) => {
	return index >= 0 && index < semesterList.length;
};
