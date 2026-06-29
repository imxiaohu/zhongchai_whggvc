// emoji 工具 - 重新导出到 utils/emoji/index.js
// 保留向后兼容，所有调用方无需修改
export {
	EMOJI_CATEGORIES,
	EMOJI_LIST,
	getRecentEmojis,
	addToRecent,
	searchEmojis,
	getEmojisByCategory,
	getAllCategoryEmojis,
	countEmojisInText,
	isEmoji
} from './emoji/index.js';
