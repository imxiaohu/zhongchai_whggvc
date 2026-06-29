<template>
	<view class="score-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="score-hero">
			<view class="score-hero-bg"></view>
			<view class="score-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<!-- 导航栏 -->
			<view class="score-hero-nav">
				<view class="score-hero-nav-left">
					<view class="score-back-btn" @tap="goBack">
						<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
					</view>
					<text class="score-hero-title">成绩查询</text>
				</view>
			</view>

			<!-- 欢迎语 -->
			<view class="score-hero-content">
				<text class="score-hero-sub">{{ heroSubtitle }}</text>
			</view>
		</view>

		<!-- 主要内容 -->
		<view class="score-content">

			<!-- 预览模式提示 -->
			<view v-if="isPreviewMode" class="score-preview-banner">
				<view class="score-preview-banner-left">
					<l-icon name="info-circle-filled" style="font-size: 18px; color: var(--warning-color);"></l-icon>
					<text class="score-preview-banner-text">当前为预览模式，显示示例数据</text>
				</view>
				<text class="score-preview-banner-btn" @tap="goToLogin">立即登录</text>
			</view>

			<!-- 查询区域 -->
			<view class="score-query-section" v-if="!isPreviewMode">
			<view class="score-semester-card">
				<view class="score-semester-left">
					<view class="score-semester-icon">
						<l-icon name="calendar" style="font-size: 20px; color: var(--primary-600);"></l-icon>
					</view>
					<view class="score-semester-info">
						<text class="score-semester-label">当前学期</text>
						<text class="score-semester-value">{{ currentSemesterName }}</text>
					</view>
				</view>
				<l-icon name="chevron-down" style="font-size: 20px; color: var(--text-tertiary);"></l-icon>
				<picker
					class="score-semester-picker"
					mode="selector"
					@change="handleSemesterChange"
					:value="semesterIndex"
					:range="semesterList"
					range-key="name"
					:disabled="loading"
				></picker>
			</view>

				<view class="score-quick-tabs">
					<view
						class="score-quick-tab"
						:class="{ 'score-quick-tab--active': isCurrentSemesterSelected }"
						@tap="quickQueryCurrentSemester"
					>本学期</view>
					<view class="score-quick-tab" @tap="quickQueryPreviousSemester">上学期</view>
					<view class="score-quick-tab" @tap="quickQueryNextSemester">下学期</view>
				</view>
			</view>

			<!-- 统计卡片 -->
			<view class="score-stats-section" v-if="hasData && !isPreviewMode">
				<view class="score-stats-grid">
					<view class="score-stat-card score-stat-card--gpa">
						<view class="score-stat-bg-decor"></view>
						<view class="score-stat-icon-wrap">
							<l-icon name="trending-up" style="font-size: 24px; color: rgba(255,255,255,0.8);"></l-icon>
						</view>
						<text class="score-stat-value">{{ semesterStats.gpa || '0.00' }}</text>
						<text class="score-stat-label">平均绩点</text>
					</view>
					<view class="score-stat-card score-stat-card--avg">
						<view class="score-stat-bg-decor"></view>
						<view class="score-stat-icon-wrap">
							<l-icon name="star-filled" style="font-size: 24px; color: rgba(255,255,255,0.8);"></l-icon>
						</view>
						<text class="score-stat-value">{{ semesterStats.averageScore || '0.0' }}</text>
						<text class="score-stat-label">平均分</text>
					</view>
					<view class="score-stat-card score-stat-card--credit">
						<view class="score-stat-bg-decor"></view>
						<view class="score-stat-icon-wrap">
							<l-icon name="education" style="font-size: 24px; color: rgba(255,255,255,0.8);"></l-icon>
						</view>
						<text class="score-stat-value">{{ semesterStats.creditTotal || '0' }}</text>
						<text class="score-stat-label">已获学分</text>
					</view>
					<view class="score-stat-card score-stat-card--pass">
						<view class="score-stat-bg-decor"></view>
						<view class="score-stat-icon-wrap">
							<l-icon name="check-circle" style="font-size: 24px; color: rgba(255,255,255,0.8);"></l-icon>
						</view>
						<text class="score-stat-value">{{ semesterStats.passRate }}%</text>
						<text class="score-stat-label">通过率</text>
					</view>
				</view>
			</view>

			<!-- 课程成绩列表 -->
			<view class="score-list-section">
				<!-- 列表头部 -->
				<view class="score-list-header" v-if="courseScores.length > 0">
					<text class="score-list-title">课程成绩</text>
					<text class="score-list-count">共 {{ courseScores.length }} 门</text>
				</view>

				<!-- 成绩列表 -->
				<view class="score-cards-list" v-if="courseScores.length > 0">
					<ScoreCard
						v-for="(course, index) in courseScores"
						:key="index"
						:course="course"
					/>
				</view>

				<!-- 加载状态 -->
				<view v-if="loading" class="score-state-card">
					<view class="score-state-spinner"></view>
					<text class="score-state-text">正在获取成绩...</text>
				</view>

				<!-- 空状态 -->
				<view v-if="!loading && courseScores.length === 0 && !isPreviewMode" class="score-state-card">
					<view class="score-state-icon score-state-icon--info">
						<l-icon name="file-text" style="font-size: 40px; color: var(--primary-500);"></l-icon>
					</view>
					<text class="score-state-title">暂无成绩记录</text>
					<text class="score-state-sub">当前学期成绩可能尚未公布</text>
				</view>
			</view>
		</view>

		<!-- 缓存提示 -->
		<cache-banner
			:visible="showCacheBanner"
			:cache-updated-at="cacheUpdatedAt"
			@close="showCacheBanner = false"
		></cache-banner>
	</view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useScorePage } from '@/composables/useScorePage.js'
import ScoreCard from '@/components/ScoreCard.vue'
import CacheBanner from '@/components/CacheBanner.vue'

const statusBarHeight = ref(20)
const navPaddingTop = ref('0px')

const {
	loading,
	semesterList,
	semesterIndex,
	courseScores,
	semesterStats,
	hasData,
	error,
	showCacheBanner,
	cacheUpdatedAt,
	isPreviewMode,
	currentSemesterName,
	isCurrentSemesterSelected,
	checkLoginStatus,
	goToLogin,
	initializeData,
	handleSemesterChange,
	quickQueryCurrentSemester,
	quickQueryPreviousSemester,
	quickQueryNextSemester,
} = useScorePage()

const heroSubtitle = computed(() => {
	if (isPreviewMode.value) return '登录后可查看真实成绩数据'
	if (hasData.value) return `共 ${courseScores.value.length} 门课程`
	return '查看各学期成绩详情'
})

function initStatusBarHeight() {
	const systemInfo = uni.getSystemInfoSync()
	statusBarHeight.value = systemInfo.statusBarHeight || 20
}

function goBack() {
	uni.navigateBack()
}

function onPageLoad() {
	console.log('成绩页面加载')
	initStatusBarHeight()
	if (!checkLoginStatus()) return
	initializeData()
}

onPageLoad()
</script>

<style lang="scss" scoped>
/* ============================================
   Score Page - Hero Style
   ============================================ */

.score-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
	position: relative;
}

/* ---- Hero Section ---- */
.score-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 220rpx;
}

.score-hero-bg {
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

.score-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.65) 0%,
		rgba(37, 99, 235, 0.4) 50%,
		rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.score-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.score-hero-nav-left {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.score-back-btn {
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

.score-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.score-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx;
}

.score-hero-sub {
	font-size: 28rpx;
	color: rgba(255, 255, 255, 0.8);
}

/* ---- Content ---- */
.score-content {
	flex: 1;
	padding: 0 24rpx;
	padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}

/* ---- Preview Banner ---- */
.score-preview-banner {
	margin-top: -24rpx;
	margin-bottom: 20rpx;
	background: rgba(255, 255, 255, 0.95);
	border-radius: 20rpx;
	padding: 20rpx 24rpx;
	display: flex;
	align-items: center;
	justify-content: space-between;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.1);
	border: 1px solid rgba(255, 255, 255, 0.8);
}

.score-preview-banner-left {
	display: flex;
	align-items: center;
	gap: 12rpx;
}

.score-preview-banner-text {
	font-size: 26rpx;
	color: var(--text-secondary);
	font-weight: 500;
}

.score-preview-banner-btn {
	font-size: 26rpx;
	font-weight: 700;
	color: var(--primary-600);
	padding: 8rpx 20rpx;
	background: var(--primary-soft);
	border-radius: 16rpx;
}

/* ---- Query Section ---- */
.score-query-section {
	padding-top: 20rpx;
	margin-bottom: 24rpx;
}

.score-semester-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 24rpx;
	display: flex;
	align-items: center;
	justify-content: space-between;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	margin-bottom: 16rpx;
	position: relative;
}

.score-semester-picker {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	opacity: 0;
	z-index: 2;
}

.score-semester-left {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.score-semester-icon {
	width: 72rpx;
	height: 72rpx;
	border-radius: 20rpx;
	background: var(--primary-soft);
	display: flex;
	align-items: center;
	justify-content: center;
}

.score-semester-info {
	display: flex;
	flex-direction: column;
	gap: 4rpx;
}

.score-semester-label {
	font-size: 22rpx;
	color: var(--text-tertiary);
	font-weight: 500;
}

.score-semester-value {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.score-quick-tabs {
	display: flex;
	gap: 12rpx;
}

.score-quick-tab {
	flex: 1;
	height: 80rpx;
	background: #fff;
	border-radius: 20rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 26rpx;
	font-weight: 600;
	color: var(--text-secondary);
	border: 1px solid rgba(148, 163, 184, 0.15);
	box-shadow: 0 2rpx 8rpx rgba(30, 64, 175, 0.04);
	transition: all 0.2s ease;

	&--active {
		background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
		color: #fff;
		border-color: transparent;
		box-shadow: 0 6rpx 20rpx rgba(37, 99, 235, 0.3);
	}

	&:active:not(.score-quick-tab--active) {
		background: var(--bg-muted);
	}
}

/* ---- Stats Section ---- */
.score-stats-section {
	margin-bottom: 24rpx;
}

.score-stats-grid {
	display: grid;
	grid-template-columns: repeat(2, 1fr);
	gap: 16rpx;
}

.score-stat-card {
	position: relative;
	padding: 24rpx;
	border-radius: 24rpx;
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	overflow: hidden;
	box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.08);

	&--gpa { background: linear-gradient(135deg, #6366f1, #818cf8); }
	&--avg { background: linear-gradient(135deg, #10b981, #34d399); }
	&--credit { background: linear-gradient(135deg, #3b82f6, #60a5fa); }
	&--pass { background: linear-gradient(135deg, #f59e0b, #fbbf24); }
}

.score-stat-bg-decor {
	position: absolute;
	top: -40rpx;
	right: -40rpx;
	width: 160rpx;
	height: 160rpx;
	background: rgba(255, 255, 255, 0.1);
	border-radius: 50%;
}

.score-stat-icon-wrap {
	margin-bottom: 8rpx;
}

.score-stat-value {
	font-size: 44rpx;
	font-weight: 800;
	color: #fff;
	line-height: 1;
	margin-bottom: 4rpx;
}

.score-stat-label {
	font-size: 22rpx;
	color: rgba(255, 255, 255, 0.8);
	font-weight: 500;
}

/* ---- List Section ---- */
.score-list-section {
	padding-bottom: 40rpx;
}

.score-list-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 16rpx;
}

.score-list-title {
	font-size: 32rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.score-list-count {
	font-size: 22rpx;
	color: var(--text-tertiary);
	font-weight: 500;
}

.score-cards-list {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
}

/* ---- State Card ---- */
.score-state-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 80rpx 32rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 16rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1px solid rgba(148, 163, 184, 0.12);
}

.score-state-spinner {
	width: 64rpx;
	height: 64rpx;
	border: 5rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: score-spin 0.8s linear infinite;
}

@keyframes score-spin {
	to { transform: rotate(360deg); }
}

.score-state-icon {
	width: 120rpx;
	height: 120rpx;
	border-radius: 32rpx;
	display: flex;
	align-items: center;
	justify-content: center;

	&--info { background: var(--primary-soft); }
}

.score-state-title {
	font-size: 36rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.score-state-sub {
	font-size: 26rpx;
	color: var(--text-tertiary);
	text-align: center;
}

.score-state-text {
	font-size: 28rpx;
	color: var(--text-secondary);
}
</style>
