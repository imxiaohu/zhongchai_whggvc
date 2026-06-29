/**
 * 首页工具函数
 * 包含服务器维护检测、问候语格式化、日期时间处理等功能
 */

import { useCourseCache } from '@/store/courseCache.js'

/**
 * 判断是否是服务器维护错误
 * @param {Error} error - 错误对象
 * @returns {boolean} 是否为服务器维护错误
 */
export function isServerMaintenanceError(error) {
	const errorMessage = error.message || '';
	const maintenanceKeywords = [
		'服务器关闭',
		'服务器维护',
		'连接超时',
		'网络错误',
		'ECONNREFUSED',
		'ETIMEDOUT',
		'Network Error',
		'timeout'
	];

	return maintenanceKeywords.some(keyword =>
		errorMessage.toLowerCase().includes(keyword.toLowerCase())
	);
}

/**
 * 获取问候语主文本
 * @param {boolean} isLoggedIn - 是否已登录
 * @returns {string} 问候语
 */
export function getGreetingText(isLoggedIn) {
	if (!isLoggedIn) return '你好';
	const h = new Date().getHours();
	if (h < 6) return '夜深了';
	if (h < 11) return '早上好';
	if (h < 13) return '中午好';
	if (h < 18) return '下午好';
	return '晚上好';
}

/**
 * 获取问候语副标题
 * @param {boolean} isLoggedIn - 是否已登录
 * @param {boolean} showBindingPrompt - 是否显示绑定提示
 * @returns {string} 副标题
 */
export function getGreetingSubtitle(isLoggedIn, showBindingPrompt) {
	if (!isLoggedIn) return '欢迎使用众柴智慧校园 · 让学习更高效';
	if (showBindingPrompt) return '请先绑定学校账号以使用完整功能';
	return '让评教、课表、成绩一站式管理';
}

/**
 * 获取头像文字（取昵称的最后一个字符）
 * @param {string} name - 用户昵称
 * @returns {string} 头像文字
 */
export function getAvatarText(name) {
	const userName = name || '同';
	return String(userName).slice(-1) || '同';
}

/**
 * 获取今日日期文字（如 周五 · 6月26日）
 * @returns {string} 格式化后的日期文字
 */
export function getTodayDateText() {
	try {
		const d = new Date();
		const weeks = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
		const week = weeks[d.getDay()];
		return `${week} · ${d.getMonth() + 1}月${d.getDate()}日`;
	} catch (e) {
		return '';
	}
}

/**
 * 获取星期几的文字描述
 * @param {Date|number} date - 日期对象或星期数字(0-6)
 * @returns {string} 星期几
 */
export function getWeekdayText(date) {
	const weeks = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
	if (typeof date === 'number') {
		return weeks[date] || '';
	}
	try {
		const d = date instanceof Date ? date : new Date(date);
		return weeks[d.getDay()] || '';
	} catch (e) {
		return '';
	}
}

/**
 * 格式化通知日期
 * @param {string} dateStr - 日期字符串
 * @returns {string} 格式化后的日期
 */
export function formatNoticeDate(dateStr) {
	if (!dateStr) return '';

	try {
		const date = new Date(dateStr);
		const now = new Date();
		const diffTime = now - date;
		const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));

		if (diffDays === 0) {
			return '今天';
		} else if (diffDays === 1) {
			return '昨天';
		} else if (diffDays < 7) {
			return `${diffDays}天前`;
		} else {
			return `${date.getMonth() + 1}-${date.getDate()}`;
		}
	} catch (error) {
		console.error('日期格式化失败:', error);
		return dateStr;
	}
}

/**
 * 初始化状态栏高度
 * @returns {number} 状态栏高度
 */
export function initStatusBarHeight() {
	try {
		// #ifdef MP-WEIXIN
		const sys = uni.getSystemInfoSync();
		if (sys.safeArea && sys.safeArea.top > 0) {
			return sys.safeArea.top;
		} else {
			return sys.statusBarHeight || 20;
		}
		// #endif

		// #ifdef APP-PLUS
		return plus.navigator.getStatusbarHeight ? plus.navigator.getStatusbarHeight() : 20;
		// #endif

		// #ifdef H5
		return 0;
		// #endif
	} catch (e) {
		return 20;
	}
}

/**
 * 刷新用户信息
 * @returns {{userInfo: Object|null, nickname: string}} 用户信息和昵称
 */
export function refreshUserInfo() {
	try {
		const userInfoStr = uni.getStorageSync('userInfo');
		if (userInfoStr) {
			const info = typeof userInfoStr === 'string' ? JSON.parse(userInfoStr) : userInfoStr;
			const nickname = info.nickName || info.nickname || info.name || info.realname || '同学';
			return { userInfo: info, nickname };
		} else {
			const rememberUser = uni.getStorageSync('rememberUser') || uni.getStorageSync('realname');
			if (rememberUser) {
				return { userInfo: null, nickname: rememberUser };
			}
		}
	} catch (e) {
		console.warn('读取用户信息失败:', e);
	}
	return { userInfo: null, nickname: '同学' };
}

/**
 * 检查功能是否可用（需要学校服务器）
 * @param {boolean} isLoggedIn - 是否已登录
 * @param {boolean} serverMaintenance - 服务器是否维护中
 * @returns {boolean} 功能是否可用
 */
export function isFeatureAvailable(isLoggedIn, serverMaintenance) {
	return isLoggedIn && !serverMaintenance;
}

/**
 * 获取功能不可用提示消息
 * @param {boolean} isLoggedIn - 是否已登录
 * @param {boolean} serverMaintenance - 服务器是否维护中
 * @returns {string} 提示消息
 */
export function getFeatureUnavailableMessage(isLoggedIn, serverMaintenance) {
	if (!isLoggedIn) {
		return '请先登录获取真实数据';
	} else if (serverMaintenance) {
		return '学校服务器维护中，功能暂时不可用';
	} else {
		return '功能暂时不可用，请稍后再试';
	}
}

/**
 * 获取当前教学周
 * @returns {Promise<number|null>} 当前教学周
 */
export async function getCurrentWeek() {
	try {
		const courseCacheStore = useCourseCache()
		const currentTimeData = await courseCacheStore.getCurrentTime()

		if (currentTimeData && currentTimeData.nowweek) {
			return currentTimeData.nowweek
		}

		return null
	} catch (error) {
		console.error('获取当前教学周失败:', error)
		return null
	}
}

/**
 * 清除加载状态的事件处理
 */
export function onClearAllLoadingStates(data) {
	console.log(`首页收到清除加载状态事件:`, data);
}
