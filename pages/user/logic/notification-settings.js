import { showToast, showLoading, hideLoading, navigateBack } from '@/pages/api/page.js';
import PreferenceItem from '../components/PreferenceItem.vue';
import SettingsItem from '../components/SettingsItem.vue';
import ScoreUpdateItem from '../components/ScoreUpdateItem.vue';
import ToggleItem from '../components/ToggleItem.vue';
import ChannelConfig from '../components/ChannelConfig.vue';
import {
	notificationManager
} from '@/api/notification.js';
import {
	NOTIFICATION_TYPES,
	CHANNEL_TYPES,
	DEFAULT_NOTIFICATION_SETTINGS,
	DEFAULT_CHANNEL_SETTINGS,
	DEFAULT_SCORE_UPDATE_SETTINGS,
	DEFAULT_COMMUNITY_SETTINGS,
	DEFAULT_SCORE_CHECK_SETTINGS,
	DEFAULT_SMS_BALANCE,
	DEFAULT_CLASS_REMINDER_SETTINGS,
	getNotificationIconType,
	getNotificationIconClass,
	getChannelIconType as getChIconType,
	getChannelIconClass as getChIconClass,
	getChannelPropertyName,
	getChannelEnabled,
	serializeNotificationSettings,
	serializeChannelSettings,
	parseNotificationSettings,
	parseChannelSettings,
	parseSMSBalance
} from '@/utils/notification.js';
import {
	COMMUNITY_NOTIFICATION_TYPES
} from '@/utils/communityNotify.js';

export default {
	mixins: [],
	components: {
		PreferenceItem,
		SettingsItem,
		ScoreUpdateItem,
		ToggleItem,
		ChannelConfig
	},
	data() {
		return {
			primaryColor: '#6366f1',
			settings: {
				...DEFAULT_NOTIFICATION_SETTINGS,
				channels: { ...DEFAULT_CHANNEL_SETTINGS },
				scoreUpdate: { ...DEFAULT_SCORE_UPDATE_SETTINGS },
				community: { ...DEFAULT_COMMUNITY_SETTINGS },
				scoreCheck: { ...DEFAULT_SCORE_CHECK_SETTINGS },
				classReminder: { ...DEFAULT_CLASS_REMINDER_SETTINGS }
			},
			hasUnsavedChanges: false,
			smsBalance: { ...DEFAULT_SMS_BALANCE },
			testingSMS: false
		};
	},

	computed: {
		notificationTypes() { return NOTIFICATION_TYPES; },
		notificationChannels() { return CHANNEL_TYPES; },
		communityNotificationTypes() { return COMMUNITY_NOTIFICATION_TYPES; },
		scoreUpdateItems() {
			return [
				{ key: 'email', title: '邮件提醒成绩更新', icon: 'mail', iconClass: 'icon-email' },
				{ key: 'sms', title: '短信提醒成绩更新', icon: 'message', iconClass: 'icon-sms' },
				{ key: 'dingtalk', title: '钉钉提醒成绩更新', icon: 'robot', iconClass: 'icon-dingtalk' }
			];
		},
		classReminderChannels() {
			return [
				{ key: 'email', name: '邮件', available: this.settings.channels.emailEnabled },
				{ key: 'sms', name: '短信', available: this.settings.channels.smsEnabled },
				{ key: 'dingtalk', name: '钉钉', available: this.settings.channels.dingTalkEnabled }
			];
		}
	},

	onLoad() {
		console.log('通知设置页面加载');
		this.loadNotificationSettings();
	},

	onShow() {
		this.loadSMSBalance();
	},

	onUnload() {
		if (this.hasUnsavedChanges) {
			uni.showModal({
				title: '未保存的修改',
				content: '确定要保存修改吗？',
				confirmText: '保存',
				cancelText: '不保存',
				success: (res) => {
					if (res.confirm) {
						this.saveSettings();
					}
				}
			});
		}
	},

	methods: {
		goBack() {
			if (this.hasUnsavedChanges) {
				uni.showModal({
					title: '未保存的修改',
					content: '确定要保存修改吗？',
					confirmText: '保存',
					cancelText: '不保存',
					success: (res) => {
						if (res.confirm) {
							this.saveSettings();
						} else {
							navigateBack();
						}
					}
				});
			} else {
				navigateBack();
			}
		},

		async loadNotificationSettings() {
			try {
				showLoading({ title: '加载中' });
				const data = await notificationManager.init();

				if (data.settings) {
					Object.assign(this.settings, parseNotificationSettings(data.settings));
				}

				if (data.smsBalance) {
					this.smsBalance = parseSMSBalance(data.smsBalance);
				}

				if (data.channels) {
					const parsed = parseChannelSettings(data.channels);
					Object.assign(this.settings.channels, parsed.channels);
					Object.assign(this.settings.scoreUpdate, parsed.scoreUpdate);
					Object.assign(this.settings.community, parsed.community);
					Object.assign(this.settings.scoreCheck, parsed.scoreCheck);
				}

				this.hasUnsavedChanges = false;
			} catch (error) {
				console.error('加载通知设置出错:', error);
			} finally {
				hideLoading();
			}
		},

		toggleNotificationType(type, value) {
			this.settings[type] = value;
			this.hasUnsavedChanges = true;
		},

		toggleNotificationChannel(channel, value) {
			const propertyName = getChannelPropertyName(channel);
			this.settings.channels[propertyName] = value;
			this.hasUnsavedChanges = true;
		},

		handleEmailInput(e) {
			this.settings.channels.emailAddress = e.detail.value;
			this.hasUnsavedChanges = true;
		},
		handlePhoneInput(e) {
			this.settings.channels.phoneNumber = e.detail.value;
			this.hasUnsavedChanges = true;
		},
		handleWebhookInput(e) {
			this.settings.channels.dingTalkWebhookURL = e.detail.value;
			this.hasUnsavedChanges = true;
		},

		toggleScoreUpdate(type, value) {
			this.settings.scoreUpdate[type] = value;
			this.hasUnsavedChanges = true;
		},

		toggleCommunityPreference(preference, value) {
			this.settings.community[preference] = value;
			this.hasUnsavedChanges = true;
		},

		toggleClassReminder(value) {
			this.settings.classReminder.enabled = value;
			this.hasUnsavedChanges = true;
		},

		updateClassReminderMinutes(value) {
			const minutes = parseInt(value, 10);
			this.settings.classReminder.minutesBefore = Number.isFinite(minutes) ? Math.max(1, minutes) : 15;
			this.hasUnsavedChanges = true;
		},

		toggleClassReminderChannel(channel) {
			const current = this.settings.classReminder.channels[channel] || false;
			this.settings.classReminder.channels[channel] = !current;
			this.hasUnsavedChanges = true;
		},

		isClassReminderChannelEnabled(channel) {
			return !!this.settings.classReminder.channels[channel];
		},

		async saveSettings() {
			try {
				showLoading({ title: '保存中...' });

				const userSettingsData = serializeNotificationSettings(this.settings);
				const channelSettingsData = serializeChannelSettings(this.settings);

				await notificationManager.updateSettings(userSettingsData);
				await notificationManager.updateChannels(channelSettingsData);

				this.hasUnsavedChanges = false;
				showToast({ title: '通知设置保存成功', icon: 'success' });
				setTimeout(() => { navigateBack(); }, 1500);
			} catch (error) {
				console.error('保存通知设置出错:', error);
			} finally {
				hideLoading();
			}
		},

		getIconType(type) {
			return getNotificationIconType(type);
		},

		getIconClass(type) {
			return getNotificationIconClass(type);
		},

		getChannelIconType(channel) {
			return getChIconType(channel);
		},

		getChannelIconClass(channel) {
			return getChIconClass(channel);
		},

		getChannelEnabled(channel) {
			return getChannelEnabled(this.settings.channels, channel);
		},

		goToRecharge() {
			uni.navigateTo({ url: '/pages/user/sms-recharge' });
		},

		async testSMS() {
			if (this.testingSMS) return;
			try {
				this.testingSMS = true;
				if (!this.settings.channels.phoneNumber) {
					showToast({ title: '请先配置手机号' });
					return;
				}
				const result = await notificationManager.testChannel('sms');
				if (result.success) {
					showToast({ title: '测试短信发送成功', icon: 'success' });
					await notificationManager.refreshSMSBalance();
					this.smsBalance = notificationManager.smsBalance;
				} else {
					showToast({ title: result.message || '发送失败' });
				}
			} catch (error) {
				console.error('测试短信发送出错:', error);
			} finally {
				this.testingSMS = false;
			}
		},

		async loadSMSBalance() {
			try {
				await notificationManager.refreshSMSBalance();
				if (notificationManager.smsBalance) {
					this.smsBalance = parseSMSBalance(notificationManager.smsBalance);
				}
			} catch (error) {
				console.error('加载短信余额出错:', error);
			}
		}
	}
};
