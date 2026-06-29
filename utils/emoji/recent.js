let recentEmojis = [];

export function getRecentEmojis() {
	return recentEmojis;
}

export function addToRecent(emojiChar) {
	const index = recentEmojis.indexOf(emojiChar);
	if (index > -1) {
		recentEmojis.splice(index, 1);
	}
	recentEmojis.unshift(emojiChar);
	if (recentEmojis.length > 20) {
		recentEmojis.pop();
	}
}

export function syncRecentToList(EMOJI_LIST) {
	EMOJI_LIST.recent = [...recentEmojis];
}
