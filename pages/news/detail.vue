<template>
	<view class="news-detail-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="detail-hero">
			<view class="detail-hero-bg"></view>
			<view class="detail-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="detail-hero-nav">
				<view class="detail-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="detail-hero-title">{{ newsDetail ? (newsDetail.title.length > 12 ? newsDetail.title.substring(0, 12) + '...' : newsDetail.title) : '通知详情' }}</text>
				<view class="detail-share-btn" @tap="shareNews" v-if="newsDetail">
					<l-icon name="share" style="font-size: 20px; color: #fff;"></l-icon>
				</view>
				<view style="width: 64rpx;" v-else></view>
			</view>

			<view class="detail-hero-content" v-if="newsDetail">
				<view class="detail-hero-meta">
					<text class="detail-hero-date">{{ formatDate(newsDetail.publishTime || newsDetail.createTime) }}</text>
					<text class="detail-hero-sep">·</text>
					<text class="detail-hero-source">{{ newsDetail.source || newsDetail.author || '系统' }}</text>
				</view>
			</view>
		</view>

		<!-- 内容区域 -->
		<view class="detail-content" v-if="newsDetail">
			<scroll-view class="detail-scroll" scroll-y="true">
				<!-- 内容卡片 -->
				<view class="detail-card">
					<!-- 标题 -->
					<view class="detail-card-title">{{ newsDetail.title }}</view>

					<!-- 分类标签 -->
					<view class="detail-card-tags" v-if="newsDetail.type">
						<text class="detail-type-tag">{{ newsDetail.type }}</text>
					</view>

					<!-- 摘要 -->
					<view class="detail-summary" v-if="newsDetail.summary">
						<l-icon class="summary-icon" name="quote" style="font-size: 20px; color: var(--primary-400);"></l-icon>
						<text class="summary-text">{{ newsDetail.summary }}</text>
					</view>

					<!-- 正文 -->
					<view class="detail-body">
						<rich-text
							:nodes="formatContent(newsDetail.content)"
							style="width: 100%; overflow: hidden;"
							@tap="handleRichTextTap"
						></rich-text>
					</view>

					<!-- 附件列表 -->
					<view class="detail-attachments" v-if="newsDetail.attachments && newsDetail.attachments.length > 0">
						<view class="detail-attachments-title">
							<l-icon name="file-attachment" style="font-size: 20px; color: var(--primary-500); margin-right: 12rpx;"></l-icon>
							<text class="detail-attachments-title-text">附件下载</text>
						</view>
						<view
							v-for="(attachment, index) in newsDetail.attachments"
							:key="index"
							class="detail-attachment"
							@tap="downloadAttachment(attachment)"
						>
							<view class="detail-attachment-icon">
								<l-icon name="file-attachment" style="font-size: 24px; color: var(--primary-500);"></l-icon>
							</view>
							<view class="detail-attachment-info">
								<text class="detail-attachment-name">{{ attachment.name || '未知文件' }}</text>
								<text class="detail-attachment-size">{{ formatFileSize(attachment.size) || '' }}</text>
							</view>
							<view class="detail-attachment-download">
								<l-icon name="download" style="font-size: 18px; color: var(--primary-500);"></l-icon>
							</view>
						</view>
					</view>
				</view>
			</scroll-view>
		</view>

		<!-- 加载状态 -->
		<view v-if="loading" class="detail-loading">
			<view class="detail-loading-spinner"></view>
			<text class="detail-loading-text">加载中...</text>
		</view>

		<!-- 错误状态 -->
		<view v-if="error && !loading" class="detail-error">
			<l-icon name="error-circle" style="font-size: 48px; color: var(--error-color);"></l-icon>
			<text class="detail-error-text">{{ error }}</text>
			<view class="detail-retry-btn" @tap="loadNewsDetail">
				<text class="detail-retry-btn-text">重试</text>
			</view>
		</view>

		<!-- 图片预览 -->
		<ImagePreview
			:visible="showImagePreview"
			:imageUrl="previewImageUrl"
			@close="closeImagePreview"
		/>
	</view>
</template>

<script>
import detailLogic from './logic/detail.js';
export default detailLogic;
</script>

<style lang="scss" scoped>
/* ============================================
   News Detail - Hero Style
   ============================================ */

.news-detail-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
	position: relative;
}

/* ---- Hero Section ---- */
.detail-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 220rpx;
}

.detail-hero-bg {
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

.detail-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.65) 0%,
		rgba(37, 99, 235, 0.4) 50%,
		rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.detail-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.detail-back-btn {
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

.detail-hero-title {
	flex: 1;
	font-size: 34rpx;
	font-weight: 800;
	color: #fff;
	text-align: center;
	padding: 0 16rpx;
}

.detail-share-btn {
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
		transform: scale(0.9);
	}
}

.detail-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 24rpx;
}

.detail-hero-meta {
	display: flex;
	align-items: center;
	gap: 12rpx;
}

.detail-hero-date,
.detail-hero-source {
	font-size: 24rpx;
	color: rgba(255, 255, 255, 0.8);
}

.detail-hero-sep {
	font-size: 24rpx;
	color: rgba(255, 255, 255, 0.4);
}

/* ---- Content ---- */
.detail-content {
	flex: 1;
	min-height: 0;
}

.detail-scroll {
	height: 100%;
}

/* ---- Detail Card ---- */
.detail-card {
	background: #fff;
	margin: -32rpx 24rpx 40rpx;
	border-radius: 28rpx;
	padding: 36rpx;
	box-shadow: 0 4rpx 24rpx rgba(30, 64, 175, 0.1);
	border: 1px solid rgba(148, 163, 184, 0.12);
	position: relative;
	z-index: 2;
}

.detail-card-title {
	font-size: 38rpx;
	font-weight: 800;
	color: var(--text-primary);
	line-height: 1.4;
	margin-bottom: 20rpx;
}

.detail-card-tags {
	margin-bottom: 20rpx;
}

.detail-type-tag {
	display: inline-flex;
	align-items: center;
	padding: 8rpx 20rpx;
	border-radius: 100rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	font-size: 22rpx;
	font-weight: 700;
	color: #fff;
}

/* ---- Summary ---- */
.detail-summary {
	display: flex;
	align-items: center;
	background: rgba(59, 102, 241, 0.06);
	border-left: 6rpx solid var(--primary-500);
	border-radius: 0 16rpx 16rpx 0;
	padding: 20rpx 24rpx;
	margin-bottom: 32rpx;

	.summary-icon {
		flex-shrink: 0;
		margin-right: 12rpx;
	}

	.summary-text {
		font-size: 28rpx;
		color: var(--text-secondary);
		line-height: 1.7;
		flex: 1;
		min-width: 0;
		word-break: break-word;
	}
}

/* ---- Body ---- */
.detail-body {
	font-size: 30rpx;
	color: var(--text-primary);
	line-height: 1.9;
}

.detail-body :deep(img) {
	max-width: 100% !important;
	height: auto !important;
	display: block;
	margin: 16rpx auto;
	border-radius: 12rpx;
}

.detail-body :deep(p) {
	margin-bottom: 16rpx;
}

/* ---- Attachments ---- */
.detail-attachments {
	margin-top: 40rpx;
	padding-top: 32rpx;
	border-top: 1px dashed rgba(148, 163, 184, 0.4);
}

.detail-attachments-title {
	display: flex;
	align-items: center;
	margin-bottom: 20rpx;

	.detail-attachments-title-text {
		font-size: 30rpx;
		font-weight: 700;
		color: var(--text-primary);
	}
}

.detail-attachment {
	display: flex;
	align-items: center;
	padding: 20rpx 24rpx;
	background: var(--bg-muted);
	border-radius: 20rpx;
	margin-bottom: 12rpx;
	border: 1px solid rgba(148, 163, 184, 0.1);
	transition: all 0.15s ease;

	&:active {
		background: var(--bg-secondary);
		transform: scale(0.99);
	}
}

.detail-attachment-icon {
	width: 80rpx;
	height: 80rpx;
	border-radius: 20rpx;
	background: rgba(59, 102, 241, 0.1);
	display: flex;
	align-items: center;
	justify-content: center;
	margin-right: 20rpx;
	flex-shrink: 0;
}

.detail-attachment-info {
	flex: 1;
	min-width: 0;
}

.detail-attachment-name {
	display: block;
	font-size: 28rpx;
	font-weight: 600;
	color: var(--text-primary);
	margin-bottom: 4rpx;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.detail-attachment-size {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.detail-attachment-download {
	width: 64rpx;
	height: 64rpx;
	border-radius: 50%;
	background: rgba(59, 102, 241, 0.1);
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
}

/* ---- Loading ---- */
.detail-loading {
	position: fixed;
	top: 0; left: 0; right: 0; bottom: 0;
	background: rgba(255, 255, 255, 0.9);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 16rpx;
	z-index: 100;

	.detail-loading-text {
		font-size: 28rpx;
		color: var(--text-secondary);
	}
}

.detail-loading-spinner {
	width: 64rpx;
	height: 64rpx;
	border: 5rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: detail-spin 0.8s linear infinite;
}

@keyframes detail-spin {
	to { transform: rotate(360deg); }
}

/* ---- Error ---- */
.detail-error {
	position: fixed;
	top: 0; left: 0; right: 0; bottom: 0;
	background: rgba(255, 255, 255, 0.95);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 20rpx;
	z-index: 100;
}

.detail-error-text {
	font-size: 28rpx;
	color: var(--text-secondary);
}

.detail-retry-btn {
	height: 80rpx;
	padding: 0 60rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 40rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 8rpx 24rpx rgba(37, 99, 235, 0.3);

	.detail-retry-btn-text {
		font-size: 28rpx;
		font-weight: 700;
		color: #fff;
	}

	&:active { transform: scale(0.95); }
}
</style>
