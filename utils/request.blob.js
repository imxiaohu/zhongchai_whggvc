// request.js - Blob 下载模块

import { BASE_URL, HEADERS } from './request.config.js';
import { handleRequestError as handleError } from './errorHandler.js';

export async function handleBlobRequest(options, resolve, reject) {
	try {
		const fullUrl = options.url.startsWith('http') ? options.url : `${BASE_URL}${options.url}`;

		const headers = {
			...options.header,
			'Accept': 'application/octet-stream, */*',
			'Accept-Language': HEADERS.ACCEPT_LANGUAGE,
			'Cache-Control': HEADERS.CACHE_CONTROL,
			'Pragma': HEADERS.PRAGMA
		};

		const token = uni.getStorageSync('token');
		if (token) {
			headers['Authorization'] = `Bearer ${token}`;
		}

		const clientId = uni.getStorageSync('clientId');
		if (clientId) {
			headers['X-Client-Id'] = clientId;
		}

		const controller = new AbortController();
		const timeout = options.timeout || 10000;

		const timeoutId = setTimeout(() => {
			controller.abort();
		}, timeout);

		const fetchOptions = {
			method: options.method || 'GET',
			headers,
			credentials: 'include',
			signal: controller.signal
		};

		if (options.data && (options.method === 'POST' || options.method === 'PUT')) {
			fetchOptions.body = JSON.stringify(options.data);
			headers['Content-Type'] = 'application/json';
		}

		const response = await fetch(fullUrl, fetchOptions);

		clearTimeout(timeoutId);

		if (!response.ok) {
			throw new Error(`HTTP ${response.status}: ${response.statusText}`);
		}

		const blob = await response.blob();

		resolve({
			data: blob,
			statusCode: response.status,
			headers: response.headers
		});

	} catch (error) {
		console.error('handleBlobRequest: blob请求失败:', error);

		if (error.name === 'AbortError') {
			const timeoutError = {
				statusCode: 408,
				message: '文件下载超时，请检查网络连接',
				isTimeout: true
			};
			handleError(timeoutError, { showDetails: false });
			reject(timeoutError);
		} else {
			reject({
				message: error.message || '下载失败',
				error
			});
		}
	}
}
