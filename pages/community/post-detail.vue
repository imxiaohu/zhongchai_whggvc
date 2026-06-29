<template>
	<view class="post-detail-page">
		<!-- 顶部蓝色渐变 Hero 导航 -->
		<view class="pd-hero">
			<view class="pd-hero-bg"></view>
			<view class="pd-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="pd-hero-nav">
				<view class="pd-back-btn" @tap="handleBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="pd-hero-title">动态详情</text>
				<view class="pd-hero-actions">
					<view class="pd-hero-icon-btn" @tap="sharePost">
						<l-icon name="share" style="font-size: 20px; color: #fff;"></l-icon>
					</view>
					<view class="pd-hero-icon-btn" @tap="showMoreActions">
						<l-icon name="more" style="font-size: 20px; color: #fff;"></l-icon>
					</view>
				</view>
			</view>
		</view>

		<!-- 帖子正文 -->
		<scroll-view class="pd-scroll" scroll-y enable-back-to-top v-if="postDetail">
			<!-- 作者信息卡片 -->
			<view class="pd-author-card">
				<view class="pd-author-avatar-wrap" @tap="goToProfile(postDetail.author?.id)">
				<UserAvatar
					class="pd-author-avatar"
					:src="postDetail.author?.avatar"
					:name="postDetail.author?.realname || postDetail.author?.nickname"
					:size="120"
				></UserAvatar>
					<view v-if="postDetail.isOfficial" class="pd-author-avatar-badge">
						<l-icon name="check-circle-filled" style="font-size: 12px; color: #fff;"></l-icon>
					</view>
				</view>
				<view class="pd-author-info">
					<view class="pd-author-row">
						<text class="pd-author-name">{{ postDetail.author?.realname || postDetail.author?.nickname || '匿名用户' }}</text>
						<view v-if="postDetail.isOfficial" class="pd-official-badge">
							<text class="pd-official-badge-text">官方</text>
						</view>
					</view>
					<text class="pd-author-meta">{{ formatTime(postDetail.publishedAt) }} · {{ postDetail.club?.name || '校园广场' }}</text>
				</view>
				<view
					v-if="!isOwnPost"
					class="pd-follow-btn"
					:class="{ 'pd-follow-btn--followed': postDetail.author?.isFollowed }"
					@tap="handleFollowToggle"
				>
					<text class="pd-follow-btn-text">{{ postDetail.author?.isFollowed ? '已关注' : '关注' }}</text>
				</view>
			</view>

			<!-- 帖子内容卡片 -->
			<view class="pd-content-card">
				<text class="pd-title">{{ postDetail.title }}</text>
				<text class="pd-body">{{ postDetail.content }}</text>

				<!-- 图片展示 -->
				<view class="pd-images" v-if="postDetail.images?.length">
					<view class="pd-images-grid" :class="getImagesGridClass(postDetail.images)">
						<image
							v-for="(img, idx) in postDetail.images.slice(0, 9)"
							:key="idx"
							class="pd-image"
							:src="img"
							mode="aspectFill"
							lazy-load
							@tap="previewImage(idx)"
						></image>
					</view>
				</view>

				<!-- 统计信息 -->
				<view class="pd-stats">
					<view class="pd-stat-item">
						<l-icon name="browse-filled" style="font-size: 16px; color: var(--text-tertiary);"></l-icon>
						<text class="pd-stat-text">{{ postDetail.viewsCount || 0 }} 阅读</text>
					</view>
					<view class="pd-stat-divider"></view>
					<view class="pd-stat-item">
						<l-icon name="heart-filled" style="font-size: 16px; color: var(--error-color);"></l-icon>
						<text class="pd-stat-text">{{ postDetail.likesCount || 0 }} 点赞</text>
					</view>
					<view class="pd-stat-divider"></view>
					<view class="pd-stat-item">
						<l-icon name="chat" style="font-size: 16px; color: var(--text-tertiary);"></l-icon>
						<text class="pd-stat-text">{{ postDetail.commentsCount || 0 }} 评论</text>
					</view>
				</view>
			</view>

			<!-- 评论区 -->
			<view class="pd-comments-section">
				<view class="pd-comments-header">
					<text class="pd-comments-title">全部评论</text>
					<text class="pd-comments-count">({{ postDetail.commentsCount || 0 }})</text>
				</view>
				<hierarchical-comments
					:post-id="postId"
					@reply="handleCommentReply"
					@click-author="handleClickAuthor"
					ref="commentsComponent"
				/>
			</view>

			<view class="pd-bottom-space"></view>
		</scroll-view>

		<!-- 加载状态 -->
		<view v-if="loading" class="pd-loading">
			<view class="pd-loading-spinner"></view>
			<text class="pd-loading-text">加载中...</text>
		</view>

		<!-- 底部交互栏 -->
		<view class="pd-bottom-bar" v-if="postDetail">
			<view class="pd-comment-box" @tap="showCommentInput">
				<l-icon name="edit" style="font-size: 18px; color: var(--text-tertiary);"></l-icon>
				<text class="pd-comment-placeholder">说点什么吧...</text>
			</view>
			<view class="pd-actions-right">
				<view class="pd-action-btn" :class="{ 'pd-action-btn--liked': postDetail.isLiked }" @tap="handleLike">
					<l-icon
						:name="postDetail.isLiked ? 'heart-filled' : 'heart'"
						:style="`font-size: 22px; color: ${postDetail.isLiked ? 'var(--error-color)' : 'var(--text-secondary)'};`"
					></l-icon>
					<text class="pd-action-text" :class="{ 'pd-action-text--liked': postDetail.isLiked }">{{ postDetail.likesCount || 0 }}</text>
				</view>
				<view class="pd-action-btn" :class="{ 'pd-action-btn--bookmarked': postDetail.isBookmarked }" @tap="handleBookmark">
					<l-icon
						:name="postDetail.isBookmarked ? 'star-filled' : 'star'"
						:style="`font-size: 22px; color: ${postDetail.isBookmarked ? '#fbbf24' : 'var(--text-secondary)'};`"
					></l-icon>
				</view>
			</view>
		</view>

		<!-- 评论输入弹窗 -->
		<comment-input
			:post-id="postId"
			@success="onCommentSuccess"
			ref="commentInput"
		/>

		<!-- 社区服务须知弹窗 -->
		<CommunityTermsModal
			v-model:visible="termsVisible"
			@agree="onTermsAgreed"
			@disagree="onTermsDisagree"
		/>

		<!-- 举报弹窗 -->
		<ReportModal
			:visible="reportVisible"
			target-type="post"
			:target-id="postId"
			@close="reportVisible = false"
			@success="reportVisible = false"
		/>
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import UserAvatar from '@/components/UserAvatar.vue'
import HierarchicalComments from '@/components/HierarchicalComments.vue'
import CommentInput from '@/components/CommentInput.vue'
import ReportModal from '@/components/ReportModal.vue'
import { getPostDetail, likePost, unlikePost, toggleBookmark, processImages, normalizePostDetailResponse, deletePost } from '../api/community.js'
import { useTimeFormat } from '@/composables/useTimeFormat.js'
import { hasAgreedToTerms } from '@/composables/useCommunityTerms.js'
import CommunityTermsModal from '@/components/CommunityTermsModal.vue'
import {onShow} from '@dcloudio/uni-app'

const { formatTime } = useTimeFormat()

// 状态
const statusBarHeight = ref(20)
const postId = ref(null)
const postDetail = ref(null)
const loading = ref(true)
const commentsComponent = ref(null)
const commentInput = ref(null)
const termsVisible = ref(false)
const reportVisible = ref(false)

function initStatusBarHeight() {
	try {
		const systemInfo = uni.getSystemInfoSync()
		statusBarHeight.value = systemInfo.statusBarHeight || 20
	} catch (e) {
		statusBarHeight.value = 20
	}
}

function getImagesGridClass(images) {
	return 'pd-grid-' + Math.min(images?.length || 0, 3)
}

const isOwnPost = computed(() => {
	const userInfoStr = uni.getStorageSync('userInfo')
	if (!userInfoStr) return false
	try {
		const user = typeof userInfoStr === 'string' ? JSON.parse(userInfoStr) : userInfoStr
		const uid = user?.id ?? user?.userId ?? user?.openid
		const authorId = postDetail.value?.author?.id
		if (!uid || !authorId) return false
		return String(authorId) === String(uid)
	} catch (e) {
		return false
	}
})

async function loadDetail() {
	loading.value = true
	try {
		const res = await getPostDetail(postId.value)
		if (res && res.success) {
			const detail = normalizePostDetailResponse(res)
			if (detail) {
				detail.images = processImages(detail.images)
			}
			postDetail.value = detail
		} else {
			postDetail.value = null
		}
	} finally {
		loading.value = false
	}
}

function previewImage(index) {
	if (!postDetail.value || !postDetail.value.images || !postDetail.value.images.length) return
	uni.previewImage({ urls: postDetail.value.images, current: index })
}

async function handleLike() {
	if (!postDetail.value) return
	try {
		const res = postDetail.value.isLiked ? await unlikePost(postId.value) : await likePost(postId.value)
		if (res && res.success) {
			postDetail.value.isLiked = !postDetail.value.isLiked
			postDetail.value.likesCount = (postDetail.value.likesCount || 0) + (postDetail.value.isLiked ? 1 : -1)
		}
	} catch (error) {
		console.error('点赞失败:', error)
	}
}

async function handleBookmark() {
	if (!postDetail.value) return
	try {
		const res = await toggleBookmark(postId.value)
		if (res) {
			const isBookmarked = res?.isBookmarked ?? res?.result?.isBookmarked
			if (typeof isBookmarked === 'boolean') postDetail.value.isBookmarked = isBookmarked
			uni.showToast({ title: isBookmarked ? '收藏成功' : '已取消收藏', icon: 'none' })
		}
	} catch (error) {
		console.error('收藏失败:', error)
		uni.showToast({ title: '收藏失败', icon: 'none' })
	}
}

async function handleFollowToggle() {
	uni.showToast({ title: '暂不支持关注', icon: 'none' })
}

function showCommentInput() {
	commentInput.value?.show()
}

function handleCommentReply(payload) {
	if (!payload) {
		showCommentInput()
		return
	}
	commentInput.value?.show(payload.comment || null, payload.parentComment || null)
}

function onCommentSuccess() {
	if (postDetail.value) {
		postDetail.value.commentsCount = (postDetail.value.commentsCount || 0) + 1
	}
	commentsComponent.value?.refresh()
}

function sharePost() {
	// #ifdef MP-WEIXIN
	uni.showToast({ title: '点击右上角分享', icon: 'none' })
	// #endif
}

function showMoreActions() {
	const options = ['举报']
	if (isOwnPost.value) options.push('编辑', '删除')
	uni.showActionSheet({
		itemList: options,
		success: (res) => {
			const cmd = options[res.tapIndex]
			if (cmd === '删除') {
				handleDelete()
			} else if (cmd === '举报') {
				handleReport()
			} else if (cmd === '编辑') {
				handleEdit()
			}
		}
	})
}

async function handleDelete() {
	uni.showModal({
		title: '确认删除',
		content: '确定要删除这篇动态吗？删除后无法恢复。',
		confirmColor: '#e74c3c',
		success: async (res) => {
			if (!res.confirm) return
			uni.showLoading({ title: '删除中...' })
			try {
				const result = await deletePost(postId.value)
				uni.hideLoading()
				if (result && result.success) {
					uni.showToast({ title: '删除成功', icon: 'success' })
					setTimeout(() => {
						uni.navigateBack()
					}, 1500)
				} else {
					uni.showToast({ title: result?.message || '删除失败', icon: 'none' })
				}
			} catch (e) {
				uni.hideLoading()
				uni.showToast({ title: '删除失败', icon: 'none' })
			}
		}
	})
}

function handleEdit() {
	uni.navigateTo({
		url: `/pages/community/create-post?postId=${postId.value}`
	})
}

function handleReport() {
	reportVisible.value = true
}

function handleBack() {
	uni.navigateBack()
}

function goToProfile(userId) {
	if (!userId) return
	uni.navigateTo({
		url: `/pages/community/user-profile?id=${userId}`
	})
}

function handleClickAuthor(author) {
	if (!author?.id) return
	goToProfile(author.id)
}

function refreshPost(updatedId, updatedData) {
	if (updatedId && String(updatedId) !== String(postId.value)) return
	if (updatedData) {
		postDetail.value = { ...postDetail.value, ...updatedData }
	} else {
		loadDetail()
	}
}

onShow(() => {
	if (postId.value) loadDetail()
})

	onMounted(async () => {
	initStatusBarHeight()
	const pages = getCurrentPages()
	const currentPage = pages[pages.length - 1]
	const options = currentPage?.options || {}
	postId.value = options.id
	if (postId.value) {
		const agreed = await hasAgreedToTerms()
		if (!agreed) {
			termsVisible.value = true
		}
		loadDetail()
	}
})

function onTermsAgreed() {
	termsVisible.value = false
}

function onTermsDisagree() {
	termsVisible.value = false
	handleBack()
}
</script>

<style lang="scss" scoped>
/* ============================================
   Post Detail Page - Hero Style
   ============================================ */

.post-detail-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
	position: relative;
}

/* ---- Hero Section ---- */
.pd-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 120rpx;
}

.pd-hero-bg {
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

.pd-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.65) 0%,
		rgba(37, 99, 235, 0.4) 50%,
		rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.pd-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.pd-back-btn {
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

.pd-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.pd-hero-actions {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.pd-hero-icon-btn {
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

/* ---- Scroll Content ---- */
.pd-scroll {
	flex: 1;
	min-height: 0;
}

/* ---- Author Card ---- */
.pd-author-card {
	display: flex;
	align-items: center;
	padding: 24rpx;
	background: #fff;
	margin: 0 24rpx;
	margin-top: 30rpx;
	border-radius: 24rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1px solid rgba(148, 163, 184, 0.12);
	margin-bottom: 16rpx;
	position: relative;
	z-index: 1;
}

.pd-author-avatar-wrap {
	position: relative;
	flex-shrink: 0;
	margin-right: 16rpx;
}

.pd-author-avatar {
	width: 96rpx;
	height: 96rpx;
	border-radius: 50%;
	background: var(--bg-muted);
	border: 4rpx solid rgba(148, 163, 184, 0.2);
}

.pd-author-avatar-badge {
	position: absolute;
	bottom: 0;
	right: 0;
	width: 36rpx;
	height: 36rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	border: 3rpx solid #fff;
}

.pd-author-info {
	flex: 1;
	overflow: hidden;
}

.pd-author-row {
	display: flex;
	align-items: center;
	gap: 10rpx;
	margin-bottom: 6rpx;
}

.pd-author-name {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.pd-official-badge {
	padding: 3rpx 10rpx;
	border-radius: 8rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	flex-shrink: 0;

	.pd-official-badge-text {
		font-size: 18rpx;
		color: #fff;
		font-weight: 600;
	}
}

.pd-author-meta {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.pd-follow-btn {
	padding: 10rpx 28rpx;
	border-radius: 100rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	box-shadow: 0 4rpx 12rpx rgba(37, 99, 235, 0.25);
	flex-shrink: 0;
	transition: all 0.2s ease;

	.pd-follow-btn-text {
		font-size: 24rpx;
		font-weight: 700;
		color: #fff;
	}

	&--followed {
		background: var(--bg-muted);
		box-shadow: none;
		.pd-follow-btn-text { color: var(--text-tertiary); }
	}

	&:active:not(.pd-follow-btn--followed) {
		transform: scale(0.95);
	}
}

/* ---- Content Card ---- */
.pd-content-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 32rpx;
	margin: 0 24rpx;
	margin-bottom: 16rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.pd-title {
	display: block;
	font-size: 40rpx;
	font-weight: 800;
	color: var(--text-primary);
	line-height: 1.4;
	margin-bottom: 20rpx;
}

.pd-body {
	display: block;
	font-size: 30rpx;
	color: var(--text-secondary);
	line-height: 1.9;
	white-space: pre-wrap;
	margin-bottom: 24rpx;
}

/* ---- Images ---- */
.pd-images {
	margin-bottom: 24rpx;
}

.pd-images-grid {
	display: grid;
	gap: 8rpx;

	&.pd-grid-1 {
		grid-template-columns: 1fr;
		.pd-image { height: 400rpx; border-radius: 16rpx; }
	}
	&.pd-grid-2 {
		grid-template-columns: repeat(2, 1fr);
		.pd-image { height: 300rpx; border-radius: 16rpx; }
	}
	&.pd-grid-3 {
		grid-template-columns: repeat(3, 1fr);
		.pd-image { height: 200rpx; border-radius: 12rpx; }
	}

	.pd-image {
		width: 100%;
		object-fit: cover;
	}
}

/* ---- Stats ---- */
.pd-stats {
	display: flex;
	align-items: center;
	padding-top: 24rpx;
	border-top: 1px solid rgba(226, 232, 240, 0.8);
}

.pd-stat-item {
	display: flex;
	align-items: center;
	gap: 8rpx;
}

.pd-stat-text {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.pd-stat-divider {
	width: 1px;
	height: 24rpx;
	background: rgba(148, 163, 184, 0.3);
	margin: 0 24rpx;
}

/* ---- Comments Section ---- */
.pd-comments-section {
	background: #fff;
	border-radius: 24rpx;
	margin: 0 24rpx;
	padding: 32rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.pd-comments-header {
	display: flex;
	align-items: center;
	gap: 8rpx;
	margin-bottom: 24rpx;
}

.pd-comments-title {
	font-size: 32rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.pd-comments-count {
	font-size: 26rpx;
	color: var(--text-tertiary);
	font-weight: 500;
}

.pd-bottom-space {
	height: 160rpx;
}

/* ---- Loading ---- */
.pd-loading {
	position: fixed;
	top: 0; left: 0; right: 0; bottom: 0;
	background: rgba(255, 255, 255, 0.9);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 24rpx;
	z-index: 1000;
}

.pd-loading-spinner {
	width: 64rpx;
	height: 64rpx;
	border: 5rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: pd-spin 0.8s linear infinite;
}

@keyframes pd-spin {
	to { transform: rotate(360deg); }
}

.pd-loading-text {
	font-size: 28rpx;
	color: var(--text-secondary);
}

/* ---- Bottom Bar ---- */
.pd-bottom-bar {
	position: fixed;
	bottom: 0;
	left: 0;
	right: 0;
	background: rgba(255, 255, 255, 0.95);
	backdrop-filter: blur(20px);
	-webkit-backdrop-filter: blur(20px);
	padding: 16rpx 24rpx calc(16rpx + env(safe-area-inset-bottom));
	border-top: 1px solid rgba(148, 163, 184, 0.15);
	z-index: 100;
	display: flex;
	align-items: center;
	gap: 16rpx;
	box-shadow: 0 -4rpx 20rpx rgba(30, 64, 175, 0.06);
}

.pd-comment-box {
	flex: 1;
	height: 80rpx;
	background: var(--bg-muted);
	border-radius: 40rpx;
	display: flex;
	align-items: center;
	padding: 0 24rpx;
	gap: 12rpx;
	border: 1px solid rgba(148, 163, 184, 0.15);

	&:active { background: var(--border-secondary); }
}

.pd-comment-placeholder {
	font-size: 26rpx;
	color: var(--text-tertiary);
}

.pd-actions-right {
	display: flex;
	align-items: center;
	gap: 8rpx;
}

.pd-action-btn {
	display: flex;
	align-items: center;
	gap: 6rpx;
	padding: 12rpx 20rpx;
	border-radius: 100rpx;
	transition: all 0.15s ease;

	&:active { background: var(--bg-muted); }
}

.pd-action-text {
	font-size: 24rpx;
	color: var(--text-secondary);

	&--liked { color: var(--error-color); }
}
</style>
