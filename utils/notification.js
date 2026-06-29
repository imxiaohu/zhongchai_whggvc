/**
 * Notification Settings Utilities
 * Notification type definitions, channel types, default settings, and helper functions
 */

/**
 * Notification type definitions
 * Each type represents a category of notifications users can enable/disable
 */
export const NOTIFICATION_TYPES = [
	{
		key: 'pushNotification',
		name: '系统推送',
		description: '接收应用内的实时推送'
	},
	{
		key: 'emailNotification',
		name: '邮件通知',
		description: '接收重要的邮件摘要'
	},
	{
		key: 'newsNotification',
		name: '通知公告',
		description: '接收学校发布的通知公告'
	},
	{
		key: 'scoreNotification',
		name: '成绩更新',
		description: '成绩更新时第一时间通知'
	},
	{
		key: 'scheduleNotification',
		name: '课表变更',
		description: '课表有变动时进行提醒'
	}
];

/**
 * Channel type definitions
 * Each channel represents a delivery method for notifications
 */
export const CHANNEL_TYPES = [
	{
		key: 'email',
		name: '电子邮件',
		description: '通过注册邮箱接收详细报告',
		available: true,
		requiresConfig: true,
		configField: 'emailAddress'
	},
	{
		key: 'sms',
		name: '手机短信',
		description: '通过手机短信接收关键摘要',
		available: true,
		requiresConfig: true,
		configField: 'phoneNumber'
	},
	{
		key: 'dingtalk',
		name: '钉钉机器人',
		description: '通过钉钉群机器人接收通知',
		available: true,
		requiresConfig: true,
		configField: 'dingTalkWebhookURL'
	}
];

/**
 * Frequency options for score check
 */
export const FREQUENCY_OPTIONS = [
	{ label: '每小时', value: 'hourly' },
	{ label: '每天', value: 'daily' },
	{ label: '每周', value: 'weekly' }
];

/**
 * Default notification settings
 */
export const DEFAULT_NOTIFICATION_SETTINGS = {
	pushNotification: true,
	emailNotification: false,
	newsNotification: true,
	scoreNotification: true,
	scheduleNotification: true
};

/**
 * Default channel settings
 */
export const DEFAULT_CHANNEL_SETTINGS = {
	emailEnabled: false,
	emailAddress: '',
	smsEnabled: false,
	phoneNumber: '',
	dingTalkEnabled: false,
	dingTalkWebhookURL: '',
	dingTalkSecret: ''
};

/**
 * Default score update settings
 */
export const DEFAULT_SCORE_UPDATE_SETTINGS = {
	email: false,
	sms: false,
	dingtalk: false
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
 * Default score check settings
 */
export const DEFAULT_SCORE_CHECK_SETTINGS = {
	enabled: false,
	frequency: 'daily',
	time: '09:00',
	semester: 'current'
};

/**
 * Default class reminder settings
 */
export const DEFAULT_CLASS_REMINDER_SETTINGS = {
	enabled: false,
	minutesBefore: 15,
	channels: {
		email: false,
		sms: false,
		dingtalk: false
	}
};

/**
 * Get icon type for a notification type
 * @param {string} type - Notification type key
 * @returns {string} Icon name
 */
export function getNotificationIconType(type) {
	switch (type) {
		case 'pushNotification': return 'notification';
		case 'emailNotification': return 'mail';
		case 'newsNotification': return 'article';
		case 'scoreNotification': return 'send';
		case 'scheduleNotification': return 'calendar';
		default: return 'notification';
	}
}

/**
 * Get icon class for a notification type
 * @param {string} type - Notification type key
 * @returns {string} CSS class name
 */
export function getNotificationIconClass(type) {
	switch (type) {
		case 'pushNotification': return 'icon-push';
		case 'emailNotification': return 'icon-email';
		case 'newsNotification': return 'icon-news';
		case 'scoreNotification': return 'icon-score';
		case 'scheduleNotification': return 'icon-schedule';
		default: return 'icon-default';
	}
}

/**
 * Get icon type for a channel
 * @param {string} channel - Channel key
 * @returns {string} Icon name
 */
export function getChannelIconType(channel) {
	switch (channel) {
		case 'email': return 'mail';
		case 'sms': return 'mail';
		case 'dingtalk': return 'robot';
		default: return 'notification';
	}
}

/**
 * Get icon class for a channel
 * @param {string} channel - Channel key
 * @returns {string} CSS class name
 */
export function getChannelIconClass(channel) {
	switch (channel) {
		case 'email': return 'icon-email';
		case 'sms': return 'icon-sms';
		case 'dingtalk': return 'icon-dingtalk';
		default: return 'icon-default';
	}
}

/**
 * Get channel property name from channel key
 * @param {string} channel - Channel key
 * @returns {string} Property name for enabled state
 */
export function getChannelPropertyName(channel) {
	switch (channel) {
		case 'email': return 'emailEnabled';
		case 'sms': return 'smsEnabled';
		case 'dingtalk': return 'dingTalkEnabled';
		default: return channel + 'Enabled';
	}
}

/**
 * Get channel enabled state from settings
 * @param {object} channels - Channels settings object
 * @param {string} channel - Channel key
 * @returns {boolean} Whether channel is enabled
 */
export function getChannelEnabled(channels, channel) {
	switch (channel) {
		case 'email': return channels.emailEnabled;
		case 'sms': return channels.smsEnabled;
		case 'dingtalk': return channels.dingTalkEnabled;
		default: return channels[channel + 'Enabled'];
	}
}

/**
 * Check if channel is available (can be enabled)
 * @param {string} channel - Channel key
 * @returns {boolean} Whether channel is available
 */
export function isChannelAvailable(channel) {
	return CHANNEL_TYPES.some(c => c.key === channel && c.available);
}

/**
 * Check if channel requires configuration
 * @param {string} channel - Channel key
 * @returns {boolean} Whether channel requires configuration
 */
export function channelRequiresConfig(channel) {
	return CHANNEL_TYPES.some(c => c.key === channel && c.requiresConfig);
}

/**
 * Get frequency index from frequency options
 * @param {string} frequency - Current frequency value
 * @returns {number} Index in frequency options
 */
export function getFrequencyIndex(frequency) {
	return FREQUENCY_OPTIONS.findIndex(option => option.value === frequency);
}

/**
 * Get frequency label by value
 * @param {string} frequency - Current frequency value
 * @returns {string} Frequency label
 */
export function getFrequencyLabel(frequency) {
	const option = FREQUENCY_OPTIONS.find(option => option.value === frequency);
	return option ? option.label : FREQUENCY_OPTIONS[0].label;
}

/**
 * Build semester options from available semesters
 * @param {Array} availableSemesters - List of available semester objects
 * @returns {Array} Options array with 'current', 'all', and semester-specific options
 */
export function buildSemesterOptions(availableSemesters = []) {
	const options = [
		{ label: '当前学期', value: 'current' },
		{ label: '全部学期', value: 'all' }
	];

	availableSemesters.forEach(semester => {
		options.push({
			label: semester.name,
			value: semester.code
		});
	});

	return options;
}

/**
 * Get semester index from options
 * @param {Array} semesterOptions - Semester options array
 * @param {string} semester - Current semester value
 * @returns {number} Index in semester options
 */
export function getSemesterIndex(semesterOptions, semester) {
	return semesterOptions.findIndex(option => option.value === semester);
}

/**
 * Get semester label by value
 * @param {Array} semesterOptions - Semester options array
 * @param {string} semester - Current semester value
 * @returns {string} Semester label
 */
export function getSemesterLabel(semesterOptions, semester) {
	const option = semesterOptions.find(option => option.value === semester);
	return option ? option.label : semesterOptions[0].label;
}

/**
 * Serialize notification settings for API submission
 * @param {object} settings - Full settings object
 * @returns {object} Serialized user settings
 */
export function serializeNotificationSettings(settings) {
	return {
		pushNotification: settings.pushNotification,
		emailNotification: settings.emailNotification,
		newsNotification: settings.newsNotification,
		scoreNotification: settings.scoreNotification,
		scheduleNotification: settings.scheduleNotification
	};
}

/**
 * Serialize channel settings for API submission
 * @param {object} settings - Full settings object
 * @returns {object} Serialized channel settings
 */
export function serializeChannelSettings(settings) {
	return {
		emailEnabled: settings.channels.emailEnabled,
		emailAddress: settings.channels.emailAddress,
		smsEnabled: settings.channels.smsEnabled,
		phoneNumber: settings.channels.phoneNumber,
		dingTalkEnabled: settings.channels.dingTalkEnabled,
		dingTalkWebhookUrl: settings.channels.dingTalkWebhookURL,
		dingTalkSecret: settings.channels.dingTalkSecret,
		scoreUpdateEmail: settings.scoreUpdate.email,
		scoreUpdateSms: settings.scoreUpdate.sms,
		scoreUpdateDingTalk: settings.scoreUpdate.dingtalk,
		communityLikeEmail: settings.community.likeEmail,
		communityLikeDingTalk: settings.community.likeDingTalk,
		communityBookmarkEmail: settings.community.bookmarkEmail,
		communityBookmarkDingTalk: settings.community.bookmarkDingTalk,
		communityCommentEmail: settings.community.commentEmail,
		communityCommentDingTalk: settings.community.commentDingTalk,
		communityCommentLikeEmail: settings.community.commentLikeEmail,
		communityCommentLikeDingTalk: settings.community.commentLikeDingTalk,
		scoreCheckEnabled: settings.scoreCheck.enabled,
		scoreCheckFrequency: settings.scoreCheck.frequency,
		scoreCheckTime: settings.scoreCheck.time,
		scoreCheckSemester: settings.scoreCheck.semester,
		classReminderEnabled: settings.classReminder.enabled,
		classReminderMinutesBefore: settings.classReminder.minutesBefore,
		classReminderChannelEmail: settings.classReminder.channels.email,
		classReminderChannelSms: settings.classReminder.channels.sms,
		classReminderChannelDingTalk: settings.classReminder.channels.dingtalk
	};
}

/**
 * Parse user notification settings from API response
 * @param {object} userSettings - User settings from API
 * @returns {object} Parsed notification settings
 */
export function parseNotificationSettings(userSettings) {
	return {
		pushNotification: userSettings.pushNotification !== undefined ? userSettings.pushNotification : true,
		emailNotification: userSettings.emailNotification !== undefined ? userSettings.emailNotification : false,
		newsNotification: userSettings.newsNotification !== undefined ? userSettings.newsNotification : true,
		scoreNotification: userSettings.scoreNotification !== undefined ? userSettings.scoreNotification : true,
		scheduleNotification: userSettings.scheduleNotification !== undefined ? userSettings.scheduleNotification : true
	};
}

/**
 * Parse channel settings from API response
 * @param {object} channelSettings - Channel settings from API
 * @returns {object} Parsed channel settings structure
 */
export function parseChannelSettings(channelSettings) {
	return {
		channels: {
			emailEnabled: channelSettings.emailEnabled !== undefined ? channelSettings.emailEnabled : false,
			emailAddress: channelSettings.emailAddress !== undefined ? channelSettings.emailAddress : '',
			smsEnabled: channelSettings.smsEnabled !== undefined ? channelSettings.smsEnabled : false,
			phoneNumber: channelSettings.phoneNumber !== undefined ? channelSettings.phoneNumber : '',
			dingTalkEnabled: channelSettings.dingTalkEnabled !== undefined ? channelSettings.dingTalkEnabled : false,
			dingTalkWebhookURL: channelSettings.dingTalkWebhookUrl !== undefined ? channelSettings.dingTalkWebhookUrl : '',
			dingTalkSecret: channelSettings.dingTalkSecret !== undefined ? channelSettings.dingTalkSecret : ''
		},
		scoreUpdate: {
			email: channelSettings.scoreUpdateEmail !== undefined ? channelSettings.scoreUpdateEmail : false,
			sms: channelSettings.scoreUpdateSms !== undefined ? channelSettings.scoreUpdateSms : false,
			dingtalk: channelSettings.scoreUpdateDingTalk !== undefined ? channelSettings.scoreUpdateDingTalk : false
		},
		community: {
			likeEmail: channelSettings.communityLikeEmail !== undefined ? channelSettings.communityLikeEmail : false,
			likeDingTalk: channelSettings.communityLikeDingTalk !== undefined ? channelSettings.communityLikeDingTalk : false,
			bookmarkEmail: channelSettings.communityBookmarkEmail !== undefined ? channelSettings.communityBookmarkEmail : false,
			bookmarkDingTalk: channelSettings.communityBookmarkDingTalk !== undefined ? channelSettings.communityBookmarkDingTalk : false,
			commentEmail: channelSettings.communityCommentEmail !== undefined ? channelSettings.communityCommentEmail : false,
			commentDingTalk: channelSettings.communityCommentDingTalk !== undefined ? channelSettings.communityCommentDingTalk : false,
			commentLikeEmail: channelSettings.communityCommentLikeEmail !== undefined ? channelSettings.communityCommentLikeEmail : false,
			commentLikeDingTalk: channelSettings.communityCommentLikeDingTalk !== undefined ? channelSettings.communityCommentLikeDingTalk : false
		},
		scoreCheck: {
			enabled: channelSettings.scoreCheckEnabled !== undefined ? channelSettings.scoreCheckEnabled : false,
			frequency: channelSettings.scoreCheckFrequency !== undefined ? channelSettings.scoreCheckFrequency : 'daily',
			time: channelSettings.scoreCheckTime !== undefined ? channelSettings.scoreCheckTime : '09:00',
			semester: channelSettings.scoreCheckSemester !== undefined ? channelSettings.scoreCheckSemester : 'current'
		},
		classReminder: {
			enabled: channelSettings.classReminderEnabled !== undefined ? channelSettings.classReminderEnabled : false,
			minutesBefore: channelSettings.classReminderMinutesBefore !== undefined ? channelSettings.classReminderMinutesBefore : 15,
			channels: {
				email: channelSettings.classReminderChannelEmail !== undefined ? channelSettings.classReminderChannelEmail : false,
				sms: channelSettings.classReminderChannelSms !== undefined ? channelSettings.classReminderChannelSms : false,
				dingtalk: channelSettings.classReminderChannelDingTalk !== undefined ? channelSettings.classReminderChannelDingTalk : false
			}
		}
	};
}

/**
 * Parse SMS balance data from API response
 * @param {object} balanceData - SMS balance from API
 * @returns {object} Parsed SMS balance
 */
export function parseSMSBalance(balanceData) {
	return {
		balance: balanceData.balance || 0,
		balanceYuan: balanceData.balanceYuan || '0.00',
		totalSpent: balanceData.totalSpent || 0,
		totalSpentYuan: balanceData.totalSpentYuan || '0.00',
		smsCost: balanceData.smsCost || 10,
		smsCostYuan: balanceData.smsCostYuan || '0.10'
	};
}

/**
 * Default SMS balance object
 */
export const DEFAULT_SMS_BALANCE = {
	balance: 0,
	balanceYuan: '0.00',
	totalSpent: 0,
	totalSpentYuan: '0.00',
	smsCost: 10,
	smsCostYuan: '0.10'
};
