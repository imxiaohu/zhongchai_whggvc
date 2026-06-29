import { EMOJI_CATEGORIES } from './categories.js';
import { smileysData } from './data-smileys.js';
import { gesturesData, heartsData } from './data-gestures.js';
import { natureData } from './data-nature.js';
import { foodData } from './data-food.js';
import { activitiesData } from './data-activities.js';
import { travelData } from './data-travel.js';
import { objectsData } from './data-objects.js';
import { symbolsData } from './data-symbols.js';
import { getRecentEmojis as _getRecentEmojis, addToRecent as _addToRecent, syncRecentToList } from './recent.js';
import { searchEmojis as _searchEmojis } from './search.js';

export { EMOJI_CATEGORIES } from './categories.js';

export const EMOJI_LIST = {
	recent: [],
	smileys: smileysData,
	gestures: gesturesData,
	hearts: heartsData,
	nature: natureData,
	food: foodData,
	activities: activitiesData,
	travel: travelData,
	objects: objectsData,
	symbols: symbolsData
};

syncRecentToList(EMOJI_LIST);

export const getRecentEmojis = _getRecentEmojis;

export function addToRecent(emojiChar) {
	_addToRecent(emojiChar);
	syncRecentToList(EMOJI_LIST);
}

export function searchEmojis(query, maxResults = 20) {
	return _searchEmojis(EMOJI_LIST, query, maxResults);
}

export function getEmojisByCategory(categoryId) {
	return EMOJI_LIST[categoryId] || [];
}

export function getAllCategoryEmojis() {
	return EMOJI_CATEGORIES.map(cat => ({
		...cat,
		emojis: EMOJI_LIST[cat.id] || []
	}));
}

export function countEmojisInText(text) {
	if (!text) return 0;
	const emojiRegex = /[\u{1F300}-\u{1F9FF}]|[\u{2600}-\u{26FF}]|[\u{2700}-\u{27BF}]|[\u{1F600}-\u{1F64F}]|[\u{1F680}-\u{1F6FF}]|[\u{1F1E0}-\u{1F1FF}]/gu;
	const matches = text.match(emojiRegex);
	return matches ? matches.length : 0;
}

export function isEmoji(char) {
	const emojiRegex = /^[\u{1F300}-\u{1F9FF}]|[\u{2600}-\u{26FF}]|[\u{2700}-\u{27BF}]|[\u{1F600}-\u{1F64F}]|[\u{1F680}-\u{1F6FF}]|[\u{1F1E0}-\u{1F1FF}]$/u;
	return emojiRegex.test(char);
}
