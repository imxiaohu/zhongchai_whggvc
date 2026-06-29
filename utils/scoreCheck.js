/**
 * Score Check Utilities
 * Score check frequency options, semester info, log formatters, and result parsers
 */

/**
 * Score check frequency options (already defined in notification.js, repeated here for standalone use)
 */
export const SCORE_CHECK_FREQUENCY_OPTIONS = [
	{ label: '每小时', value: 'hourly' },
	{ label: '每天', value: 'daily' },
	{ label: '每周', value: 'weekly' }
];

/**
 * Default score check settings
 */
export const DEFAULT_SCORE_CHECK = {
	enabled: false,
	frequency: 'daily',
	time: '09:00',
	semester: 'current'
};

/**
 * Log type definitions
 */
export const LOG_TYPES = {
	BATCH_CACHE: 'batch_cache',
	SCORE_UPDATE: 'score_update',
	SCORE_CHECK: 'score_check'
};

/**
 * Log status definitions
 */
export const LOG_STATUS = {
	SUCCESS: 'success',
	FAILED: 'failed',
	PENDING: 'pending'
};

/**
 * Get log status text
 * @param {string} status - Log status
 * @returns {string} Status text in Chinese
 */
export function getLogStatusText(status) {
	switch (status) {
		case 'success': return '成功';
		case 'failed': return '失败';
		case 'pending': return '待处理';
		default: return status || '未知';
	}
}

/**
 * Get log status CSS class
 * @param {string} status - Log status
 * @returns {string} CSS class name
 */
export function getLogStatusClass(status) {
	switch (status) {
		case 'success': return 'success';
		case 'error': return 'error';
		default: return '';
	}
}

/**
 * Get log type text
 * @param {string} type - Log type
 * @returns {string} Type text in Chinese
 */
export function getLogTypeText(type) {
	switch (type) {
		case 'batch_cache': return '批量缓存';
		case 'score_update': return '成绩更新';
		case 'score_check': return '成绩检查';
		default: return type || '未知类型';
	}
}

/**
 * Format log time for display
 * @param {string} timeStr - ISO time string
 * @returns {string} Formatted time string
 */
export function formatLogTime(timeStr) {
	if (!timeStr) return '';
	try {
		const date = new Date(timeStr);
		if (isNaN(date.getTime())) return '';
		return date.toLocaleString('zh-CN', {
			year: 'numeric',
			month: '2-digit',
			day: '2-digit',
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit'
		});
	} catch (e) {
		return '';
	}
}

/**
 * Parse score check logs from API response
 * @param {object} response - API response
 * @returns {object} Parsed logs with pagination
 */
export function parseScoreCheckLogs(response) {
	if (!response.success) {
		return {
			logs: [],
			pagination: { page: 1, pageSize: 20, total: 0, pages: 0 }
		};
	}

	const data = response.result;
	return {
		logs: data.logs || [],
		pagination: {
			page: data.page || 1,
			pageSize: data.pageSize || 20,
			total: data.total || 0,
			pages: data.pages || 0
		}
	};
}

/**
 * Default pagination object
 */
export const DEFAULT_PAGINATION = {
	page: 1,
	pageSize: 20,
	total: 0,
	pages: 0
};

/**
 * Parse test notification result message
 * @param {object} result - Test notification result
 * @returns {string} Human-readable message
 */
export function parseTestResultMessage(result) {
	if (result.success) {
		const data = result.result;
		return data.message || `测试完成，成功 ${data.successCount}/${data.totalCount} 个通知渠道`;
	}
	return result.message || '测试失败';
}

/**
 * Validate score check settings
 * @param {object} scoreCheck - Score check settings
 * @returns {object} Validation result with isValid and error message
 */
export function validateScoreCheckSettings(scoreCheck) {
	if (!scoreCheck) {
		return { isValid: false, error: '成绩检查设置不能为空' };
	}

	if (scoreCheck.frequency && !SCORE_CHECK_FREQUENCY_OPTIONS.some(opt => opt.value === scoreCheck.frequency)) {
		return { isValid: false, error: '无效的检查频率' };
	}

	if (scoreCheck.time) {
		const timeRegex = /^([01]\d|2[0-3]):([0-5]\d)$/;
		if (!timeRegex.test(scoreCheck.time)) {
			return { isValid: false, error: '时间格式无效，请使用 HH:mm 格式' };
		}
	}

	return { isValid: true };
}

/**
 * Format test result summary
 * @param {object} testData - Test result data
 * @returns {string} Formatted summary
 */
export function formatTestResultSummary(testData) {
	const { successCount = 0, totalCount = 0, results = [] } = testData;

	let summary = `成功 ${successCount}/${totalCount}`;
	if (results.length > 0) {
		const failedChannels = results.filter(r => !r.success).map(r => r.channel);
		if (failedChannels.length > 0) {
			summary += ` (失败: ${failedChannels.join(', ')})`;
		}
	}
	return summary;
}
