export function searchEmojis(EMOJI_LIST, query, maxResults = 20) {
	if (!query) return [];

	const lowerQuery = query.toLowerCase();
	const results = [];
	const seen = new Set();

	for (const [categoryId, emojis] of Object.entries(EMOJI_LIST)) {
		if (categoryId === 'recent') continue;

		for (const emoji of emojis) {
			const key = emoji.char;
			if (seen.has(key)) continue;

			if (emoji.name.toLowerCase().includes(lowerQuery)) {
				results.push(emoji);
				seen.add(key);
				continue;
			}

			for (const keyword of emoji.keywords) {
				if (keyword.toLowerCase().includes(lowerQuery)) {
					results.push(emoji);
					seen.add(key);
					break;
				}
			}

			if (results.length >= maxResults) break;
		}

		if (results.length >= maxResults) break;
	}

	return results;
}
