<template>
	<view class="pl-root">
		<!-- Loading Skeleton -->
		<template v-if="loading && posts.length === 0">
			<view v-for="i in 3" :key="'skeleton-'+i" class="pl-card pl-card--skeleton">
				<view class="pl-card-header">
					<view class="pl-avatar pl-avatar--skeleton"></view>
					<view class="pl-author-info">
						<view class="pl-skeleton-line pl-skeleton-line--w30"></view>
						<view class="pl-skeleton-line pl-skeleton-line--w20 pl-skeleton-line--mt6"></view>
					</view>
				</view>
				<view class="pl-card-body">
					<view class="pl-skeleton-line pl-skeleton-line--w80"></view>
					<view class="pl-skeleton-line pl-skeleton-line--w100 pl-skeleton-line--mt16"></view>
					<view class="pl-skeleton-line pl-skeleton-line--w55 pl-skeleton-line--mt12"></view>
				</view>
			</view>
		</template>

		<!-- Post Items -->
		<view
			v-for="(post, index) in posts"
			:key="post.id || index"
			class="pl-card"
			:style="{ animationDelay: (index * 0.06) + 's' }"
			@tap="$emit('postClick', post)"
		>
			<!-- Card Header -->
			<view class="pl-card-header">
				<UserAvatar
					class="pl-avatar"
					:src="post.author?.avatar"
					:name="post.author?.realname || post.author?.nickname"
					:size="80"
					@tap.stop="$emit('clickAuthor', post.author)"
					@error="e => e.target && (e.target.src = '')"
				></UserAvatar>
				<view class="pl-author-info" @tap.stop="$emit('clickAuthor', post.author)">
					<view class="pl-author-row">
						<text class="pl-author-name">{{ post.author?.realname || post.author?.nickname || '匿名用户' }}</text>
						<view v-if="post.isOfficial" class="pl-official-badge">
							<l-icon name="check-circle-filled" size="12" color="#fff"></l-icon>
							<text class="pl-official-text">官方</text>
						</view>
					</view>
					<text class="pl-post-time">{{ formatTime(post.publishedAt || post.createdAt) }}</text>
				</view>
				<view v-if="post.isTop" class="pl-top-badge">
					<l-icon name="pin" size="18" color="var(--warning-color)"></l-icon>
				</view>
			</view>

			<!-- Card Body -->
			<view class="pl-card-body">
				<text class="pl-post-title">{{ post.title }}</text>
				<text v-if="post.summary" class="pl-post-summary">{{ post.summary }}</text>

				<!-- Images -->
				<view v-if="post.images && post.images.length > 0" class="pl-images">
					<view class="pl-images-grid" :class="getImageGridClass(post.images.length)">
						<image
							v-for="(image, imgIndex) in post.images.slice(0, 9)"
							:key="imgIndex"
							class="pl-image"
							:src="image"
							mode="aspectFill"
							lazy-load
							@tap.stop="previewImage(post.images, imgIndex)"
						></image>
					</view>
				</view>

				<!-- Tags -->
				<view class="pl-tags">
					<view class="pl-type-tag" :class="getTypeClass(post.type)">
						{{ getTypeText(post.type) }}
					</view>
					<text v-if="post.club?.name" class="pl-club-tag">{{ post.club.name }}</text>
				</view>
			</view>

			<!-- Card Footer -->
			<view class="pl-card-footer">
				<view class="pl-stats">
					<view class="pl-stat-item">
						<l-icon name="browse-filled" size="14" color="var(--text-tertiary)"></l-icon>
						<text class="pl-stat-text">{{ formatNumber(post.viewsCount || 0) }}</text>
					</view>
					<view class="pl-stat-item">
						<l-icon name="chat" size="14" color="var(--text-tertiary)"></l-icon>
						<text class="pl-stat-text">{{ formatNumber(post.commentsCount || 0) }}</text>
					</view>
				</view>
				<view class="pl-actions">
					<view
						class="pl-action-btn"
						:class="{ 'pl-action-btn--liked': post.isLiked }"
						@tap.stop="$emit('likePost', post)"
					>
						<l-icon
							:name="post.isLiked ? 'heart-filled' : 'heart'"
							size="18"
							:color="post.isLiked ? 'var(--error-color)' : 'var(--text-secondary)'"
						></l-icon>
						<text class="pl-action-text" :class="{ 'pl-action-text--liked': post.isLiked }">{{ formatNumber(post.likesCount || 0) }}</text>
					</view>
					<view class="pl-action-btn" @tap.stop="sharePost(post)">
						<l-icon name="share" size="18" color="var(--text-secondary)"></l-icon>
					</view>
				</view>
			</view>
		</view>

		<!-- Loading More -->
		<view v-if="loading && posts.length > 0" class="pl-loading-more">
			<view class="pl-loading-spinner"></view>
			<text class="pl-loading-text">加载中...</text>
		</view>

		<!-- Empty State -->
		<view v-if="!loading && posts.length === 0" class="pl-empty-state">
			<view class="pl-empty-illustration">
				<view class="pl-empty-circle pl-empty-circle--1"></view>
				<view class="pl-empty-circle pl-empty-circle--2"></view>
				<view class="pl-empty-icon-center">
					<l-icon name="file-text" size="52" color="var(--primary-400)"></l-icon>
				</view>
			</view>
			<text class="pl-empty-title">暂无相关动态</text>
			<text class="pl-empty-subtitle">快来发布第一条内容吧</text>
		</view>
	</view>
</template>

<script setup>
import UserAvatar from '@/components/UserAvatar.vue'
import { useTimeFormat } from '@/composables/useTimeFormat.js'
import lIcon from '@/uni_modules/lime-icon/components/l-icon/l-icon.vue'

defineProps({
	loading: { type: Boolean, default: false },
	posts: { type: Array, default: () => [] }
})

defineEmits(['postClick', 'likePost', 'clickAuthor'])

const { formatTime, formatNumber } = useTimeFormat()

function getTypeText(type) {
	const typeMap = { 'article': '文章', 'announcement': '公告', 'activity': '校园活动' }
	return typeMap[type] || '文章'
}

function getTypeClass(type) { return `pl-type-${type || 'article'}` }

function getImageGridClass(count) {
	if (count === 1) return 'pl-grid-1'
	if (count === 2) return 'pl-grid-2'
	if (count <= 4) return 'pl-grid-4'
	return 'pl-grid-9'
}

function previewImage(images, current) {
	const idx = typeof current === 'number' ? current : images.indexOf(current)
	uni.previewImage({ urls: images, current: images[idx] || current })
}

function sharePost(post) {
	// #ifdef MP-WEIXIN
	uni.vibrateShort()
	uni.showToast({ title: '点击右上角分享', icon: 'none' })
	// #endif
	// #ifndef MP-WEIXIN
	uni.setClipboardData({ data: `【${post.title}】 查看详情`, success: () => uni.showToast({ title: '链接已复制', icon: 'success' }) })
	// #endif
}
</script>

<style lang="scss" scoped>
.pl-root {
	padding: 16rpx 24rpx;
}

/* Skeleton */
.pl-card--skeleton {
	.pl-skeleton-line {
		height: 22rpx;
		background: linear-gradient(90deg, var(--bg-muted) 25%, rgba(148, 163, 184, 0.12) 50%, var(--bg-muted) 75%);
		background-size: 200% 100%;
		border-radius: 8rpx;
		animation: pl-shimmer 1.5s infinite;
		&--w20 { width: 40%; }
		&--w30 { width: 55%; }
		&--w55 { width: 55%; }
		&--w80 { width: 80%; }
		&--w100 { width: 100%; }
		&--mt6 { margin-top: 6rpx; }
		&--mt12 { margin-top: 12rpx; }
		&--mt16 { margin-top: 16rpx; }
	}
}

@keyframes pl-shimmer {
	0% { background-position: 200% 0; }
	100% { background-position: -200% 0; }
}

/* Card */
.pl-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 28rpx;
	margin-bottom: 20rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	transition: all 0.2s var(--ease-out);
	animation: pl-fadeIn 0.4s ease-out both;

	&:active {
		transform: scale(0.99);
		background: #f8fafc;
	}
}

@keyframes pl-fadeIn {
	from { opacity: 0; transform: translateY(10rpx); }
	to { opacity: 1; transform: translateY(0); }
}

/* Card Header */
.pl-card-header {
	display: flex;
	align-items: center;
	margin-bottom: 20rpx;
}

.pl-avatar {
	width: 80rpx;
	height: 80rpx;
	border-radius: 50%;
	background: var(--bg-muted);
	flex-shrink: 0;
	border: 3rpx solid rgba(148, 163, 184, 0.12);
}

.pl-author-info {
	flex: 1;
	margin-left: 16rpx;
	overflow: hidden;
}

.pl-author-row {
	display: flex;
	align-items: center;
	gap: 8rpx;
}

.pl-author-name {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.pl-official-badge {
	display: flex;
	align-items: center;
	gap: 4rpx;
	padding: 4rpx 10rpx;
	border-radius: 8rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	flex-shrink: 0;
}

.pl-official-text {
	font-size: 18rpx;
	color: #fff;
	font-weight: 600;
}

.pl-post-time {
	font-size: 22rpx;
	color: var(--text-tertiary);
	margin-top: 4rpx;
	display: block;
}

.pl-top-badge { flex-shrink: 0; }

/* Card Body */
.pl-card-body { margin-bottom: 20rpx; }

.pl-post-title {
	font-size: 32rpx;
	font-weight: 800;
	color: var(--text-primary);
	line-height: 1.4;
	margin-bottom: 8rpx;
	display: block;
}

.pl-post-summary {
	font-size: 26rpx;
	color: var(--text-secondary);
	line-height: 1.6;
	display: -webkit-box;
	-webkit-box-orient: vertical;
	-webkit-line-clamp: 3;
	line-clamp: 3;
	overflow: hidden;
	margin-bottom: 12rpx;
}

/* Images */
.pl-images { margin-top: 12rpx; }

.pl-images-grid {
	display: grid;
	gap: 8rpx;

	&.pl-grid-1 .pl-image { width: 100%; height: 360rpx; border-radius: 16rpx; }
	&.pl-grid-2 { grid-template-columns: repeat(2, 1fr); .pl-image { height: 260rpx; border-radius: 16rpx; } }
	&.pl-grid-4 { grid-template-columns: repeat(2, 1fr); .pl-image { height: 200rpx; border-radius: 16rpx; } }
	&.pl-grid-9 { grid-template-columns: repeat(3, 1fr); .pl-image { height: 160rpx; border-radius: 12rpx; } }

	.pl-image { width: 100%; object-fit: cover; }
}

/* Tags */
.pl-tags {
	display: flex;
	align-items: center;
	gap: 12rpx;
	margin-top: 12rpx;
}

.pl-type-tag {
	padding: 4rpx 14rpx;
	border-radius: 100rpx;
	font-size: 22rpx;
	font-weight: 600;
}

.pl-type-article { background: rgba(99, 102, 241, 0.1); color: var(--primary-500); }
.pl-type-announcement { background: rgba(239, 68, 68, 0.1); color: var(--error-color); }
.pl-type-activity { background: rgba(16, 185, 129, 0.1); color: var(--success-color); }

.pl-club-tag {
	padding: 4rpx 14rpx;
	border-radius: 100rpx;
	font-size: 22rpx;
	background: rgba(148, 163, 184, 0.1);
	color: var(--text-secondary);
	font-weight: 500;
}

/* Card Footer */
.pl-card-footer {
	display: flex;
	justify-content: space-between;
	align-items: center;
	border-top: 1px solid rgba(226, 232, 240, 0.6);
	padding-top: 16rpx;
}

.pl-stats { display: flex; gap: 24rpx; }

.pl-stat-item { display: flex; align-items: center; gap: 6rpx; }

.pl-stat-text { font-size: 22rpx; color: var(--text-tertiary); }

.pl-actions { display: flex; gap: 8rpx; }

.pl-action-btn {
	display: flex;
	align-items: center;
	gap: 6rpx;
	padding: 8rpx 16rpx;
	border-radius: 100rpx;
	transition: all 0.15s ease;

	&:active { background: var(--bg-muted); }
	&--liked { color: var(--error-color); }
}

.pl-action-text {
	font-size: 22rpx;
	color: var(--text-secondary);
	&--liked { color: var(--error-color); }
}

/* Loading More */
.pl-loading-more {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 32rpx 0;
	gap: 12rpx;
}

.pl-loading-spinner {
	width: 48rpx;
	height: 48rpx;
	border: 4rpx solid rgba(99, 102, 241, 0.12);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: pl-spin 0.8s linear infinite;
}

@keyframes pl-spin { to { transform: rotate(360deg); } }

.pl-loading-text { font-size: 24rpx; color: var(--text-tertiary); }

/* Empty State */
.pl-empty-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 80rpx 0;
	gap: 16rpx;
}

.pl-empty-illustration {
	position: relative;
	width: 180rpx;
	height: 180rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 8rpx;
}

.pl-empty-circle {
	position: absolute;
	border-radius: 50%;
	&--1 { width: 180rpx; height: 180rpx; background: rgba(59, 102, 241, 0.05); animation: pl-emp-pulse 3s ease-in-out infinite; }
	&--2 { width: 130rpx; height: 130rpx; background: rgba(59, 102, 241, 0.08); animation: pl-emp-pulse 3s ease-in-out infinite 0.6s; }
}

@keyframes pl-emp-pulse {
	0%, 100% { transform: scale(1); opacity: 1; }
	50% { transform: scale(1.06); opacity: 0.7; }
}

.pl-empty-icon-center {
	position: relative;
	z-index: 1;
	width: 110rpx;
	height: 110rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, rgba(59, 102, 241, 0.1), rgba(59, 102, 241, 0.04));
	border: 1px solid rgba(59, 102, 241, 0.15);
	display: flex;
	align-items: center;
	justify-content: center;
}

.pl-empty-title {
	font-size: 32rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.pl-empty-subtitle {
	font-size: 26rpx;
	color: var(--text-tertiary);
}
</style>
