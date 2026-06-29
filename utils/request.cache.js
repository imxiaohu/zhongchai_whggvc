// request.js - HTTP 缓存管理

import {
	HTTP_CACHE_PREFIX,
	NO_CACHE_PATHS,
	BASE_URL
} from './request.config.js';

export function stripBaseUrl(url) {
	if (!url) return '';
	if (url.startsWith(BASE_URL)) return url.slice(BASE_URL.length);
	return url;
}

export function isCacheableGet(options) {
	const method = (options.method || 'GET').toUpperCase();
	if (method !== 'GET') return false;
	if (options.responseType && options.responseType !== 'text') return false;
	const url = stripBaseUrl(options.url);
	if (!url) return false;
	if (NO_CACHE_PATHS.some((p) => url.includes(p))) return false;
	return options.cache !== false;
}

export function stableStringify(obj) {
	if (!obj || typeof obj !== 'object' || Array.isArray(obj)) return JSON.stringify(obj || {});
	const keys = Object.keys(obj).sort();
	const next = {};
	for (const k of keys) next[k] = obj[k];
	return JSON.stringify(next);
}

export function buildCacheKey(options) {
	const method = (options.method || 'GET').toUpperCase();
	const url = stripBaseUrl(options.url);
	const query = stableStringify(options.data);
	return `${HTTP_CACHE_PREFIX}${method}:${url}?${query}`;
}

export function formatDateTime(ts) {
	const d = new Date(ts);
	const pad = (n) => String(n).padStart(2, '0');
	return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

export function readHttpCache(cacheKey, allowStale) {
	try {
		const raw = uni.getStorageSync(cacheKey);
		if (!raw) return null;
		const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw;
		if (!parsed || !parsed.data || !parsed.expiresAt) return null;
		const now = Date.now();
		if (!allowStale && now > parsed.expiresAt) return null;
		return parsed;
	} catch (e) {
		return null;
	}
}

export function writeHttpCache(cacheKey, data, ttlSec) {
	try {
		if (!data || typeof data !== 'object') return;
		const now = Date.now();
		const record = {
			data,
			cachedAt: now,
			expiresAt: now + ttlSec * 1000
		};
		uni.setStorageSync(cacheKey, JSON.stringify(record));
	} catch (e) {
		return;
	}
}

export function decorateCachedResponse(record) {
	const cacheUpdatedAt = formatDateTime(record.cachedAt);
	const base = record.data;
	if (base && typeof base === 'object' && !Array.isArray(base)) {
		return {
			...base,
			fromCache: true,
			dataSourceType: base.dataSourceType || 'database',
			cacheUpdatedAt: base.cacheUpdatedAt || cacheUpdatedAt
		};
	}
	return {
		result: base,
		success: true,
		code: 200,
		message: '缓存命中',
		fromCache: true,
		dataSourceType: 'database',
		cacheUpdatedAt
	};
}

export function shouldFallbackToCache(error) {
	if (!error) return false;
	const status = error.statusCode || error.status;
	if (status === 401) return true;
	if (status === 408) return true;
	if (status === 503) return true;
	if (status >= 500 && status < 600) return true;
	if (error.isTimeout) return true;
	const msg = (error.message || error.errMsg || '').toLowerCase();
	const needles = ['network', 'timeout', 'eof', 'connection', 'failed', '失败', '超时'];
	return needles.some((n) => msg.includes(n));
}
