/**
 * 新闻/通知工具函数
 * 包含新闻类型定义、格式化、过滤等功能
 */

/**
 * 新闻类型定义
 */
export const NEWS_TYPES = {
	NOTICE: 'notice',       // 通知公告
	NEWS: 'news',           // 新闻动态
	EVENT: 'event',         // 活动通知
	SYSTEM: 'system'        // 系统消息
};

/**
 * 新闻类型颜色映射
 */
export const NEWS_TYPE_COLORS = {
	[NEWS_TYPES.NOTICE]: '#ef4444',  // 红色 - 通知
	[NEWS_TYPES.NEWS]: '#3b82f6',     // 蓝色 - 新闻
	[NEWS_TYPES.EVENT]: '#10b981',   // 绿色 - 活动
	[NEWS_TYPES.SYSTEM]: '#8b5cf6'   // 紫色 - 系统
};

/**
 * 新闻类型文字映射
 */
export const NEWS_TYPE_LABELS = {
	[NEWS_TYPES.NOTICE]: '通知',
	[NEWS_TYPES.NEWS]: '新闻',
	[NEWS_TYPES.EVENT]: '活动',
	[NEWS_TYPES.SYSTEM]: '系统'
};

/**
 * 获取默认来源列表
 */
export const DEFAULT_SOURCES = ['系统', '管理员', '教务处'];

/**
 * 格式化新闻/通知项
 * @param {Object} notice - 原始通知数据
 * @returns {Object} 格式化后的通知数据
 */
export function formatNewsItem(notice) {
	return {
		id: notice.id,
		title: notice.title || notice.name || '无标题',
		publishTime: notice.publishTime || notice.createtime || notice.pubdate,
		source: notice.source || notice.departname || notice.author || '系统',
		isRead: notice.isRead ?? (Math.random() > 0.5),
		isTop: notice.isTop || false,
		type: notice.type || NEWS_TYPES.NOTICE,
		summary: notice.summary || notice.desc || '',
		content: notice.content || notice.body || '',
		images: notice.images || notice.imgs || [],
		views: notice.views || notice.viewCount || 0,
		likes: notice.likes || 0
	};
}

/**
 * 批量格式化新闻列表
 * @param {Array} notices - 原始通知列表
 * @returns {Array} 格式化后的通知列表
 */
export function formatNewsList(notices) {
	if (!Array.isArray(notices)) {
		return [];
	}
	return notices.map(formatNewsItem);
}

/**
 * 计算未读新闻数量
 * @param {Array} notices - 通知列表
 * @returns {number} 未读数量
 */
export function getUnreadCount(notices) {
	if (!Array.isArray(notices)) {
		return 0;
	}
	return notices.filter(notice => !notice.isRead).length;
}

/**
 * 计算未读徽章显示文本
 * @param {number} count - 未读数量
 * @returns {string} 徽章文本
 */
export function getUnreadBadgeText(count) {
	if (count <= 0) return '';
	if (count > 99) return '99+';
	return String(count);
}

/**
 * 按类型过滤新闻
 * @param {Array} notices - 通知列表
 * @param {string} type - 新闻类型
 * @returns {Array} 过滤后的列表
 */
export function filterNewsByType(notices, type) {
	if (!Array.isArray(notices)) {
		return [];
	}
	if (!type || type === 'all') {
		return notices;
	}
	return notices.filter(notice => notice.type === type);
}

/**
 * 获取置顶新闻
 * @param {Array} notices - 通知列表
 * @returns {Array} 置顶新闻列表
 */
export function getTopNews(notices) {
	if (!Array.isArray(notices)) {
		return [];
	}
	return notices.filter(notice => notice.isTop);
}

/**
 * 获取普通新闻（非置顶）
 * @param {Array} notices - 通知列表
 * @returns {Array} 非置顶新闻列表
 */
export function getNormalNews(notices) {
	if (!Array.isArray(notices)) {
		return [];
	}
	return notices.filter(notice => !notice.isTop);
}

/**
 * 按时间排序（最新的在前）
 * @param {Array} notices - 通知列表
 * @returns {Array} 排序后的列表
 */
export function sortNewsByTime(notices) {
	if (!Array.isArray(notices)) {
		return [];
	}
	return [...notices].sort((a, b) => {
		const timeA = new Date(a.publishTime || 0).getTime();
		const timeB = new Date(b.publishTime || 0).getTime();
		return timeB - timeA;
	});
}

/**
 * 创建预览通知数据
 * @returns {Array} 预览通知列表
 */
export function createPreviewNotices() {
	return [
		{
			id: 'preview-notice-1',
			title: '欢迎使用评教系统',
			publishTime: new Date().toISOString(),
			source: '系统',
			isRead: false,
			isTop: true,
			isPreview: true
		}
	];
}

/**
 * 获取新闻类型对应的颜色
 * @param {string} type - 新闻类型
 * @returns {string} 颜色值
 */
export function getNewsTypeColor(type) {
	return NEWS_TYPE_COLORS[type] || NEWS_TYPE_COLORS[NEWS_TYPES.NOTICE];
}

/**
 * 获取新闻类型对应的文字标签
 * @param {string} type - 新闻类型
 * @returns {string} 类型文字
 */
export function getNewsTypeLabel(type) {
	return NEWS_TYPE_LABELS[type] || NEWS_TYPES.NOTICE;
}
