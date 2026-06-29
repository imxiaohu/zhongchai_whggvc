// request.js - 辅助函数

import { BASE_URL, CONTENT_TYPES } from './request.config.js';

export function processBaseUrl(url) {
	if (!url.startsWith('http')) {
		return BASE_URL + url;
	}
	return url;
}

export function initializeHeaders(options) {
	return options.header || {};
}

export function getAcceptHeader(url) {
	if (url.includes('/validateCode')) {
		return CONTENT_TYPES.IMAGE;
	}
	if (url.includes('/scloud/login')) {
		return CONTENT_TYPES.HTML;
	}
	return 'application/json, text/plain, */*';
}

export function setContentType(options) {
	if (options.method === 'POST') {
		if (isComplexData(options.data)) {
			options.header['content-type'] = CONTENT_TYPES.JSON;
		} else {
			options.header['content-type'] = CONTENT_TYPES.FORM;
		}
	}
}

export function isComplexData(data) {
	return Array.isArray(data) ||
		   (typeof data === 'object' && data !== null && Object.keys(data).length > 0);
}
