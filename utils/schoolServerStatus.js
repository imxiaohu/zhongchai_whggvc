/**
 * 学校服务器状态管理工具
 */

import { request } from './request.js';

// 服务器状态缓存
let serverStatusCache = {
	isAlive: true,
	lastCheck: null,
	cacheTime: 0
};

// 缓存有效期（5分钟）
const CACHE_DURATION = 5 * 60 * 1000;

/**
 * 检查学校服务器状态
 * @param {boolean} forceCheck 是否强制检查（忽略缓存）
 * @returns {Promise<Object>} 服务器状态信息
 */
export async function checkSchoolServerStatus(forceCheck = false) {
	const now = Date.now();
	
	// 如果缓存有效且不强制检查，返回缓存数据
	if (!forceCheck && serverStatusCache.cacheTime && (now - serverStatusCache.cacheTime < CACHE_DURATION)) {
		return serverStatusCache;
	}
	
	try {
		const response = await request({
			url: '/api/health/school-server',
			method: 'GET',
			timeout: 5000 // 5秒超时
		});
		
		if (response && response.result) {
			serverStatusCache = {
				...response.result,
				cacheTime: now
			};
		}
		
		return serverStatusCache;
	} catch (error) {
		console.warn('检查学校服务器状态失败:', error);
		// 如果检查失败，假设服务器不可用
		serverStatusCache = {
			isAlive: false,
			lastCheck: new Date().toISOString(),
			errorMsg: '无法连接到服务器',
			cacheTime: now
		};
		return serverStatusCache;
	}
}

/**
 * 获取维护信息
 * @returns {Promise<Object>} 维护信息
 */
export async function getMaintenanceInfo() {
	try {
		const response = await request({
			url: '/api/health/maintenance-info',
			method: 'GET',
			timeout: 5000
		});
		
		return response?.result || {};
	} catch (error) {
		console.warn('获取维护信息失败:', error);
		return {
			isServerAlive: false,
			maintenanceMsg: '学校服务器正在维护，请稍后再试'
		};
	}
}

/**
 * 强制检查学校服务器状态
 * @returns {Promise<Object>} 检查结果
 */
export async function forceCheckSchoolServer() {
	try {
		const response = await request({
			url: '/api/health/school-server/check',
			method: 'POST',
			timeout: 10000 // 10秒超时
		});
		
		// 清除缓存，强制下次重新检查
		serverStatusCache.cacheTime = 0;
		
		return response;
	} catch (error) {
		console.warn('强制检查学校服务器失败:', error);
		throw error;
	}
}

/**
 * 处理API响应，检查是否为服务器维护错误
 * @param {Object} error 错误对象
 * @returns {boolean} 是否为服务器维护错误
 */
export function isServerMaintenanceError(error) {
	// 检查HTTP状态码
	if (error.statusCode === 503 || error.status === 503) {
		return true;
	}
	
	// 检查错误消息
	const message = error.message || error.errMsg || '';
	if (message.includes('学校服务器正在维护') || 
		message.includes('服务器暂时关闭') ||
		message.includes('server maintenance')) {
		return true;
	}
	
	// 检查响应数据
	if (error.data && error.data.code === 503) {
		return true;
	}
	
	return false;
}

/**
 * 显示服务器维护提示
 * @param {Object} maintenanceInfo 维护信息
 */
export function showMaintenanceMessage(maintenanceInfo = {}) {
	const message = maintenanceInfo.maintenanceMsg || '学校服务器正在维护，请稍后再试';
	const lastAlive = maintenanceInfo.lastAlive;
	
	let detailMessage = message;
	if (lastAlive) {
		detailMessage += `\n最后连接时间: ${lastAlive}`;
	}
	
	// 使用uni-app的模态框显示
	uni.showModal({
		title: '服务器维护',
		content: detailMessage,
		showCancel: true,
		cancelText: '知道了',
		confirmText: '重试',
		success: (res) => {
			if (res.confirm) {
				// 用户点击重试，强制检查服务器状态
				forceCheckSchoolServer().then(() => {
					uni.showToast({
						title: '已重新检查服务器状态',
						icon: 'success'
					});
				}).catch(() => {
					uni.showToast({
						title: '服务器仍在维护中',
						icon: 'none'
					});
				});
			}
		}
	});
}

/**
 * 在请求拦截器中使用的错误处理函数
 * @param {Object} error 请求错误
 * @returns {Promise} 处理后的Promise
 */
export function handleServerMaintenanceError(error) {
	if (isServerMaintenanceError(error)) {
		// 异步显示维护提示（不影响主流程）
		getMaintenanceInfo().then(maintenanceInfo => {
			showMaintenanceMessage(maintenanceInfo);
		}).catch(err => {
			console.warn('获取维护信息失败，使用默认信息:', err);
			showMaintenanceMessage({
				maintenanceMsg: '学校服务器正在维护，请稍后再试'
			});
		});

		// 返回 rejected Promise，让外层 catch 捕获
		return Promise.reject({
			...error,
			isMaintenanceError: true,
			message: '学校服务器正在维护'
		});
	}

	// 不是维护错误，返回 rejected Promise
	return Promise.reject(error);
}

/**
 * 获取服务器状态的简单检查（同步）
 * @returns {boolean} 服务器是否可用
 */
export function isSchoolServerAlive() {
	const now = Date.now();
	
	// 如果缓存有效，返回缓存状态
	if (serverStatusCache.cacheTime && (now - serverStatusCache.cacheTime < CACHE_DURATION)) {
		return serverStatusCache.isAlive;
	}
	
	// 缓存无效，默认认为服务器可用（避免阻塞）
	return true;
}

/**
 * 清除服务器状态缓存
 */
export function clearServerStatusCache() {
	serverStatusCache = {
		isAlive: true,
		lastCheck: null,
		cacheTime: 0
	};
}

export default {
	checkSchoolServerStatus,
	getMaintenanceInfo,
	forceCheckSchoolServer,
	isServerMaintenanceError,
	showMaintenanceMessage,
	handleServerMaintenanceError,
	isSchoolServerAlive,
	clearServerStatusCache
};
