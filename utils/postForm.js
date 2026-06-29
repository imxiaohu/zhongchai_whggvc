/**
 * 帖子表单工具类
 * 提供表单验证和载荷构建功能
 */

/**
 * 验证帖子表单
 * @param {Object} form 表单数据 { title, content, images }
 * @returns {Object} { valid: boolean, errors: string[] }
 */
export function validatePostForm(form) {
	const errors = [];

	if (!form) {
		return {
			valid: false,
			errors: ['表单数据不能为空']
		};
	}

	// 验证标题
	if (!form.title || typeof form.title !== 'string') {
		errors.push('请输入标题');
	} else if (form.title.trim().length === 0) {
		errors.push('标题不能为空');
	} else if (form.title.trim().length > 100) {
		errors.push('标题不能超过100个字符');
	}

	// 验证内容
	if (!form.content || typeof form.content !== 'string') {
		errors.push('请输入内容');
	} else if (form.content.trim().length === 0) {
		errors.push('内容不能为空');
	} else if (form.content.trim().length > 5000) {
		errors.push('内容不能超过5000个字符');
	}

	// 验证图片（可选）
	if (form.images && Array.isArray(form.images)) {
		if (form.images.length > 20) {
			errors.push('最多只能上传20张图片');
		}

		// 验证每个图片URL
		for (let i = 0; i < form.images.length; i++) {
			const img = form.images[i];
			if (!img || typeof img !== 'string' || img.trim().length === 0) {
				errors.push(`第${i + 1}张图片URL无效`);
			}
		}
	}

	// 验证帖子类型
	const validTypes = ['article', 'announcement', 'activity'];
	if (form.type && !validTypes.includes(form.type)) {
		errors.push('无效的帖子类型');
	}

	return {
		valid: errors.length === 0,
		errors
	};
}

/**
 * 构建帖子提交载荷
 * @param {Object} form 表单数据
 * @param {Array} images 图片URL数组
 * @param {string} postType 帖子类型
 * @param {number|string} clubId 社团ID
 * @returns {Object} API提交数据
 */
export function buildPostPayload(form, images = [], postType = 'article', clubId = null) {
	const payload = {
		title: form.title.trim(),
		content: form.content.trim(),
		summary: form.content.trim().substring(0, 200),
		images: JSON.stringify(images || []),
		type: postType || form.type || 'article'
	};

	// 处理社团ID
	if (clubId) {
		payload.clubId = clubId;
	}

	// 处理官方公告标记
	if (!clubId && form.isOfficial) {
		payload.isOfficial = true;
	}

	return payload;
}

/**
 * 验证社团权限
 * @param {string|number} userRole 用户角色
 * @param {string} action 操作类型 ('create', 'edit', 'delete', 'announce')
 * @returns {Object} { allowed: boolean, message: string }
 */
export function validateClubPermission(userRole, action) {
	const adminActions = ['create', 'edit', 'delete', 'announce', 'manage'];
	const memberActions = ['create', 'edit_own', 'delete_own'];

	// 管理员可以执行所有操作
	const adminRoles = ['admin', 'owner', 'super_admin'];
	if (adminRoles.includes(userRole)) {
		return {
			allowed: true,
			message: '权限验证通过'
		};
	}

	// 检查具体操作权限
	if (adminActions.includes(action)) {
		return {
			allowed: false,
			message: '需要管理员权限才能执行此操作'
		};
	}

	if (memberActions.includes(action)) {
		const memberRoles = ['member', 'editor'];
		if (memberRoles.includes(userRole)) {
			return {
				allowed: true,
				message: '权限验证通过'
			};
		}
	}

	return {
		allowed: false,
		message: '您没有执行此操作的权限'
	};
}

/**
 * 格式化帖子数据用于显示
 * @param {Object} post 原始帖子数据
 * @returns {Object} 格式化后的数据
 */
export function formatPostForDisplay(post) {
	if (!post) return null;

	return {
		id: post.id,
		title: post.title,
		content: post.content,
		summary: post.summary || post.content?.substring(0, 200) || '',
		images: parsePostImages(post.images),
		type: post.type,
		clubId: post.clubId,
		authorId: post.authorId,
		authorName: post.authorName,
		authorAvatar: post.authorAvatar,
		createdAt: formatDate(post.createdAt),
		updatedAt: formatDate(post.updatedAt),
		likes: post.likes || 0,
		comments: post.comments || 0,
		views: post.views || 0,
		isLiked: post.isLiked || false,
		isBookmarked: post.isBookmarked || false
	};
}

/**
 * 解析帖子图片
 * @param {string|Array} images 图片数据
 * @returns {Array} 图片URL数组
 */
export function parsePostImages(images) {
	if (!images) return [];

	if (Array.isArray(images)) {
		return images;
	}

	if (typeof images === 'string') {
		try {
			return JSON.parse(images);
		} catch (e) {
			// 如果不是JSON字符串，可能是一组用逗号分隔的URL
			return images.split(',').filter(url => url.trim());
		}
	}

	return [];
}

/**
 * 格式化日期
 * @param {string|Date} date 日期
 * @returns {string} 格式化后的日期字符串
 */
export function formatDate(date) {
	if (!date) return '';

	const d = new Date(date);
	const now = new Date();
	const diff = now - d;

	// 1分钟内
	if (diff < 60000) {
		return '刚刚';
	}

	// 1小时内
	if (diff < 3600000) {
		return Math.floor(diff / 60000) + '分钟前';
	}

	// 24小时内
	if (diff < 86400000) {
		return Math.floor(diff / 3600000) + '小时前';
	}

	// 7天内
	if (diff < 604800000) {
		return Math.floor(diff / 86400000) + '天前';
	}

	// 超过7天显示日期
	const year = d.getFullYear();
	const month = String(d.getMonth() + 1).padStart(2, '0');
	const day = String(d.getDate()).padStart(2, '0');

	if (year === now.getFullYear()) {
		return `${month}-${day}`;
	}

	return `${year}-${month}-${day}`;
}

/**
 * 获取帖子类型显示名称
 * @param {string} type 帖子类型
 * @returns {string} 显示名称
 */
export function getPostTypeName(type) {
	const typeNames = {
		article: 'Article',
		announcement: 'Announcement',
		activity: 'Activity'
	};

	return typeNames[type] || 'Article';
}

/**
 * 获取帖子类型对应的图标
 * @param {string} type 帖子类型
 * @returns {string} 图标名称
 */
export function getPostTypeIcon(type) {
	const typeIcons = {
		article: 'file-text',
		announcement: 'megaphone',
		activity: 'calendar'
	};

	return typeIcons[type] || 'file-text';
}

/**
 * 清理表单数据
 * @param {Object} form 原始表单数据
 * @returns {Object} 清理后的表单数据
 */
export function cleanFormData(form) {
	return {
		title: (form.title || '').trim(),
		content: (form.content || '').trim(),
		images: form.images || [],
		type: form.type || 'article'
	};
}

/**
 * 检测表单是否有变更
 * @param {Object} original 原始数据
 * @param {Object} current 当前数据
 * @returns {boolean} 是否有变更
 */
export function hasFormChanged(original, current) {
	if (!original || !current) return false;

	return (
		original.title !== current.title ||
		original.content !== current.content ||
		JSON.stringify(original.images) !== JSON.stringify(current.images) ||
		original.type !== current.type
	);
}

export default {
	validatePostForm,
	buildPostPayload,
	validateClubPermission,
	formatPostForDisplay,
	parsePostImages,
	formatDate,
	getPostTypeName,
	getPostTypeIcon,
	cleanFormData,
	hasFormChanged
};
