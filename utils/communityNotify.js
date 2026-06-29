/**
 * Community Notification Helpers
 * Community notification type mappings, settings parsers, and channel options
 */

/**
 * Community notification type definitions
 */
export const COMMUNITY_NOTIFICATION_TYPES = [
	{
		key: 'like',
		name: '点赞通知',
		icon: 'heart-filled',
		iconClass: 'like-icon',
		channels: ['email', 'dingtalk']
	},
	{
		key: 'bookmark',
		name: '收藏通知',
		icon: 'star-filled',
		iconClass: 'bookmark-icon',
		channels: ['email', 'dingtalk']
	},
	{
		key: 'comment',
		name: '评论通知',
		icon: 'message-circle',
		iconClass: 'comment-icon',
		channels: ['email', 'dingtalk']
	},
	{
		key: 'commentLike',
		name: '评论点赞',
		icon: 'thumb-up',
		iconClass: 'comment-like-icon',
		channels: ['email', 'dingtalk']
	}
];

/**
 * Community notification preference mappings
 * Maps UI preference keys to internal setting keys and channel types
 */
export const COMMUNITY_PREFERENCE_MAP = {
	likeEmail: { type: 'like', channel: 'email' },
	likeDingTalk: { type: 'like', channel: 'dingtalk' },
	bookmarkEmail: { type: 'bookmark', channel: 'email' },
	bookmarkDingTalk: { type: 'bookmark', channel: 'dingtalk' },
	commentEmail: { type: 'comment', channel: 'email' },
	commentDingTalk: { type: 'comment', channel: 'dingtalk' },
	commentLikeEmail: { type: 'commentLike', channel: 'email' },
	commentLikeDingTalk: { type: 'commentLike', channel: 'dingtalk' }
};

/**
 * Default community notification settings
 */
export const DEFAULT_COMMUNITY_SETTINGS = {
	likeEmail: false,
	likeDingTalk: false,
	bookmarkEmail: false,
	bookmarkDingTalk: false,
	commentEmail: false,
	commentDingTalk: false,
	commentLikeEmail: false,
	commentLikeDingTalk: false
};

/**
 * Channel availability check for community notifications
 * @param {string} channel - Channel key ('email' or 'dingtalk')
 * @param {object} channels - Channel settings object
 * @returns {boolean} Whether the channel is enabled
 */
export function isCommunityChannelAvailable(channel, channels) {
	switch (channel) {
		case 'email': return channels.emailEnabled;
		case 'dingtalk': return channels.dingTalkEnabled;
		default: return false;
	}
}

/**
 * Get community notification channel options
 * @param {object} channels - Channel settings object
 * @returns {Array} Array of channel options with availability status
 */
export function getCommunityChannelOptions(channels) {
	return [
		{
			key: 'email',
			label: '邮件',
			available: channels.emailEnabled
		},
		{
			key: 'dingtalk',
			label: '钉钉',
			available: channels.dingTalkEnabled
		}
	];
}

/**
 * Build community settings from API response
 * @param {object} channelSettings - Channel settings from API
 * @returns {object} Parsed community settings
 */
export function parseCommunitySettings(channelSettings) {
	return {
		likeEmail: channelSettings.communityLikeEmail !== undefined ? channelSettings.communityLikeEmail : false,
		likeDingTalk: channelSettings.communityLikeDingTalk !== undefined ? channelSettings.communityLikeDingTalk : false,
		bookmarkEmail: channelSettings.communityBookmarkEmail !== undefined ? channelSettings.communityBookmarkEmail : false,
		bookmarkDingTalk: channelSettings.communityBookmarkDingTalk !== undefined ? channelSettings.communityBookmarkDingTalk : false,
		commentEmail: channelSettings.communityCommentEmail !== undefined ? channelSettings.communityCommentEmail : false,
		commentDingTalk: channelSettings.communityCommentDingTalk !== undefined ? channelSettings.communityCommentDingTalk : false,
		commentLikeEmail: channelSettings.communityCommentLikeEmail !== undefined ? channelSettings.communityCommentLikeEmail : false,
		commentLikeDingTalk: channelSettings.communityCommentLikeDingTalk !== undefined ? channelSettings.communityCommentLikeDingTalk : false
	};
}

/**
 * Serialize community settings for API submission
 * @param {object} community - Community settings object
 * @returns {object} Serialized community settings
 */
export function serializeCommunitySettings(community) {
	return {
		communityLikeEmail: community.likeEmail,
		communityLikeDingTalk: community.likeDingTalk,
		communityBookmarkEmail: community.bookmarkEmail,
		communityBookmarkDingTalk: community.bookmarkDingTalk,
		communityCommentEmail: community.commentEmail,
		communityCommentDingTalk: community.commentDingTalk,
		communityCommentLikeEmail: community.commentLikeEmail,
		communityCommentLikeDingTalk: community.commentLikeDingTalk
	};
}

/**
 * Get notification type by key
 * @param {string} key - Notification type key
 * @returns {object|null} Notification type object
 */
export function getCommunityNotificationType(key) {
	return COMMUNITY_NOTIFICATION_TYPES.find(t => t.key === key) || null;
}

/**
 * Check if any community notifications are enabled
 * @param {object} community - Community settings object
 * @returns {boolean} Whether any community notifications are enabled
 */
export function hasAnyCommunityNotification(community) {
	return Object.values(community).some(value => value === true);
}

/**
 * Get enabled channels for a community notification type
 * @param {string} typeKey - Community notification type key
 * @param {object} community - Community settings object
 * @returns {Array} Array of enabled channel keys
 */
export function getEnabledChannelsForType(typeKey, community) {
	const channels = [];
	const prefix = typeKey;

	if (community[`${prefix}Email`]) channels.push('email');
	if (community[`${prefix}DingTalk`]) channels.push('dingtalk');

	return channels;
}

/**
 * Toggle community preference setting
 * @param {object} community - Community settings object
 * @param {string} preference - Preference key to toggle
 * @param {boolean} value - New value
 * @returns {object} Updated community settings
 */
export function toggleCommunityPreference(community, preference, value) {
	return {
		...community,
		[preference]: value
	};
}
