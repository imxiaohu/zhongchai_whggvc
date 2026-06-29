/**
 * 分享相关API
 */

import { request } from './request.js';

/**
 * 记录分享统计
 * @param {Object} data 分享数据
 * @param {string} data.type 分享类型 (club, post, etc.)
 * @param {string} data.targetId 目标ID
 * @param {string} data.source 分享来源 (friend, timeline, qq, weibo, etc.)
 * @param {string} data.platform 平台 (mp-weixin, h5, app-plus)
 */
export function trackShare(data) {
	return request({
		url: '/api/share/track',
		method: 'POST',
		data: {
			type: data.type,
			targetId: data.targetId,
			source: data.source,
			platform: data.platform,
			timestamp: new Date().toISOString(),
			userAgent: navigator.userAgent || '',
			// #ifdef MP-WEIXIN
			scene: data.scene || '',
			// #endif
		}
	});
}

/**
 * 获取分享统计
 * @param {Object} params 查询参数
 * @param {string} params.type 分享类型
 * @param {string} params.targetId 目标ID
 * @param {string} params.startDate 开始日期
 * @param {string} params.endDate 结束日期
 */
export function getShareStats(params) {
	return request({
		url: '/api/share/stats',
		method: 'GET',
		data: params
	});
}

/**
 * 获取热门分享内容
 * @param {Object} params 查询参数
 * @param {string} params.type 内容类型
 * @param {number} params.limit 限制数量
 * @param {string} params.timeRange 时间范围 (day, week, month)
 */
export function getHotShares(params) {
	return request({
		url: '/api/share/hot',
		method: 'GET',
		data: params
	});
}

/**
 * 生成分享海报
 * @param {Object} data 海报数据
 * @param {string} data.type 内容类型
 * @param {string} data.targetId 目标ID
 * @param {Object} data.config 海报配置
 */
export function generateSharePoster(data) {
	return request({
		url: '/api/share/poster',
		method: 'POST',
		data: data
	});
}
