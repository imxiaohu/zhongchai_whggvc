// request.js - 响应处理模块

import { handleRequestError as handleError } from './errorHandler.js';
import { showMaintenanceMessage, getMaintenanceInfo } from './schoolServerStatus.js';
import { clearSchoolLoginData } from './request.auth.js';
import { stripBaseUrl } from './request.cache.js';

// 标记请求是否正在重试
let isRetrying = false;

export function getIsRetrying() { return isRetrying; }
export function setIsRetrying(v) { isRetrying = v; }

const NO_REDIRECT_PAGES = [
	'index/index', 'login/login', 'user/bind', 'user/index',
	'evaluation/list', 'evaluation/swipe', 'evaluation/detail', 'community/index'
];

function getCurrentRoute() {
	const pages = getCurrentPages();
	const currentPage = pages[pages.length - 1];
	return currentPage ? currentPage.route : '';
}

function shouldNotRedirect() {
	return NO_REDIRECT_PAGES.some(page => getCurrentRoute().includes(page));
}

export function handleResponseStatus(response) {
	if (response.statusCode === 401) {
		return false;
	} else if (response.statusCode === 503) {
		const error = {
			statusCode: response.statusCode,
			status: response.statusCode,
			message: response.data?.message || '学校服务器正在维护',
			data: response.data,
			isMaintenanceError: true
		};
		// 异步获取维护信息并显示弹窗（不阻塞错误传播）
		getMaintenanceInfo().then(info => showMaintenanceMessage(info)).catch(() => {
			showMaintenanceMessage({ maintenanceMsg: '学校服务器正在维护，请稍后再试' });
		});
		// throw 让 uni-request 拦截器捕获并路由到 fail 路径
		throw error;
	} else if (response.statusCode !== 200) {
		return handleErrorResponse(response);
	}
	return response;
}

export function handleErrorResponse(response) {
	let errorMessage = '服务器响应错误：' + response.statusCode;

	if (response.data) {
		if (typeof response.data === 'object' && response.data.message) {
			errorMessage = response.data.message;
		} else if (typeof response.data === 'string') {
			try {
				const parsedData = JSON.parse(response.data);
				if (parsedData.message) errorMessage = parsedData.message;
			} catch (e) {
				errorMessage = response.data;
			}
		}
	}

	const error = {
		statusCode: response.statusCode,
		message: errorMessage,
		data: response.data
	};

	if (response.statusCode >= 500 && response.statusCode < 600) {
		handleError(error, { showDetails: false });
	} else {
		uni.showToast({ title: errorMessage, icon: 'none' });
	}

	return false;
}

export function handleRequestFailure(error) {
	console.error('请求失败:', error);
	try {
		handleServerMaintenanceError(error);
		return false;
	} catch (maintenanceError) {
		if (!maintenanceError.isMaintenanceError) {
			handleError(error, { showDetails: false });
		}
		return false;
	}
}

export function handleRequestUnauthorized(reject) {
	const userInfoStr = uni.getStorageSync('userInfo');
	const isWechatLogin = !!userInfoStr;

	clearSchoolLoginData();

	if (isWechatLogin) {
		uni.setStorageSync('userInfo', userInfoStr);
	}

	if (shouldNotRedirect()) {
		reject({ message: '会话已失效，请重新登录', statusCode: 401, isTokenInvalid: true });
		return;
	}

	const hasWechatLogin = !!userInfoStr;

	if (hasWechatLogin) {
		uni.showToast({ title: '学校账号会话失效，请重新绑定', icon: 'none', duration: 3000 });
		reject({ message: '学校账号会话失效，请重新绑定', statusCode: 401, isTokenInvalid: true });
		return;
	}

	if (!isRetrying) {
		isRetrying = true;
		uni.showToast({ title: '登录已过期，请重新登录', icon: 'none' });

		setTimeout(() => {
			uni.reLaunch({ url: '/pages/login/login' });
			setTimeout(() => { isRetrying = false; }, 500);
		}, 1500);
	}

	reject({ message: '会话已失效，请重新登录' });
}

export function handleSuccessResponse(res, resolve, reject) {
	if (isEmptyObjectResponse(res.data)) {
		console.warn('服务器返回空对象，可能会话已失效');
		handleEmptyResponse(reject);
		return;
	}
	resolve(res.data);
}

export function isEmptyObjectResponse(data) {
	return typeof data === 'object' && !Array.isArray(data) && Object.keys(data).length === 0;
}

export function handleEmptyResponse(reject) {
	if (shouldNotRedirect()) {
		reject({ message: '登录状态异常，请重新登录', statusCode: 401, isTokenInvalid: true });
		return;
	}

	const userInfoStr = uni.getStorageSync('userInfo');
	const hasWechatLogin = !!userInfoStr;

	if (hasWechatLogin) {
		reject({ message: '学校账号会话失效，请重新绑定', statusCode: 401, isTokenInvalid: true });
		return;
	}

	if (!isRetrying) {
		isRetrying = true;
		setTimeout(() => {
			uni.reLaunch({ url: '/pages/login/login' });
			setTimeout(() => { isRetrying = false; }, 500);
		}, 500);
	}

	reject({ message: '登录状态异常，请重新登录' });
}

export function isValidateCodeRequest(url) {
	return url.includes('/validateCode');
}

export function isMobileLoginRequest(url) {
	return url.includes('/sys/mLogin');
}

export function handleMobileLoginResponse(res, options) {
	if (options.url.includes('/sys/mLogin')) {
		if (res.statusCode === 200 && res.data && res.data.result && res.data.result.token) {
			uni.setStorageSync('token', res.data.result.token);
			uni.setStorageSync('loginType', 'mobile');
		}
	}
}

export function handleRequestError(err, reject) {
	console.error('请求失败 (in request):', err);

	if (err.errMsg && err.errMsg.includes('timeout')) {
		const timeoutError = {
			statusCode: 408,
			message: '请求超时，服务器无响应',
			isTimeout: true,
			originalError: err
		};
		handleError(timeoutError, { showDetails: false });
		reject(timeoutError);
	} else {
		handleError(err, { showDetails: false });
		reject(err);
	}
}

export function handleErrorStatusResponse(res, reject) {
	let errorMessage = '请求失败';

	if (res.data) {
		if (typeof res.data === 'object' && res.data.message) {
			errorMessage = res.data.message;
		} else if (typeof res.data === 'string') {
			try {
				const parsedData = JSON.parse(res.data);
				if (parsedData.message) errorMessage = parsedData.message;
			} catch (e) {
				errorMessage = res.data;
			}
		}
	}

	const error = {
		statusCode: res.statusCode,
		message: errorMessage,
		data: res.data
	};

	reject(error);
}

export function maybePersistAuthFromResponse(url, data) {
	try {
		const u = stripBaseUrl(url);
		const isLogin = u.includes('/scloud/login') || u.includes('/sys/mLogin');
		if (!isLogin) return;
		const token = data && data.result && data.result.token;
		if (!token) return;
		uni.setStorageSync('token', token);
		if (u.includes('/sys/mLogin')) {
			uni.setStorageSync('loginType', 'mobile');
		} else {
			uni.setStorageSync('loginType', 'school');
		}
	} catch (e) {
		return;
	}
}
