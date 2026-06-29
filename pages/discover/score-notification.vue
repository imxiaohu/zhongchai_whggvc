<template>
	<view class="sn-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="sn-hero">
			<view class="sn-hero-bg"></view>
			<view class="sn-hero-overlay"></view>

			<view class="sn-hero-nav">
				<view class="sn-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="sn-hero-title">成绩通知</text>
				<view style="width: 64rpx;"></view>
			</view>

			<view class="sn-hero-content">
				<text class="sn-hero-sub">SCORE NOTIFICATION</text>
			</view>
		</view>

		<!-- 内容区域 -->
		<scroll-view class="sn-scroll" scroll-y="true">
			<view class="sn-content">
				<!-- 送达渠道卡片 -->
				<view class="sn-card">
					<view class="sn-card-header">
						<view class="sn-card-header-icon sn-card-header-icon--green">
							<l-icon name="send" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<view class="sn-card-header-text">
							<text class="sn-card-title">送达渠道</text>
							<text class="sn-card-sub">配置成绩通知的送达方式</text>
						</view>
					</view>
					<view class="sn-settings-list">
						<view class="sn-settings-group" v-for="channel in notificationChannels" :key="channel.key">
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

				<!-- 成绩主动查询卡片 -->
				<view class="sn-card">
					<view class="sn-card-header">
						<view class="sn-card-header-icon sn-card-header-icon--orange">
							<l-icon name="search" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<view class="sn-card-header-text">
							<text class="sn-card-title">成绩主动查询</text>
							<text class="sn-card-sub">配置系统自动为您查询成绩的频率</text>
						</view>
					</view>
					<view class="sn-settings-list">
						<ToggleItem
							title="启用自动查询"
							icon="check-circle"
							iconClass="icon-check"
							:checked="settings.scoreCheck.enabled"
							:color="primaryColor"
							@change="toggleScoreCheck"
						/>
						<ScoreCheckConfig
							v-if="settings.scoreCheck.enabled"
							:frequencyOptions="frequencyOptions"
							:frequencyIndex="getFrequencyIndex()"
							:frequencyLabel="getFrequencyLabel()"
							:time="settings.scoreCheck.time"
							:semesterOptions="semesterOptions"
							:semesterIndex="getSemesterIndex()"
							:semesterLabel="getSemesterLabel()"
							:loadingSemesters="loadingSemesters"
							:enabled="settings.scoreCheck.enabled"
							@frequency-change="updateScoreCheckFrequency"
							@time-change="updateScoreCheckTime"
							@semester-change="updateScoreCheckSemester"
							@test="testScoreCheck"
							@view-logs="showScoreCheckLogs"
						/>
					</view>
				</view>

				<!-- 保存按钮 -->
				<view class="sn-save-container">
					<view
						class="sn-save-btn"
						:class="{ 'sn-save-btn--disabled': !hasUnsavedChanges }"
						@tap="saveSettings"
					>
						<l-icon name="save" style="font-size: 20px; color: #fff; margin-right: 12rpx;"></l-icon>
						<text class="sn-save-btn-text">保存更改</text>
					</view>
				</view>
			</view>
		</scroll-view>

		<!-- 成绩检查日志弹窗 -->
		<view v-if="showLogsModal" class="sn-modal-overlay" @tap="closeLogsModal">
			<view class="sn-modal" @tap.stop>
				<view class="sn-modal-header">
					<view class="sn-modal-title-row">
						<l-icon name="article" style="font-size: 20px; color: var(--primary-500); margin-right: 12rpx;"></l-icon>
						<text class="sn-modal-title">成绩检查日志</text>
					</view>
					<view class="sn-modal-close" @tap="closeLogsModal">
						<l-icon name="close" style="font-size: 20px; color: var(--text-tertiary);"></l-icon>
					</view>
				</view>

				<view class="sn-modal-body">
					<view v-if="logsLoading" class="sn-modal-loading">
						<view class="sn-modal-spinner"></view>
						<text class="sn-modal-loading-text">正在获取日志...</text>
					</view>

					<scroll-view v-else class="sn-logs-scroll" scroll-y="true">
						<view v-if="scoreCheckLogs.length === 0" class="sn-logs-empty">
							<l-icon name="file-text-off" style="font-size: 48px; color: var(--text-tertiary); margin-bottom: 16rpx;"></l-icon>
							<text class="sn-logs-empty-text">暂无日志记录</text>
						</view>
						<NotificationLogItem v-for="log in scoreCheckLogs" :key="log.id" :log="log" />
					</scroll-view>

					<view v-if="logsPagination.total > 0" class="sn-pagination">
						<view class="sn-pagination-info">
							<text class="sn-pagination-page">第 {{ logsPagination.page }} / {{ logsPagination.pages }} 页</text>
							<text class="sn-pagination-total">共 {{ logsPagination.total }} 条</text>
						</view>
						<view class="sn-pagination-btns">
							<view
								class="sn-page-btn"
								:class="{ 'sn-page-btn--disabled': logsPagination.page <= 1 }"
								@tap="logsPagination.page > 1 && loadScoreCheckLogs(logsPagination.page - 1)"
							>
								<l-icon name="chevron-left" style="font-size: 18px; color: var(--text-primary);"></l-icon>
							</view>
							<view
								class="sn-page-btn"
								:class="{ 'sn-page-btn--disabled': logsPagination.page >= logsPagination.pages }"
								@tap="logsPagination.page < logsPagination.pages && loadScoreCheckLogs(logsPagination.page + 1)"
							>
								<l-icon name="chevron-right" style="font-size: 18px; color: var(--text-primary);"></l-icon>
							</view>
						</view>
					</view>
				</view>
			</view>
		</view>
	</view>
</template>

<script>
import ToggleItem from '@/pages/user/components/ToggleItem.vue';
import scoreNotificationLogic from './logic/score-notification.js';
export default {
	...scoreNotificationLogic,
	components: {
		...scoreNotificationLogic.components,
		ToggleItem
	}
};
</script>

<style lang="scss" scoped>
/* ============================================
   Score Notification Page - Hero Style
   ============================================ */

.sn-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	position: relative;
	overflow: hidden;
}

/* ---- Hero Section ---- */
.sn-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 200rpx;
}

.sn-hero-bg {
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

.sn-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.65) 0%,
		rgba(37, 99, 235, 0.4) 50%,
		rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.sn-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.sn-back-btn {
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

.sn-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.sn-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 8rpx;
}

.sn-hero-sub {
	font-size: 20rpx;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.6);
	letter-spacing: 2px;
}

/* ---- Scroll ---- */
.sn-scroll {
	height: calc(100vh - 200rpx);
}

/* ---- Content ---- */
.sn-content {
	padding: 24rpx;
	padding-bottom: 60rpx;
}

/* ---- Card ---- */
.sn-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 32rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1px solid rgba(148, 163, 184, 0.12);
}

.sn-card-header {
	display: flex;
	align-items: flex-start;
	gap: 20rpx;
	margin-bottom: 28rpx;
}

.sn-card-header-icon {
	width: 80rpx;
	height: 80rpx;
	border-radius: 24rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;

	&--green { background: linear-gradient(135deg, #10b981, #059669); }
	&--orange { background: linear-gradient(135deg, #f97316, #ea580c); }
}

.sn-card-header-text {
	flex: 1;
	padding-top: 4rpx;
}

.sn-card-title {
	display: block;
	font-size: 32rpx;
	font-weight: 800;
	color: var(--text-primary);
	margin-bottom: 6rpx;
}

.sn-card-sub {
	display: block;
	font-size: 22rpx;
	color: var(--text-tertiary);
}

/* ---- Settings List ---- */
.sn-settings-list {
	display: flex;
	flex-direction: column;
	gap: 12rpx;
}

.sn-settings-group {
	display: flex;
	flex-direction: column;
}

/* ---- Save Button ---- */
.sn-save-container {
	margin-top: 8rpx;
}

	.sn-save-btn {
	height: 100rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 50rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 8rpx 32rpx rgba(37, 99, 235, 0.3);
	transition: all 0.2s ease;

	.sn-save-btn-text {
		font-size: 32rpx;
		font-weight: 800;
		color: #fff;
	}

	&--disabled {
		background: var(--bg-muted);
		box-shadow: none;
		.sn-save-btn-text { color: var(--text-tertiary); }
	}

	&:active:not(.sn-save-btn--disabled) {
		transform: translateY(2rpx);
		box-shadow: 0 4rpx 16rpx rgba(37, 99, 235, 0.25);
	}
}

/* ---- Modal ---- */
.sn-modal-overlay {
	position: fixed;
	top: 0; left: 0; right: 0; bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 1000;
	backdrop-filter: blur(4px);
}

.sn-modal {
	width: 92%;
	max-height: 80vh;
	background: #fff;
	border-radius: 28rpx;
	display: flex;
	flex-direction: column;
	overflow: hidden;
	box-shadow: 0 20rpx 60rpx rgba(0, 0, 0, 0.2);
}

.sn-modal-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 32rpx;
	border-bottom: 1px solid rgba(226, 232, 240, 0.8);
}

.sn-modal-title-row {
	display: flex;
	align-items: center;
}

.sn-modal-title {
	font-size: 34rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.sn-modal-close {
	width: 64rpx;
	height: 64rpx;
	border-radius: 50%;
	background: var(--bg-muted);
	display: flex;
	align-items: center;
	justify-content: center;

	&:active { opacity: 0.7; }
}

.sn-modal-body {
	flex: 1;
	overflow: hidden;
	display: flex;
	flex-direction: column;
}

.sn-modal-loading {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 80rpx 0;
	gap: 20rpx;
}

.sn-modal-spinner {
	width: 60rpx;
	height: 60rpx;
	border: 4rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: sn-spin 0.8s linear infinite;
}

@keyframes sn-spin {
	to { transform: rotate(360deg); }
}

.sn-modal-loading-text {
	font-size: 26rpx;
	color: var(--text-tertiary);
}

.sn-logs-scroll {
	flex: 1;
	padding: 24rpx;
}

.sn-logs-empty {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 80rpx 0;
}

.sn-logs-empty-text {
	font-size: 26rpx;
	color: var(--text-tertiary);
}

.sn-pagination {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 20rpx 32rpx;
	border-top: 1px solid rgba(226, 232, 240, 0.8);
}

.sn-pagination-info {
	display: flex;
	flex-direction: column;
	gap: 4rpx;
}

.sn-pagination-page {
	font-size: 24rpx;
	color: var(--text-secondary);
}

.sn-pagination-total {
	font-size: 20rpx;
	color: var(--text-tertiary);
}

.sn-pagination-btns {
	display: flex;
	gap: 16rpx;
}

.sn-page-btn {
	width: 72rpx;
	height: 72rpx;
	border-radius: 16rpx;
	background: var(--bg-muted);
	display: flex;
	align-items: center;
	justify-content: center;

	&--disabled { opacity: 0.3; }
	&:active:not(.sn-page-btn--disabled) { background: var(--primary-soft); }
}
</style>
