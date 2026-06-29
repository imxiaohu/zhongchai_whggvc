<template>
	<view class="evaluation-page">
		<!-- 顶部蓝色渐变 Hero 区域 -->
		<view class="eval-hero">
			<view class="eval-hero-bg"></view>
			<view class="eval-hero-overlay"></view>

			<!-- 状态栏占位 -->
			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<!-- 顶部导航栏 -->
			<view class="eval-hero-nav">
				<view class="eval-hero-nav-left">
					<view class="eval-hero-back" @tap="handleBack">
						<l-icon name="chevron-left" style="font-size: 22px; color: #fff;"></l-icon>
					</view>
					<text class="eval-hero-title">课程评教</text>
				</view>
				<view class="eval-hero-nav-right">
					<view class="eval-hero-icon-btn" @tap="goToSwipeEvaluation">
						<l-icon name="edit" style="font-size: 20px; color: #fff;"></l-icon>
					</view>
					<view class="eval-hero-icon-btn" :class="{'eval-hero-icon-btn--disabled': !hasUnevaluatedItems}" @tap="quickEvaluateAll">
						<l-icon name="star" style="font-size: 20px; color: #fff;"></l-icon>
					</view>
				</view>
			</view>

			<!-- 欢迎语 -->
			<view class="eval-hero-content">
				<text class="eval-hero-subtitle">{{ heroSubtitle }}</text>
			</view>
		</view>

		<!-- 主要内容区域 -->
		<view class="eval-content" :style="{ paddingTop: '0' }">

			<!-- 预览模式提示 -->
			<view v-if="isPreviewMode" class="eval-preview-banner">
				<view class="eval-preview-banner-left">
					<l-icon name="info-circle-filled" style="font-size: 18px; color: var(--warning-color);"></l-icon>
					<text class="eval-preview-banner-text">当前为预览模式，显示示例数据</text>
				</view>
				<text class="eval-preview-banner-btn" @tap="goToLogin">立即登录</text>
			</view>

			<view class="eval-page-content">
				<!-- 统计概览卡片 -->
				<view v-if="!loading && !error && evaluationList.length > 0" class="eval-summary-card">
					<view class="eval-summary-bg-decor1"></view>
					<view class="eval-summary-bg-decor2"></view>
					<view class="eval-summary-item">
						<text class="eval-summary-value">{{ evaluationList.length }}</text>
						<text class="eval-summary-label">全部课程</text>
					</view>
					<view class="eval-summary-divider"></view>
					<view class="eval-summary-item">
						<text class="eval-summary-value" style="color: #86efac;">{{ completedCount }}</text>
						<text class="eval-summary-label">已完成</text>
					</view>
					<view class="eval-summary-divider"></view>
					<view class="eval-summary-item">
						<text class="eval-summary-value" style="color: #fcd34d;">{{ pendingCount }}</text>
						<text class="eval-summary-label">待评教</text>
					</view>
				</view>

				<!-- 加载状态 -->
				<view v-if="loading" class="eval-status-card">
					<view class="eval-status-spinner"></view>
					<text class="eval-status-title">加载中...</text>
				</view>

				<!-- 错误状态 -->
				<view v-else-if="error" class="eval-status-card">
					<view class="eval-status-icon eval-status-icon--error">
						<l-icon name="error-circle" size="40" color="var(--error-color)"></l-icon>
					</view>
					<text class="eval-status-title">{{ error }}</text>
					<button class="eval-retry-btn" @tap="loadEvaluationList">重新加载</button>
				</view>

				<!-- 全部完成状态 -->
				<view v-else-if="evaluationList.length === 0" class="eval-status-card">
					<view class="eval-status-icon eval-status-icon--success">
						<l-icon name="check-circle" size="40" color="var(--success-color)"></l-icon>
					</view>
					<text class="eval-status-title">全部完成</text>
					<text class="eval-status-subtitle">暂无待评课程</text>
				</view>

				<!-- 课程列表 -->
				<view v-else class="eval-list">
					<view
						v-for="(item, index) in evaluationList"
						:key="index"
						class="eval-card"
						:class="{
							'eval-card--evaluated': item.evaluated,
							'eval-card--draft': getCacheStatusMixin(item.id)
						}"
						@tap="goToEvaluate(item)"
					>
						<!-- 左侧状态条 -->
						<view class="eval-card-status-bar"
							:class="{
								'eval-card-status-bar--completed': item.evaluated,
								'eval-card-status-bar--draft': getCacheStatusMixin(item.id),
								'eval-card-status-bar--pending': !item.evaluated && !getCacheStatusMixin(item.id)
							}"
						></view>

						<view class="eval-card-body">
							<!-- Header: 状态标签 + 箭头 -->
							<view class="eval-card-header">
								<view class="eval-card-tag"
									:class="{
										'eval-card-tag--completed': item.evaluated,
										'eval-card-tag--draft': getCacheStatusMixin(item.id),
										'eval-card-tag--pending': !item.evaluated && !getCacheStatusMixin(item.id)
									}">
									<view class="eval-card-tag-dot"
										:class="{
											'eval-card-tag-dot--completed': item.evaluated,
											'eval-card-tag-dot--draft': getCacheStatusMixin(item.id),
											'eval-card-tag-dot--pending': !item.evaluated && !getCacheStatusMixin(item.id)
										}">
									</view>
									<text class="eval-card-tag-text">{{ getCacheStatusMixin(item.id) ? '草稿' : (item.evaluated ? '已完成' : '待评教') }}</text>
								</view>
								<l-icon name="chevron-right" size="18" color="var(--text-tertiary)"></l-icon>
							</view>

							<!-- 课程名称 -->
							<view class="eval-card-title">
								<text>{{ item.name }}</text>
							</view>

							<!-- 底部信息 -->
							<view class="eval-card-footer">
								<view class="eval-card-info">
									<view class="eval-card-row">
										<l-icon name="user-circle-filled" size="14" class="eval-card-icon"></l-icon>
										<text>{{ item.teacherName }}</text>
									</view>
									<view class="eval-card-row">
										<l-icon name="time" size="14" class="eval-card-icon"></l-icon>
										<text>{{ formatTimeRangeMixin(item.startTime, item.endTime) }}</text>
									</view>
								</view>
							</view>
						</view>
					</view>
				</view>
			</view>
		</view>

		<!-- 悬浮删除草稿按钮 -->
		<view class="eval-fab" v-if="hasDraftData">
			<view class="eval-fab-btn eval-fab-btn--danger" @tap="showClearCacheConfirm = true">
				<l-icon name="delete" size="24" color="#fff"></l-icon>
			</view>
		</view>

		<!-- 一键评教浮动按钮 -->
		<l-fav
			v-if="!loading && !error && evaluationList.length > 0"
			:visible="!isPreviewMode && pendingCount > 0"
			:pending-count="pendingCount"
			show-badge
			style="margin-bottom: 80rpx;"
			@click="quickEvaluateAll"
		/>

		<!-- 清空草稿确认弹窗 -->
		<view v-if="showClearCacheConfirm" class="eval-modal-overlay" @tap="cancelClearCache">
			<view class="eval-modal-container" @tap.stop>
				<view class="eval-modal-body">
					<text class="eval-modal-title">清空草稿</text>
					<text class="eval-modal-message">确定要清空所有草稿吗？此操作不可撤销。</text>
				</view>
				<view class="eval-modal-footer">
					<view class="eval-modal-btn" @tap="cancelClearCache">取消</view>
					<view class="eval-modal-btn eval-modal-btn--confirm" @tap="confirmClearCache">确定</view>
				</view>
			</view>
		</view>
	</view>
</template>

<script>
import { ref, computed, onMounted } from 'vue';
import { onPullDownRefresh, onTabItemTap, onShareAppMessage, onShareTimeline } from '@dcloudio/uni-app';
import { useEvaluationList } from '@/composables/useEvaluationList.js';
import { shouldAutoRefreshData } from '@/utils/errorHandler.js';
import CustomNavBar from '@/components/CustomNavBar.vue';
import lFav from '@/components/l-fav.vue';
import shareManager from '@/utils/shareManager.js';
import { onShow } from '@dcloudio/uni-app'
import unreadMessageManager from '@/utils/unreadMessageManager.js';

const CONSTANTS = {
	REFRESH_INTERVAL: 3000,
	AUTO_REFRESH_INTERVAL: 60000
};

export default {
	components: {
		CustomNavBar,
		lFav
	},
	setup() {
		const statusBarHeight = ref(20);
		const navPaddingTop = ref('0px');

		const {
			loading,
			error,
			evaluationList,
			evaluationCache,
			showClearCacheConfirm,
			lastRefreshTime,
			isPreviewMode,
			refreshUnreadMessages,
			showPreviewData,
			goToLogin,
			refreshEvaluationData,
			loadEvaluationList,
			loadAllEvaluationData,
			loadCacheDataMixin,
			getCacheStatusMixin,
			confirmClearCache,
			cancelClearCache,
			goToSwipeEvaluation,
			quickEvaluateAll,
			handleBackClick,
			handleTitleClick,
			handleNavHeightReady,
			formatTimeRangeMixin,
			goToEvaluate
		} = useEvaluationList();

		const clientId = ref('');

		// 计算属性
		const heroSubtitle = computed(() => {
			if (isPreviewMode.value) return '登录后可查看真实评教数据'
			if (pendingCount.value > 0) return `还有 ${pendingCount.value} 门课程待评教`
			if (completedCount.value > 0) return '已完成全部课程评教'
			return '暂无需要评教的课程'
		});

		const hasUnevaluatedItems = computed(() =>
			evaluationList.value.some(item => !item.evaluated)
		);
		const hasDraftData = computed(() =>
			!isPreviewMode.value && Object.keys(evaluationCache.value).length > 0
		);
		const completedCount = computed(() =>
			evaluationList.value.filter(item => item.evaluated).length
		);
		const pendingCount = computed(() =>
			evaluationList.value.filter(item => !item.evaluated).length
		);

		// 初始化状态栏高度
		function initStatusBarHeight() {
			try {
				const systemInfo = uni.getSystemInfoSync();
				statusBarHeight.value = systemInfo.statusBarHeight || 20;
			} catch (e) {
				statusBarHeight.value = 20;
			}
		}

		function handleBack() {
			uni.navigateBack();
		}

		onMounted(() => {
			initStatusBarHeight();
			uni.setNavigationBarTitle({ title: '评教' });
			const token = uni.getStorageSync('token');
			if (!token) {
				showPreviewData();
				return;
			}
			clientId.value = uni.getStorageSync('clientId') || '';
			loadEvaluationList().then(() => {
				loadCacheDataMixin();
				loadAllEvaluationData();
			});
		});

		onShow(() => {
			refreshUnreadMessages();
			loadCacheDataMixin();
			try {
				const shouldRefresh = shouldAutoRefreshData();
				if (shouldRefresh) {
					refreshEvaluationData();
				} else {
					const now = Date.now();
					if (now - lastRefreshTime.value > CONSTANTS.AUTO_REFRESH_INTERVAL) {
						refreshEvaluationData();
					}
				}
			} catch (error) {
				const now = Date.now();
				if (now - lastRefreshTime.value > CONSTANTS.AUTO_REFRESH_INTERVAL) {
					refreshEvaluationData();
				}
			}
		});

		onPullDownRefresh(() => {
			refreshEvaluationData();
		});

		onTabItemTap(() => {
			const now = Date.now();
			if (now - lastRefreshTime.value > CONSTANTS.REFRESH_INTERVAL) {
				refreshEvaluationData();
			}
		});

		// #ifdef MP-WEIXIN
		onShareAppMessage(() => {
			const completedCountVal = evaluationList.value.filter(item => item.evaluated).length;
			const totalCount = evaluationList.value.length;
			return shareManager.generateShareConfig({
				title: '课程评教 - 众柴智慧校园',
				path: '/pages/evaluation/list',
				imageUrl: '/static/images/share-evaluation.png',
				trackingData: { page: 'evaluation_list', feature: 'evaluation_share', totalCount, completedCount: completedCountVal, pendingCount: totalCount - completedCountVal, isPreview: isPreviewMode.value, hasData: evaluationList.value.length > 0 }
			});
		});

		onShareTimeline(() => {
			const completedCountVal = evaluationList.value.filter(item => item.evaluated).length;
			const totalCount = evaluationList.value.length;
			return shareManager.generateTimelineConfig({
				title: '课程评教 - 众柴智慧校园，轻松完成课程评价',
				query: '',
				imageUrl: '/static/images/share-evaluation.png',
				trackingData: { page: 'evaluation_list', feature: 'evaluation_timeline_share', totalCount, completedCount: completedCountVal, isPreview: isPreviewMode.value, hasData: evaluationList.value.length > 0 }
			});
		});
		// #endif

		return {
			statusBarHeight,
			navPaddingTop,
			loading,
			error,
			evaluationList,
			clientId,
			evaluationCache,
			showClearCacheConfirm,
			lastRefreshTime,
			isPreviewMode,
			heroSubtitle,
			refreshUnreadMessages,
			showPreviewData,
			goToLogin,
			refreshEvaluationData,
			loadEvaluationList,
			loadAllEvaluationData,
			loadCacheDataMixin,
			getCacheStatusMixin,
			confirmClearCache,
			cancelClearCache,
			goToSwipeEvaluation,
			quickEvaluateAll,
			handleBack,
			handleBackClick,
			handleTitleClick,
			handleNavHeightReady,
			hasUnevaluatedItems,
			hasDraftData,
			completedCount,
			pendingCount,
			formatTimeRangeMixin,
			goToEvaluate
		};
	}
};
</script>

<style lang="scss" scoped>
/* ============================================
   Evaluation List Page - Home Hero Style
   ============================================ */

.evaluation-page {
	flex: 1;
	width: 100%;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
	font-family: -apple-system, "SF Pro Text", "PingFang SC", sans-serif;
	position: relative;
	min-height: 100vh;
}

/* ---- Hero Section ---- */
.eval-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 280rpx;
}

.eval-hero-bg {
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

.eval-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.65) 0%,
		rgba(37, 99, 235, 0.4) 50%,
		rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.eval-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.eval-hero-nav-left {
	display: flex;
	align-items: center;
	gap: 16rpx;
	flex: 1;
}

.eval-hero-back {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.18);
	border-radius: 50%;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.25);
	flex-shrink: 0;

	&:active {
		background: rgba(255, 255, 255, 0.28);
		transform: scale(0.95);
	}
}

.eval-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.eval-hero-nav-right {
	display: flex;
	align-items: center;
	gap: 16rpx;
	flex-shrink: 0;
}

.eval-hero-icon-btn {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.18);
	border-radius: 50%;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.25);
	transition: all 0.2s var(--ease-out);

	&:active {
		background: rgba(255, 255, 255, 0.28);
		transform: scale(0.9);
	}

	&--disabled {
		opacity: 0.4;
		pointer-events: none;
	}
}

.eval-hero-content {
	position: relative;
	z-index: 2;
	padding: 8rpx 32rpx 0;
}

.eval-hero-subtitle {
	font-size: 28rpx;
	color: rgba(255, 255, 255, 0.8);
	font-weight: 400;
}

/* ---- Content Area ---- */
.eval-content {
	flex: 1;
	padding: 0 24rpx;
	padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
}

/* ---- Preview Banner ---- */
.eval-preview-banner {
	margin-bottom: 20rpx;
	background: rgba(255, 255, 255, 0.95);
	border-radius: 20rpx;
	padding: 20rpx 24rpx;
	display: flex;
	align-items: center;
	justify-content: space-between;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.1);
	border: 1px solid rgba(255, 255, 255, 0.8);
	margin-top: -24rpx;
}

.eval-preview-banner-left {
	display: flex;
	align-items: center;
	gap: 12rpx;
}

.eval-preview-banner-text {
	font-size: 26rpx;
	color: var(--text-secondary);
	font-weight: 500;
}

.eval-preview-banner-btn {
	font-size: 26rpx;
	font-weight: 700;
	color: var(--primary-600);
	padding: 8rpx 20rpx;
	background: var(--primary-soft);
	border-radius: 16rpx;
}

/* ---- Page Content ---- */
.eval-page-content {
	padding-top: 20rpx;
}

/* ---- Summary Card ---- */
.eval-summary-card {
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 28rpx;
	padding: 32rpx;
	display: flex;
	justify-content: space-around;
	align-items: center;
	margin-bottom: 24rpx;
	box-shadow: 0 8rpx 32rpx rgba(37, 99, 235, 0.3);
	position: relative;
	overflow: hidden;
}

.eval-summary-bg-decor1 {
	position: absolute;
	top: -60rpx;
	right: -60rpx;
	width: 200rpx;
	height: 200rpx;
	background: rgba(255, 255, 255, 0.08);
	border-radius: 50%;
}

.eval-summary-bg-decor2 {
	position: absolute;
	bottom: -80rpx;
	left: -40rpx;
	width: 240rpx;
	height: 240rpx;
	background: rgba(255, 255, 255, 0.05);
	border-radius: 50%;
}

.eval-summary-item {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 8rpx;
	position: relative;
	z-index: 1;
}

.eval-summary-value {
	font-size: 56rpx;
	font-weight: 800;
	color: #fff;
	line-height: 1;
}

.eval-summary-label {
	font-size: 22rpx;
	color: rgba(255, 255, 255, 0.8);
	font-weight: 500;
}

.eval-summary-divider {
	width: 1px;
	height: 80rpx;
	background: rgba(255, 255, 255, 0.25);
	position: relative;
	z-index: 1;
}

/* ---- Status Card ---- */
.eval-status-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 80rpx 32rpx;
	margin-bottom: 24rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 16rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1px solid rgba(148, 163, 184, 0.12);
}

.eval-status-spinner {
	width: 64rpx;
	height: 64rpx;
	border: 5rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: eval-spin 0.8s linear infinite;
}

@keyframes eval-spin {
	to { transform: rotate(360deg); }
}

.eval-status-icon {
	width: 120rpx;
	height: 120rpx;
	border-radius: 32rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 8rpx;

	&--success { background: var(--success-soft); }
	&--error { background: var(--error-soft); }
}

.eval-status-title {
	font-size: 36rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.eval-status-subtitle {
	font-size: 26rpx;
	color: var(--text-secondary);
	text-align: center;
}

.eval-retry-btn {
	margin-top: 8rpx;
	padding: 20rpx 48rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	color: #fff;
	border-radius: 44rpx;
	font-size: 28rpx;
	font-weight: 700;
	border: none;
	box-shadow: 0 6rpx 24rpx rgba(37, 99, 235, 0.3);

	&:active { transform: translateY(2rpx); opacity: 0.9; }
}

/* ---- List ---- */
.eval-list {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
}

/* ---- Card ---- */
.eval-card {
	background: #fff;
	border-radius: 24rpx;
	display: flex;
	overflow: hidden;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	transition: all 0.2s var(--ease-out);

	&:active {
		transform: scale(0.99);
		background: #f8fafc;
	}
}

.eval-card-status-bar {
	width: 8rpx;
	flex-shrink: 0;

	&--completed { background: linear-gradient(180deg, var(--success-color), #059669); }
	&--draft { background: linear-gradient(180deg, var(--primary-500), var(--primary-600)); }
	&--pending { background: linear-gradient(180deg, var(--warning-color), #d97706); }
}

.eval-card-body {
	flex: 1;
	padding: 24rpx;
}

.eval-card-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 16rpx;
}

.eval-card-tag {
	display: flex;
	align-items: center;
	gap: 8rpx;
	padding: 6rpx 16rpx;
	border-radius: 100rpx;
	font-size: 22rpx;
	font-weight: 600;

	&--completed { background: var(--success-soft); color: var(--success-color); }
	&--draft { background: var(--primary-soft); color: var(--primary-600); }
	&--pending { background: var(--warning-soft); color: var(--warning-color); }
}

.eval-card-tag-dot {
	width: 12rpx;
	height: 12rpx;
	border-radius: 50%;

	&--completed { background: var(--success-color); }
	&--draft { background: var(--primary-500); }
	&--pending { background: var(--warning-color); }
}

.eval-card-tag-text {
	font-size: 22rpx;
}

.eval-card-title {
	font-size: 32rpx;
	font-weight: 700;
	color: var(--text-primary);
	line-height: 1.4;
	margin-bottom: 20rpx;
}

.eval-card-footer {
	border-top: 1px solid rgba(226, 232, 240, 0.8);
	padding-top: 16rpx;
}

.eval-card-info {
	display: flex;
	flex-direction: column;
	gap: 8rpx;
}

.eval-card-row {
	display: flex;
	align-items: center;
	gap: 8rpx;
	font-size: 24rpx;
	color: var(--text-secondary);
}

.eval-card-icon {
	color: var(--text-tertiary);
}

/* ---- FAB ---- */
.eval-fab {
	position: fixed;
	right: 32rpx;
	bottom: calc(160rpx + env(safe-area-inset-bottom));
	z-index: 100;
}

.eval-fab-btn {
	width: 100rpx;
	height: 100rpx;
	border-radius: 28rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 8rpx 24rpx rgba(239, 68, 68, 0.35);
	transition: all 0.3s var(--ease-spring);

	&:active { transform: scale(0.9); }

	&--danger {
		background: linear-gradient(135deg, var(--error-color), #dc2626);
	}
}

/* ---- Modal ---- */
.eval-modal-overlay {
	position: fixed;
	top: 0; left: 0; right: 0; bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	display: flex;
	justify-content: center;
	align-items: center;
	z-index: 1000;
	animation: eval-fadeIn 0.2s ease;
}

@keyframes eval-fadeIn {
	from { opacity: 0; }
	to { opacity: 1; }
}

.eval-modal-container {
	width: 600rpx;
	background: #fff;
	border-radius: 28rpx;
	overflow: hidden;
	box-shadow: 0 12rpx 40rpx rgba(0, 0, 0, 0.15);
}

.eval-modal-body {
	padding: 48rpx 40rpx;
	text-align: center;
}

.eval-modal-title {
	font-size: 36rpx;
	font-weight: 800;
	color: var(--text-primary);
	margin-bottom: 16rpx;
	display: block;
}

.eval-modal-message {
	font-size: 28rpx;
	color: var(--text-secondary);
	line-height: 1.6;
	display: block;
}

.eval-modal-footer {
	display: flex;
	border-top: 1px solid rgba(226, 232, 240, 0.8);
}

.eval-modal-btn {
	flex: 1;
	height: 96rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 30rpx;
	font-weight: 600;
	color: var(--text-secondary);

	&:active { background: var(--bg-muted); }

	&--confirm {
		color: var(--error-color);
		font-weight: 800;
	}

	&:not(:last-child) { border-right: 1px solid rgba(226, 232, 240, 0.8); }
}
</style>
