<template>
	<view class="ns-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="ns-hero">
			<view class="ns-hero-bg"></view>
			<view class="ns-hero-overlay"></view>

			<view class="ns-hero-nav">
				<view class="ns-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="ns-hero-title">通知设置</text>
				<view style="width: 64rpx;"></view>
			</view>

			<view class="ns-hero-content">
				<text class="ns-hero-sub">NOTIFICATION SETTINGS</text>
			</view>
		</view>

		<!-- 内容区域 -->
		<scroll-view class="ns-scroll" scroll-y="true">
			<view class="ns-content">
				<!-- 通知类型卡片 -->
				<view class="ns-card">
					<view class="ns-card-header">
						<view class="ns-card-header-icon ns-card-header-icon--blue">
							<l-icon name="notification" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<view class="ns-card-header-text">
							<text class="ns-card-title">通知类型</text>
							<text class="ns-card-sub">选择您希望收到的通知类型</text>
						</view>
					</view>
					<view class="ns-settings-list">
						<SettingsItem
							v-for="type in notificationTypes"
							:key="type.key"
							:title="type.name"
							:description="type.description"
							:icon="getIconType(type.key)"
							:iconClass="getIconClass(type.key)"
							:enabled="settings[type.key]"
							:color="primaryColor"
							@toggle="(val) => toggleNotificationType(type.key, val)"
						/>
					</view>
				</view>

				<!-- 送达渠道卡片 -->
				<view class="ns-card">
					<view class="ns-card-header">
						<view class="ns-card-header-icon ns-card-header-icon--green">
							<l-icon name="send" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<view class="ns-card-header-text">
							<text class="ns-card-title">送达渠道</text>
							<text class="ns-card-sub">配置通知的送达方式</text>
						</view>
					</view>
					<view class="ns-settings-list">
						<view class="ns-settings-group" v-for="channel in notificationChannels" :key="channel.key">
							<SettingsItem
								:title="channel.name"
								:description="channel.description"
								:icon="getChannelIconType(channel.key)"
								:iconClass="getChannelIconClass(channel.key)"
								:enabled="getChannelEnabled(channel.key)"
								:available="channel.available"
								:color="primaryColor"
								@toggle="(val) => toggleNotificationChannel(channel.key, val)"
							/>
							<ChannelConfig
								v-if="getChannelEnabled(channel.key)"
								:channelKey="channel.key"
								:settings="settings"
								:smsBalance="smsBalance"
								:testingSMS="testingSMS"
								@email-input="handleEmailInput"
								@phone-input="handlePhoneInput"
								@webhook-input="handleWebhookInput"
								@recharge="goToRecharge"
								@test-sms="testSMS"
							/>
						</view>
					</view>
				</view>

				<!-- 上课提醒卡片 -->
				<view class="ns-card">
					<view class="ns-card-header">
						<view class="ns-card-header-icon ns-card-header-icon--indigo">
							<l-icon name="clock" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<view class="ns-card-header-text">
							<text class="ns-card-title">上课提醒</text>
							<text class="ns-card-sub">上课前通过已开启渠道提醒你</text>
						</view>
					</view>
					<view class="ns-settings-list">
						<view class="ns-class-reminder">
							<view class="ns-class-reminder-row">
								<text class="ns-class-reminder-label">开启上课提醒</text>
								<switch :checked="settings.classReminder.enabled" @change="(e) => toggleClassReminder(e.detail.value)" color="#6366f1" />
							</view>
							<view v-if="settings.classReminder.enabled" class="ns-class-reminder-config">
								<view class="ns-class-reminder-item">
									<text class="ns-class-reminder-item-label">提前提醒时间</text>
									<view class="ns-class-reminder-input-wrap">
										<input
											class="ns-class-reminder-input"
											type="number"
											:value="settings.classReminder.minutesBefore"
											@input="(e) => updateClassReminderMinutes(e.detail.value)"
											placeholder="分钟"
										/>
										<text class="ns-class-reminder-unit">分钟</text>
									</view>
								</view>
								<view class="ns-class-reminder-channels">
									<text class="ns-class-reminder-channels-label">提醒渠道</text>
									<view class="ns-class-reminder-channels-list">
										<view
											v-for="channel in classReminderChannels"
											:key="channel.key"
											class="ns-class-reminder-channel"
											:class="{ 'ns-class-reminder-channel--disabled': !channel.available }"
											@tap="toggleClassReminderChannel(channel.key)"
										>
											<view class="ns-class-reminder-channel-check" :class="{ 'ns-class-reminder-channel-check--active': isClassReminderChannelEnabled(channel.key) }">
												<text v-if="isClassReminderChannelEnabled(channel.key)" class="ns-class-reminder-channel-check-icon">✓</text>
											</view>
											<text class="ns-class-reminder-channel-name">{{ channel.name }}</text>
										</view>
									</view>
								</view>
							</view>
						</view>
					</view>
				</view>

				<!-- 成绩更新提醒卡片 -->
				<view class="ns-card">
					<view class="ns-card-header">
						<view class="ns-card-header-icon ns-card-header-icon--yellow">
							<l-icon name="chart-bar" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<view class="ns-card-header-text">
							<text class="ns-card-title">成绩更新提醒</text>
							<text class="ns-card-sub">配置成绩更新时的通知方式</text>
						</view>
					</view>
					<view class="ns-settings-list">
						<ScoreUpdateItem
							v-for="item in scoreUpdateItems"
							:key="item.key"
							:title="item.title"
							:icon="item.icon"
							:iconClass="item.iconClass"
							:enabled="settings.scoreUpdate[item.key]"
							:channelEnabled="item.key === 'email' ? settings.channels.emailEnabled : item.key === 'sms' ? settings.channels.smsEnabled : settings.channels.dingTalkEnabled"
							:color="primaryColor"
							@toggle="(val) => toggleScoreUpdate(item.key, val)"
						/>
					</view>
				</view>

				<!-- 社区互动通知卡片 -->
				<view class="ns-card">
					<view class="ns-card-header">
						<view class="ns-card-header-icon ns-card-header-icon--purple">
							<l-icon name="user-group" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<view class="ns-card-header-text">
							<text class="ns-card-title">社区通知设置</text>
							<text class="ns-card-sub">配置社区互动相关的通知</text>
						</view>
					</view>
					<view class="ns-pref-grid">
						<PreferenceItem
							v-for="item in communityNotificationTypes"
							:key="item.key"
							:type="item.key"
							:title="item.name"
							:icon="item.icon"
							:iconClass="item.iconClass"
							:settings="settings"
							:channels="item.channels"
							:color="primaryColor"
							@change="toggleCommunityPreference"
						/>
					</view>
				</view>

				<!-- 保存按钮 -->
				<view class="ns-save-container">
					<view
						class="ns-save-btn"
						:class="{ 'ns-save-btn--disabled': !hasUnsavedChanges }"
						@tap="saveSettings"
					>
						<l-icon name="save" style="font-size: 20px; color: #fff; margin-right: 12rpx;"></l-icon>
						<text class="ns-save-btn-text">保存更改</text>
					</view>
				</view>
			</view>
		</scroll-view>
	</view>
</template>

<script>
import notificationSettingsLogic from './logic/notification-settings.js';
export default {
	...notificationSettingsLogic
};
</script>

<style lang="scss" scoped>
/* ============================================
   Notification Settings - Hero Style
   ============================================ */

.ns-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	position: relative;
	overflow: hidden;
}

/* ---- Hero Section ---- */
.ns-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 200rpx;
}

.ns-hero-bg {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		#1e3a8a 0%,
		#1e40af 25%,
		#2563eb 55%,
		#3b82f6 75%,
		#93c5fd 100%);
	z-index: 0;
}

.ns-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.65) 0%,
		rgba(37, 99, 235, 0.4) 50%,
		rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.ns-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.ns-back-btn {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.18);
	border-radius: 50%;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.25);

	&:active {
		background: rgba(255, 255, 255, 0.28);
		transform: scale(0.95);
	}
}

.ns-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.ns-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 8rpx;
}

.ns-hero-sub {
	font-size: 20rpx;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.6);
	letter-spacing: 2px;
}

/* ---- Scroll ---- */
.ns-scroll {
	height: calc(100vh - 200rpx);
}

/* ---- Content ---- */
.ns-content {
	padding: 24rpx;
	padding-bottom: 60rpx;
}

/* ---- Card ---- */
.ns-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 32rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1px solid rgba(148, 163, 184, 0.12);
}

.ns-card-header {
	display: flex;
	align-items: flex-start;
	gap: 20rpx;
	margin-bottom: 28rpx;
}

.ns-card-header-icon {
	width: 80rpx;
	height: 80rpx;
	border-radius: 24rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;

	&--blue { background: linear-gradient(135deg, #3b82f6, #2563eb); }
	&--green { background: linear-gradient(135deg, #10b981, #059669); }
	&--yellow { background: linear-gradient(135deg, #f59e0b, #d97706); }
	&--purple { background: linear-gradient(135deg, #8b5cf6, #7c3aed); }
	&--orange { background: linear-gradient(135deg, #f97316, #ea580c); }
}

.ns-card-header-text {
	flex: 1;
	padding-top: 4rpx;
}

.ns-card-title {
	display: block;
	font-size: 32rpx;
	font-weight: 800;
	color: var(--text-primary);
	margin-bottom: 6rpx;
}

.ns-card-sub {
	display: block;
	font-size: 22rpx;
	color: var(--text-tertiary);
}

/* ---- Settings List ---- */
.ns-settings-list {
	display: flex;
	flex-direction: column;
	gap: 12rpx;
}

.ns-settings-group {
	display: flex;
	flex-direction: column;
}

/* ---- Preference Grid ---- */
.ns-pref-grid {
	display: grid;
	grid-template-columns: repeat(2, 1fr);
	gap: 16rpx;
}

/* ---- Save Button ---- */
.ns-save-container {
	margin-top: 8rpx;
}

.ns-save-btn {
	height: 100rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 50rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 8rpx 32rpx rgba(37, 99, 235, 0.3);
	transition: all 0.2s ease;

	.ns-save-btn-text {
		font-size: 32rpx;
		font-weight: 800;
		color: #fff;
	}

	&--disabled {
		background: var(--bg-muted);
		box-shadow: none;
		.ns-save-btn-text { color: var(--text-tertiary); }
	}

	&:active:not(.ns-save-btn--disabled) {
		transform: translateY(2rpx);
		box-shadow: 0 4rpx 16rpx rgba(37, 99, 235, 0.25);
	}
}

/* ---- Modal ---- */
.ns-modal-overlay {
	position: fixed;
	top: 0; left: 0; right: 0; bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 1000;
	backdrop-filter: blur(4px);
}

.ns-modal {
	width: 92%;
	max-height: 80vh;
	background: #fff;
	border-radius: 28rpx;
	display: flex;
	flex-direction: column;
	overflow: hidden;
	box-shadow: 0 20rpx 60rpx rgba(0, 0, 0, 0.2);
}

.ns-modal-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 32rpx;
	border-bottom: 1px solid rgba(226, 232, 240, 0.8);
}

.ns-modal-title-row {
	display: flex;
	align-items: center;
}

.ns-modal-title {
	font-size: 34rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.ns-modal-close {
	width: 64rpx;
	height: 64rpx;
	border-radius: 50%;
	background: var(--bg-muted);
	display: flex;
	align-items: center;
	justify-content: center;

	&:active { opacity: 0.7; }
}

.ns-modal-body {
	flex: 1;
	overflow: hidden;
	display: flex;
	flex-direction: column;
}

.ns-modal-loading {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 80rpx 0;
	gap: 20rpx;
}

.ns-modal-spinner {
	width: 60rpx;
	height: 60rpx;
	border: 4rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: ns-spin 0.8s linear infinite;
}

@keyframes ns-spin {
	to { transform: rotate(360deg); }
}

.ns-modal-loading-text {
	font-size: 26rpx;
	color: var(--text-tertiary);
}

.ns-logs-scroll {
	flex: 1;
	padding: 24rpx;
}

.ns-logs-empty {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 80rpx 0;
}

.ns-logs-empty-text {
	font-size: 26rpx;
	color: var(--text-tertiary);
}

.ns-pagination {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 20rpx 32rpx;
	border-top: 1px solid rgba(226, 232, 240, 0.8);
}

.ns-pagination-info {
	display: flex;
	flex-direction: column;
	gap: 4rpx;
}

.ns-pagination-page {
	font-size: 24rpx;
	color: var(--text-secondary);
}

.ns-pagination-total {
	font-size: 20rpx;
	color: var(--text-tertiary);
}

.ns-pagination-btns {
	display: flex;
	gap: 16rpx;
}

.ns-page-btn {
	width: 72rpx;
	height: 72rpx;
	border-radius: 16rpx;
	background: var(--bg-muted);
	display: flex;
	align-items: center;
	justify-content: center;

	&--disabled { opacity: 0.3; }
	&:active:not(.ns-page-btn--disabled) { background: var(--primary-soft); }
}

/* ========== 上课提醒 ========== */
.ns-class-reminder {
	display: flex;
	flex-direction: column;
	gap: 20rpx;

	&-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	&-label {
		font-size: 28rpx;
		font-weight: 600;
		color: var(--text-primary);
	}

	&-config {
		display: flex;
		flex-direction: column;
		gap: 20rpx;
		padding-top: 8rpx;
	}

	&-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	&-item-label {
		font-size: 26rpx;
		color: var(--text-secondary);
	}

	&-input-wrap {
		display: flex;
		align-items: center;
		gap: 12rpx;
	}

	&-input {
		width: 140rpx;
		height: 64rpx;
		padding: 0 20rpx;
		border-radius: 12rpx;
		border: 1px solid var(--border-primary);
		background: var(--bg-card);
		text-align: right;
		font-size: 26rpx;
		color: var(--text-primary);
	}

	&-unit {
		font-size: 24rpx;
		color: var(--text-tertiary);
	}

	&-channels {
		display: flex;
		flex-direction: column;
		gap: 12rpx;
	}

	&-channels-label {
		font-size: 24rpx;
		color: var(--text-tertiary);
	}

	&-channels-list {
		display: flex;
		flex-wrap: wrap;
		gap: 16rpx;
	}

	&-channel {
		display: flex;
		align-items: center;
		gap: 10rpx;
		padding: 12rpx 20rpx;
		border-radius: 999rpx;
		border: 1px solid var(--border-primary);
		background: var(--bg-card);

		&--disabled {
			opacity: 0.45;
		}
	}

	&-channel-check {
		width: 32rpx;
		height: 32rpx;
		border-radius: 50%;
		border: 1px solid var(--border-primary);
		display: flex;
		align-items: center;
		justify-content: center;
		background: #fff;

		&--active {
			background: var(--primary-color);
			border-color: var(--primary-color);
		}
	}

	&-channel-check-icon {
		font-size: 20rpx;
		color: #fff;
		font-weight: 700;
	}

	&-channel-name {
		font-size: 24rpx;
		color: var(--text-primary);
	}
}
</style>
