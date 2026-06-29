// 同步相关API接口

// 导入请求工具类
import { request as utilRequest } from '../utils/request.js';

// 注意：现在使用统一的请求工具类，不需要单独的URL和token处理

// 通用请求方法（使用统一的请求工具类）
function request(options) {
	// 使用统一的请求工具类，它会自动处理认证、错误处理等
	return utilRequest({
		...options,
		// 确保URL是相对路径，让request.js处理BASE_URL
		url: options.url
	});
}

// 获取同步设置
export function getSyncSettings() {
	return request({
		url: '/api/sync/settings',
		method: 'GET'
	});
}

// 更新同步设置
export function updateSyncSettings(settings) {
	return request({
		url: '/api/sync/settings',
		method: 'POST',
		data: settings
	});
}

// 获取同步状态
export function getSyncStatus() {
	return request({
		url: '/api/sync/status',
		method: 'GET'
	});
}

// 手动触发同步
export function manualSync() {
	return request({
		url: '/api/sync/manual',
		method: 'POST'
	});
}

// 获取同步日志
export function getSyncLogs(limit = 20) {
	return request({
		url: `/api/sync/logs?limit=${limit}`,
		method: 'GET'
	});
}

// 同步设置管理类
export class SyncManager {
	constructor() {
		this.settings = null;
		this.status = null;
		this.logs = [];
	}

	// 初始化，加载所有同步相关数据
	async init() {
		try {
			const [settingsRes, statusRes, logsRes] = await Promise.allSettled([
				getSyncSettings(),
				getSyncStatus(),
				getSyncLogs(10)
			]);

			if (settingsRes.status === 'fulfilled') {
				this.settings = settingsRes.value.result;
			}

			if (statusRes.status === 'fulfilled') {
				this.status = statusRes.value.result;
			}

			if (logsRes.status === 'fulfilled') {
				this.logs = logsRes.value.result || [];
			}

			return {
				settings: this.settings,
				status: this.status,
				logs: this.logs
			};
		} catch (error) {
			console.error('初始化同步管理器失败:', error);
			throw error;
		}
	}

	// 启用/禁用同步
	async toggleSync(enabled) {
		if (!this.settings) {
			throw new Error('同步设置未加载');
		}

		const updatedSettings = {
			...this.settings,
			enabled
		};

		const result = await updateSyncSettings(updatedSettings);
		this.settings = result.result;
		return this.settings;
	}

	// 更新同步频率
	async updateFrequency(frequency) {
		if (!this.settings) {
			throw new Error('同步设置未加载');
		}

		const updatedSettings = {
			...this.settings,
			frequency
		};

		const result = await updateSyncSettings(updatedSettings);
		this.settings = result.result;
		return this.settings;
	}

	// 更新时间范围
	async updateTimeRange(timeRange) {
		if (!this.settings) {
			throw new Error('同步设置未加载');
		}

		const updatedSettings = {
			...this.settings,
			timeRange
		};

		const result = await updateSyncSettings(updatedSettings);
		this.settings = result.result;
		return this.settings;
	}

	// 切换自动重试
	async toggleAutoRetry(autoRetryEnabled) {
		if (!this.settings) {
			throw new Error('同步设置未加载');
		}

		const updatedSettings = {
			...this.settings,
			autoRetryEnabled
		};

		const result = await updateSyncSettings(updatedSettings);
		this.settings = result.result;
		return this.settings;
	}

	// 执行手动同步
	async performManualSync() {
		const result = await manualSync();
		// 刷新状态
		await this.refreshStatus();
		return result;
	}

	// 刷新同步状态
	async refreshStatus() {
		try {
			const result = await getSyncStatus();
			this.status = result.result;
			return this.status;
		} catch (error) {
			console.error('刷新同步状态失败:', error);
			throw error;
		}
	}

	// 刷新同步日志
	async refreshLogs(limit = 20) {
		try {
			const result = await getSyncLogs(limit);
			this.logs = result.result || [];
			return this.logs;
		} catch (error) {
			console.error('刷新同步日志失败:', error);
			throw error;
		}
	}

	// 获取同步状态文本
	getStatusText() {
		if (!this.settings || !this.status) {
			return '未知状态';
		}

		if (!this.settings.enabled) {
			return '未启用';
		}

		const statusMap = {
			'idle': '空闲',
			'syncing': '同步中...',
			'success': '同步成功',
			'failed': '同步失败'
		};

		return statusMap[this.status.syncStatus] || '未知状态';
	}

	// 获取频率文本
	getFrequencyText() {
		if (!this.settings) return '';

		const textMap = {
			'daily': '每天',
			'weekly': '每周',
			'every2days': '每2天',
			'every3days': '每3天'
		};

		return textMap[this.settings.frequency] || '每天';
	}

	// 格式化时间
	formatTime(timeStr) {
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
}

// 创建全局同步管理器实例
export const syncManager = new SyncManager();
