/**
 * 时间格式化 Composable
 * 仅提供中文时间格式化功能
 */

export function useTimeFormat() {
	function formatTime(time) {
		if (!time) return '';

		try {
			const date = new Date(time);
			const now = new Date();
			const diff = now - date;

			if (diff < 60000) {
				return '刚刚';
			}

			if (diff < 3600000) {
				const minutes = Math.floor(diff / 60000);
				return `${minutes}分钟前`;
			}

			if (diff < 86400000) {
				const hours = Math.floor(diff / 3600000);
				return `${hours}小时前`;
			}

			if (diff < 604800000) {
				const days = Math.floor(diff / 86400000);
				return `${days}天前`;
			}

			return formatDate(date);
		} catch (error) {
			console.error('格式化时间失败:', error);
			return new Date(time).toLocaleDateString();
		}
	}

	function formatDate(date) {
		try {
			const year = date.getFullYear();
			const month = String(date.getMonth() + 1).padStart(2, '0');
			const day = String(date.getDate()).padStart(2, '0');
			return `${year}-${month}-${day}`;
		} catch (error) {
			console.error('格式化日期失败:', error);
			return date.toLocaleDateString();
		}
	}

	function formatNumber(num) {
		if (!num || num === 0) return '0';

		if (num < 1000) {
			return num.toString();
		}

		if (num < 10000) {
			return (num / 1000).toFixed(1) + 'k';
		}

		return (num / 10000).toFixed(1) + 'w';
	}

	function formatCount(count, key) {
		const countValue = count || 0;

		if (key && key.includes('articleCount')) {
			return `${countValue} 篇文章`;
		} else if (key && key.includes('memberCount')) {
			return `${countValue} 名成员`;
		} else if (key && key.includes('viewCount')) {
			return `${countValue} 次浏览`;
		} else if (key && key.includes('likeCount')) {
			return `${countValue} 个赞`;
		} else if (key && key.includes('commentCount')) {
			return `${countValue} 条评论`;
		} else if (key && key.includes('readerCount')) {
			return `${countValue} 人阅读`;
		}

		return `${countValue}`;
	}

	return {
		formatTime,
		formatDate,
		formatNumber,
		formatCount
	};
}
