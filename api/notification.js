// 通知设置相关API接口

// 导入请求工具类
import { request as utilRequest } from '../utils/request.js';

// 通用请求方法（使用统一的请求工具类）
function request(options) {
	// 使用统一的请求工具类，它会自动处理认证、错误处理等
	return utilRequest({
		...options,
		// 确保URL是相对路径，让request.js处理BASE_URL
		url: options.url
	});
}

// ==================== 通知设置相关API ====================

// 获取用户通知设置
export function getNotificationSettings() {
	return request({
		url: '/api/user/settings',
		method: 'GET'
	});
}

// 更新用户通知设置
export function updateNotificationSettings(settings) {
	return request({
		url: '/api/user/settings',
		method: 'POST',
		data: settings
	});
}

// 获取可用学期列表
export function getAvailableSemesters() {
	return request({
		url: '/api/notification-settings/semesters',
		method: 'GET'
	});
}

// 获取当前学期信息
export function getCurrentSemester() {
	return request({
		url: '/api/notification-settings/current-semester',
		method: 'GET'
	});
}

// 测试成绩检查功能
export function testScoreCheck() {
	return request({
		url: '/api/notification-settings/test-score-check',
		method: 'POST'
	});
}

// 获取通知日志
export function getNotificationLogs(params = {}) {
	const queryString = new URLSearchParams(params).toString();
	const url = queryString ? `/api/notification-settings/logs?${queryString}` : '/api/notification-settings/logs';

	return request({
		url,
		method: 'GET'
	});
}

// 获取成绩检查日志
export function getScoreCheckLogs(params = {}) {
	const queryString = new URLSearchParams(params).toString();
	const url = queryString ? `/api/notification-settings/score-check-logs?${queryString}` : '/api/notification-settings/score-check-logs';

	return request({
		url,
		method: 'GET'
	});
}

// ==================== 多渠道通知相关API ====================

// 获取通知渠道配置
export function getNotificationChannels() {
	return request({
		url: '/api/notification-channels',
		method: 'GET'
	});
}

// 更新通知渠道配置
export function updateNotificationChannels(channels) {
	return request({
		url: '/api/notification-channels',
		method: 'PUT',
		data: channels
	});
}

// 测试通知渠道
export function testNotificationChannel(type) {
	return request({
		url: `/api/notification-channels/test/${type}`,
		method: 'POST'
	});
}

// ==================== 短信相关API ====================

// 获取短信余额
export function getSMSBalance() {
	return request({
		url: '/api/sms/balance',
		method: 'GET'
	});
}

// 获取短信交易记录
export function getSMSTransactions(params = {}) {
	const queryString = new URLSearchParams(params).toString();
	const url = queryString ? `/api/sms/transactions?${queryString}` : '/api/sms/transactions';
	
	return request({
		url,
		method: 'GET'
	});
}

// 短信充值
export function rechargeSMS(amount) {
	return request({
		url: '/api/sms/recharge',
		method: 'POST',
		data: { amount }
	});
}

// 测试短信发送
export function testSMS() {
	return request({
		url: '/api/notification-channels/test/sms',
		method: 'POST'
	});
}

// ==================== 通知管理类 ====================

export class NotificationManager {
	constructor() {
		this.settings = null;
		this.channels = null;
		this.smsBalance = null;
		this.currentSemester = null;
		this.availableSemesters = [];
	}

	// 初始化，加载所有通知相关数据
	async init() {
		try {
			const [settingsRes, channelsRes, balanceRes] = await Promise.allSettled([
				getNotificationSettings(),
				getNotificationChannels(),
				getSMSBalance()
			]);

			if (settingsRes.status === 'fulfilled') {
				this.settings = settingsRes.value.result;
			}

			if (channelsRes.status === 'fulfilled') {
				this.channels = channelsRes.value.result;
			}

			if (balanceRes.status === 'fulfilled') {
				this.smsBalance = balanceRes.value.result;
			}

			return {
				settings: this.settings,
				channels: this.channels,
				smsBalance: this.smsBalance
			};
		} catch (error) {
			console.error('初始化通知管理器失败:', error);
			throw error;
		}
	}

	// 加载学期信息
	async loadSemesterInfo() {
		try {
			const [currentRes, availableRes] = await Promise.allSettled([
				getCurrentSemester(),
				getAvailableSemesters()
			]);

			if (currentRes.status === 'fulfilled') {
				this.currentSemester = currentRes.value.result;
			}

			if (availableRes.status === 'fulfilled') {
				this.availableSemesters = availableRes.value.result || [];
			}

			return {
				currentSemester: this.currentSemester,
				availableSemesters: this.availableSemesters
			};
		} catch (error) {
			console.error('加载学期信息失败:', error);
			throw error;
		}
	}

	// 更新通知设置
	async updateSettings(newSettings) {
		const result = await updateNotificationSettings(newSettings);
		this.settings = result.result;
		return this.settings;
	}

	// 更新通知渠道
	async updateChannels(newChannels) {
		const result = await updateNotificationChannels(newChannels);
		this.channels = result.result;
		return this.channels;
	}

	// 刷新短信余额
	async refreshSMSBalance() {
		try {
			const result = await getSMSBalance();
			this.smsBalance = result.result;
			return this.smsBalance;
		} catch (error) {
			console.error('刷新短信余额失败:', error);
			throw error;
		}
	}

	// 执行测试通知
	async performTestNotification() {
		const result = await testScoreCheck();
		return result;
	}

	// 测试特定渠道
	async testChannel(type) {
		const result = await testNotificationChannel(type);
		return result;
	}

	// 获取通知渠道状态文本
	getChannelStatusText(channel) {
		if (!this.channels) return '未知';

		const channelConfig = this.channels[`${channel}Enabled`];
		return channelConfig ? '已启用' : '未启用';
	}

	// 格式化余额显示
	formatBalance(balance) {
		if (typeof balance !== 'number') return '0.00';
		return (balance / 100).toFixed(2);
	}

	// 计算可发送短信数量
	calculateSMSCount(balance, cost = 10) {
		if (typeof balance !== 'number' || typeof cost !== 'number') return 0;
		return Math.floor(balance / cost);
	}
}

// 创建全局通知管理器实例
export const notificationManager = new NotificationManager();
