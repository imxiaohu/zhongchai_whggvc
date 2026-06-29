import { showToast, showLoading, hideLoading, navigateBack } from '@/pages/api/page.js';
import SettingsItem from '@/pages/user/components/SettingsItem.vue';
import NotificationLogItem from '@/pages/user/components/NotificationLogItem.vue';
import ChannelConfig from '@/pages/user/components/ChannelConfig.vue';
import ScoreCheckConfig from '@/pages/user/components/ScoreCheckConfig.vue';
import {
	getScoreCheckLogs,
	notificationManager
} from '@/api/notification.js';
import {
	CHANNEL_TYPES,
	FREQUENCY_OPTIONS,
	DEFAULT_CHANNEL_SETTINGS,
	DEFAULT_SCORE_UPDATE_SETTINGS,
	DEFAULT_SCORE_CHECK_SETTINGS,
	DEFAULT_SMS_BALANCE,
	getChannelIconType as getChIconType,
	getChannelIconClass as getChIconClass,
	getChannelPropertyName,
	getChannelEnabled,
	getFrequencyIndex as getFreqIdx,
	getFrequencyLabel as getFreqLbl,
	buildSemesterOptions,
	getSemesterIndex as getSemIdx,
	getSemesterLabel as getSemLbl,
	serializeChannelSettings,
	parseChannelSettings,
	parseSMSBalance
} from '@/utils/notification.js';
import {
	getLogStatusText,
	getLogStatusClass,
	getLogTypeText,
	formatLogTime,
	parseScoreCheckLogs,
	parseTestResultMessage,
	DEFAULT_PAGINATION
} from '@/utils/scoreCheck.js';

export default {
	components: {
		SettingsItem,
		NotificationLogItem,
		ChannelConfig,
		ScoreCheckConfig
	},
	data() {
		return {
			primaryColor: '#6366f1',
			settings: {
				channels: { ...DEFAULT_CHANNEL_SETTINGS },
				scoreUpdate: { ...DEFAULT_SCORE_UPDATE_SETTINGS },
				scoreCheck: { ...DEFAULT_SCORE_CHECK_SETTINGS }
			},
			hasUnsavedChanges: false,
			smsBalance: { ...DEFAULT_SMS_BALANCE },
			testingSMS: false,
			availableSemesters: [],
			currentSemester: null,
			loadingSemesters: false,
			showLogsModal: false,
			scoreCheckLogs: [],
			logsLoading: false,
			logsPagination: { ...DEFAULT_PAGINATION }
		};
	},

	computed: {
		notificationChannels() { return CHANNEL_TYPES; },
		frequencyOptions() { return FREQUENCY_OPTIONS; },
		semesterOptions() {
			return buildSemesterOptions(this.availableSemesters);
		}
	},

	onLoad() {
		console.log('成绩通知页面加载');
		this.loadNotificationData();
		this.loadSemesterInfo();
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

		async loadNotificationData() {
			try {
				showLoading({ title: '加载中' });
				const data = await notificationManager.init();

				if (data.smsBalance) {
					this.smsBalance = parseSMSBalance(data.smsBalance);
				}

				if (data.channels) {
					const parsed = parseChannelSettings(data.channels);
					Object.assign(this.settings.channels, parsed.channels);
					Object.assign(this.settings.scoreUpdate, parsed.scoreUpdate);
					Object.assign(this.settings.scoreCheck, parsed.scoreCheck);
				}

				this.hasUnsavedChanges = false;
			} catch (error) {
				console.error('加载成绩通知设置出错:', error);
			} finally {
				hideLoading();
			}
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

		toggleScoreCheck(value) {
			this.settings.scoreCheck.enabled = value;
			this.hasUnsavedChanges = true;
		},

		updateScoreCheckFrequency(index) {
			this.settings.scoreCheck.frequency = this.frequencyOptions[index].value;
			this.hasUnsavedChanges = true;
		},

		updateScoreCheckTime(value) {
			this.settings.scoreCheck.time = value;
			this.hasUnsavedChanges = true;
		},

		updateScoreCheckSemester(index) {
			this.settings.scoreCheck.semester = this.semesterOptions[index].value;
			this.hasUnsavedChanges = true;
		},

		async saveSettings() {
			try {
				showLoading({ title: '保存中...' });
				const channelSettingsData = serializeChannelSettings(this.settings);
				await notificationManager.updateChannels(channelSettingsData);
				this.hasUnsavedChanges = false;
				showToast({ title: '成绩通知设置保存成功', icon: 'success' });
				setTimeout(() => { navigateBack(); }, 1500);
			} catch (error) {
				console.error('保存成绩通知设置出错:', error);
			} finally {
				hideLoading();
			}
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

		getFrequencyIndex() {
			return getFreqIdx(this.settings.scoreCheck.frequency);
		},

		getFrequencyLabel() {
			return getFreqLbl(this.settings.scoreCheck.frequency);
		},

		getSemesterIndex() {
			return getSemIdx(this.semesterOptions, this.settings.scoreCheck.semester);
		},

		getSemesterLabel() {
			return getSemLbl(this.semesterOptions, this.settings.scoreCheck.semester);
		},

		async loadSemesterInfo() {
			try {
				this.loadingSemesters = true;
				const semesterInfo = await notificationManager.loadSemesterInfo();
				this.currentSemester = semesterInfo.currentSemester;
				this.availableSemesters = semesterInfo.availableSemesters;
			} catch (error) {
				console.error('加载学期信息出错:', error);
			} finally {
				this.loadingSemesters = false;
			}
		},

		async testScoreCheck() {
			try {
				showLoading({ title: '正在测试成绩检查...' });
				const result = await notificationManager.performTestNotification();
				if (result.success) {
					showToast({ title: parseTestResultMessage(result), icon: 'success' });
				} else {
					showToast({ title: result.message || '测试失败，请稍后重试', icon: 'none' });
				}
			} catch (error) {
				console.error('测试成绩检查出错:', error);
				showToast({ title: '测试请求失败', icon: 'none' });
			} finally {
				hideLoading();
			}
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
		},

		async showScoreCheckLogs() {
			try {
				this.logsLoading = true;
				this.showLogsModal = true;
				await this.loadScoreCheckLogs();
			} catch (error) {
				console.error('显示成绩检查日志出错:', error);
			}
		},

		async loadScoreCheckLogs(page = 1) {
			try {
				this.logsLoading = true;
				const response = await getScoreCheckLogs({ page: page, pageSize: this.logsPagination.pageSize });
				if (response.success) {
					const { logs, pagination } = parseScoreCheckLogs(response);
					this.scoreCheckLogs = logs;
					this.logsPagination = pagination;
				} else {
					showToast({ title: response.message || '加载日志失败' });
				}
			} catch (error) {
				console.error('加载成绩检查日志出错:', error);
			} finally {
				this.logsLoading = false;
			}
		},

		closeLogsModal() {
			this.showLogsModal = false;
			this.logsLoading = false;
			this.scoreCheckLogs = [];
			this.logsPagination = { page: 1, pageSize: 20, total: 0, pages: 0 };
		},

		formatLogTime(timeStr) {
			return formatLogTime(timeStr);
		},

		getLogStatusText(status) {
			return getLogStatusText(status);
		},

		getLogStatusClass(status) {
			return getLogStatusClass(status);
		},

		getLogTypeText(type) {
			return getLogTypeText(type);
		}
	}
};
