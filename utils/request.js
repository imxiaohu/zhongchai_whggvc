// request.js - 请求工具主入口
// 职责：拦截器注册 + request/get/post 主方法 + 内部请求包装器

import { handleRequestError as handleError } from './errorHandler.js';
import { handleServerMaintenanceError } from './schoolServerStatus.js';
import { BASE_URL, CONTENT_TYPES, HEADERS, MAX_RETRY_COUNT, RETRY_DELAY } from './request.config.js';
import { processBaseUrl, initializeHeaders, getAcceptHeader, setContentType } from './request.helpers.js';
import { addClientId, addAuthentication, saveClientIdFromResponse } from './request.auth.js';
import {
	stripBaseUrl,
	isCacheableGet,
	buildCacheKey,
	readHttpCache,
	writeHttpCache,
	decorateCachedResponse,
	shouldFallbackToCache
} from './request.cache.js';
import {
	handleResponseStatus,
	handleRequestError,
	handleRequestUnauthorized,
	handleSuccessResponse,
	handleErrorStatusResponse,
	isValidateCodeRequest,
	handleMobileLoginResponse,
	handleEmptyResponse,
	maybePersistAuthFromResponse,
	setIsRetrying
} from './request.response.js';
import { handleBlobRequest } from './request.blob.js';

const retryQueue = new Set();

async function requestWithRetry(options, resolve, reject, retryCount = 0) {
	try {
		await uni.request({
			...options,
			timeout: options.timeout || 10000,
			success: (res) => {
				handleRequestSuccess(res, options, resolve, reject);
			},
			fail: (err) => {
				handleRequestError(err, reject);
			}
		});
	} catch (error) {
		if ((error.errMsg?.includes('network') || error.errMsg?.includes('timeout')) &&
			retryCount < MAX_RETRY_COUNT && !isRetrying()) {
			setIsRetrying(true);
			const requestId = `${options.url}_${Date.now()}`;
			if (!retryQueue.has(requestId)) {
				retryQueue.add(requestId);
				console.log(`请求失败，${RETRY_DELAY / 1000}秒后重试 (${retryCount + 1}/${MAX_RETRY_COUNT})`, options.url);
				setTimeout(async () => {
					retryQueue.delete(requestId);
					setIsRetrying(false);
					await requestWithRetry(options, resolve, reject, retryCount + 1);
				}, RETRY_DELAY);
				return;
			}
		}
		reject(error);
	}
}

let isRetryingFlag = false;
function isRetrying() { return isRetryingFlag; }

function handleRequestSuccess(res, options, resolve, reject) {
	saveClientIdFromRequestResponse(res);
	handleMobileLoginResponse(res, options);

	if (isValidateCodeRequest(options.url)) {
		resolve(res);
		return;
	}

	if (res.statusCode === 401) {
		handleRequestUnauthorized(reject);
		return;
	}

	if (res.statusCode === 200) {
		handleSuccessResponse(res, resolve, reject);
	} else {
		handleErrorStatusResponse(res, reject);
	}
}

function saveClientIdFromRequestResponse(res) {
	const clientId = res.header?.['x-client-id'] || res.header?.['X-Client-Id'];
	if (clientId) {
		uni.setStorageSync('clientId', clientId);
	}
}

// 请求拦截器
uni.addInterceptor('request', {
	invoke(options) {
		options.url = processBaseUrl(options.url);
		options.header = initializeHeaders(options);
		options.header = {
			...options.header,
			'Accept': getAcceptHeader(options.url),
			'Accept-Language': HEADERS.ACCEPT_LANGUAGE,
			'Cache-Control': HEADERS.CACHE_CONTROL,
			'Pragma': HEADERS.PRAGMA
		};
		addClientId(options);
		addAuthentication(options);
		setContentType(options);
		options.withCredentials = true;
		return options;
	},

	success(response) {
		return handleResponseStatus(response);
	},

	fail(error) {
		return handleRequestFailure(error);
	}
});

// 主请求方法
export const request = (options) => {
	return new Promise(async (resolve, reject) => {
		const timeout = options.timeout || 10000;
		let isResolved = false;

		const cacheEnabled = isCacheableGet(options);
		const cacheKey = cacheEnabled ? buildCacheKey(options) : null;
		const ttlSec = typeof options.cacheTTL === 'number' ? options.cacheTTL : 600;
		const networkType = await getNetworkTypeSafe();
		const offline = networkType === 'none';
		if (offline && cacheEnabled && cacheKey) {
			const cached = readHttpCache(cacheKey, true);
			if (cached) {
				resolve(decorateCachedResponse(cached));
				return;
			}
		}

		const timeoutTimer = setTimeout(() => {
			if (!isResolved) {
				isResolved = true;
				const timeoutError = {
					statusCode: 408,
					message: '服务器连接超时，请检查网络连接',
					isTimeout: true
				};
				handleError(timeoutError, { showDetails: false });
				reject(timeoutError);
			}
		}, timeout);

		const wrappedResolve = (result) => {
			if (!isResolved) {
				isResolved = true;
				clearTimeout(timeoutTimer);
				if (cacheEnabled && cacheKey && result && typeof result === 'object') {
					writeHttpCache(cacheKey, result, ttlSec);
				}
				maybePersistAuthFromResponse(options.url, result);
				resolve(result);
			}
		};

		const wrappedReject = (error) => {
			if (!isResolved) {
				isResolved = true;
				clearTimeout(timeoutTimer);
				if (cacheEnabled && cacheKey && shouldFallbackToCache(error)) {
					const cached = readHttpCache(cacheKey, true);
					if (cached) {
						resolve(decorateCachedResponse(cached));
						return;
					}
				}
				reject(error);
			}
		};

		if (options.responseType === 'blob') {
			handleBlobRequest(options, wrappedResolve, wrappedReject);
		} else {
			requestWithRetry(options, wrappedResolve, wrappedReject);
		}
	});
};

export const get = (url, data = {}, options = {}) => {
	return request({ url, data, method: 'GET', timeout: 10000, ...options });
};

export const post = (url, data = {}, options = {}) => {
	return request({ url, data, method: 'POST', timeout: 10000, ...options });
};

function getNetworkTypeSafe() {
	return new Promise((resolve) => {
		try {
			uni.getNetworkType({
				success: (res) => resolve(res.networkType || 'unknown'),
				fail: () => resolve('unknown')
			});
		} catch (e) {
			resolve('unknown');
		}
	});
}

function handleRequestFailure(error) {
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

// 此模块不直接定义 handleErrorStatusResponse，统一从 request.response.js 导入
