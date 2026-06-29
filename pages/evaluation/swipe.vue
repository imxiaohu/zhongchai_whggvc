<template>
	<view class="swipe-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="swipe-hero">
			<view class="swipe-hero-bg"></view>
			<view class="swipe-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<!-- 导航栏 -->
			<view class="swipe-hero-nav">
				<view class="swipe-hero-nav-left">
					<view class="swipe-back-btn" @tap="navigateBack">
						<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
					</view>
					<view class="swipe-progress-ring">
						<text class="swipe-progress-current">{{ currentIndex + 1 }}</text>
						<text class="swipe-progress-sep">/</text>
						<text class="swipe-progress-total">{{ evaluationList.length }}</text>
					</view>
				</view>
				<view class="swipe-hero-nav-right">
					<view class="swipe-hero-icon-btn" @tap="quickEvaluateAndSubmitAll" v-if="!loading && hasUnevaluatedItems">
						<l-icon name="star-filled" style="font-size: 20px; color: #fff;"></l-icon>
					</view>
				</view>
			</view>

			<!-- 标题区 -->
			<view class="swipe-hero-content">
				<text class="swipe-hero-title">课程评教</text>
				<text class="swipe-hero-sub">{{ hasUnevaluatedItems ? '认真评价，助力教学' : '全部课程已完成' }}</text>
			</view>
		</view>

		<!-- 滑动区域 -->
		<swiper
			class="swipe-swiper"
			:current="currentIndex"
			@change="onSwiperChange"
			v-if="!loading && evaluationList.length > 0"
		>
			<swiper-item v-for="(item, index) in evaluationList" :key="item.id">
				<scroll-view scroll-y class="swipe-scroll">
					<!-- 课程信息卡片 -->
					<view class="swipe-course-card">
						<view class="swipe-course-card-inner">
							<!-- 顶部标签 + 箭头 -->
							<view class="swipe-course-header">
								<view class="swipe-status-badge"
									:class="isCompleted(item.id) ? 'swipe-status-badge--done' : 'swipe-status-badge--pending'">
									<l-icon
										:name="isCompleted(item.id) ? 'check-circle-filled' : 'clock'"
										style="font-size: 14px;"
									></l-icon>
									<text class="swipe-status-badge-text">{{ isCompleted(item.id) ? '已完成' : '待评价' }}</text>
								</view>
								<view class="swipe-card-index">{{ index + 1 }} / {{ evaluationList.length }}</view>
							</view>

							<!-- 课程名称 -->
							<text class="swipe-course-name">{{ item.name }}</text>

							<!-- 教师与时间 -->
							<view class="swipe-course-meta">
								<view class="swipe-meta-item">
									<l-icon name="user-circle-filled" style="font-size: 14px; color: var(--text-tertiary);"></l-icon>
									<text>{{ item.teacherName }}</text>
								</view>
								<view class="swipe-meta-item">
									<l-icon name="calendar" style="font-size: 14px; color: var(--text-tertiary);"></l-icon>
									<text>{{ formatTimeRange(item.startTime, item.endTime) }}</text>
								</view>
							</view>

							<!-- 进度条 -->
							<view class="swipe-progress-section">
								<view class="swipe-progress-label">
									<text class="swipe-progress-label-text">评分进度</text>
									<text class="swipe-progress-label-value">{{ calculateScorePercentage(item.id) }}%</text>
								</view>
								<view class="swipe-progress-track">
									<view
										class="swipe-progress-fill"
										:class="getScoreClass(calculateScorePercentage(item.id))"
										:style="{ width: calculateScorePercentage(item.id) + '%' }"
									></view>
								</view>
							</view>
						</view>
					</view>

					<!-- 评分指标区 -->
					<view class="swipe-eval-section">
						<view class="swipe-eval-section-header">
							<text class="swipe-eval-section-title">评价指标</text>
							<view class="swipe-eval-section-action" @tap="setMaxScore(item.id)">
								<l-icon name="check-circle-filled" style="font-size: 16px; color: var(--primary-500);"></l-icon>
								<text class="swipe-eval-section-action-text">本课满分</text>
							</view>
						</view>

						<!-- 评分项 -->
						<view class="swipe-norm-list">
							<view
								v-for="(norm, normIndex) in getNormList(item.id)"
								:key="normIndex"
								class="swipe-norm-card"
							>
								<view class="swipe-norm-info">
									<view class="swipe-norm-number">{{ normIndex + 1 }}</view>
									<view class="swipe-norm-text">
										<text class="swipe-norm-name">{{ norm.name }}</text>
										<text v-if="norm.content" class="swipe-norm-desc">{{ norm.content }}</text>
									</view>
								</view>
								<view class="swipe-score-row">
									<view
										v-for="score in 5"
										:key="score"
										class="swipe-score-btn"
										:class="{
										'swipe-score-btn--active': norm.score === score,
										'swipe-score-btn--high': norm.score === score && score >= 4,
										'swipe-score-btn--mid': norm.score === score && score === 3,
										'swipe-score-btn--low': norm.score === score && score <= 2
									}"
									@tap="selectScore(item.id, normIndex, score)"
								>
									<text class="swipe-score-btn-text">{{ score }}</text>
								</view>
								</view>
							</view>
						</view>

						<!-- 评语区 -->
						<view class="swipe-comment-section">
							<view class="swipe-comment-header">
								<text class="swipe-comment-title">评价建议</text>
								<text class="swipe-comment-hint">（选填）</text>
							</view>
							<view class="swipe-textarea-wrap">
								<textarea
									class="swipe-textarea"
									:value="getComment(item.id)"
									@input="updateComment(item.id, $event.detail.value)"
									placeholder="分享你的学习体验，帮助老师改进教学..."
									:maxlength="CONSTANTS.COMMENT_MAX_LENGTH"
									placeholder-class="swipe-textarea-placeholder"
								></textarea>
								<text class="swipe-char-count">{{ getComment(item.id).length }}/{{ CONSTANTS.COMMENT_MAX_LENGTH }}</text>
							</view>

							<!-- 快捷评语 -->
							<view class="swipe-quick-chips">
								<view
									v-for="(comment, ci) in quickComments"
									:key="ci"
									class="swipe-quick-chip"
									@tap="selectQuickComment(item.id, comment)"
								>
									<text class="swipe-quick-chip-text">{{ comment }}</text>
								</view>
							</view>
						</view>
					</view>

					<view class="swipe-bottom-space"></view>
				</scroll-view>
			</swiper-item>
		</swiper>

		<!-- 底部操作栏 -->
		<view class="swipe-bottom-bar" v-if="!loading && evaluationList.length > 0">
			<view class="swipe-bottom-inner">
				<view class="swipe-btn-secondary" @tap="saveCurrentEvaluation">
					<l-icon name="save" style="font-size: 18px; color: var(--text-secondary);"></l-icon>
					<text class="swipe-btn-text">保存草稿</text>
				</view>
				<view
					class="swipe-btn-primary"
					:class="{ 'swipe-btn-primary--disabled': !isCompleted(evaluationList[currentIndex]?.id) }"
					@tap="submitCurrentEvaluation"
				>
					<l-icon name="check-circle-filled" style="font-size: 18px; color: #fff;"></l-icon>
					<text class="swipe-btn-text">提交评价</text>
				</view>
			</view>
		</view>

		<!-- 加载状态 -->
		<view v-if="loading" class="swipe-loading">
			<view class="swipe-loading-spinner"></view>
			<text class="swipe-loading-text">加载中...</text>
		</view>
	</view>
</template>

<script>
import swipeLogic from './logic/swipe.js';
export default swipeLogic;
</script>

<style lang="scss" scoped>
/* ============================================
   Swipe Evaluation Page - Hero Style
   ============================================ */

.swipe-page {
	position: relative;
	width: 100%;
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
}

/* ---- Hero Section ---- */
.swipe-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 240rpx;
}

.swipe-hero-bg {
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

.swipe-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.65) 0%,
		rgba(37, 99, 235, 0.4) 50%,
		rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.swipe-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.swipe-hero-nav-left {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.swipe-back-btn {
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

.swipe-progress-ring {
	display: flex;
	align-items: baseline;
	gap: 4rpx;
}

.swipe-progress-current {
	font-size: 40rpx;
	font-weight: 800;
	color: #fff;
	line-height: 1;
}

.swipe-progress-sep {
	font-size: 24rpx;
	color: rgba(255, 255, 255, 0.6);
	margin: 0 4rpx;
}

.swipe-progress-total {
	font-size: 28rpx;
	color: rgba(255, 255, 255, 0.8);
}

.swipe-hero-nav-right {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.swipe-hero-icon-btn {
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
}

.swipe-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 8rpx;
}

.swipe-hero-title {
	display: block;
	font-size: 40rpx;
	font-weight: 800;
	color: #fff;
	line-height: 1.3;
}

.swipe-hero-sub {
	display: block;
	font-size: 26rpx;
	color: rgba(255, 255, 255, 0.75);
	margin-top: 4rpx;
}

/* ---- Swiper ---- */
.swipe-swiper {
	flex: 1;
	min-height: 0;
}

.swipe-scroll {
	width: 100%;
	height: 100%;
	box-sizing: border-box;
	padding: 24rpx;
	padding-bottom: calc(160rpx + env(safe-area-inset-bottom));
}

/* ---- Course Card ---- */
.swipe-course-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 32rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.1);
	border: 1px solid rgba(148, 163, 184, 0.12);
	overflow: hidden;
}

.swipe-course-card-inner {
	border-left: 8rpx solid var(--primary-500);
	padding-left: 24rpx;
}

.swipe-course-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 16rpx;
}

.swipe-status-badge {
	display: flex;
	align-items: center;
	gap: 8rpx;
	padding: 6rpx 16rpx;
	border-radius: 100rpx;
	font-size: 22rpx;
	font-weight: 600;

	&--done {
		background: var(--success-soft);
		color: var(--success-color);
	}
	&--pending {
		background: var(--warning-soft);
		color: var(--warning-color);
	}
}

.swipe-card-index {
	font-size: 22rpx;
	color: var(--text-tertiary);
	font-weight: 500;
}

.swipe-course-name {
	display: block;
	font-size: 36rpx;
	font-weight: 800;
	color: var(--text-primary);
	line-height: 1.4;
	margin-bottom: 16rpx;
}

.swipe-course-meta {
	display: flex;
	flex-wrap: wrap;
	gap: 20rpx;
	margin-bottom: 24rpx;
}

.swipe-meta-item {
	display: flex;
	align-items: center;
	gap: 8rpx;
	font-size: 24rpx;
	color: var(--text-secondary);
}

.swipe-progress-section {
	border-top: 1px solid rgba(226, 232, 240, 0.8);
	padding-top: 20rpx;
}

.swipe-progress-label {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 12rpx;
}

.swipe-progress-label-text {
	font-size: 24rpx;
	color: var(--text-tertiary);
	font-weight: 500;
}

.swipe-progress-label-value {
	font-size: 24rpx;
	font-weight: 700;
	color: var(--primary-600);
}

.swipe-progress-track {
	height: 10rpx;
	background: var(--bg-muted);
	border-radius: 5rpx;
	overflow: hidden;
}

.swipe-progress-fill {
	height: 100%;
	border-radius: 5rpx;
	transition: width 0.4s var(--ease-out), background 0.3s ease;

	&.score-high { background: linear-gradient(90deg, var(--success-color), #34d399); }
	&.score-medium { background: linear-gradient(90deg, var(--primary-500), var(--primary-400)); }
	&.score-low { background: linear-gradient(90deg, var(--warning-color), #fbbf24); }
	&.score-none { background: var(--text-tertiary); }
}

/* ---- Eval Section ---- */
.swipe-eval-section {
	margin-bottom: 24rpx;
}

.swipe-eval-section-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 20rpx;
}

.swipe-eval-section-title {
	font-size: 32rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.swipe-eval-section-action {
	display: flex;
	align-items: center;
	gap: 6rpx;
	font-size: 24rpx;
	color: var(--primary-500);
	font-weight: 600;
	padding: 8rpx 16rpx;
	border-radius: 100rpx;
	background: var(--primary-soft);

	&:active { opacity: 0.7; }
}

/* ---- Norm Cards ---- */
.swipe-norm-list {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
	margin-bottom: 32rpx;
}

.swipe-norm-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 24rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.swipe-norm-info {
	display: flex;
	gap: 16rpx;
	margin-bottom: 20rpx;
}

.swipe-norm-number {
	width: 48rpx;
	height: 48rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, var(--primary-500), var(--primary-600));
	color: #fff;
	font-size: 24rpx;
	font-weight: 700;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
}

.swipe-norm-text {
	flex: 1;
}

.swipe-norm-name {
	display: block;
	font-size: 28rpx;
	font-weight: 600;
	color: var(--text-primary);
	line-height: 1.4;
	margin-bottom: 4rpx;
}

.swipe-norm-desc {
	display: block;
	font-size: 22rpx;
	color: var(--text-tertiary);
	line-height: 1.5;
}

/* ---- Score Buttons ---- */
.swipe-score-row {
	display: grid;
	grid-template-columns: repeat(5, 1fr);
	gap: 12rpx;
}

.swipe-score-btn {
	height: 88rpx;
	background: #fff;
	border: 2rpx solid rgba(148, 163, 184, 0.3);
	border-radius: 20rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	transition: all 0.2s var(--ease-spring);

	.swipe-score-btn-text {
		font-size: 36rpx;
		font-weight: 800;
		color: var(--text-secondary);
		transition: color 0.2s ease;
	}

	&:active { transform: scale(0.95); }

	&--active {
		border-color: transparent;
		box-shadow: 0 6rpx 20rpx rgba(0, 0, 0, 0.15);
		transform: scale(1.05);

		.swipe-score-btn-text { color: #fff; }

		&.swipe-score-btn--high {
			background: linear-gradient(135deg, #10b981, #34d399);
			box-shadow: 0 6rpx 20rpx rgba(16, 185, 129, 0.35);
		}
		&.swipe-score-btn--mid {
			background: linear-gradient(135deg, var(--primary-500), var(--primary-400));
			box-shadow: 0 6rpx 20rpx rgba(59, 130, 246, 0.35);
		}
		&.swipe-score-btn--low {
			background: linear-gradient(135deg, #f59e0b, #fbbf24);
			box-shadow: 0 6rpx 20rpx rgba(245, 158, 11, 0.35);
		}
	}
}

/* ---- Comment Section ---- */
.swipe-comment-section {
	margin-bottom: 24rpx;
}

.swipe-comment-header {
	display: flex;
	align-items: center;
	gap: 8rpx;
	margin-bottom: 16rpx;
}

.swipe-comment-title {
	font-size: 32rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.swipe-comment-hint {
	font-size: 24rpx;
	color: var(--text-tertiary);
}

.swipe-textarea-wrap {
	position: relative;
	margin-bottom: 16rpx;
}

.swipe-textarea {
	width: 100%;
	height: 200rpx;
	background: #fff;
	border: 1px solid rgba(148, 163, 184, 0.2);
	border-radius: 20rpx;
	padding: 24rpx;
	font-size: 28rpx;
	color: var(--text-primary);
	box-sizing: border-box;
	line-height: 1.6;
	box-shadow: 0 2rpx 8rpx rgba(30, 64, 175, 0.04);
	transition: border-color 0.2s ease;

	&:focus {
		border-color: var(--primary-400);
		box-shadow: 0 0 0 4rpx rgba(59, 130, 246, 0.1);
	}
}

.swipe-textarea-placeholder {
	color: var(--text-tertiary);
	font-size: 28rpx;
}

.swipe-char-count {
	position: absolute;
	right: 20rpx;
	bottom: 20rpx;
	font-size: 20rpx;
	color: var(--text-tertiary);
}

/* ---- Quick Chips ---- */
.swipe-quick-chips {
	display: flex;
	flex-wrap: wrap;
	gap: 12rpx;
}

.swipe-quick-chip {
	padding: 10rpx 24rpx;
	background: #fff;
	border-radius: 100rpx;
	border: 1px solid rgba(148, 163, 184, 0.25);
	box-shadow: 0 2rpx 8rpx rgba(30, 64, 175, 0.04);

	.swipe-quick-chip-text {
		font-size: 24rpx;
		color: var(--text-secondary);
		font-weight: 500;
	}

	&:active {
		background: var(--primary-soft);
		border-color: var(--primary-400);
		.swipe-quick-chip-text { color: var(--primary-600); }
	}
}

/* ---- Bottom Space ---- */
.swipe-bottom-space {
	height: 20rpx;
}

/* ---- Bottom Action Bar ---- */
.swipe-bottom-bar {
	position: fixed;
	bottom: 0;
	left: 0;
	right: 0;
	background: rgba(255, 255, 255, 0.92);
	backdrop-filter: blur(20px);
	-webkit-backdrop-filter: blur(20px);
	padding: 20rpx 32rpx calc(20rpx + env(safe-area-inset-bottom));
	border-top: 1px solid rgba(148, 163, 184, 0.15);
	z-index: 100;
	box-shadow: 0 -4rpx 20rpx rgba(30, 64, 175, 0.08);
}

.swipe-bottom-inner {
	display: flex;
	gap: 16rpx;
}

.swipe-btn-secondary {
	flex: 0 0 220rpx;
	height: 96rpx;
	background: var(--bg-muted);
	border-radius: 48rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 10rpx;
	border: 1px solid rgba(148, 163, 184, 0.2);

	.swipe-btn-text {
		font-size: 28rpx;
		font-weight: 700;
		color: var(--text-secondary);
	}

	&:active { opacity: 0.8; transform: scale(0.98); }
}

.swipe-btn-primary {
	flex: 1;
	height: 96rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 48rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 10rpx;
	box-shadow: 0 8rpx 24rpx rgba(37, 99, 235, 0.35);

	.swipe-btn-text {
		font-size: 30rpx;
		font-weight: 800;
		color: #fff;
	}

	&:active:not(.swipe-btn-primary--disabled) {
		transform: translateY(2rpx);
		box-shadow: 0 4rpx 12rpx rgba(37, 99, 235, 0.25);
	}

	&--disabled {
		background: var(--bg-muted);
		box-shadow: none;

		.swipe-btn-text { color: var(--text-tertiary); }
	}
}

/* ---- Loading ---- */
.swipe-loading {
	position: fixed;
	top: 0; left: 0; right: 0; bottom: 0;
	background: rgba(255, 255, 255, 0.9);
	backdrop-filter: blur(4px);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 24rpx;
	z-index: 1000;
}

.swipe-loading-spinner {
	width: 64rpx;
	height: 64rpx;
	border: 5rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: swipe-spin 0.8s linear infinite;
}

@keyframes swipe-spin {
	to { transform: rotate(360deg); }
}

.swipe-loading-text {
	font-size: 28rpx;
	color: var(--text-secondary);
	font-weight: 500;
}
</style>
