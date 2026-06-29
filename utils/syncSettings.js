/**
 * Sync Settings Utilities
 * Sync settings configuration, formatters, and helper functions
 */

// Sync frequency options
export const SYNC_FREQUENCY_OPTIONS = [
	{ text: '每天同步一次', value: 'daily' },
	{ text: '每周同步一次', value: 'weekly' },
	{ text: '每两天同步一次', value: 'every2days' },
	{ text: '每三天同步一次', value: 'every3days' }
];

// Time range presets
export const TIME_RANGE_PRESETS = [
	{ text: '上午 08:30-12:00', value: '08:30-12:00', desc: '适合课表空档同步' },
	{ text: '中午 12:00-14:00', value: '12:00-14:00', desc: '午间低峰同步' },
	{ text: '下午 14:00-18:00', value: '14:00-18:00', desc: '下午窗口同步' },
	{ text: '晚上 18:00-22:20', value: '18:00-22:20', desc: '晚间稳定时段' },
	{ text: '全天 08:30-22:20', value: '08:30-22:20', desc: '全天开放' }
];

// Sync status configurations
export const SYNC_STATUS_CONFIG = {
	idle: { icon: 'check-circle', color: '#6b7280', title: '等待同步' },
	syncing: { icon: 'refresh', color: '#3b82f6', title: '正在同步' },
	success: { icon: 'check-circle', color: '#10b981', title: '同步成功' },
	failed: { icon: 'close-circle', color: '#ef4444', title: '同步失败' }
};

// Server time constraints
export const SERVER_TIME = {
	START: '08:30',
	END: '22:20'
};

/**
 * Get sync frequency text
 * @param {string} frequency - Frequency value
 * @returns {string} Display text
 */
export function getFrequencyText(frequency) {
	const textMap = {
		'daily': '每天',
		'weekly': '每周',
		'every2days': '每两天',
		'every3days': '每三天'
	};
	return textMap[frequency] || '未设置';
}

/**
 * Get sync frequency description
 * @param {string} frequency - Frequency value
 * @returns {string} Description text
 */
export function getFrequencyDescription(frequency) {
	const descMap = {
		'daily': '每天同步一次',
		'weekly': '每周同步一次',
		'every2days': '每两天同步一次',
		'every3days': '每三天同步一次'
	};
	return descMap[frequency] || '自定义频率';
}

/**
 * Get status icon
 * @param {string} status - Sync status
 * @returns {string} Icon name
 */
export function getStatusIcon(status) {
	return SYNC_STATUS_CONFIG[status]?.icon || 'help-circle';
}

/**
 * Get status color
 * @param {string} status - Sync status
 * @returns {string} Color hex
 */
export function getStatusColor(status) {
	return SYNC_STATUS_CONFIG[status]?.color || '#6b7280';
}

/**
 * Get status title
 * @param {string} status - Sync status
 * @param {boolean} enabled - Whether sync is enabled
 * @returns {string} Status title
 */
export function getStatusTitle(status, enabled) {
	if (!enabled) return '同步未开启';
	return SYNC_STATUS_CONFIG[status]?.title || '未知状态';
}

/**
 * Get status icon class for animation
 * @param {string} status - Sync status
 * @returns {string} CSS class name
 */
export function getStatusIconClass(status) {
	return status === 'syncing' ? 'rotating' : '';
}

/**
 * Format relative time (for history display)
 * @param {string|number|Date} timeStr - Time to format
 * @returns {string} Formatted relative time
 */
export function formatRelativeTime(timeStr) {
	if (!timeStr) return '';

	const date = new Date(timeStr);
	const now = new Date();
	const diff = now - date;
	const diffHours = Math.floor(diff / (1000 * 60 * 60));
	const diffDays = Math.floor(diffHours / 24);

	if (diffHours < 1) {
		return '刚刚';
	} else if (diffHours < 24) {
		return `${diffHours}小时前`;
	} else if (diffDays < 7) {
		return `${diffDays}天前`;
	} else {
		return date.toLocaleDateString() + ' ' + date.toLocaleTimeString().slice(0, 5);
	}
}

/**
 * Format next sync time with friendly display
 * @param {string|number|Date} timeStr - Time to format
 * @returns {string} Formatted time
 */
export function formatNextSyncTime(timeStr) {
	if (!timeStr) return '';

	const date = new Date(timeStr);
	const now = new Date();

	if (isNaN(date.getTime())) {
		return '时间格式无效';
	}

	if (date.toDateString() === now.toDateString()) {
		return `今天 ${date.toLocaleTimeString().slice(0, 5)}`;
	}

	const tomorrow = new Date(now);
	tomorrow.setDate(tomorrow.getDate() + 1);
	if (date.toDateString() === tomorrow.toDateString()) {
		return `明天 ${date.toLocaleTimeString().slice(0, 5)}`;
	}

	const dayAfterTomorrow = new Date(now);
	dayAfterTomorrow.setDate(dayAfterTomorrow.getDate() + 2);
	if (date.toDateString() === dayAfterTomorrow.toDateString()) {
		return `后天 ${date.toLocaleTimeString().slice(0, 5)}`;
	}

	const month = date.getMonth() + 1;
	const day = date.getDate();
	const hours = date.getHours().toString().padStart(2, '0');
	const minutes = date.getMinutes().toString().padStart(2, '0');

	return `${month}月${day}日 ${hours}:${minutes}`;
}

/**
 * Format time range for display
 * @param {string} timeRange - Time range string (e.g., "08:30-12:00")
 * @returns {string} Formatted time range
 */
export function formatTimeRange(timeRange) {
	if (!timeRange) return '未设置时间范围';

	const presetMap = {
		'08:30-12:00': '上午 08:30-12:00',
		'12:00-14:00': '中午 12:00-14:00',
		'14:00-18:00': '下午 14:00-18:00',
		'18:00-22:20': '晚上 18:00-22:20',
		'08:30-22:20': '全天 08:30-22:20'
	};

	if (presetMap[timeRange]) {
		return presetMap[timeRange];
	}

	const [startTime, endTime] = timeRange.split('-');
	return `${startTime} - ${endTime}`;
}

/**
 * Convert time string to minutes
 * @param {string} timeStr - Time string (e.g., "08:30")
 * @returns {number} Minutes since midnight
 */
export function timeToMinutes(timeStr) {
	const [hours, minutes] = timeStr.split(':').map(Number);
	return hours * 60 + minutes;
}

/**
 * Validate time range is within server hours
 * @param {string} startTime - Start time
 * @param {string} endTime - End time
 * @returns {boolean} Whether time range is valid
 */
export function validateTimeRange(startTime, endTime) {
	const start = timeToMinutes(startTime);
	const end = timeToMinutes(endTime);

	const serverStart = timeToMinutes(SERVER_TIME.START);
	const serverEnd = timeToMinutes(SERVER_TIME.END);

	if (start < serverStart || start > serverEnd || end < serverStart || end > serverEnd) {
		return false;
	}

	return start < end;
}

/**
 * Get auto retry description
 * @param {number} maxRetryCount - Maximum retry count
 * @returns {string} Description text
 */
export function getAutoRetryDescription(maxRetryCount) {
	return `自动重试${maxRetryCount}次，直至成功`;
}

/**
 * Get default sync settings
 * @returns {Object} Default settings object
 */
export function getDefaultSyncSettings() {
	return {
		enabled: false,
		frequency: 'daily',
		timeRange: '08:30-22:20',
		autoRetryEnabled: true,
		maxRetryCount: 3
	};
}

/**
 * Get default sync status
 * @returns {Object} Default status object
 */
export function getDefaultSyncStatus() {
	return {
		syncStatus: 'idle',
		lastSyncAt: null,
		nextSyncAt: null,
		lastSyncMessage: '',
		coursesCount: 0
	};
}
