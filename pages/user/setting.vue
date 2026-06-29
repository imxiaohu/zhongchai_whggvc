<template>
	<view class="setting-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="setting-hero">
			<view class="setting-hero-bg"></view>
			<view class="setting-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="setting-hero-nav">
				<view class="setting-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="setting-hero-title">设置</text>
				<view style="width: 64rpx;"></view>
			</view>

			<view class="setting-hero-content">
				<text class="setting-hero-sub">SETTINGS</text>
			</view>
		</view>

		<!-- 内容区域 -->
		<scroll-view class="setting-scroll" scroll-y>
			<view class="setting-content">
				<!-- 用户信息卡片 -->
				<view class="setting-user-card" @tap="showUserDetail">
					<view class="setting-user-avatar">
						<text class="setting-user-avatar-text">{{ getUserInitialMixin() }}</text>
					</view>
					<view class="setting-user-info">
						<text class="setting-user-name">{{ username || '未登录' }}</text>
						<text class="setting-user-id">{{ studentId || '未绑定学号' }}</text>
					</view>
					<l-icon name="chevron-right" style="font-size: 20px; color: var(--text-tertiary);"></l-icon>
				</view>

				<!-- 账号与同步 -->
				<view class="setting-section">
					<view class="setting-section-title">账号与同步</view>
					<view class="setting-group">
						<view class="setting-item" @tap="navigateToBindAccount">
							<view class="setting-item-icon setting-item-icon--blue">
								<l-icon name="user-setting" style="font-size: 20px; color: #fff;"></l-icon>
							</view>
							<text class="setting-item-text">学校账号绑定</text>
							<text class="setting-item-value">{{ isBound ? '已绑定' : '未绑定' }}</text>
							<l-icon name="chevron-right" style="font-size: 18px; color: var(--text-tertiary);"></l-icon>
						</view>
						<view class="setting-item" @tap="showSyncSettings">
							<view class="setting-item-icon setting-item-icon--green">
								<l-icon name="refresh" style="font-size: 20px; color: #fff;"></l-icon>
							</view>
							<text class="setting-item-text">服务器同步</text>
							<text class="setting-item-value">{{ getSyncStatusTextMixin() }}</text>
							<l-icon name="chevron-right" style="font-size: 18px; color: var(--text-tertiary);"></l-icon>
						</view>
					</view>
				</view>

				<!-- 通知 -->
				<view class="setting-section">
					<view class="setting-section-title">通知</view>
					<view class="setting-group">
						<view class="setting-item" @tap="navigateToNotificationSettings">
							<view class="setting-item-icon setting-item-icon--orange">
								<l-icon name="notification" style="font-size: 20px; color: #fff;"></l-icon>
							</view>
							<text class="setting-item-text">通知设置</text>
							<l-icon name="chevron-right" style="font-size: 18px; color: var(--text-tertiary);"></l-icon>
						</view>
					</view>
				</view>

			<!-- 首页展示 -->
			<view class="setting-section">
				<view class="setting-section-title">首页展示</view>
				<view class="setting-group">
					<view class="setting-item" @tap="toggleClassLayout">
						<view class="setting-item-icon setting-item-icon--blue">
							<l-icon name="view-column" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<text class="setting-item-text">今日课程布局</text>
						<text class="setting-item-value">{{ classLayoutText }}</text>
						<l-icon name="chevron-right" style="font-size: 18px; color: var(--text-tertiary);"></l-icon>
					</view>
					<view class="setting-item">
						<view class="setting-item-icon setting-item-icon--green">
							<l-icon name="clock" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<view class="setting-item-text-wrap">
							<text class="setting-item-text">仅显示待上课</text>
							<text class="setting-item-desc">开启后，当日已过课程不再显示</text>
						</view>
						<switch
							:checked="hidePastClasses"
							@change="toggleHidePastClasses($event.detail.value)"
							color="#10b981"
							style="transform: scale(0.75);"
						/>
					</view>
				</view>
			</view>

			<!-- 通用 -->
			<view class="setting-section">
				<view class="setting-section-title">通用</view>
				<view class="setting-group">
					<view class="setting-item" @tap="clearCache">
						<view class="setting-item-icon setting-item-icon--red">
							<l-icon name="delete" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<text class="setting-item-text">清空缓存</text>
						<text class="setting-item-value">{{ cacheSize }} MB</text>
						<l-icon name="chevron-right" style="font-size: 18px; color: var(--text-tertiary);"></l-icon>
					</view>
						<view class="setting-item" @tap="checkUpdate">
							<view class="setting-item-icon setting-item-icon--purple">
								<l-icon name="cloud-download" style="font-size: 20px; color: #fff;"></l-icon>
							</view>
							<text class="setting-item-text">检查更新</text>
							<text class="setting-item-value">v{{ appVersion }}</text>
							<view class="setting-update-dot" v-if="hasUpdate"></view>
							<l-icon name="chevron-right" style="font-size: 18px; color: var(--text-tertiary);"></l-icon>
						</view>
						<view class="setting-item" @tap="showAbout">
							<view class="setting-item-icon setting-item-icon--yellow">
								<l-icon name="info-circle" style="font-size: 20px; color: #fff;"></l-icon>
							</view>
							<text class="setting-item-text">关于我们</text>
							<l-icon name="chevron-right" style="font-size: 18px; color: var(--text-tertiary);"></l-icon>
						</view>
					</view>
				</view>

				<!-- 退出登录 -->
				<view class="setting-logout-section">
					<view class="setting-logout-btn" @tap="logout">
						<text class="setting-logout-btn-text">退出当前账号</text>
					</view>
				</view>

				<!-- 版权信息 -->
				<view class="setting-copyright">
					<text class="setting-copyright-name">众柴智慧校园 v{{ appVersion }}</text>
					<text class="setting-copyright-copy">2026 xiaohu</text>
				</view>
			</view>
		</scroll-view>
	</view>
</template>

<script>
import { ref } from 'vue';
import { onLoad as onUniLoad, onShow as onUniShow, onHide as onUniHide, onUnload as onUniUnload } from '@dcloudio/uni-app';
import { useUserSettings } from '@/composables/useUserSettings.js';

const CLASS_LAYOUT_KEY = 'todayClassLayout';
const HIDE_PAST_KEY = 'hidePastClasses';

export default {
	setup() {
		const statusBarHeight = ref(20);
		const classLayout = ref('vertical');
		const hidePastClasses = ref(false);

		const {
			username,
			studentId,
			isBound,
			userTags,
			userDetail,
			appVersion,
			cacheSize,
			hasUpdate,
			updateInfo,
			updateProgress,
			isUpdating,
			syncSettings,
			syncStatus,
			loadUserInfo,
			calculateCacheSizeMixin,
			getAppVersion,
			checkHasUpdate,
			clearCache,
			checkUpdate,
			showAbout,
			logout,
			loadSyncSettings,
			getSyncStatusTextMixin,
			showSyncSettings,
			loadUserSettings,
			showUserDetail,
			navigateToBindAccount,
			showBindingInfo,
			confirmRebind,
			confirmUnbind,
			unbindSchoolAccount,
			getUserInitialMixin
		} = useUserSettings();

		const classLayoutText = ref('竖向');

		function initStatusBar() {
			try {
				const systemInfo = uni.getSystemInfoSync();
				statusBarHeight.value = systemInfo.statusBarHeight || 20;
			} catch (e) {
				statusBarHeight.value = 20;
			}
		}

		function loadDisplaySettings() {
			const layout = uni.getStorageSync(CLASS_LAYOUT_KEY);
			classLayout.value = layout || 'vertical';
			classLayoutText.value = classLayout.value === 'horizontal' ? '横向' : '竖向';

			const hidePast = uni.getStorageSync(HIDE_PAST_KEY);
			hidePastClasses.value = hidePast === 'true';
		}

		function toggleClassLayout() {
			const layouts = ['vertical', 'horizontal'];
			const texts = { vertical: '竖向', horizontal: '横向' };
			const currentIndex = layouts.indexOf(classLayout.value);
			const nextIndex = (currentIndex + 1) % layouts.length;
			classLayout.value = layouts[nextIndex];
			classLayoutText.value = texts[classLayout.value];
			uni.setStorageSync(CLASS_LAYOUT_KEY, classLayout.value);
			uni.showToast({ title: `已切换为${texts[classLayout.value]}布局`, icon: 'none', duration: 1500 });
		}

		function toggleHidePastClasses(value) {
			hidePastClasses.value = value;
			uni.setStorageSync(HIDE_PAST_KEY, String(value));
			uni.showToast({
				title: value ? '已开启，仅显示待上课' : '已关闭，显示全部课程',
				icon: 'none',
				duration: 1500
			});
		}

		function markFromSettings() {
			uni.setStorageSync('fromSettingsPage', 'true');
			uni.setStorageSync('fromSettingsPageTime', Date.now().toString());
		}

		function goBack() {
			uni.navigateBack();
		}

		function navigateToNotificationSettings() {
			uni.navigateTo({
				url: '/pages/user/notification-settings'
			});
		}

		onUniLoad(() => {
			console.log('设置页面加载');
			loadUserInfo();
			calculateCacheSizeMixin();
			getAppVersion();
			checkHasUpdate();
			loadSyncSettings();
			loadUserSettings();
			initStatusBar();
			loadDisplaySettings();
		});

		onUniShow(() => {
			loadUserInfo();
			loadSyncSettings();
		});

		onUniHide(() => {
			console.log('设置页面隐藏');
			markFromSettings();
		});

		onUniUnload(() => {
			console.log('设置页面卸载');
			markFromSettings();
		});

		return {
			statusBarHeight,
			classLayout,
			classLayoutText,
			hidePastClasses,
			username,
			studentId,
			isBound,
			userTags,
			userDetail,
			appVersion,
			cacheSize,
			hasUpdate,
			updateInfo,
			updateProgress,
			isUpdating,
			syncSettings,
			syncStatus,
			loadUserInfo,
			calculateCacheSizeMixin,
			getAppVersion,
			checkHasUpdate,
			clearCache,
			checkUpdate,
			showAbout,
			logout,
			loadSyncSettings,
			getSyncStatusTextMixin,
			showSyncSettings,
			loadUserSettings,
			showUserDetail,
			navigateToBindAccount,
			navigateToNotificationSettings,
			showBindingInfo,
			confirmRebind,
			confirmUnbind,
			unbindSchoolAccount,
			getUserInitialMixin,
			goBack,
			loadDisplaySettings,
			toggleClassLayout,
			toggleHidePastClasses
		};
	}
};
</script>

<style lang="scss" scoped>
/* ============================================
   Settings Page - Hero Style
   ============================================ */

.setting-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
}

/* ---- Hero Section ---- */
.setting-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 200rpx;
}

.setting-hero-bg {
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

.setting-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.65) 0%,
		rgba(37, 99, 235, 0.4) 50%,
		rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.setting-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.setting-back-btn {
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

.setting-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.setting-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 8rpx;
}

.setting-hero-sub {
	font-size: 20rpx;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.6);
	letter-spacing: 2px;
}

/* ---- Scroll ---- */
.setting-scroll {
	flex: 1;
	min-height: 0;
}

.setting-content {
	padding: 24rpx;
}

/* ---- User Card ---- */
.setting-user-card {
	display: flex;
	align-items: center;
	background: #fff;
	border-radius: 28rpx;
	padding: 32rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1px solid rgba(148, 163, 184, 0.12);

	&:active { opacity: 0.8; }
}

.setting-user-avatar {
	width: 100rpx;
	height: 100rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, var(--primary-500), var(--primary-700));
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
	box-shadow: 0 4rpx 16rpx rgba(59, 130, 246, 0.3);

	.setting-user-avatar-text {
		font-size: 40rpx;
		font-weight: 700;
		color: #fff;
	}
}

.setting-user-info {
	flex: 1;
	margin-left: 24rpx;
}

.setting-user-name {
	display: block;
	font-size: 34rpx;
	font-weight: 800;
	color: var(--text-primary);
	margin-bottom: 8rpx;
}

.setting-user-id {
	display: block;
	font-size: 24rpx;
	color: var(--text-tertiary);
}

/* ---- Section ---- */
.setting-section {
	margin-bottom: 24rpx;
}

.setting-section-title {
	font-size: 22rpx;
	font-weight: 700;
	color: var(--text-tertiary);
	text-transform: uppercase;
	letter-spacing: 1px;
	margin-bottom: 12rpx;
	margin-left: 8rpx;
}

/* ---- Group ---- */
.setting-group {
	background: #fff;
	border-radius: 28rpx;
	overflow: hidden;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.12);
}

/* ---- Item ---- */
.setting-item {
	display: flex;
	align-items: center;
	padding: 28rpx 24rpx;
	position: relative;
	transition: background 0.15s ease;

	&:not(:last-child)::after {
		content: '';
		position: absolute;
		bottom: 0;
		left: 104rpx;
		right: 24rpx;
		height: 1px;
		background: rgba(226, 232, 240, 0.6);
	}

	&:active { background: var(--bg-muted); }
}

.setting-item-icon {
	width: 72rpx;
	height: 72rpx;
	border-radius: 20rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;

	&--blue { background: linear-gradient(135deg, #6366f1, #818cf8); }
	&--green { background: linear-gradient(135deg, #10b981, #34d399); }
	&--red { background: linear-gradient(135deg, #f43f5e, #fb7185); }
	&--purple { background: linear-gradient(135deg, #8b5cf6, #a78bfa); }
	&--yellow { background: linear-gradient(135deg, #f59e0b, #fbbf24); }
	&--orange { background: linear-gradient(135deg, #f97316, #fb923c); }
}

.setting-item-text {
	flex: 1;
	margin-left: 20rpx;
	font-size: 30rpx;
	font-weight: 600;
	color: var(--text-primary);
}

.setting-item-value {
	font-size: 26rpx;
	color: var(--text-tertiary);
	margin-right: 12rpx;
}

.setting-item-desc {
	font-size: 22rpx;
	color: var(--text-tertiary);
	margin-top: 4rpx;
}

.setting-item-text-wrap {
	display: flex;
	flex-direction: column;
	flex: 1;
	margin-left: 20rpx;
}

.setting-update-dot {
	width: 16rpx;
	height: 16rpx;
	background: var(--error-color);
	border-radius: 50%;
	margin-right: 8rpx;
}

/* ---- Logout ---- */
.setting-logout-section {
	margin-top: 32rpx;
}

.setting-logout-btn {
	height: 96rpx;
	background: #fff;
	border-radius: 28rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	border: 1px solid rgba(148, 163, 184, 0.12);
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);

	.setting-logout-btn-text {
		font-size: 30rpx;
		font-weight: 700;
		color: var(--error-color);
	}

	&:active {
		background: rgba(244, 63, 94, 0.05);
		transform: scale(0.98);
	}
}

/* ---- Copyright ---- */
.setting-copyright {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 4rpx;
	margin-top: 48rpx;
	margin-bottom: 40rpx;
}

.setting-copyright-name {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.setting-copyright-copy {
	font-size: 20rpx;
	color: var(--text-tertiary);
}
</style>
