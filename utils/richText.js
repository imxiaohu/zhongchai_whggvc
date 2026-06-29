/**
 * 富文本处理工具 - 从 news/detail.vue 提取
 * 处理 HTML 内容渲染、图片提取、URL 转换
 */

/**
 * 从 HTML 富文本中提取所有图片 URL
 */
export function extractImagesFromHtml(htmlContent) {
	if (!htmlContent) return [];

	const imgRegex = /<img[^>]+src=["']([^"']+)["']/gi;
	const images = [];
	let match;

	while ((match = imgRegex.exec(htmlContent)) !== null) {
		const src = match[1];
		if (src && !src.startsWith('data:')) {
			images.push(src);
		}
	}

	return images;
}

/**
 * 处理富文本内容中的相对路径
 * @param {string} html - 原始 HTML 内容
 * @param {string} baseUrl - 基础 URL
 * @returns {string} 处理后的 HTML
 */
export function processRelativeUrls(html, baseUrl) {
	if (!html) return '';

	let processed = html
		.replace(/src="(\/[^"]+)"/g, (match, src) => {
			if (src.startsWith('//')) {
				return `src="https:${src}"`;
			}
			if (src.startsWith('/')) {
				const base = baseUrl.endsWith('/') ? baseUrl.slice(0, -1) : baseUrl;
				return `src="${base}${src}"`;
			}
			return match;
		})
		.replace(/href="(\/[^"]+)"/g, (match, href) => {
			if (href.startsWith('//')) {
				return `href="https:${href}"`;
			}
			if (href.startsWith('/')) {
				const base = baseUrl.endsWith('/') ? baseUrl.slice(0, -1) : baseUrl;
				return `href="${base}${href}"`;
			}
			return match;
		});

	return processed;
}

/**
 * HTML 安全过滤（移除危险标签和属性）
 */
export function sanitizeHtml(html) {
	if (!html) return '';

	let result = html
		.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
		.replace(/<style\b[^<]*(?:(?!<\/style>)<[^<]*)*<\/style>/gi, '')
		.replace(/<!--[\s\S]*?-->/g, '')

		.replace(/on\w+="[^"]*"/gi, '')
		.replace(/on\w+='[^']*'/gi, '')
		.replace(/on\w+=\S+/gi, '')

		.replace(/\bjavascript\s*:/gi, '')
		.replace(/\bdata\s*:/gi, '')
		.replace(/\bvbscript\s*:/gi, '')
		.replace(/\bexpression\s*\(/gi, '')

		.replace(/<iframe\b[^<]*(?:(?!<\/iframe>)<[^<]*)*<\/iframe>/gi, '')
		.replace(/<object\b[^<]*(?:(?!<\/object>)<[^<]*)*<\/object>/gi, '')
		.replace(/<embed\b[^>]*>/gi, '')
		.replace(/<form\b[^<]*(?:(?!<\/form>)<[^<]*)*<\/form>/gi, '')
		.replace(/<svg\b[^<]*(?:(?!<\/svg>)<[^<]*)*<\/svg>/gi, '')
		.replace(/<math\b[^<]*(?:(?!<\/math>)<[^<]*)*<\/math>/gi, '')
		.replace(/<base\b[^>]*>/gi, '')
		.replace(/<link\b[^>]*>/gi, '')
		.replace(/<meta\b[^>]*>/gi, '')

		.replace(/style\s*=\s*"[^"]*expression\s*\([^"]*"/gi, '')
		.replace(/style\s*=\s*'[^']*expression\s*\([^']*'/gi, '')

		.replace(/href\s*=\s*"[^"]*javascript\s*:/gi, 'href="javascript:void(0)"')
		.replace(/href\s*=\s*'[^']*javascript\s*:/gi, "href='javascript:void(0)'")
		.replace(/src\s*=\s*"[^"]*javascript\s*:/gi, 'src="javascript:void(0)"')
		.replace(/src\s*=\s*'[^']*javascript\s*:/gi, "src='javascript:void(0)'");

	return result;
}

/**
 * 截断文本并添加省略号
 */
export function truncateText(text, maxLength = 100) {
	if (!text || text.length <= maxLength) return text;
	return text.slice(0, maxLength) + '...';
}

/**
 * 格式化附件大小显示
 */
export function formatFileSize(bytes) {
	if (!bytes || bytes === 0) return '0 B';
	const k = 1024;
	const sizes = ['B', 'KB', 'MB', 'GB'];
	const i = Math.floor(Math.log(bytes) / Math.log(k));
	return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

/**
 * 获取文件类型图标类名
 */
export function getFileTypeIcon(filename) {
	if (!filename) return 'icon-file-text';

	const ext = filename.split('.').pop().toLowerCase();

	const iconMap = {
		pdf: 'icon-file-pdf',
		doc: 'icon-file-word',
		docx: 'icon-file-word',
		xls: 'icon-file-excel',
		xlsx: 'icon-file-excel',
		ppt: 'icon-file-powerpoint',
		pptx: 'icon-file-powerpoint',
		zip: 'icon-file-archive',
		rar: 'icon-file-archive',
		txt: 'icon-file-text',
		jpg: 'icon-image',
		jpeg: 'icon-image',
		png: 'icon-image',
		gif: 'icon-image',
		mp3: 'icon-music',
		mp4: 'icon-video',
		avi: 'icon-video'
	};

	return iconMap[ext] || 'icon-file-text';
}

/**
 * 获取文件类型颜色
 */
export function getFileTypeColor(filename) {
	if (!filename) return '#999';

	const ext = filename.split('.').pop().toLowerCase();

	const colorMap = {
		pdf: '#E74C3C',
		doc: '#3498DB',
		docx: '#3498DB',
		xls: '#27AE60',
		xlsx: '#27AE60',
		ppt: '#E67E22',
		pptx: '#E67E22',
		zip: '#9B59B6',
		rar: '#9B59B6'
	};

	return colorMap[ext] || '#999';
}

/**
 * 判断是否为图片文件
 */
export function isImageFile(filename) {
	if (!filename) return false;
	const ext = filename.split('.').pop().toLowerCase();
	return ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg'].includes(ext);
}

/**
 * 判断是否为视频文件
 */
export function isVideoFile(filename) {
	if (!filename) return false;
	const ext = filename.split('.').pop().toLowerCase();
	return ['mp4', 'avi', 'mov', 'wmv', 'flv', 'mkv'].includes(ext);
}

/**
 * 从 URL 中提取文件名
 */
export function getFileNameFromUrl(url) {
	if (!url) return '';
	try {
		const urlObj = new URL(url);
		const pathname = urlObj.pathname;
		return decodeURIComponent(pathname.split('/').pop());
	} catch {
		return url.split('/').pop();
	}
}

/**
 * 清理 HTML 标签，获取纯文本
 */
export function stripHtmlTags(html) {
	if (!html) return '';
	return html
		.replace(/<[^>]*>/g, '')
		.replace(/&nbsp;/g, ' ')
		.replace(/&amp;/g, '&')
		.replace(/&lt;/g, '<')
		.replace(/&gt;/g, '>')
		.replace(/&quot;/g, '"')
		.trim();
}
