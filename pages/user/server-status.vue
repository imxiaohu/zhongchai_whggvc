<template>
	<view class="ss-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="ss-hero">
			<view class="ss-hero-bg"></view>
			<view class="ss-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="ss-hero-nav">
				<view class="ss-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="ss-hero-title">服务器状态</text>
				<view class="ss-refresh-btn" @tap="refreshStatus" :class="{ 'ss-refresh-btn--loading': isRefreshing }">
					<l-icon name="refresh" style="font-size: 20px; color: #fff;"></l-icon>
				</view>
			</view>

			<view class="ss-hero-content">
				<text class="ss-hero-sub">SERVER STATUS</text>
			</view>
		</view>

		<!-- 内容区域 -->
		<scroll-view class="ss-scroll" scroll-y>
			<view class="ss-content">
				<!-- 状态卡片 -->
				<view class="ss-status-card" :class="serverStatus.isAlive ? 'ss-status-card--online' : 'ss-status-card--offline'">
					<view class="ss-status-card-header">
						<view class="ss-status-card-header-left">
							<view class="ss-status-pill" :class="serverStatus.isAlive ? 'ss-status-pill--online' : 'ss-status-pill--offline'">
								<l-icon
									:name="serverStatus.isAlive ? 'check-circle-filled' : 'error-circle-filled'"
									style="font-size: 14px;"
								></l-icon>
								<text>{{ serverStatus.isAlive ? '运行正常' : '连接异常' }}</text>
							</view>
						</view>
					</view>

					<view class="ss-status-grid">
						<view class="ss-status-grid-item">
							<text class="ss-status-grid-label">状态</text>
							<view class="ss-status-grid-row">
								<view class="ss-status-dot" :class="serverStatus.isAlive ? 'ss-status-dot--online' : 'ss-status-dot--offline'"></view>
								<text class="ss-status-grid-value">{{ serverStatus.isAlive ? '在线' : '离线' }}</text>
							</view>
						</view>
						<view class="ss-status-grid-item">
							<text class="ss-status-grid-label">最后检查</text>
							<text class="ss-status-grid-value">{{ formatTime(serverStatus.lastCheck) }}</text>
						</view>
						<view class="ss-status-grid-item">
							<text class="ss-status-grid-label">最后在线</text>
							<text class="ss-status-grid-value">{{ formatTime(serverStatus.lastAlive) }}</text>
						</view>
						<view class="ss-status-grid-item" v-if="serverStatus.responseTime">
							<text class="ss-status-grid-label">响应时间</text>
							<text class="ss-status-grid-value" :class="getResponseTimeClass(serverStatus.responseTime)">
								{{ serverStatus.responseTime }}ms
							</text>
						</view>
						<view class="ss-status-grid-item" v-if="serverStatus.errorCount > 0">
							<text class="ss-status-grid-label">错误次数</text>
							<text class="ss-status-grid-value ss-status-grid-value--error">{{ serverStatus.errorCount }}</text>
						</view>
					</view>

					<view class="ss-error-box" v-if="!serverStatus.isAlive && serverStatus.errorMsg">
						<view class="ss-error-header">
							<l-icon name="info-circle" style="font-size: 18px; color: var(--error-color);"></l-icon>
							<text class="ss-error-title">连接问题</text>
						</view>
						<text class="ss-error-message">{{ serverStatus.errorMsg }}</text>
					</view>
				</view>

				<!-- 服务说明卡片 -->
				<view class="ss-info-card">
					<view class="ss-info-card-header">
						<view class="ss-info-card-icon">
							<l-icon name="info-circle" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<view class="ss-info-card-text">
							<text class="ss-info-card-title">服务说明</text>
							<text class="ss-info-card-sub">离线时将使用本地缓存数据</text>
						</view>
					</view>
					<view class="ss-info-list">
						<view class="ss-info-item">
							<view class="ss-info-bullet"></view>
							<text class="ss-info-text">学校服务器用于获取课程表、成绩等数据</text>
						</view>
						<view class="ss-info-item">
							<view class="ss-info-bullet"></view>
							<text class="ss-info-text">系统每 5 分钟自动检查一次服务器状态</text>
						</view>
						<view class="ss-info-item">
							<view class="ss-info-bullet"></view>
							<text class="ss-info-text">服务器维护期间可能无法获取最新数据</text>
						</view>
					</view>
				</view>

				<!-- 操作按钮 -->
				<view class="ss-actions">
					<view class="ss-action-primary" :class="{ 'ss-action-primary--disabled': isChecking }" @tap="forceCheck">
						<l-icon name="lock-on" style="font-size: 20px; color: #fff; margin-right: 12rpx;"></l-icon>
						<text class="ss-action-btn-text">{{ isChecking ? '检查中...' : '立即检查' }}</text>
					</view>
					<view class="ss-action-secondary" @tap="viewMaintenanceInfo">
						<l-icon name="tools" style="font-size: 20px; color: var(--primary-600); margin-right: 12rpx;"></l-icon>
						<text class="ss-action-btn-text">查看维护信息</text>
					</view>
				</view>

				<!-- 历史记录卡片 -->
				<view class="ss-history-card" v-if="statusHistory.length > 0">
					<view class="ss-info-card-header">
						<view class="ss-info-card-icon ss-info-card-icon--green">
							<l-icon name="history" style="font-size: 20px; color: #fff;"></l-icon>
						</view>
						<view class="ss-info-card-text">
							<text class="ss-info-card-title">最近状态</text>
							<text class="ss-info-card-sub">最近 {{ statusHistory.length }} 次检查</text>
						</view>
					</view>

					<view class="ss-history-list">
						<view class="ss-history-item" v-for="(item, index) in statusHistory" :key="index">
							<view class="ss-history-dot" :class="item.isAlive ? 'ss-history-dot--online' : 'ss-history-dot--offline'"></view>
							<view class="ss-history-main">
								<text class="ss-history-status">{{ item.isAlive ? '在线' : '离线' }}</text>
								<text class="ss-history-time">{{ formatTime(item.time) }}</text>
							</view>
							<text class="ss-history-response" :class="getResponseTimeClass(item.responseTime)" v-if="item.responseTime">
								{{ item.responseTime }}ms
							</text>
						</view>
					</view>
				</view>
			</view>
		</scroll-view>
	</view>
</template>

<script>
import { ref, onMounted } from 'vue';
import CustomNavBar from '../../components/CustomNavBar.vue';
import { onUnload } from '@dcloudio/uni-app';
import { checkSchoolServerStatus, forceCheckSchoolServer, getMaintenanceInfo } from '../../utils/schoolServerStatus.js';

export default {
	components: { CustomNavBar },
	setup() {
		const statusBarHeight = ref(20);
		const navPaddingTop = ref('0px');

		const serverStatus = ref({
			isAlive: true,
			lastCheck: null,
			lastAlive: null,
			responseTime: null,
			errorMsg: null,
			errorCount: 0
		});
		const isRefreshing = ref(false);
		const isChecking = ref(false);
		const statusHistory = ref([]);
		let autoRefreshTimer = null;

		function initStatusBar() {
			try {
				const systemInfo = uni.getSystemInfoSync();
				statusBarHeight.value = systemInfo.statusBarHeight || 20;
			} catch (e) {
				statusBarHeight.value = 20;
			}
		}

		function handleNavHeightReady(navInfo) {
			navPaddingTop.value = navInfo.heightPx;
		}

		function addToHistory(status) {
			statusHistory.value.unshift({
				isAlive: status.isAlive,
				time: status.lastCheck,
				responseTime: status.responseTime
			});
			if (statusHistory.value.length > 10) {
				statusHistory.value = statusHistory.value.slice(0, 10);
			}
		}

		async function loadServerStatus() {
			try {
				const status = await checkSchoolServerStatus();
				serverStatus.value = status;
				addToHistory(status);
			} catch (error) {
				console.error('获取服务器状态失败:', error);
				uni.showToast({ title: '获取状态失败', icon: 'none' });
			}
		}

		async function refreshStatus() {
			if (isRefreshing.value) return;
			isRefreshing.value = true;
			try {
				await loadServerStatus();
			} finally {
				isRefreshing.value = false;
			}
		}

		async function forceCheck() {
			if (isChecking.value) return;
			isChecking.value = true;
			try {
				await forceCheckSchoolServer();
				uni.showToast({ title: '检查完成', icon: 'success' });
				setTimeout(() => { loadServerStatus(); }, 1000);
			} catch (error) {
				console.error('强制检查失败:', error);
				uni.showToast({ title: '检查失败', icon: 'none' });
			} finally {
				isChecking.value = false;
			}
		}

		async function viewMaintenanceInfo() {
			try {
				const maintenanceInfo = await getMaintenanceInfo();
				const content = [
					`服务器状态: ${maintenanceInfo.isServerAlive ? '在线' : '离线'}`,
					`最后检查: ${maintenanceInfo.lastCheck || '未知'}`,
					`最后在线: ${maintenanceInfo.lastAlive || '未知'}`,
					maintenanceInfo.maintenanceMsg || '暂无维护信息'
				].join('\n');
				uni.showModal({ title: '维护信息', content, showCancel: false });
			} catch (error) {
				console.error('获取维护信息失败:', error);
				uni.showToast({ title: '获取信息失败', icon: 'none' });
			}
		}

		function formatTime(timeStr) {
			if (!timeStr) return '未知';
			try {
				const date = new Date(timeStr);
				const now = new Date();
				const diff = now - date;
				if (diff < 60000) return '刚刚';
				if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`;
				if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`;
				return date.toLocaleDateString();
			} catch (e) {
				return timeStr;
			}
		}

		function getResponseTimeClass(responseTime) {
			if (!responseTime) return '';
			if (responseTime < 500) return 'ss-response--good';
			if (responseTime < 1000) return 'ss-response--normal';
			return 'ss-response--slow';
		}

		function goBack() {
			uni.navigateBack();
		}

		onMounted(() => {
			initStatusBar();
			loadServerStatus();
			autoRefreshTimer = setInterval(() => { loadServerStatus(); }, 60000);
		});

		onUnload(() => {
			if (autoRefreshTimer) {
				clearInterval(autoRefreshTimer);
				autoRefreshTimer = null;
			}
		});

		return {
			statusBarHeight,
			navPaddingTop,
			serverStatus,
			isRefreshing,
			isChecking,
			statusHistory,
			refreshStatus,
			forceCheck,
			viewMaintenanceInfo,
			formatTime,
			getResponseTimeClass,
			goBack,
			handleNavHeightReady
		};
	}
};
</script>

<style lang="scss" scoped>
/* ============================================
   Server Status Page - Hero Style
   ============================================ */

.ss-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
}

/* ---- Hero Section ---- */
.ss-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 200rpx;
}

.ss-hero-bg {
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

.ss-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.65) 0%,
		rgba(37, 99, 235, 0.4) 50%,
		rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.ss-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.ss-back-btn {
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

.ss-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.ss-refresh-btn {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.18);
	border-radius: 50%;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.25);

	&--loading {
		animation: ss-spin 1s linear infinite;
	}

	&:active {
		background: rgba(255, 255, 255, 0.28);
		transform: scale(0.9);
	}
}

@keyframes ss-spin {
	to { transform: rotate(360deg); }
}

.ss-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 8rpx;
}

.ss-hero-sub {
	font-size: 20rpx;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.6);
	letter-spacing: 2px;
}

/* ---- Scroll ---- */
.ss-scroll {
	flex: 1;
	min-height: 0;
}

.ss-content {
	padding: 24rpx;
}

/* ---- Status Card ---- */
.ss-status-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 32rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1px solid rgba(148, 163, 184, 0.12);
}

.ss-status-card--online {
	border-top: 6rpx solid var(--success-color);
}

.ss-status-card--offline {
	border-top: 6rpx solid var(--error-color);
}

.ss-status-card-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 28rpx;
}

.ss-status-card-header-left {
	flex: 1;
}

.ss-status-pill {
	display: inline-flex;
	align-items: center;
	gap: 8rpx;
	padding: 10rpx 24rpx;
	border-radius: 100rpx;
	font-size: 26rpx;
	font-weight: 700;

	&--online {
		background: var(--success-soft);
		color: var(--success-color);
	}
	&--offline {
		background: rgba(239, 68, 68, 0.1);
		color: var(--error-color);
	}
}

.ss-status-grid {
	display: grid;
	grid-template-columns: repeat(2, 1fr);
	gap: 16rpx;
}

.ss-status-grid-item {
	background: var(--bg-muted);
	border-radius: 16rpx;
	padding: 20rpx;
}

.ss-status-grid-label {
	display: block;
	font-size: 22rpx;
	color: var(--text-tertiary);
	margin-bottom: 8rpx;
}

.ss-status-grid-row {
	display: flex;
	align-items: center;
	gap: 10rpx;
}

.ss-status-dot {
	width: 16rpx;
	height: 16rpx;
	border-radius: 50%;

	&--online { background: var(--success-color); }
	&--offline { background: var(--error-color); }
}

.ss-status-grid-value {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);

	&--error { color: var(--error-color); }
}

.ss-response--good { color: var(--success-color); }
.ss-response--normal { color: var(--warning-color); }
.ss-response--slow { color: var(--error-color); }

/* ---- Error Box ---- */
.ss-error-box {
	margin-top: 24rpx;
	background: rgba(239, 68, 68, 0.06);
	border: 1px solid rgba(239, 68, 68, 0.18);
	border-radius: 16rpx;
	padding: 20rpx;
}

.ss-error-header {
	display: flex;
	align-items: center;
	gap: 10rpx;
	margin-bottom: 10rpx;
}

.ss-error-title {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--error-color);
}

.ss-error-message {
	font-size: 26rpx;
	color: var(--text-secondary);
	line-height: 1.5;
}

/* ---- Info Card ---- */
.ss-info-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 32rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1px solid rgba(148, 163, 184, 0.12);
}

.ss-info-card-header {
	display: flex;
	align-items: flex-start;
	gap: 20rpx;
	margin-bottom: 24rpx;
}

.ss-info-card-icon {
	width: 80rpx;
	height: 80rpx;
	border-radius: 24rpx;
	background: linear-gradient(135deg, #3b82f6, #2563eb);
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;

	&--green { background: linear-gradient(135deg, #10b981, #059669); }
}

.ss-info-card-text {
	flex: 1;
}

.ss-info-card-title {
	display: block;
	font-size: 32rpx;
	font-weight: 800;
	color: var(--text-primary);
	margin-bottom: 6rpx;
}

.ss-info-card-sub {
	display: block;
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.ss-info-list {
	display: flex;
	flex-direction: column;
	gap: 14rpx;
}

.ss-info-item {
	display: flex;
	align-items: flex-start;
	gap: 12rpx;
}

.ss-info-bullet {
	width: 12rpx;
	height: 12rpx;
	border-radius: 50%;
	background: var(--primary-500);
	margin-top: 10rpx;
	flex-shrink: 0;
}

.ss-info-text {
	font-size: 26rpx;
	color: var(--text-secondary);
	line-height: 1.6;
}

/* ---- Actions ---- */
.ss-actions {
	display: flex;
	gap: 16rpx;
	margin-bottom: 24rpx;
}

.ss-action-primary {
	flex: 1;
	height: 96rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 24rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 8rpx 24rpx rgba(37, 99, 235, 0.3);

	.ss-action-btn-text {
		font-size: 30rpx;
		font-weight: 700;
		color: #fff;
	}

	&--disabled {
		opacity: 0.6;
	}

	&:active:not(.ss-action-primary--disabled) {
		transform: scale(0.98);
	}
}

.ss-action-secondary {
	flex: 1;
	height: 96rpx;
	background: #fff;
	border-radius: 24rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	border: 1px solid rgba(148, 163, 184, 0.12);
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);

	.ss-action-btn-text {
		font-size: 30rpx;
		font-weight: 700;
		color: var(--text-primary);
	}

	&:active {
		background: var(--bg-muted);
	}
}

/* ---- History Card ---- */
.ss-history-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 32rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1px solid rgba(148, 163, 184, 0.12);
}

.ss-history-list {
	display: flex;
	flex-direction: column;
}

.ss-history-item {
	display: flex;
	align-items: center;
	gap: 16rpx;
	padding: 20rpx 0;
	border-bottom: 1px solid rgba(226, 232, 240, 0.6);

	&:last-child { border-bottom: none; }
}

.ss-history-dot {
	width: 16rpx;
	height: 16rpx;
	border-radius: 50%;
	flex-shrink: 0;

	&--online { background: var(--success-color); }
	&--offline { background: var(--error-color); }
}

.ss-history-main {
	flex: 1;
	display: flex;
	flex-direction: column;
	gap: 4rpx;
}

.ss-history-status {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.ss-history-time {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.ss-history-response {
	font-size: 26rpx;
	font-weight: 700;
}
</style>
