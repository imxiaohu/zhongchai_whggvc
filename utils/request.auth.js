// request.js - 认证与客户端标识

import { SIGN_MAP } from './request.config.js';

function generateTimestamp() {
	const now = new Date();
	const year = now.getFullYear();
	const month = String(now.getMonth() + 1).padStart(2, '0');
	const day = String(now.getDate()).padStart(2, '0');
	const hour = String(now.getHours()).padStart(2, '0');
	const minute = String(now.getMinutes()).padStart(2, '0');
	const second = String(now.getSeconds()).padStart(2, '0');
	return `${year}${month}${day}${hour}${minute}${second}`;
}

function getSignature(url) {
	for (const [key, value] of Object.entries(SIGN_MAP)) {
		if (key !== 'default' && url.includes(key)) {
			return value;
		}
	}
	return SIGN_MAP.default;
}

export { generateTimestamp, getSignature };

export function addClientId(options) {
	let clientId = uni.getStorageSync('clientId');

	if (shouldUseSchoolClientId(options.url)) {
		const schoolClientId = uni.getStorageSync('schoolClientId');
		if (schoolClientId) {
			clientId = schoolClientId;
		}
	}

	if (clientId) {
		options.header['X-Client-Id'] = clientId.includes(',') ? clientId.split(',')[0].trim() : clientId;
	}
}

export function shouldUseSchoolClientId(url) {
	return url.includes('/scloud/') &&
		   url !== '/scloud/init' &&
		   url !== '/scloud/validateCode' &&
		   url !== '/scloud/login';
}

export function addAuthentication(options) {
	const token = uni.getStorageSync('token');

	if (!token) {
		if (needsAuthentication(options.url)) {
			console.warn('请求需要授权，但未找到token');
		}
		return;
	}

	if (needsAuthentication(options.url)) {
		addBearerAuth(options, token);
	}
}

export function isCommunityAPI(url) {
	return url.includes('/api/community/') ||
		   url.includes('/api/clubs') ||
		   url.includes('/api/posts') ||
		   url.includes('/api/comments') ||
		   url.includes('/api/bookmarks') ||
		   url.includes('/api/reports') ||
		   url.includes('/api/notifications') ||
		   url.includes('/api/users/') ||
		   url.includes('/api/upload');
}

export function needsAuthentication(url) {
	return url.includes('/api/user/') ||
		   url.includes('/api/sync/') ||
		   url.includes('/api/settings/') ||
		   url.includes('/api/auth/') ||
		   url.includes('/api/health/') ||
		   url.includes('/api/pc/') ||
		   url.includes('/scloud/educational/') ||
		   url.includes('/scloud/courseTimetableDetail/') ||
		   url.includes('/scloud/courseTimetable/') ||
		   (url.includes('/scloudoa/') && !url.includes('/scloudoa/sys/mLogin')) ||
		   (url.includes('/api/m/') && !url.includes('/api/m/sys/mLogin')) ||
		   url.includes('/api/bookmarks') ||
		   url.includes('/api/reports') ||
		   url.includes('/api/notifications') ||
		   url.includes('/api/notification-') ||
		   url.includes('/api/sms/') ||
		   isCommunityAPI(url);
}

export function addAccessTokenAuth(options, token) {
	options.header['x-access-token'] = token;
	const timestamp = generateTimestamp();
	options.header['x-timestamp'] = timestamp;
	options.header['x-sign'] = getSignature(options.url);
}

export function addBearerAuth(options, token) {
	options.header['Authorization'] = `Bearer ${token}`;
}

export function clearSchoolLoginData() {
	const preserveClientId = uni.getStorageSync('clientId');
	const savedUsername = uni.getStorageSync('saved_username');
	const savedPassword = uni.getStorageSync('saved_password');
	const rememberPassword = uni.getStorageSync('remember_password');
	const userInfoStr = uni.getStorageSync('userInfo');

	uni.removeStorageSync('token');
	uni.removeStorageSync('loginType');
	uni.removeStorageSync('schoolClientId');

	if (preserveClientId) uni.setStorageSync('clientId', preserveClientId);
	if (rememberPassword) {
		uni.setStorageSync('saved_username', savedUsername);
		uni.setStorageSync('saved_password', savedPassword);
		uni.setStorageSync('remember_password', true);
	}
	if (userInfoStr) {
		uni.setStorageSync('userInfo', userInfoStr);
	}
}

export function saveClientIdFromResponse(response) {
	let clientId = null;

	const headerClientId = response.header?.['x-client-id'] || response.header?.['X-Client-Id'];
	if (headerClientId) {
		clientId = headerClientId;
	}

	if (!clientId && response.data) {
		let data = response.data;

		if (typeof data === 'string') {
			try {
				data = JSON.parse(data);
			} catch (e) { /* ignore */ }
		}

		if (typeof data === 'object' && data !== null) {
			if (data.clientId) {
				clientId = data.clientId;
			} else if (data.result && data.result.clientId) {
				clientId = data.result.clientId;
			}
		}
	}

	if (clientId) {
		uni.setStorageSync('clientId', clientId);
	}
}

export function saveClientIdFromRequestResponse(res) {
	const clientId = res.header?.['x-client-id'] || res.header?.['X-Client-Id'];
	if (clientId) {
		uni.setStorageSync('clientId', clientId);
	}
}
