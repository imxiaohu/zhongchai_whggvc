/**
 * 用户 @提及 工具函数
 * 提供 @提及 的提取、格式化、搜索等功能
 */

// @提及的正则表达式
export const MENTION_REGEX = /@(\w+)/g;

/**
 * 从文本中提取所有 @提及
 * @param {string} text - 输入文本
 * @returns {Array<{username: string, start: number, end: number}>} 提及列表
 */
export function extractMentions(text) {
	if (!text) return [];
	
	const mentions = [];
	let match;
	const regex = new RegExp(MENTION_REGEX.source, 'g');
	
	while ((match = regex.exec(text)) !== null) {
		mentions.push({
			username: match[1],
			start: match.index,
			end: match.index + match[0].length,
			fullMatch: match[0]
		});
	}
	
	return mentions;
}

/**
 * 从文本中提取 @后的查询词（光标位置附近）
 * @param {string} text - 输入文本
 * @param {number} cursorPosition - 光标位置
 * @returns {string|null} 查询词或null
 */
export function extractMentionQuery(text, cursorPosition) {
	if (!text || cursorPosition === undefined) return null;
	
	// 从光标位置向前查找@
	const beforeCursor = text.substring(0, cursorPosition);
	const lastAtIndex = beforeCursor.lastIndexOf('@');
	
	if (lastAtIndex === -1) return null;
	
	// 检查@和光标之间是否有空格
	const textBetween = beforeCursor.substring(lastAtIndex);
	if (/\s/.test(textBetween)) return null;
	
	// 获取@后面的查询词
	const query = beforeCursor.substring(lastAtIndex + 1);
	
	// 查询词只能是字母、数字、下划线
	if (!/^\w*$/.test(query)) return null;
	
	return query;
}

/**
 * 格式化提及用于显示（高亮）
 * @param {string} text - 输入文本
 * @returns {Array<{type: 'text'|'mention', content: string}>} 格式化后的片段
 */
export function formatMentionsForDisplay(text) {
	if (!text) return [];
	
	const segments = [];
	let lastIndex = 0;
	let match;
	const regex = new RegExp(MENTION_REGEX.source, 'g');
	
	while ((match = regex.exec(text)) !== null) {
		// 添加 @ 之前的文本
		if (match.index > lastIndex) {
			segments.push({
				type: 'text',
				content: text.substring(lastIndex, match.index)
			});
		}
		
		// 添加提及
		segments.push({
			type: 'mention',
			content: match[0],
			username: match[1]
		});
		
		lastIndex = match.index + match[0].length;
	}
	
	// 添加剩余文本
	if (lastIndex < text.length) {
		segments.push({
			type: 'text',
			content: text.substring(lastIndex)
		});
	}
	
	return segments;
}

/**
 * 获取用户显示名称
 * @param {Object} user - 用户对象
 * @returns {string} 显示名称
 */
export function getUserDisplayName(user) {
	if (!user) return '匿名用户';
	return user.realname || user.nickname || user.username || '匿名用户';
}

/**
 * 生成 @提及 文本
 * @param {Object} user - 用户对象
 * @returns {string} @提及文本
 */
export function createMentionText(user) {
	return `@${getUserDisplayName(user)} `;
}

/**
 * 在文本中替换指定位置的 @提及
 * @param {string} text - 原文本
 * @param {number} start - 起始位置
 * @param {number} end - 结束位置
 * @param {Object} user - 用户对象
 * @returns {string} 替换后的文本
 */
export function replaceMention(text, start, end, user) {
	const before = text.substring(0, start);
	const after = text.substring(end);
	return before + createMentionText(user) + after;
}

/**
 * 检测文本中是否包含 @提及
 * @param {string} text - 输入文本
 * @returns {boolean} 是否包含提及
 */
export function hasMentions(text) {
	if (!text) return false;
	return MENTION_REGEX.test(text);
}

/**
 * 获取 @提及 的数量
 * @param {string} text - 输入文本
 * @returns {number} 提及数量
 */
export function countMentions(text) {
	if (!text) return 0;
	const mentions = extractMentions(text);
	return mentions.length;
}

/**
 * 过滤用户列表用于 @提及 自动完成
 * @param {Array} users - 用户列表
 * @param {string} query - 搜索查询
 * @param {number} limit - 返回数量限制
 * @returns {Array} 匹配的用户列表
 */
export function filterUsersForMention(users, query, limit = 5) {
	if (!users || !Array.isArray(users)) return [];
	if (!query) return users.slice(0, limit);
	
	const lowerQuery = query.toLowerCase();
	
	return users.filter(user => {
		const username = (user.username || '').toLowerCase();
		const nickname = (user.nickname || '').toLowerCase();
		const realname = (user.realname || '').toLowerCase();
		
		return username.includes(lowerQuery) ||
			nickname.includes(lowerQuery) ||
			realname.includes(lowerQuery);
	}).slice(0, limit);
}

/**
 * 获取 mock 用户列表（用于测试）
 * @returns {Array} mock 用户列表
 */
export function getMockMentionUsers() {
	return [
		{ id: 1, username: 'zhangsan', nickname: '张三', realname: '张三', avatar: '/static/logo.png' },
		{ id: 2, username: 'lisi', nickname: '李四', realname: '李四', avatar: '/static/logo.png' },
		{ id: 3, username: 'wangwu', nickname: '王五', realname: '王五', avatar: '/static/logo.png' },
		{ id: 4, username: 'zhaoliu', nickname: '赵六', realname: '赵六', avatar: '/static/logo.png' },
		{ id: 5, username: 'admin', nickname: '管理员', realname: '系统管理员', avatar: '/static/logo.png' }
	];
}

/**
 * 搜索用户用于 @提及 自动完成
 * @param {string} query - 搜索查询
 * @returns {Promise<Array>} 匹配的用户列表
 */
export async function searchMentionUsers(query) {
	// 模拟API延迟
	await new Promise(resolve => setTimeout(resolve, 200));
	
	const mockUsers = getMockMentionUsers();
	return filterUsersForMention(mockUsers, query);
}

/**
 * 从评论对象中获取提及的用户ID列表
 * @param {Object} comment - 评论对象
 * @returns {Array<number>} 用户ID列表
 */
export function getMentionedUserIds(comment) {
	if (!comment || !comment.mentionedUsers) return [];
	return comment.mentionedUsers.map(u => u.id);
}

/**
 * 清理文本中的无效提及（用户不存在的）
 * @param {string} text - 输入文本
 * @param {Array} validUsers - 有效用户列表
 * @returns {string} 清理后的文本
 */
export function cleanInvalidMentions(text, validUsers) {
	if (!text) return '';
	
	const validUsernames = new Set(validUsers.map(u => u.username.toLowerCase()));
	const mentions = extractMentions(text);
	
	let cleanedText = text;
	// 从后向前替换，避免位置偏移问题
	const sortedMentions = [...mentions].sort((a, b) => b.start - a.start);
	
	for (const mention of sortedMentions) {
		if (!validUsernames.has(mention.username.toLowerCase())) {
			cleanedText = cleanedText.substring(0, mention.start) + 
				cleanedText.substring(mention.end);
		}
	}
	
	return cleanedText;
}
