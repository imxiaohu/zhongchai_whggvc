/**
 * 用户主页相关工具函数
 */

/**
 * 学院名称映射
 */
export const FACULTY_MAP = {
	'1': '计算机学院',
	'2': '经济管理学院',
};

/**
 * 获取学院名称
 * @param {string|number} facultyId 学院ID
 * @returns {string} 学院名称
 */
export const getFacultyName = (facultyId) => {
	if (!facultyId) return '';
	return FACULTY_MAP[facultyId] || '';
};

/**
 * 获取用户名字首字母
 * @param {Object} userInfo 用户信息
 * @returns {string} 首字母
 */
export const getUserInitial = (userInfo) => {
	const name = userInfo?.name || userInfo?.realname || '';
	if (!name) return '?';
	return name.charAt(0).toUpperCase();
};

/**
 * 格式化用户信息显示
 * @param {Object} userInfo 用户信息
 * @param {string} loginType 登录类型
 * @returns {Object} 格式化后的用户信息
 */
export const formatUserInfoDisplay = (userInfo, loginType) => {
	if (loginType === 'mobile' || loginType === 'school') {
		return {
			avatar: userInfo.avatar || '',
			name: userInfo.realname || userInfo.name || '普通用户',
			studentId: userInfo.username || userInfo.studentId || '',
			college: userInfo.college || getFacultyName(userInfo.facultyId) || userInfo.organizationName || '武汉光谷职业学院',
			className: userInfo.className || ''
		};
	} else if (loginType === 'wechat') {
		if (userInfo.hasSchoolAccount && userInfo.realname) {
			return {
				avatar: userInfo.avatar || userInfo.avatarUrl || '',
				name: userInfo.realname || userInfo.nickname || '微信用户',
				studentId: userInfo.username || '',
				college: userInfo.college || getFacultyName(userInfo.facultyId) || '',
				className: userInfo.className || ''
			};
		} else {
			return {
				avatar: userInfo.avatarUrl || userInfo.avatar || '',
				name: userInfo.nickname || '微信用户',
				studentId: '未绑定',
				college: '',
				className: ''
			};
		}
	} else {
		return {
			avatar: userInfo.avatar || '',
			name: userInfo.realname || userInfo.name || userInfo.nickname || '普通用户',
			studentId: userInfo.username || userInfo.studentId || '',
			college: userInfo.college || '',
			className: userInfo.className || ''
		};
	}
};

/**
 * 获取默认用户信息
 * @returns {Object} 默认用户信息
 */
export const getDefaultUserInfo = () => ({
	avatar: '',
	name: '未登录',
	studentId: '未绑定',
	college: '',
	className: ''
});

/**
 * 获取默认用户统计数据
 * @returns {Object} 默认统计数据
 */
export const getDefaultUserStats = () => ({
	averageScore: '',
	creditTotal: '',
	gpa: '',
	pendingEvaluationCount: null,
	semester: ''
});

/**
 * 获取默认服务器状态
 * @returns {Object} 默认服务器状态
 */
export const getDefaultServerStatus = () => ({
	isAlive: false,
	lastCheck: null,
	lastAlive: null,
	responseTime: null,
	errorMsg: null
});

/**
 * 格式化用户统计数据
 * @param {Object} stats 原始统计数据
 * @returns {Object} 格式化后的统计数据
 */
export const formatUserStats = (stats) => {
	const averageScoreRaw = stats?.averageScore;
	const creditTotalRaw = stats?.creditTotal;
	const gpaRaw = stats?.gpa;

	return {
		averageScore: typeof averageScoreRaw === 'number'
			? averageScoreRaw.toFixed(2)
			: (typeof averageScoreRaw === 'string' ? averageScoreRaw : ''),
		creditTotal: typeof creditTotalRaw === 'number'
			? creditTotalRaw.toFixed(2)
			: (typeof creditTotalRaw === 'string' ? creditTotalRaw : ''),
		gpa: typeof gpaRaw === 'number'
			? gpaRaw.toFixed(2)
			: (typeof gpaRaw === 'string' ? gpaRaw : ''),
		pendingEvaluationCount: typeof stats?.pendingEvaluationCount === 'number'
			? stats.pendingEvaluationCount
			: null,
		semester: stats?.semester || ''
	};
};

/**
 * 检查是否应该获取用户统计数据
 * @param {boolean} isLoggedIn 是否已登录
 * @param {string} loginType 登录类型
 * @param {boolean} isSchoolAccountBound 学校账号是否已绑定
 * @returns {boolean} 是否应该获取统计数据
 */
export const shouldFetchUserStats = (isLoggedIn, loginType, isSchoolAccountBound) => {
	if (!isLoggedIn) return false;
	return loginType === 'mobile' || loginType === 'school' || isSchoolAccountBound;
};

/**
 * 从存储加载用户信息
 * @returns {Object} 用户信息对象
 */
export const loadUserInfoFromStorage = () => {
	const token = uni.getStorageSync('token');
	const isLoggedIn = !!token;
	const userInfoStr = uni.getStorageSync('userInfo');

	let userInfo = null;
	if (userInfoStr && token) {
		try {
			userInfo = JSON.parse(userInfoStr);
		} catch (e) {
			console.error('解析用户信息失败:', e);
		}
	}

	return { token, isLoggedIn, userInfo };
};

/**
 * 格式化日期时间
 * @param {Date|string|number} date 日期
 * @returns {string} 格式化后的日期字符串
 */
export const formatDateTime = (date) => {
	if (!date) return '';
	const d = new Date(date);
	const year = d.getFullYear();
	const month = String(d.getMonth() + 1).padStart(2, '0');
	const day = String(d.getDate()).padStart(2, '0');
	const hours = String(d.getHours()).padStart(2, '0');
	const minutes = String(d.getMinutes()).padStart(2, '0');
	return `${year}-${month}-${day} ${hours}:${minutes}`;
};

/**
 * 格式化相对时间
 * @param {Date|string|number} date 日期
 * @returns {string} 相对时间字符串
 */
export const formatRelativeTime = (date) => {
	if (!date) return '';
	const now = Date.now();
	const timestamp = new Date(date).getTime();
	const diff = now - timestamp;

	const minutes = Math.floor(diff / 60000);
	const hours = Math.floor(diff / 3600000);
	const days = Math.floor(diff / 86400000);

	if (minutes < 1) return '刚刚';
	if (minutes < 60) return `${minutes}分钟前`;
	if (hours < 24) return `${hours}小时前`;
	if (days < 7) return `${days}天前`;

	return formatDateTime(date);
};

/**
 * 导航路由映射
 */
export const NAVIGATION_ROUTES = {
	score: '/pages/user/score',
	bind: '/pages/user/bind',
	settings: '/pages/user/setting',
	notificationCenter: '/pages/user/notification-center',
	bookmark: '/pages/community/bookmark-list',
	serverStatus: '/pages/user/server-status',
	profileEdit: '/pages/user/profile-edit',
	login: '/pages/login/login'
};

/**
 * 获取导航路由
 * @param {string} routeKey 路由键
 * @returns {string} 路由路径
 */
export const getNavigationRoute = (routeKey) => NAVIGATION_ROUTES[routeKey] || '';

/**
 * 验证登录状态并导航
 * @param {boolean} isLoggedIn 是否已登录
 * @param {Function} redirectToLogin 跳转到登录页的函数
 * @returns {boolean} 是否继续执行
 */
export const validateLoginAndNavigate = (isLoggedIn, redirectToLogin) => {
	if (!isLoggedIn) {
		redirectToLogin();
		return false;
	}
	return true;
};

/**
 * 获取绑定菜单文本
 * @param {boolean} isBound 是否已绑定
 * @returns {string} 菜单文本
 */
export const getBindMenuText = (isBound) => isBound ? '学校账号管理' : '绑定学校账号';

/**
 * 获取绑定图标类型
 * @param {boolean} isBound 是否已绑定
 * @returns {string} 图标类型
 */
export const getBindIconType = (isBound) => isBound ? 'auth-filled' : 'person';
