<template>
	<view class="sync-page">
		<!-- Hero 区域 -->
		<view class="hero-section">
			<view class="hero-bg"></view>
			<view class="hero-bg-overlay"></view>
			<view class="hero-statusbar" :style="{ height: statusBarHeight + 'px' }"></view>

			<!-- 导航栏 -->
			<view class="hero-nav">
				<view class="sync-nav__back" @tap="goBack">
					<l-icon name="arrow-left" size="18" color="#fff"></l-icon>
				</view>
				<text class="sync-nav__title">服务器同步</text>
				<view class="sync-nav__placeholder"></view>
			</view>

			<!-- Hero 内容 -->
			<view class="hero-content">
				<view class="hero-greeting">
					<text class="hero-title">数据同步管理</text>
					<text class="hero-subtitle">保持课程数据与教务系统实时一致</text>
				</view>
			</view>
		</view>

		<!-- 主体内容区域 -->
		<view class="sync-content">

			<!-- 同步状态卡片 -->
			<view class="sync-status-card" :class="`sync-status-card--${syncStatus.syncStatus}`">
				<view class="sync-status-card__header">
					<view class="sync-status-card__icon-wrap">
						<view class="sync-status-card__icon-bg"></view>
						<view class="sync-status-card__icon" :class="{'sync-status-card__icon--spin': syncStatus.syncStatus === 'syncing'}">
							<l-icon :name="getStatusIcon()" size="28" :color="getStatusColor()"></l-icon>
						</view>
					</view>
					<view class="sync-status-card__info">
						<text class="sync-status-card__title">{{ getStatusTitle() }}</text>
						<text class="sync-status-card__desc">{{ getStatusDescription() }}</text>
					</view>
					<view
						class="sync-status-card__refresh"
						v-if="syncSettings.enabled && syncStatus.syncStatus !== 'syncing'"
						@tap="manualSyncAction"
					>
						<l-icon name="refresh" size="16" color="#fff"></l-icon>
					</view>
				</view>

				<view class="sync-status-card__stats" v-if="syncStatus.lastSyncAt">
					<view class="sync-status-card__stat">
						<text class="sync-status-card__stat-value">{{ syncStatus.coursesCount || 0 }}</text>
						<text class="sync-status-card__stat-label">同步课程</text>
					</view>
					<view class="sync-status-card__stat-divider"></view>
					<view class="sync-status-card__stat">
						<text class="sync-status-card__stat-value">{{ formatTime(syncStatus.lastSyncAt) }}</text>
						<text class="sync-status-card__stat-label">最后同步</text>
					</view>
					<view class="sync-status-card__stat-divider" v-if="syncSettings.enabled && syncStatus.nextSyncAt"></view>
					<view class="sync-status-card__stat" v-if="syncSettings.enabled && syncStatus.nextSyncAt">
						<text class="sync-status-card__stat-value">{{ formatNextSyncTime(syncStatus.nextSyncAt) }}</text>
						<text class="sync-status-card__stat-label">下次同步</text>
					</view>
				</view>
			</view>

			<!-- 同步设置卡片 -->
			<view class="sync-card">
				<view class="sync-card__header">
					<view class="sync-card__header-left">
						<text class="sync-card__title">同步设置</text>
						<text class="sync-card__title-sub">SYNC SETTINGS</text>
					</view>
				</view>

				<!-- 启用同步 -->
				<view class="sync-setting-row">
					<view class="sync-setting-row__left">
						<view class="sync-setting-row__icon sync-setting-row__icon--purple">
							<l-icon name="cloud-download" size="16" color="#fff"></l-icon>
						</view>
						<view class="sync-setting-row__text">
							<text class="sync-setting-row__label">启用同步</text>
							<text class="sync-setting-row__desc">开启后自动同步教务系统课程数据</text>
						</view>
					</view>
					<view class="sync-setting-row__switch">
						<switch :checked="syncSettings.enabled" @change="onSyncEnabledChange" color="var(--primary-500)" />
					</view>
				</view>

				<!-- 同步频率 -->
				<view class="sync-setting-row" v-if="syncSettings.enabled">
					<view class="sync-setting-row__left">
						<view class="sync-setting-row__icon sync-setting-row__icon--blue">
							<l-icon name="time" size="16" color="#fff"></l-icon>
						</view>
						<view class="sync-setting-row__text">
							<text class="sync-setting-row__label">同步频率</text>
							<text class="sync-setting-row__desc">{{ getFrequencyDescriptionValue() }}</text>
						</view>
					</view>
					<view class="sync-setting-row__arrow" @tap="showFrequencyPicker">
						<text class="sync-setting-row__arrow-text">{{ getFrequencyTextValue() }}</text>
						<l-icon name="chevron-right" size="14" color="var(--text-tertiary)"></l-icon>
					</view>
				</view>

				<!-- 同步时间段 -->
				<view class="sync-setting-row" v-if="syncSettings.enabled">
					<view class="sync-setting-row__left">
						<view class="sync-setting-row__icon sync-setting-row__icon--amber">
							<l-icon name="time" size="16" color="#fff"></l-icon>
						</view>
						<view class="sync-setting-row__text">
							<text class="sync-setting-row__label">同步时间段</text>
							<text class="sync-setting-row__desc">设置允许自动同步的时间范围</text>
						</view>
					</view>
					<view class="sync-setting-row__arrow" @tap="showTimeRangePicker">
						<text class="sync-setting-row__arrow-text">{{ formatTimeRange(syncSettings.timeRange) }}</text>
						<l-icon name="chevron-right" size="14" color="var(--text-tertiary)"></l-icon>
					</view>
				</view>

			<!-- 自动重试 -->
			<view class="sync-setting-row">
				<view class="sync-setting-row__left">
					<view class="sync-setting-row__icon sync-setting-row__icon--green">
						<l-icon name="refresh" size="16" color="#fff"></l-icon>
					</view>
					<view class="sync-setting-row__text">
						<text class="sync-setting-row__label">自动重试</text>
						<text class="sync-setting-row__desc">{{ getAutoRetryDescriptionValue() }}</text>
					</view>
				</view>
				<view class="sync-setting-row__switch">
					<switch :checked="syncSettings.autoRetryEnabled" @change="onAutoRetryChange" color="var(--primary-500)" />
				</view>
			</view>

			<!-- 个人基础信息缓存 -->
			<view class="sync-setting-row">
				<view class="sync-setting-row__left">
					<view class="sync-setting-row__icon sync-setting-row__icon--red">
						<l-icon name="info-circle" size="16" color="#fff"></l-icon>
					</view>
					<view class="sync-setting-row__text">
						<text class="sync-setting-row__label">个人基础信息缓存</text>
						<text class="sync-setting-row__desc">{{ getPersonalInfoDescription() }}</text>
					</view>
				</view>
				<view class="sync-setting-row__switch">
					<switch :checked="syncSettings.personalInfoSyncEnabled" @change="onPersonalInfoSyncChange" color="var(--primary-500)" />
				</view>
			</view>

			<!-- 缓存状态提示（开启后显示） -->
			<view class="sync-info-banner" v-if="syncSettings.personalInfoSyncEnabled">
				<view class="sync-info-banner__dot" :class="getCacheStatusClass()"></view>
				<text class="sync-info-banner__text">{{ getCacheStatusText() }}</text>
			</view>
		</view>

			<!-- 同步日志卡片 -->
			<view class="sync-card" v-if="syncLogs.length > 0">
				<view class="sync-card__header">
					<view class="sync-card__header-left">
						<text class="sync-card__title">同步日志</text>
						<text class="sync-card__title-sub">SYNC LOGS</text>
					</view>
					<view class="sync-card__action" @tap="loadSyncLogs">
						<l-icon name="refresh" size="14" color="var(--primary-500)"></l-icon>
						<text>刷新</text>
					</view>
				</view>

				<view class="sync-log-list">
					<view class="sync-log-item" v-for="log in syncLogs" :key="log.id">
						<view class="sync-log-item__header">
							<view class="sync-log-item__badge" :class="`sync-log-item__badge--${log.status}`">
								<view class="sync-log-item__badge-dot"></view>
								<text>{{ log.syncType === 'auto' ? '自动' : '手动' }}</text>
							</view>
							<text class="sync-log-item__time">{{ formatTime(log.syncedAt) }}</text>
						</view>
						<text class="sync-log-item__msg">{{ log.message }}</text>
						<view class="sync-log-item__meta" v-if="log.coursesSync > 0 || log.duration > 0">
							<view class="sync-log-item__tag" v-if="log.coursesSync > 0">
								<l-icon name="book-open" size="11" color="var(--text-tertiary)"></l-icon>
								<text>{{ log.coursesSync }} 门课程</text>
							</view>
							<view class="sync-log-item__tag" v-if="log.duration > 0">
								<l-icon name="time" size="11" color="var(--text-tertiary)"></l-icon>
								<text>{{ log.duration }}ms</text>
							</view>
						</view>
						<view class="sync-log-item__error" v-if="log.status === 'failed' && log.errorDetail">
							<view class="sync-log-item__error-header">
								<l-icon name="error-circle" size="13" color="#ef4444"></l-icon>
								<text>错误详情</text>
							</view>
							<text class="sync-log-item__error-text">{{ log.errorDetail }}</text>
						</view>
					</view>
				</view>
			</view>

			<!-- 说明信息卡片 -->
			<view class="sync-info-card">
				<view class="sync-info-card__header">
					<view class="sync-info-card__icon">
						<l-icon name="info-circle" size="14" color="var(--primary-500)"></l-icon>
					</view>
					<text class="sync-info-card__title">同步说明</text>
				</view>
				<view class="sync-info-card__list">
					<view class="sync-info-card__item">
						<view class="sync-info-card__bullet"></view>
						<text class="sync-info-card__item-text">自动将学校教务系统的课程表和成绩数据同步到本地，数据加密存储，保护个人隐私</text>
					</view>
					<view class="sync-info-card__item">
						<view class="sync-info-card__bullet"></view>
						<text class="sync-info-card__item-text">仅在您设定的时间段内自动同步，避免在学习或休息时段打扰正常使用</text>
					</view>
					<view class="sync-info-card__item">
						<view class="sync-info-card__bullet"></view>
						<text class="sync-info-card__item-text">同步失败时支持自动重试，默认最多重试3次，确保数据可靠同步</text>
					</view>
					<view class="sync-info-card__item">
						<view class="sync-info-card__bullet"></view>
						<text class="sync-info-card__item-text">建议开启同步功能，随时保持课程数据与学校教务系统一致</text>
					</view>
				</view>
			</view>
		</view>
	</view>
</template>

<script>
import { ref } from 'vue';
import { useSyncSettings } from '@/composables/useSyncSettings.js';
import CustomNavBar from '@/components/CustomNavBar.vue';
import lIcon from '@/uni_modules/lime-icon/components/l-icon/l-icon.vue';
import {
	SYNC_STATUS_CONFIG,
	formatRelativeTime,
	formatNextSyncTime,
	formatTimeRange
} from '@/utils/syncSettings.js';

export default {
	components: {
		CustomNavBar,
		lIcon
	},
	setup() {
		const navPaddingTop = ref('0px');

		const {
			syncSettings,
			syncStatus,
			syncLogs,
			loading,
			loadSyncSettings,
			loadSyncLogs,
			saveSyncSettings,
			manualSyncAction,
			onSyncEnabledChange,
			onAutoRetryChange,
			onPersonalInfoSyncChange,
			showFrequencyPicker,
			showTimeRangePicker,
			showPresetTimes,
			getFrequencyTextValue,
			getFrequencyDescriptionValue,
			getAutoRetryDescriptionValue,
			goBack,
			bindLifecycle
		} = useSyncSettings();

		bindLifecycle();

		const statusBarHeight = ref(uni.getSystemInfoSync().statusBarHeight || 20);

		const handleNavHeightReady = (navInfo) => {
			navPaddingTop.value = navInfo.heightPx;
		};

		const getStatusIcon = () => {
			const statusMap = {
				idle: 'cloud',
				syncing: 'refresh',
				success: 'check-circle',
				failed: 'error-circle'
			};
			return statusMap[syncStatus.value.syncStatus] || 'help-circle';
		};

		const getStatusColor = () => {
			return '#fff';
		};

		const getStatusTitle = () => {
			if (!syncSettings.value.enabled) return '同步未开启';
			return SYNC_STATUS_CONFIG[syncStatus.value.syncStatus]?.title || '未知状态';
		};

		const getStatusDescription = () => {
			if (!syncSettings.value.enabled) return '点击上方开关开启自动同步';
			return syncStatus.value.lastSyncMessage || '暂无同步记录';
		};

		const getPersonalInfoDescription = () => {
			if (!syncSettings.value.personalInfoSyncEnabled) return '关闭后不会定时获取个人信息';
			const status = syncSettings.value.personalInfoCacheStatus;
			if (status === 'paused') return '已暂停（超过2天未使用）';
			if (status === 'resumed') return '已恢复（下次登录时生效）';
			return '定时从学校服务器缓存个人基础信息';
		};

		const getCacheStatusClass = () => {
			const status = syncSettings.value.personalInfoCacheStatus;
			if (status === 'paused') return 'dot--paused';
			if (status === 'resumed') return 'dot--resumed';
			return 'dot--active';
		};

		const getCacheStatusText = () => {
			const status = syncSettings.value.personalInfoCacheStatus;
			if (status === 'paused') return '已暂停：超过 2 天未使用 App，缓存将在下次登录后自动恢复';
			if (status === 'resumed') return '已恢复：下次打开 App 时将重新开始缓存';
			return '缓存中：每 6 小时自动刷新，超 2 天不活跃自动暂停';
		};

		const formatTime = (timeStr) => formatRelativeTime(timeStr);

		return {
			navPaddingTop,
			statusBarHeight,
			syncSettings,
			syncStatus,
			syncLogs,
			loading,
			loadSyncSettings,
			loadSyncLogs,
			saveSyncSettings,
			manualSyncAction,
			onSyncEnabledChange,
			onAutoRetryChange,
			onPersonalInfoSyncChange,
			showFrequencyPicker,
			showTimeRangePicker,
			showPresetTimes,
			getFrequencyTextValue,
			getFrequencyDescriptionValue,
			getAutoRetryDescriptionValue,
			getPersonalInfoDescription,
			getCacheStatusClass,
			getCacheStatusText,
			goBack,
			handleNavHeightReady,
			getStatusIcon,
			getStatusColor,
			getStatusTitle,
			getStatusDescription,
			formatTime,
			formatNextSyncTime,
			formatTimeRange
		};
	}
};
</script>

<style lang="scss" scoped>
/* Hero 渐变背景（与首页一致） */
.hero-section {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 340rpx;
}

.hero-bg {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: linear-gradient(180deg,
		#1e3a8a 0%,
		#1e40af 25%,
		#2563eb 55%,
		#3b82f6 78%,
		#93c5fd 100%);
	z-index: 0;
}

.hero-bg-overlay {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.7) 0%,
		rgba(37, 99, 235, 0.45) 50%,
		rgba(147, 197, 253, 0.15) 100%);
	z-index: 1;
}

.hero-statusbar {
	width: 100%;
	flex-shrink: 0;
}

.hero-nav {
	position: relative;
	z-index: 10;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.sync-nav__back {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.2);
	border-radius: 50%;
	backdrop-filter: blur(8px);
	-webkit-backdrop-filter: blur(8px);
	border: 1rpx solid rgba(255, 255, 255, 0.25);
	flex-shrink: 0;
}

.sync-nav__back:active {
	background: rgba(255, 255, 255, 0.3);
	transform: scale(0.92);
}

.sync-nav__title {
	font-size: 34rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.sync-nav__placeholder {
	width: 64rpx;
	flex-shrink: 0;
}

.hero-content {
	position: relative;
	z-index: 10;
	padding: 8rpx 32rpx 0;
}

.hero-greeting {
	margin-bottom: 0;
}

.hero-title {
	display: block;
	font-size: 44rpx;
	font-weight: 800;
	color: #fff;
	line-height: 1.3;
}

.hero-subtitle {
	display: block;
	font-size: 26rpx;
	color: rgba(255, 255, 255, 0.75);
	margin-top: 6rpx;
	font-weight: 400;
}

/* 页面内容（与首页 content-area 一致） */
.sync-content {
	position: relative;
	z-index: 1;
	padding: 0 24rpx;
	padding-bottom: calc(48rpx + env(safe-area-inset-bottom));
	margin-top: -40rpx;
}

/* 同步状态卡片（蓝底渐变风格） */
.sync-status-card {
	background: linear-gradient(135deg, #1e40af, #2563eb, #3b82f6);
	border-radius: 28rpx;
	padding: 32rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 8rpx 32rpx rgba(37, 99, 235, 0.3);
	position: relative;
	overflow: hidden;

	&::before {
		content: '';
		position: absolute;
		top: -40rpx;
		right: -40rpx;
		width: 200rpx;
		height: 200rpx;
		border-radius: 50%;
		background: radial-gradient(circle, rgba(255, 255, 255, 0.08) 0%, transparent 70%);
		pointer-events: none;
	}
}

.sync-status-card__header {
	display: flex;
	align-items: center;
	gap: 24rpx;
	margin-bottom: 24rpx;
	position: relative;
	z-index: 1;
}

.sync-status-card__icon-wrap {
	position: relative;
	flex-shrink: 0;
}

.sync-status-card__icon-bg {
	position: absolute;
	inset: -8rpx;
	border-radius: 24rpx;
	background: rgba(255, 255, 255, 0.1);
}

.sync-status-card__icon {
	width: 80rpx;
	height: 80rpx;
	border-radius: 20rpx;
	background: rgba(255, 255, 255, 0.2);
	border: 1rpx solid rgba(255, 255, 255, 0.25);
	display: flex;
	align-items: center;
	justify-content: center;
	position: relative;
	z-index: 1;
	backdrop-filter: blur(8px);
	-webkit-backdrop-filter: blur(8px);

	&--spin {
		animation: iconSpin 1.2s linear infinite;
	}
}

@keyframes iconSpin {
	from { transform: rotate(0deg); }
	to { transform: rotate(360deg); }
}

.sync-status-card__info {
	flex: 1;
	min-width: 0;
	position: relative;
	z-index: 1;
}

.sync-status-card__title {
	display: block;
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	margin-bottom: 4rpx;
}

.sync-status-card__desc {
	display: block;
	font-size: 26rpx;
	color: rgba(255, 255, 255, 0.75);
	line-height: 1.4;
}

.sync-status-card__refresh {
	width: 64rpx;
	height: 64rpx;
	border-radius: 16rpx;
	background: rgba(255, 255, 255, 0.2);
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
	backdrop-filter: blur(8px);
	-webkit-backdrop-filter: blur(8px);
	border: 1rpx solid rgba(255, 255, 255, 0.2);
	position: relative;
	z-index: 1;
	transition: all 0.15s ease;

	&:active {
		background: rgba(255, 255, 255, 0.3);
		transform: scale(0.92);
	}
}

.sync-status-card__stats {
	display: flex;
	align-items: center;
	gap: 0;
	background: rgba(255, 255, 255, 0.12);
	border-radius: 16rpx;
	padding: 20rpx;
	backdrop-filter: blur(12px);
	-webkit-backdrop-filter: blur(12px);
	border: 1rpx solid rgba(255, 255, 255, 0.15);
	position: relative;
	z-index: 1;
}

.sync-status-card__stat {
	flex: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 4rpx;
}

.sync-status-card__stat-value {
	font-size: 28rpx;
	font-weight: 700;
	color: #fff;
}

.sync-status-card__stat-label {
	font-size: 22rpx;
	color: rgba(255, 255, 255, 0.6);
}

.sync-status-card__stat-divider {
	width: 1rpx;
	height: 48rpx;
	background: rgba(255, 255, 255, 0.2);
	flex-shrink: 0;
}

/* 白卡片通用样式（与首页 section-card 一致） */
.sync-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 32rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1rpx solid rgba(148, 163, 184, 0.12);
	overflow: hidden;
}

.sync-card__header {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	margin-bottom: 28rpx;
}

.sync-card__header-left {
	flex: 1;
}

.sync-card__title {
	display: block;
	font-size: 32rpx;
	font-weight: 800;
	color: var(--text-primary);
	line-height: 1.2;
}

.sync-card__title-sub {
	display: block;
	font-size: 20rpx;
	font-weight: 600;
	color: var(--text-tertiary);
	letter-spacing: 1px;
	margin-top: 4rpx;
}

.sync-card__action {
	font-size: 26rpx;
	color: var(--primary-500);
	font-weight: 600;
	flex-shrink: 0;
	align-self: center;
	display: flex;
	align-items: center;
	gap: 4rpx;
}

/* 设置行 */
.sync-setting-row {
	display: flex;
	align-items: center;
	padding: 28rpx 0;
	border-bottom: 1rpx solid rgba(226, 232, 240, 0.7);
	transition: background 0.15s ease;

	&:first-child { padding-top: 4rpx; }
	&:last-child { border-bottom: none; padding-bottom: 4rpx; }

	&:active {
		background: rgba(99, 102, 241, 0.04);
		margin: 0 -32rpx;
		padding-left: 32rpx;
		padding-right: 32rpx;
		border-radius: 12rpx;
	}
}

.sync-setting-row__left {
	flex: 1;
	display: flex;
	align-items: center;
	gap: 20rpx;
	min-width: 0;
}

.sync-setting-row__icon {
	width: 64rpx;
	height: 64rpx;
	border-radius: 16rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;

	&--purple { background: linear-gradient(135deg, #6366f1, #8b5cf6); box-shadow: 0 4rpx 12rpx rgba(99, 102, 241, 0.3); }
	&--blue { background: linear-gradient(135deg, #3b82f6, #60a5fa); box-shadow: 0 4rpx 12rpx rgba(59, 130, 246, 0.3); }
	&--amber { background: linear-gradient(135deg, #f59e0b, #fbbf24); box-shadow: 0 4rpx 12rpx rgba(245, 158, 11, 0.3); }
	&--green { background: linear-gradient(135deg, #10b981, #34d399); box-shadow: 0 4rpx 12rpx rgba(16, 185, 129, 0.3); }
	&--red { background: linear-gradient(135deg, #ef4444, #f87171); box-shadow: 0 4rpx 12rpx rgba(239, 68, 68, 0.3); }
}

.sync-setting-row__text {
	flex: 1;
	min-width: 0;
}

.sync-setting-row__label {
	display: block;
	font-size: 30rpx;
	font-weight: 600;
	color: var(--text-primary);
	margin-bottom: 4rpx;
}

.sync-setting-row__desc {
	display: block;
	font-size: 24rpx;
	color: var(--text-tertiary);
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.sync-setting-row__switch {
	flex-shrink: 0;
	margin-left: 16rpx;
	switch { transform: scale(0.85); }
}

.sync-setting-row__arrow {
	display: flex;
	align-items: center;
	gap: 6rpx;
	flex-shrink: 0;
	margin-left: 16rpx;
	padding: 8rpx 16rpx;
	background: var(--bg-secondary);
	border-radius: 12rpx;
	border: 1rpx solid rgba(226, 232, 240, 0.5);
	transition: background 0.15s ease;

	&:active { background: var(--bg-muted); }
}

.sync-setting-row__arrow-text {
	font-size: 26rpx;
	color: var(--primary-500);
	font-weight: 700;
}

/* 日志列表 */
.sync-log-list {
	padding: 0;
}

.sync-log-item {
	padding: 24rpx 0;
	border-bottom: 1rpx solid rgba(226, 232, 240, 0.7);

	&:last-child { border-bottom: none; }
}

.sync-log-item__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 10rpx;
}

.sync-log-item__badge {
	display: flex;
	align-items: center;
	gap: 8rpx;
	font-size: 22rpx;
	font-weight: 600;
	padding: 4rpx 14rpx;
	border-radius: 100rpx;
	background: rgba(100, 116, 139, 0.1);
	color: var(--text-secondary);
}

.sync-log-item__badge-dot {
	width: 6rpx;
	height: 6rpx;
	border-radius: 50%;
	background: var(--text-tertiary);
}

.sync-log-item__badge--success {
	background: rgba(16, 185, 129, 0.1);
	color: #059669;
	.sync-log-item__badge-dot { background: #10b981; }
}

.sync-log-item__badge--failed {
	background: rgba(239, 68, 68, 0.1);
	color: #dc2626;
	.sync-log-item__badge-dot { background: #ef4444; }
}

.sync-log-item__badge--syncing {
	background: rgba(99, 102, 241, 0.1);
	color: var(--primary-600);
	.sync-log-item__badge-dot {
		background: var(--primary-500);
		animation: dotPulse 1.5s ease-in-out infinite;
	}
}

@keyframes dotPulse {
	0%, 100% { opacity: 1; }
	50% { opacity: 0.3; }
}

.sync-log-item__time {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.sync-log-item__msg {
	display: block;
	font-size: 28rpx;
	color: var(--text-primary);
	line-height: 1.5;
	margin-bottom: 10rpx;
}

.sync-log-item__meta {
	display: flex;
	gap: 16rpx;
}

.sync-log-item__tag {
	display: flex;
	align-items: center;
	gap: 6rpx;
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.sync-log-item__error {
	margin-top: 12rpx;
	padding: 16rpx;
	background: rgba(239, 68, 68, 0.05);
	border-radius: 12rpx;
	border-left: 4rpx solid #ef4444;
}

.sync-log-item__error-header {
	display: flex;
	align-items: center;
	gap: 8rpx;
	font-size: 24rpx;
	font-weight: 600;
	color: #ef4444;
	margin-bottom: 8rpx;
}

.sync-log-item__error-text {
	font-size: 24rpx;
	color: var(--text-secondary);
	line-height: 1.5;
	word-break: break-all;
}

/* 说明信息 */
.sync-info-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 32rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1rpx solid rgba(148, 163, 184, 0.12);
}

.sync-info-card__header {
	display: flex;
	align-items: center;
	gap: 16rpx;
	margin-bottom: 24rpx;
}

.sync-info-card__icon {
	width: 52rpx;
	height: 52rpx;
	border-radius: 14rpx;
	background: linear-gradient(135deg, rgba(99, 102, 241, 0.12), rgba(99, 102, 241, 0.04));
	border: 1rpx solid rgba(99, 102, 241, 0.12);
	display: flex;
	align-items: center;
	justify-content: center;
}

.sync-info-card__title {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.sync-info-card__list {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
}

.sync-info-card__item {
	display: flex;
	align-items: flex-start;
	gap: 16rpx;

	.sync-info-card__item-text {
		font-size: 26rpx;
		color: var(--text-secondary);
		line-height: 1.5;
	}
}

.sync-info-banner {
	display: flex;
	align-items: center;
	gap: 12rpx;
	padding: 16rpx 24rpx;
	margin: 0 -32rpx;
	margin-top: -12rpx;
	margin-bottom: 8rpx;
	background: rgba(99, 102, 241, 0.06);
	border-top: 1rpx solid rgba(99, 102, 241, 0.1);
}

.sync-info-banner__dot {
	width: 10rpx;
	height: 10rpx;
	border-radius: 50%;
	flex-shrink: 0;
	background: #10b981;

	&.dot--paused { background: #f59e0b; }
	&.dot--resumed { background: #3b82f6; }
	&.dot--inactive { background: #9ca3af; }
}

.sync-info-banner__text {
	font-size: 24rpx;
	color: var(--text-secondary);
	line-height: 1.4;
}
</style>
