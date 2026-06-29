<template>
	<view class="bm-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="bm-hero">
			<view class="bm-hero-bg"></view>
			<view class="bm-hero-overlay"></view>

			<!-- 装饰性圆形 -->
			<view class="bm-hero-orb bm-hero-orb--1"></view>
			<view class="bm-hero-orb bm-hero-orb--2"></view>
			<view class="bm-hero-orb bm-hero-orb--3"></view>

			<!-- 装饰性斜线 -->
			<view class="bm-hero-slash bm-hero-slash--1"></view>
			<view class="bm-hero-slash bm-hero-slash--2"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

				<view class="bm-hero-nav">
				<view class="bm-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="bm-hero-title">我的收藏</text>
				<view class="bm-hero-actions">
					<view v-if="posts.length > 0" class="bm-count-badge">
						<text class="bm-count-text">{{ total }}</text>
					</view>
				</view>
			</view>

			<view class="bm-hero-content">
				<view class="bm-hero-stat">
					<view class="bm-hero-stat-icon">
						<l-icon name="bookmark" style="font-size: 20px; color: #fff;"></l-icon>
					</view>
					<view class="bm-hero-stat-info">
						<text class="bm-hero-stat-num">{{ total }}</text>
						<text class="bm-hero-stat-label">条收藏内容</text>
					</view>
				</view>
				<text class="bm-hero-hint">长按卡片可快速取消收藏</text>
			</view>
		</view>

		<!-- 列表区域 -->
		<scroll-view
			class="bm-scroll"
			scroll-y
			@scrolltolower="loadMore"
			:refresher-enabled="true"
			:refresher-triggered="refreshing"
			@refresherrefresh="onRefresh"
		>
			<view class="bm-content">
				<!-- 加载骨架屏 -->
				<view v-if="loading && posts.length === 0" class="bm-skeleton-list">
					<view class="bm-skeleton-card" v-for="i in 4" :key="i">
						<view class="bm-skeleton-title"></view>
						<view class="bm-skeleton-lines">
							<view class="bm-skeleton-line"></view>
							<view class="bm-skeleton-line bm-skeleton-line--short"></view>
						</view>
						<view class="bm-skeleton-images">
							<view class="bm-skeleton-img" v-for="j in 3" :key="j"></view>
						</view>
						<view class="bm-skeleton-meta">
							<view class="bm-skeleton-avatar"></view>
							<view class="bm-skeleton-chip"></view>
						</view>
					</view>
				</view>

				<!-- 空状态 -->
				<view v-else-if="!loading && posts.length === 0" class="bm-empty-state">
					<view class="bm-empty-illustration">
						<view class="bm-empty-circle">
							<l-icon name="no-result" style="font-size: 56px; color: var(--primary-400);"></l-icon>
						</view>
						<view class="bm-empty-ring bm-empty-ring--1"></view>
						<view class="bm-empty-ring bm-empty-ring--2"></view>
					</view>
					<text class="bm-empty-title">暂无收藏</text>
					<text class="bm-empty-desc">遇到喜欢的内容，点一下收藏就会出现在这里</text>
					<view class="bm-empty-btn" @tap="goToExplore">
						<l-icon name="compass" style="font-size: 18px; margin-right: 8px; color: #fff;"></l-icon>
						<text class="bm-empty-btn-text">去逛逛</text>
					</view>
				</view>

				<!-- 收藏卡片列表 -->
				<view v-else class="bm-list">
					<view
						class="bm-card"
						v-for="(post, index) in posts"
						:key="post.id"
						:style="{ animationDelay: (index * 60) + 'ms' }"
						@tap="goToPostDetail(post)"
					>
						<!-- 卡片顶部：类型标签 + 书签按钮 -->
						<view class="bm-card-top">
							<view class="bm-card-meta-row">
								<view class="bm-card-type" :class="getTypeClass(post.type)">
									<l-icon :name="getTypeIcon(post.type)" style="font-size: 11px; margin-right: 4rpx;"></l-icon>
									<text>{{ getTypeLabel(post.type) }}</text>
								</view>
								<text class="bm-card-time">{{ formatTime(post.createdAt) }}</text>
							</view>
						</view>

						<!-- 标题 -->
						<text class="bm-card-title">{{ post.title }}</text>

						<!-- 内容摘要 -->
						<text v-if="post.content" class="bm-card-summary">{{ getPostSummary(post.content) }}</text>

						<!-- 图片网格 -->
						<view v-if="getPostImages(post.images).length > 0" class="bm-card-images" :class="getImagesColClass(post.images)">
							<image
								v-for="(image, idx) in getPostImages(post.images).slice(0, 3)"
								:key="idx"
								:src="image"
								class="bm-card-image"
								mode="aspectFill"
								@error="onImageError"
							/>
							<view v-if="getPostImages(post.images).length > 3" class="bm-card-image-more">
								<text class="bm-card-image-more-text">+{{ getPostImages(post.images).length - 3 }}</text>
							</view>
						</view>

						<!-- 卡片底部 -->
						<view class="bm-card-footer">
							<view class="bm-card-author">
								<UserAvatar
									:src="post.author?.avatar"
									:name="post.author?.realname || post.author?.nickname || post.author?.username"
									:size="56"
									class="bm-card-avatar"
								></UserAvatar>
								<view class="bm-card-author-info">
									<text class="bm-card-author-name">{{ post.author?.realname || post.author?.nickname || post.author?.username }}</text>
									<view class="bm-card-stats">
										<view class="bm-card-stat">
											<l-icon name="heart" style="font-size: 12px; color: var(--like-color); margin-right: 3rpx;"></l-icon>
											<text class="bm-card-stat-text">{{ formatNumber(post.likesCount) }}</text>
										</view>
										<view class="bm-card-stat">
											<l-icon name="chat" style="font-size: 12px; color: var(--comment-color); margin-right: 3rpx;"></l-icon>
											<text class="bm-card-stat-text">{{ formatNumber(post.commentsCount) }}</text>
										</view>
									</view>
								</view>
							</view>
							<view class="bm-card-bookmark" @tap.stop>
								<BookmarkButton
									:post-id="post.id"
									:initial-bookmarked="true"
									@bookmark-changed="onBookmarkChanged"
								/>
							</view>
						</view>

						<!-- 卡片左边装饰条 -->
						<view class="bm-card-accent" :class="getTypeClass(post.type)"></view>
					</view>
				</view>

				<!-- 加载更多 -->
				<view v-if="hasMore && posts.length > 0" class="bm-load-more">
					<view v-if="loadingMore" class="bm-load-more-spinner"></view>
					<text class="bm-load-more-text">{{ loadingMore ? '加载更多...' : '上拉加载更多' }}</text>
				</view>

				<!-- 到底了 -->
				<view v-if="!hasMore && posts.length > 0" class="bm-bottom-tip">
					<view class="bm-bottom-line"></view>
					<text class="bm-bottom-text">已展示全部收藏</text>
					<view class="bm-bottom-line"></view>
				</view>

				<!-- 底部安全区 -->
				<view class="bm-safe-bottom"></view>
			</view>
		</scroll-view>

		<!-- 社区服务须知弹窗 -->
		<CommunityTermsModal
			v-model:visible="termsVisible"
			@agree="onTermsAgreed"
			@disagree="onTermsDisagree"
		/>
	</view>
</template>

<script setup>
import UserAvatar from '@/components/UserAvatar.vue'
import { ref, onMounted } from 'vue'
import { getBookmarksList, formatTime, formatNumber, processImages } from '../api/community.js'
import BookmarkButton from '../../components/BookmarkButton.vue'
import { onShow } from '@dcloudio/uni-app'
import { hasAgreedToTerms } from '@/composables/useCommunityTerms.js'
import CommunityTermsModal from '@/components/CommunityTermsModal.vue'

const statusBarHeight = ref(20)
const termsVisible = ref(false)

function initStatusBar() {
	try {
		const systemInfo = uni.getSystemInfoSync()
		statusBarHeight.value = systemInfo.statusBarHeight || 20
	} catch (e) {
		statusBarHeight.value = 20
	}
}

const posts = ref([])
const total = ref(0)
const loading = ref(false)
const refreshing = ref(false)
const loadingMore = ref(false)
const hasMore = ref(true)
const page = ref(1)
const pageSize = ref(10)
const totalPages = ref(1)
const hasLoadedOnce = ref(false)

function goBack() {
	uni.navigateBack()
}

async function loadBookmarks(isRefresh = false) {
	if (loading.value && !isRefresh) return
	loading.value = true

	try {
		const response = await getBookmarksList({
			page: isRefresh ? 1 : page.value,
			pageSize: pageSize.value
		})

		if (response.success) {
			const result = response.result || {}
			const pagination = result.pagination || {}
			const newPosts = result.posts || []

			if (isRefresh) {
				posts.value = newPosts
				page.value = 2
			} else {
				posts.value.push(...newPosts)
				page.value++
			}

			total.value = pagination.total || total.value
			totalPages.value = pagination.totalPages || totalPages.value
			const currentPage = pagination.page || (isRefresh ? 1 : page.value - 1)
			hasMore.value = currentPage < totalPages.value
			hasLoadedOnce.value = true
		} else {
			throw new Error(response.message || '获取收藏列表失败')
		}
	} catch (error) {
		console.error('加载收藏列表失败:', error)
		uni.showToast({ title: error.message || '加载失败', icon: 'none' })
	} finally {
		loading.value = false
		refreshing.value = false
		loadingMore.value = false
	}
}

function refreshBookmarks() {
	loadBookmarks(true)
}

function onRefresh() {
	refreshing.value = true
	refreshBookmarks()
}

function loadMore() {
	if (hasMore.value && !loading.value && !loadingMore.value) {
		loadingMore.value = true
		loadBookmarks()
	}
}

function onBookmarkChanged(data) {
	if (!data.isBookmarked) {
		posts.value = posts.value.filter(post => post.id !== data.postId)
		total.value = Math.max(0, total.value - 1)
	}
}

function goToPostDetail(post) {
	uni.navigateTo({ url: `/pages/community/post-detail?id=${post.id}` })
}

function goToExplore() {
	uni.switchTab({ url: '/pages/community/index' })
}

function getPostSummary(content) {
	if (!content) return ''
	return content.length > 100 ? content.substring(0, 100) + '...' : content
}

function getPostImages(images) {
	return processImages(images)
}

function getImagesColClass(images) {
	return 'bm-card-images--' + Math.min(getPostImages(images).length, 3)
}

function onImageError(e) {
	console.error('图片加载失败:', e)
}

function getTypeLabel(type) {
	const map = { activity: '活动', article: '文章', help: '求助', lost: '失物' }
	return map[type] || '内容'
}

function getTypeIcon(type) {
	const map = { activity: 'activity', article: 'book', help: 'help-circle', lost: 'search' }
	return map[type] || 'bookmark'
}

function getTypeClass(type) {
	const map = {
		activity: 'bm-type--activity',
		article: 'bm-type--article',
		help: 'bm-type--help',
		lost: 'bm-type--lost'
	}
	return map[type] || 'bm-type--default'
}

onMounted(async () => {
	initStatusBar()
	const agreed = await hasAgreedToTerms()
	if (!agreed) {
		termsVisible.value = true
		return
	}
	loadBookmarks()
})

function onTermsAgreed() {
	termsVisible.value = false
	loadBookmarks()
}

function onTermsDisagree() {
	termsVisible.value = false
	uni.switchTab({ url: '/pages/index/index' })
}

onShow(() => {
	if (hasLoadedOnce.value) refreshBookmarks()
})
</script>

<style lang="scss" scoped>
/* ============================================
   Bookmark List - Optimized Hero Style
   ============================================ */

.bm-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
}

/* ---- Hero Section ---- */
.bm-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 280rpx;
	padding-bottom: 24rpx;
}

/* 渐变背景 */
.bm-hero-bg {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(160deg,
		#1e3a8a 0%,
		#1e40af 20%,
		#2563eb 50%,
		#3b82f6 75%,
		#60a5fa 100%);
	z-index: 0;
}

/* 遮罩层 */
.bm-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: radial-gradient(ellipse 120% 80% at 30% 60%,
		rgba(147, 197, 253, 0.18) 0%,
		transparent 60%);
	z-index: 1;
}

/* 装饰性光点 */
.bm-hero-orb {
	position: absolute;
	border-radius: 50%;
	z-index: 1;
	pointer-events: none;
}

.bm-hero-orb--1 {
	width: 200rpx;
	height: 200rpx;
	top: -60rpx;
	right: -40rpx;
	background: radial-gradient(circle, rgba(255,255,255,0.15) 0%, transparent 70%);
}

.bm-hero-orb--2 {
	width: 300rpx;
	height: 300rpx;
	bottom: -100rpx;
	left: -80rpx;
	background: radial-gradient(circle, rgba(147,197,253,0.2) 0%, transparent 70%);
}

.bm-hero-orb--3 {
	width: 160rpx;
	height: 160rpx;
	top: 80rpx;
	right: 120rpx;
	background: radial-gradient(circle, rgba(255,255,255,0.1) 0%, transparent 70%);
}

/* 装饰性斜线 */
.bm-hero-slash {
	position: absolute;
	z-index: 1;
	pointer-events: none;
	border-radius: 4rpx;
	opacity: 0.12;
}

.bm-hero-slash--1 {
	width: 400rpx;
	height: 3rpx;
	top: 120rpx;
	right: -80rpx;
	background: linear-gradient(90deg, transparent, rgba(255,255,255,0.8), transparent);
	transform: rotate(-25deg);
}

.bm-hero-slash--2 {
	width: 300rpx;
	height: 2rpx;
	bottom: 60rpx;
	left: -60rpx;
	background: linear-gradient(90deg, transparent, rgba(255,255,255,0.6), transparent);
	transform: rotate(-15deg);
}

/* 导航栏 */
.bm-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.bm-back-btn {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.18);
	border-radius: 50%;
	backdrop-filter: blur(8px);
	-webkit-backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.25);

	&:active {
		background: rgba(255, 255, 255, 0.28);
		transform: scale(0.95);
	}
}

.bm-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
	text-shadow: 0 2rpx 8rpx rgba(0,0,0,0.15);
}

.bm-hero-actions {
	width: 64rpx;
	display: flex;
	align-items: center;
	justify-content: flex-end;
}

.bm-count-badge {
	min-width: 40rpx;
	height: 40rpx;
	padding: 0 10rpx;
	background: rgba(255,255,255,0.22);
	border-radius: 100rpx;
	display: flex;
	align-items: center;
	justify-content: center;

	.bm-count-text {
		font-size: 22rpx;
		font-weight: 800;
		color: #fff;
	}
}

/* Hero 内容区 */
.bm-hero-content {
	position: relative;
	z-index: 2;
	padding: 8rpx 32rpx 0;
}

.bm-hero-stat {
	display: flex;
	align-items: center;
	gap: 16rpx;
	margin-bottom: 12rpx;
}

.bm-hero-stat-icon {
	width: 72rpx;
	height: 72rpx;
	border-radius: 50%;
	background: rgba(255,255,255,0.2);
	border: 1px solid rgba(255,255,255,0.3);
	display: flex;
	align-items: center;
	justify-content: center;
	backdrop-filter: blur(4px);
}

.bm-hero-stat-info {
	display: flex;
	flex-direction: column;
	gap: 2rpx;
}

.bm-hero-stat-num {
	font-size: 52rpx;
	font-weight: 900;
	color: #fff;
	line-height: 1;
	letter-spacing: -1px;
	text-shadow: 0 2rpx 12rpx rgba(0,0,0,0.2);
}

.bm-hero-stat-label {
	font-size: 26rpx;
	color: rgba(255,255,255,0.8);
	font-weight: 500;
}

.bm-hero-hint {
	font-size: 22rpx;
	color: rgba(255,255,255,0.55);
}

/* ---- Scroll ---- */
.bm-scroll {
	flex: 1;
	min-height: 0;
}

.bm-content {
	padding: 24rpx;
	padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}

/* ---- Skeleton ---- */
.bm-skeleton-list {
	display: flex;
	flex-direction: column;
	gap: 20rpx;
}

.bm-skeleton-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 28rpx;
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.bm-skeleton-title {
	height: 36rpx;
	border-radius: 12rpx;
	background: linear-gradient(90deg, var(--bg-muted) 0%, #e8edf4 50%, var(--bg-muted) 100%);
	background-size: 200% 100%;
	animation: bm-shimmer 1.4s ease-in-out infinite;
	margin-bottom: 16rpx;
}

.bm-skeleton-lines {
	display: flex;
	flex-direction: column;
	gap: 12rpx;
	margin-bottom: 16rpx;
}

.bm-skeleton-line {
	height: 24rpx;
	border-radius: 12rpx;
	background: linear-gradient(90deg, var(--bg-muted) 0%, #e8edf4 50%, var(--bg-muted) 100%);
	background-size: 200% 100%;
	animation: bm-shimmer 1.4s ease-in-out infinite;

	&--short { width: 65%; animation-delay: 0.15s; }
}

.bm-skeleton-images {
	display: flex;
	gap: 12rpx;
	margin-bottom: 16rpx;
}

.bm-skeleton-img {
	width: 196rpx;
	height: 196rpx;
	border-radius: 16rpx;
	background: linear-gradient(90deg, var(--bg-muted) 0%, #e8edf4 50%, var(--bg-muted) 100%);
	background-size: 200% 100%;
	animation: bm-shimmer 1.4s ease-in-out infinite;
}

.bm-skeleton-meta {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.bm-skeleton-avatar {
	width: 48rpx;
	height: 48rpx;
	border-radius: 50%;
	background: linear-gradient(90deg, var(--bg-muted) 0%, #e8edf4 50%, var(--bg-muted) 100%);
	background-size: 200% 100%;
	animation: bm-shimmer 1.4s ease-in-out infinite;
}

.bm-skeleton-chip {
	width: 140rpx;
	height: 24rpx;
	border-radius: 100rpx;
	background: linear-gradient(90deg, var(--bg-muted) 0%, #e8edf4 50%, var(--bg-muted) 100%);
	background-size: 200% 100%;
	animation: bm-shimmer 1.4s ease-in-out infinite;
}

@keyframes bm-shimmer {
	0% { background-position: 200% 0; }
	100% { background-position: -200% 0; }
}

/* ---- Empty State ---- */
.bm-empty-state {
	padding: 60rpx 40rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	text-align: center;
}

.bm-empty-illustration {
	position: relative;
	width: 200rpx;
	height: 200rpx;
	margin-bottom: 32rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.bm-empty-circle {
	width: 140rpx;
	height: 140rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, var(--primary-50), var(--primary-100));
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 8rpx 32rpx rgba(59, 130, 246, 0.15);
}

.bm-empty-ring {
	position: absolute;
	border-radius: 50%;
	border: 1px solid var(--primary-200);
	opacity: 0.4;
}

.bm-empty-ring--1 {
	width: 180rpx;
	height: 180rpx;
	animation: bm-pulse-ring 3s ease-in-out infinite;
}

.bm-empty-ring--2 {
	width: 220rpx;
	height: 220rpx;
	animation: bm-pulse-ring 3s ease-in-out infinite 1.5s;
}

@keyframes bm-pulse-ring {
	0%, 100% { transform: scale(1); opacity: 0.3; }
	50% { transform: scale(1.05); opacity: 0.5; }
}

.bm-empty-title {
	font-size: 38rpx;
	font-weight: 800;
	color: var(--text-primary);
	margin-bottom: 12rpx;
}

.bm-empty-desc {
	font-size: 26rpx;
	color: var(--text-tertiary);
	line-height: 1.7;
	max-width: 480rpx;
}

.bm-empty-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	height: 88rpx;
	padding: 0 56rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 44rpx;
	margin-top: 36rpx;
	box-shadow: 0 8rpx 24rpx rgba(37, 99, 235, 0.35);

	.bm-empty-btn-text {
		font-size: 30rpx;
		font-weight: 700;
		color: #fff;
	}

	&:active {
		transform: scale(0.97);
		box-shadow: 0 4rpx 12rpx rgba(37, 99, 235, 0.3);
	}
}

/* ---- Card List ---- */
.bm-list {
	display: flex;
	flex-direction: column;
	gap: 20rpx;
}

/* ---- Card ---- */
.bm-card {
	position: relative;
	background: #fff;
	border-radius: 28rpx;
	padding: 28rpx;
	padding-left: 36rpx;
	border: 1px solid rgba(148, 163, 184, 0.1);
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.06);
	overflow: hidden;
	transition: all 0.22s var(--ease-out);
	animation: bm-card-enter 0.35s var(--ease-out) both;

	&:active {
		transform: scale(0.99);
		box-shadow: 0 2rpx 10rpx rgba(30, 64, 175, 0.08);
	}
}

@keyframes bm-card-enter {
	from {
		opacity: 0;
		transform: translateY(16rpx);
	}
	to {
		opacity: 1;
		transform: translateY(0);
	}
}

/* 卡片左边装饰条 */
.bm-card-accent {
	position: absolute;
	left: 0;
	top: 0;
	bottom: 0;
	width: 6rpx;
	border-radius: 0 4rpx 4rpx 0;

	&.bm-type--article { background: linear-gradient(180deg, var(--type-article), var(--primary-400)); }
	&.bm-type--activity { background: linear-gradient(180deg, var(--type-activity), #34d399); }
	&.bm-type--help { background: linear-gradient(180deg, var(--type-discussion), #a78bfa); }
	&.bm-type--lost { background: linear-gradient(180deg, var(--primary-500), var(--primary-300)); }
	&.bm-type--default { background: linear-gradient(180deg, var(--text-tertiary), var(--bg-muted)); }
}

/* 卡片顶部元信息行 */
.bm-card-top {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 16rpx;
}

.bm-card-meta-row {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

/* 类型标签 */
.bm-card-type {
	display: inline-flex;
	align-items: center;
	padding: 5rpx 14rpx;
	border-radius: 100rpx;
	font-size: 22rpx;
	font-weight: 700;
	letter-spacing: 0.3px;

	&.bm-type--article {
		color: #4338ca;
		background: linear-gradient(135deg, rgba(99, 102, 241, 0.12), rgba(99, 102, 241, 0.06));
		border: 1px solid rgba(99, 102, 241, 0.2);
	}
	&.bm-type--activity {
		color: #059669;
		background: linear-gradient(135deg, rgba(16, 185, 129, 0.12), rgba(16, 185, 129, 0.05));
		border: 1px solid rgba(16, 185, 129, 0.2);
	}
	&.bm-type--help {
		color: #d97706;
		background: linear-gradient(135deg, rgba(245, 158, 11, 0.12), rgba(245, 158, 11, 0.05));
		border: 1px solid rgba(245, 158, 11, 0.2);
	}
	&.bm-type--lost {
		color: #2563eb;
		background: linear-gradient(135deg, rgba(37, 99, 235, 0.12), rgba(37, 99, 235, 0.05));
		border: 1px solid rgba(37, 99, 235, 0.2);
	}
	&.bm-type--default {
		color: var(--text-secondary);
		background: var(--bg-muted);
		border: 1px solid var(--border-light);
	}
}

.bm-card-time {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

/* 标题 */
.bm-card-title {
	display: -webkit-box;
	-webkit-line-clamp: 2;
	line-clamp: 2;
	-webkit-box-orient: vertical;
	overflow: hidden;
	font-size: 32rpx;
	font-weight: 800;
	color: var(--text-primary);
	line-height: 1.45;
	margin-bottom: 12rpx;
}

/* 摘要 */
.bm-card-summary {
	display: -webkit-box;
	-webkit-line-clamp: 3;
	line-clamp: 3;
	-webkit-box-orient: vertical;
	overflow: hidden;
	font-size: 26rpx;
	color: var(--text-secondary);
	line-height: 1.65;
	margin-bottom: 16rpx;
}

/* ---- Images ---- */
.bm-card-images {
	display: flex;
	gap: 10rpx;
	position: relative;
	margin-bottom: 20rpx;

	&--1 .bm-card-image { width: 100%; height: 320rpx; }
	&--2 .bm-card-image { width: 50%; height: 240rpx; }
	&--3 .bm-card-image { width: 33.33%; height: 200rpx; }
}

.bm-card-image {
	border-radius: 16rpx;
	object-fit: cover;
	background: var(--bg-muted);
	flex: 1;
}

.bm-card-image-more {
	position: absolute;
	right: 0;
	top: 0;
	width: 33.33%;
	height: 200rpx;
	border-radius: 16rpx;
	background: rgba(0, 0, 0, 0.45);
	display: flex;
	align-items: center;
	justify-content: center;
	backdrop-filter: blur(2rpx);

	.bm-card-image-more-text {
		font-size: 32rpx;
		font-weight: 800;
		color: #fff;
		text-shadow: 0 2rpx 8rpx rgba(0,0,0,0.3);
	}
}

/* ---- Footer ---- */
.bm-card-footer {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding-top: 16rpx;
	border-top: 1rpx solid var(--border-secondary);
}

.bm-card-author {
	display: flex;
	align-items: center;
	gap: 12rpx;
	min-width: 0;
}

.bm-card-avatar {
	width: 44rpx !important;
	height: 44rpx !important;
	flex-shrink: 0;
}

.bm-card-author-info {
	display: flex;
	flex-direction: column;
	gap: 3rpx;
	min-width: 0;
}

.bm-card-author-name {
	font-size: 25rpx;
	font-weight: 700;
	color: var(--text-primary);
	max-width: 200rpx;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.bm-card-stats {
	display: flex;
	align-items: center;
	gap: 12rpx;
}

.bm-card-stat {
	display: flex;
	align-items: center;

	.bm-card-stat-text {
		font-size: 21rpx;
		color: var(--text-tertiary);
	}
}

.bm-card-bookmark {
	flex-shrink: 0;
}

/* ---- Load More ---- */
.bm-load-more {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 12rpx;
	padding: 32rpx 0;
}

.bm-load-more-spinner {
	width: 32rpx;
	height: 32rpx;
	border: 3rpx solid rgba(59, 130, 246, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: bm-spin 0.8s linear infinite;
}

@keyframes bm-spin {
	to { transform: rotate(360deg); }
}

.bm-load-more-text {
	font-size: 24rpx;
	color: var(--text-tertiary);
}

/* 底部提示 */
.bm-bottom-tip {
	display: flex;
	align-items: center;
	gap: 20rpx;
	padding: 8rpx 0 32rpx;
}

.bm-bottom-line {
	flex: 1;
	height: 1rpx;
	background: linear-gradient(90deg, transparent, var(--border-light), transparent);
}

.bm-bottom-text {
	font-size: 22rpx;
	color: var(--text-tertiary);
	white-space: nowrap;
}

/* 底部安全区 */
.bm-safe-bottom {
	height: env(safe-area-inset-bottom);
}
</style>
