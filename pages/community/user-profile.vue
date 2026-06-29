<!-- webpackChunkName: "community-user-profile" -->
<template>
	<view class="up-page">
		<!-- 顶部 Hero 区域 -->
		<view class="up-hero" :class="{ 'up-hero--skeleton': loading && !userProfile }">
			<view class="up-hero-bg" :style="heroBgStyle"></view>
			<view class="up-hero-overlay"></view>

			<!-- 状态栏占位 -->
			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<!-- 顶部导航栏 -->
			<view class="up-hero-nav">
				<view class="up-back-btn" @tap="handleBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="up-hero-title">{{ userProfile?.realname || userProfile?.nickname || '用户主页' }}</text>
				<view class="up-more-btn" @tap="showMoreActions">
					<l-icon name="more" style="font-size: 20px; color: #fff;"></l-icon>
				</view>
			</view>

			<!-- 用户信息区 -->
			<view class="up-user-section" v-if="userProfile">
				<view class="up-avatar-wrap">
					<UserAvatar
						class="up-avatar"
						:src="userProfile.avatar"
						:name="userProfile.realname || userProfile.nickname"
						:size="140"
					></UserAvatar>
					<view v-if="userProfile.isOfficial" class="up-avatar-badge">
						<l-icon name="check-circle-filled" style="font-size: 12px; color: #fff;"></l-icon>
					</view>
				</view>

				<view class="up-user-info">
					<view class="up-name-row">
						<text class="up-nickname">{{ userProfile.realname || userProfile.nickname || '匿名用户' }}</text>
						<view v-if="userProfile.isOfficial" class="up-official-badge">
							<text class="up-official-badge-text">官方认证</text>
						</view>
					</view>
					<text class="up-username" v-if="userProfile.username">@{{ userProfile.username }}</text>
					<text class="up-bio" v-if="userProfile.bio">{{ userProfile.bio }}</text>
				</view>

				<!-- 操作按钮 -->
				<view class="up-actions-row" v-if="userProfile">
					<!-- 非自己显示关注按钮 -->
					<view v-if="!isOwnProfile" class="up-follow-btn" :class="{ 'up-follow-btn--followed': userProfile.isFollowed, 'up-follow-btn--mutual': userProfile.isMutual }" @tap="handleFollowToggle">
						<l-icon :name="userProfile.isFollowed ? (userProfile.isMutual ? 'account-multiple' : 'check') : 'add'" size="16" color="#fff"></l-icon>
						<text class="up-follow-btn-text">{{ userProfile.isMutual ? '互相关注' : (userProfile.isFollowed ? '已关注' : '关注') }}</text>
					</view>
					<!-- 自己显示编辑按钮 -->
					<view v-else class="up-edit-btn" @tap="handleEditProfile">
						<l-icon name="edit" size="16" color="var(--primary-color)"></l-icon>
						<text class="up-edit-btn-text">编辑资料</text>
					</view>
					<view class="up-msg-btn" v-if="!isOwnProfile" @tap="handleSendMessage">
						<l-icon name="chat" size="16" color="var(--text-secondary)"></l-icon>
					</view>
				</view>
			</view>

			<!-- 骨架屏用户区 -->
			<view class="up-user-section up-user-section--skeleton" v-else-if="loading">
				<view class="up-avatar-skeleton"></view>
				<view class="up-info-skeleton">
					<view class="up-name-skeleton"></view>
					<view class="up-bio-skeleton"></view>
				</view>
			</view>
		</view>

		<!-- 统计栏 -->
		<view class="up-stats-bar" v-if="userProfile">
			<view class="up-stat-item" @tap="scrollToPosts">
				<text class="up-stat-num">{{ userProfile.postsCount || 0 }}</text>
				<text class="up-stat-label">发布</text>
			</view>
			<view class="up-stat-divider"></view>
			<view class="up-stat-item" @tap="scrollToFollowers">
				<text class="up-stat-num">{{ userProfile.followersCount || 0 }}</text>
				<text class="up-stat-label">粉丝</text>
			</view>
			<view class="up-stat-divider"></view>
			<view class="up-stat-item" @tap="scrollToFollowing">
				<text class="up-stat-num">{{ userProfile.followingCount || 0 }}</text>
				<text class="up-stat-label">关注</text>
			</view>
			<view class="up-stat-divider"></view>
			<view class="up-stat-item">
				<text class="up-stat-num">{{ userProfile.likesCount || 0 }}</text>
				<text class="up-stat-label">获赞</text>
			</view>
		</view>

		<!-- 内容区域 -->
		<scroll-view
			class="up-content"
			id="posts-section"
			scroll-y
			@scrolltolower="loadMore"
			:refresher-enabled="true"
			:refresher-triggered="postsRefreshing"
			@refresherrefresh="onPostsRefresh"
		>
			<!-- 内容类型切换 -->
			<view class="up-tabs">
				<view
					v-for="tab in tabs"
					:key="tab.key"
					class="up-tab"
					:class="{ 'up-tab--active': currentTab === tab.key }"
					@tap="switchTab(tab.key)"
				>
					<l-icon :name="tab.icon" size="18" :color="currentTab === tab.key ? 'var(--primary-color)' : 'var(--text-tertiary)'"></l-icon>
					<text class="up-tab-text">{{ tab.name }}</text>
					<view class="up-tab-indicator" v-if="currentTab === tab.key"></view>
				</view>
			</view>

			<!-- 加载中 -->
			<view v-if="loadingPosts" class="up-loading">
				<view class="up-spinner"></view>
				<text class="up-loading-text">加载中...</text>
			</view>

			<!-- 图文瀑布流 -->
			<view v-else-if="currentTab === 'grid' && userPosts.length > 0" class="up-posts-grid">
				<view
					v-for="post in userPosts"
					:key="post.id"
					class="up-post-grid-item"
					:class="{ 'up-post-grid-item--text': !hasGridCover(post) }"
					:style="!hasGridCover(post) ? { backgroundColor: getGridBgColor(post) } : {}"
					@tap="goToPostDetail(post)"
				>
					<image
						v-if="hasGridCover(post)"
						class="up-post-cover"
						:src="getPostCover(post)"
						mode="aspectFill"
						lazy-load
					></image>
					<text v-else class="up-post-text-preview">{{ getGridTextPreview(post) }}</text>
					<view class="up-post-overlay">
						<view class="up-post-meta">
							<l-icon name="heart-filled" size="12" color="#fff"></l-icon>
							<text class="up-post-stat-text">{{ post.likesCount || 0 }}</text>
							<l-icon name="chat" size="12" color="#fff"></l-icon>
							<text class="up-post-stat-text">{{ post.commentsCount || 0 }}</text>
						</view>
					</view>
				</view>
			</view>

			<!-- 列表模式 -->
			<view v-else-if="currentTab === 'list' && userPosts.length > 0" class="up-posts-list">
				<view
					v-for="post in userPosts"
					:key="post.id"
					class="up-post-list-item"
					@tap="goToPostDetail(post)"
				>
					<view class="up-post-list-content">
						<text class="up-post-list-title">{{ post.title }}</text>
						<text class="up-post-list-summary" v-if="post.summary">{{ post.summary }}</text>
						<view class="up-post-list-footer">
							<text class="up-post-list-time">{{ formatTime(post.publishedAt) }}</text>
							<view class="up-post-list-stats">
								<l-icon name="heart-filled" size="12" color="var(--error-color)"></l-icon>
								<text class="up-post-list-stat">{{ post.likesCount || 0 }}</text>
								<l-icon name="chat" size="12" color="var(--text-tertiary)"></l-icon>
								<text class="up-post-list-stat">{{ post.commentsCount || 0 }}</text>
							</view>
						</view>
					</view>
					<image
						v-if="post.images && post.images.length > 0"
						class="up-post-list-cover"
						:src="getPostCover(post)"
						mode="aspectFill"
					></image>
				</view>
			</view>

			<!-- 空状态 -->
			<view v-else-if="!loadingPosts" class="up-empty">
				<image class="up-empty-icon" src="/static/images/empty-posts.png" mode="aspectFit"></image>
				<text class="up-empty-text">{{ isOwnProfile ? '还没有发布过动态' : '暂无发布内容' }}</text>
				<view v-if="isOwnProfile" class="up-empty-action" @tap="goToCreatePost">
					<text class="up-empty-action-text">发布动态</text>
				</view>
			</view>

			<!-- 加载更多 -->
			<view class="up-load-more" v-if="userPosts.length > 0 && hasMore">
				<text class="up-load-more-text" @tap="loadMore">加载更多</text>
			</view>

			<view class="up-bottom-space"></view>
		</scroll-view>

		<!-- 悬浮发布按钮（仅自己） -->
		<view class="up-fab" v-if="isOwnProfile && currentTab !== 'skeleton'">
			<view class="up-fab-btn" @tap="goToCreatePost">
				<l-icon name="add" size="28" color="#fff"></l-icon>
			</view>
		</view>

		<!-- 社区服务须知弹窗 -->
		<CommunityTermsModal
			v-model:visible="termsVisible"
			@agree="onTermsAgreed"
			@disagree="onTermsDisagree"
		/>

		<!-- 举报用户弹窗 -->
		<ReportModal
			v-if="reportVisible"
			:visible="reportVisible"
			targetType="user"
			:targetId="userId"
			@close="reportVisible = false"
			@success="onReportSuccess"
		/>
	</view>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import UserAvatar from '@/components/UserAvatar.vue'
import ReportModal from '@/components/ReportModal.vue'
import { getUserProfile, getUserPosts, followUser, unfollowUser, processImages, formatTime as formatTimeUtil, blockUser, unblockUser, checkBlockStatus } from '../api/community.js'
import { hasAgreedToTerms } from '@/composables/useCommunityTerms.js'
import CommunityTermsModal from '@/components/CommunityTermsModal.vue'

// 状态
const statusBarHeight = ref(20)
const userId = ref(null)
const userProfile = ref(null)
const userPosts = ref([])
const loading = ref(false)
const loadingPosts = ref(false)
const postsRefreshing = ref(false)
const currentTab = ref('grid')
const tabs = [
	{ key: 'grid', name: '图文', icon: 'image' },
	{ key: 'list', name: '列表', icon: 'format-list-bulleted' }
]
const page = ref(1)
const pageSize = ref(20)
const hasMore = ref(false)
const currentUserId = ref(null)
const termsVisible = ref(false)
const reportVisible = ref(false)
const isBlocked = ref(false)

function initStatusBarHeight() {
	try {
		const systemInfo = uni.getSystemInfoSync()
		statusBarHeight.value = systemInfo.statusBarHeight || 20
	} catch (e) {
		statusBarHeight.value = 20
	}
}

function getCurrentUserId() {
	try {
		const userInfoStr = uni.getStorageSync('userInfo')
		if (!userInfoStr) return null
		const user = typeof userInfoStr === 'string' ? JSON.parse(userInfoStr) : userInfoStr
		return user?.id ?? user?.userId ?? user?.openid ?? null
	} catch (e) {
		return null
	}
}

const isOwnProfile = computed(() => {
	if (!userId.value || !currentUserId.value) return false
	return String(userId.value) === String(currentUserId.value)
})

const heroBgStyle = computed(() => {
	if (!userProfile.value) return {}
	const bgColor = userProfile.value.coverColor || userProfile.value.bgColor
	if (bgColor) {
		return { background: bgColor }
	}
	return {}
})

function formatTime(time) {
	return formatTimeUtil(time)
}

function getPostCover(post) {
	if (post.coverImage) return post.coverImage
	if (post.images && post.images.length > 0) return post.images[0]
	return null
}

const GRID_BG_COLORS = [
	'#DBEAFE', '#BFDBFE', '#93C5FD', '#A5F3FC',
	'#D1FAE5', '#BBF7D0', '#FEF3C7', '#FDE68A',
	'#E9D5FF', '#DDD6FE', '#FECACA', '#FED7AA',
	'#FEF08A', '#F9A8D4', '#C7D2FE', '#6EE7B7',
]

function getGridBgColor(post) {
	const id = post.id || post.title || Math.random()
	const hash = String(id).split('').reduce((acc, c) => acc + c.charCodeAt(0), 0)
	return GRID_BG_COLORS[hash % GRID_BG_COLORS.length]
}

function getGridTextPreview(post) {
	const title = post.title || ''
	const content = post.summary || post.content || ''
	if (title.length >= 40) return title.slice(0,60)
	if (title.length + content.length > 60) {
		return title + content.slice(0, 60 - title.length)
	}
	return title + content
}

function hasGridCover(post) {
	return !!(post.coverImage || (post.images && post.images.length > 0))
}

async function fetchUserProfile() {
	loading.value = true
	try {
		const res = await getUserProfile(userId.value)
		if (res && res.success && res.result) {
			userProfile.value = res.result
		} else if (res && res.result) {
			userProfile.value = res.result
		}
		// 检查是否屏蔽了该用户
		if (!isOwnProfile.value) {
			checkBlockStatusApi()
		}
	} catch (error) {
		console.error('获取用户信息失败:', error)
	} finally {
		loading.value = false
	}
}

async function checkBlockStatusApi() {
	try {
		const res = await checkBlockStatus(userId.value)
		if (res && res.result) {
			isBlocked.value = res.result.isBlocked || false
		}
	} catch (e) {
		// 忽略错误
	}
}

async function fetchUserPosts() {
	loadingPosts.value = true
	try {
		const res = await getUserPosts(userId.value, { page: page.value, pageSize: pageSize.value })
		const records = res?.result?.records || res?.result || []
		if (page.value === 1) {
			userPosts.value = records.map(p => ({
				...p,
				images: processImages(p.images)
			}))
		} else {
			userPosts.value = [...userPosts.value, ...records.map(p => ({
				...p,
				images: processImages(p.images)
			}))]
		}
		const total = res?.result?.total || res?.total || records.length
		hasMore.value = userPosts.value.length < total
	} catch (error) {
		console.error('获取用户帖子失败:', error)
	} finally {
		loadingPosts.value = false
		postsRefreshing.value = false
	}
}

async function onPostsRefresh() {
	page.value = 1
	await fetchUserPosts()
}

async function loadMore() {
	if (!hasMore.value || loadingPosts.value) return
	page.value++
	await fetchUserPosts()
}

function switchTab(key) {
	currentTab.value = key
}

function handleBack() {
	uni.navigateBack()
}

function goToPostDetail(post) {
	uni.navigateTo({ url: `/pages/community/post-detail?id=${post.id}` })
}

function goToCreatePost() {
	uni.navigateTo({ url: '/pages/community/create-post' })
}

function handleEditProfile() {
	uni.navigateTo({ url: '/pages/user/profile-edit' })
}

function handleSendMessage() {
	uni.showToast({ title: '私信功能开发中', icon: 'none' })
}

async function handleFollowToggle() {
	if (!userProfile.value) return
	try {
		const wasFollowed = userProfile.value.isFollowed
		const wasMutual = userProfile.value.isMutual
		const res = wasFollowed
			? await unfollowUser(userId.value)
			: await followUser(userId.value)
		if (res && res.success) {
			userProfile.value.isFollowed = !wasFollowed
			userProfile.value.followersCount = (userProfile.value.followersCount || 0) + (userProfile.value.isFollowed ? 1 : -1)
			userProfile.value.isMutual = false // 取消关注后互相关注状态必定消失，关注后需要重新获取才能确定是否互相关注
			uni.showToast({
				title: userProfile.value.isFollowed ? '关注成功' : '已取消关注',
				icon: 'success'
			})
			// 关注成功后重新获取个人信息以确认互相关注状态
			if (userProfile.value.isFollowed && !wasMutual) {
				fetchUserProfile()
			}
		}
	} catch (error) {
		console.error('关注操作失败:', error)
		uni.showToast({ title: '操作失败', icon: 'none' })
	}
}

function showReportModal() {
	reportVisible.value = true
}

function onReportSuccess() {
	reportVisible.value = false
	uni.showToast({ title: '举报已提交', icon: 'success' })
}

function showMoreActions() {
	let options
	if (isOwnProfile.value) {
		options = ['设置']
	} else if (isBlocked.value) {
		options = ['解除屏蔽', '举报用户']
	} else {
		options = ['屏蔽用户', '举报用户']
	}
	uni.showActionSheet({
		itemList: options,
		success: (res) => {
			const cmd = options[res.tapIndex]
			if (cmd === '举报用户') {
				showReportModal()
			} else if (cmd === '屏蔽用户') {
				handleBlockUser()
			} else if (cmd === '解除屏蔽') {
				handleUnblockUser()
			} else if (cmd === '设置') {
				uni.navigateTo({ url: '/pages/user/setting' })
			}
		}
	})
}

function handleBlockUser() {
	uni.showModal({
		title: '屏蔽用户',
		content: '确定要屏蔽此用户吗？屏蔽后你将无法看到对方的内容。',
		confirmColor: 'var(--primary-color)',
		success: async (res) => {
			if (res.confirm) {
				try {
					const result = await blockUser(userId.value)
					if (result && result.success) {
						uni.showToast({ title: '已屏蔽', icon: 'success' })
						setTimeout(() => handleBack(), 1000)
					}
				} catch (e) {
					uni.showToast({ title: '操作失败', icon: 'none' })
				}
			}
		}
	})
}

async function handleUnblockUser() {
	try {
		const res = await unblockUser(userId.value)
		if (res && res.success) {
			isBlocked.value = false
			uni.showToast({ title: '已解除屏蔽', icon: 'success' })
		}
	} catch (e) {
		uni.showToast({ title: '操作失败', icon: 'none' })
	}
}

function scrollToPosts() {
	nextTick(() => {
		uni.createSelectorQuery()
			.select('#posts-section')
			.boundingClientRect(data => {
				if (data) {
					uni.pageScrollTo({ scrollTop: data.top - 100, duration: 300 })
				}
			})
			.exec()
	})
}

function scrollToFollowers() {
	uni.navigateTo({ url: `/pages/community/follow-list?userId=${userId.value}&type=followers` })
}

function scrollToFollowing() {
	uni.navigateTo({ url: `/pages/community/follow-list?userId=${userId.value}&type=following` })
}

onMounted(async () => {
	initStatusBarHeight()
	currentUserId.value = getCurrentUserId()

	const pages = getCurrentPages()
	const currentPage = pages[pages.length - 1]
	const options = currentPage?.options || {}
	userId.value = options.id || options.userId

	if (userId.value) {
		const agreed = await hasAgreedToTerms()
		if (!agreed) {
			termsVisible.value = true
		}
		fetchUserProfile()
		fetchUserPosts()
	} else {
		uni.showToast({ title: '用户信息获取失败', icon: 'none' })
		setTimeout(() => handleBack(), 1500)
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
   Community User Profile Page - 小红书风格
   ============================================ */

.up-page {
	width: 100%;
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
	position: relative;
}

/* ---- Hero Section ---- */
.up-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 560rpx;
}

.up-hero-bg {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		#1e3a8a 0%,
		#1e40af 20%,
		#2563eb 40%,
		#3b82f6 60%,
		#60a5fa 80%,
		#93c5fd 100%);
	z-index: 0;
}

.up-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.5) 0%,
		rgba(37, 99, 235, 0.3) 40%,
		rgba(147, 197, 253, 0.05) 100%);
	z-index: 1;
}

/* ---- Nav ---- */
.up-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 24rpx;
	height: 88rpx;
}

.up-back-btn,
.up-more-btn {
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

.up-hero-title {
	font-size: 32rpx;
	font-weight: 700;
	color: #fff;
	max-width: 400rpx;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

/* ---- User Info ---- */
.up-user-section {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 32rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	margin-top: -40rpx;
}

.up-avatar-wrap {
	position: relative;
	margin-bottom: 16rpx;
}

.up-avatar {
	border: 6rpx solid #fff;
	box-shadow: 0 8rpx 32rpx rgba(0, 0, 0, 0.15);
}

.up-avatar-badge {
	position: absolute;
	bottom: 4rpx;
	right: 4rpx;
	width: 44rpx;
	height: 44rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	border: 4rpx solid #fff;
}

.up-user-info {
	display: flex;
	flex-direction: column;
	align-items: center;
	text-align: center;
	max-width: 600rpx;
}

.up-name-row {
	display: flex;
	align-items: center;
	gap: 10rpx;
	margin-bottom: 6rpx;
}

.up-nickname {
	font-size: 40rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: -0.5px;
	text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.15);
}

.up-official-badge {
	padding: 4rpx 12rpx;
	border-radius: 8rpx;
	background: linear-gradient(135deg, #fbbf24, #f59e0b);
	flex-shrink: 0;

	.up-official-badge-text {
		font-size: 18rpx;
		color: #fff;
		font-weight: 700;
	}
}

.up-username {
	font-size: 24rpx;
	color: rgba(255, 255, 255, 0.7);
	margin-bottom: 12rpx;
}

.up-bio {
	font-size: 26rpx;
	color: rgba(255, 255, 255, 0.85);
	line-height: 1.6;
	margin-top: 8rpx;
	max-width: 500rpx;
}

.up-actions-row {
	display: flex;
	align-items: center;
	gap: 16rpx;
	margin-top: 24rpx;
}

.up-follow-btn {
	display: flex;
	align-items: center;
	gap: 8rpx;
	padding: 12rpx 36rpx;
	border-radius: 100rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	box-shadow: 0 4rpx 16rpx rgba(37, 99, 235, 0.3);
	transition: all 0.2s ease;

	.up-follow-btn-text {
		font-size: 26rpx;
		font-weight: 700;
		color: #fff;
	}

	&:active { transform: scale(0.95); }

	&--followed {
		background: rgba(255, 255, 255, 0.2);
		backdrop-filter: blur(10px);
		border: 1px solid rgba(255, 255, 255, 0.3);
		box-shadow: none;

		.up-follow-btn-text { color: #fff; }
	}

	&--mutual {
		background: linear-gradient(135deg, #059669, #10b981);
		border: none;
		box-shadow: 0 4rpx 12rpx rgba(5, 150, 105, 0.3);

		.up-follow-btn-text { color: #fff; }
	}
}

.up-edit-btn {
	display: flex;
	align-items: center;
	gap: 8rpx;
	padding: 12rpx 36rpx;
	border-radius: 100rpx;
	background: rgba(255, 255, 255, 0.85);
	backdrop-filter: blur(10px);
	border: 1px solid rgba(255, 255, 255, 0.6);
	box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.1);
	transition: all 0.2s ease;

	.up-edit-btn-text {
		font-size: 26rpx;
		font-weight: 700;
		color: var(--primary-color);
	}

	&:active { transform: scale(0.95); }
}

.up-msg-btn {
	width: 72rpx;
	height: 72rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.2);
	backdrop-filter: blur(10px);
	border: 1px solid rgba(255, 255, 255, 0.25);
	border-radius: 50%;
	transition: all 0.2s ease;

	&:active { transform: scale(0.95); }
}

/* ---- Skeleton ---- */
.up-user-section--skeleton {
	margin-top: 0;
	padding-top: 32rpx;

	.up-avatar-skeleton {
		width: 140rpx;
		height: 140rpx;
		border-radius: 50%;
		background: rgba(255, 255, 255, 0.2);
		margin-bottom: 16rpx;
	}

	.up-info-skeleton {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12rpx;
	}

	.up-name-skeleton {
		width: 200rpx;
		height: 40rpx;
		border-radius: 8rpx;
		background: rgba(255, 255, 255, 0.2);
	}

	.up-bio-skeleton {
		width: 300rpx;
		height: 26rpx;
		border-radius: 8rpx;
		background: rgba(255, 255, 255, 0.15);
	}
}

/* ---- Stats Bar ---- */
.up-stats-bar {
	display: flex;
	align-items: center;
	justify-content: space-around;
	padding: 24rpx 0;
	background: #fff;
	margin: 0 24rpx;
	border-radius: 24rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	position: relative;
	z-index: 3;
	margin-top: -24rpx;
}

.up-stat-item {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 4rpx;
	flex: 1;

	&:active { opacity: 0.7; }
}

.up-stat-num {
	font-size: 36rpx;
	font-weight: 800;
	color: var(--text-primary);
	letter-spacing: -0.5px;
}

.up-stat-label {
	font-size: 22rpx;
	color: var(--text-tertiary);
	font-weight: 500;
}

.up-stat-divider {
	width: 1px;
	height: 48rpx;
	background: rgba(148, 163, 184, 0.2);
}

/* ---- Content ---- */
.up-content {
	flex: 1;
	padding: 0 24rpx;
	margin-top: 24rpx;
	box-sizing: border-box;
}

/* ---- Tabs ---- */
.up-tabs {
	display: flex;
	gap: 8rpx;
	margin-bottom: 24rpx;
}

.up-tab {
	flex: 1;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 8rpx;
	padding: 16rpx 0;
	border-radius: 16rpx;
	background: #fff;
	border: 1px solid rgba(148, 163, 184, 0.1);
	transition: all 0.2s ease;
	position: relative;

	&:active { transform: scale(0.97); }

	&--active {
		background: linear-gradient(135deg, rgba(99, 102, 241, 0.08), rgba(99, 102, 241, 0.04));
		border-color: var(--primary-color);
		box-shadow: 0 4rpx 12rpx rgba(99, 102, 241, 0.15);
	}
}

.up-tab-text {
	font-size: 26rpx;
	font-weight: 600;
	color: var(--text-tertiary);

	.up-tab--active & {
		color: var(--primary-color);
	}
}

.up-tab-indicator {
	position: absolute;
	bottom: 0;
	left: 50%;
	transform: translateX(-50%);
	width: 40rpx;
	height: 4rpx;
	background: var(--primary-color);
	border-radius: 2rpx;
}

/* ---- Grid Posts ---- */
.up-posts-grid {
	display: grid;
	grid-template-columns: repeat(2, 1fr);
	gap: 6rpx;
}

.up-post-grid-item {
	position: relative;
	aspect-ratio: 3 / 4;
	overflow: hidden;
	border-radius: 6rpx;
}

.up-post-grid-item--text {
	display: flex;
	align-items: flex-start;
	justify-content: flex-start;
	padding: 16rpx;
}

.up-post-text-preview {
	font-size: 30rpx;
	font-weight: 800;
	color: #292828;
	line-height: 1.6;
	overflow: hidden;
	text-overflow: ellipsis;
	display: -webkit-box;
	-webkit-line-clamp: 8;
	-webkit-box-orient: vertical;
	word-break: break-all;
	padding: 20rpx;
}

.up-post-cover {
	width: 100%;
	height: 100%;
	object-fit: cover;
}

.up-post-overlay {
	position: absolute;
	bottom: 0;
	left: 0;
	right: 0;
	padding: 8rpx 10rpx;
	background: linear-gradient(transparent, rgba(0, 0, 0, 0.4));
}

.up-post-meta {
	display: flex;
	align-items: center;
	gap: 6rpx;
}

.up-post-stat-text {
	font-size: 20rpx;
	color: #fff;
	font-weight: 600;
}

/* ---- List Posts ---- */
.up-posts-list {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
}

.up-post-list-item {
	display: flex;
	align-items: flex-start;
	gap: 16rpx;
	padding: 20rpx;
	background: #fff;
	border-radius: 20rpx;
	box-shadow: 0 2rpx 12rpx rgba(30, 64, 175, 0.05);
	border: 1px solid rgba(148, 163, 184, 0.08);

	&:active { opacity: 0.9; }
}

.up-post-list-content {
	flex: 1;
	display: flex;
	flex-direction: column;
	gap: 8rpx;
	min-height: 0;
}

.up-post-list-title {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
	line-height: 1.4;
	overflow: hidden;
	text-overflow: ellipsis;
	display: -webkit-box;
	-webkit-line-clamp: 2;
	-webkit-box-orient: vertical;
}

.up-post-list-summary {
	font-size: 24rpx;
	color: var(--text-secondary);
	line-height: 1.5;
	overflow: hidden;
	text-overflow: ellipsis;
	display: -webkit-box;
	-webkit-line-clamp: 2;
	-webkit-box-orient: vertical;
}

.up-post-list-footer {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-top: auto;
	padding-top: 8rpx;
}

.up-post-list-time {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.up-post-list-stats {
	display: flex;
	align-items: center;
	gap: 6rpx;
}

.up-post-list-stat {
	font-size: 22rpx;
	color: var(--text-tertiary);
	margin-left: 4rpx;
}

.up-post-list-cover {
	width: 180rpx;
	height: 140rpx;
	border-radius: 25px;
	flex-shrink: 0;
	object-fit: cover;
}

/* ---- Loading ---- */
.up-loading {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 80rpx 0;
	gap: 16rpx;
}

.up-spinner {
	width: 56rpx;
	height: 56rpx;
	border: 4rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-color);
	border-radius: 50%;
	animation: up-spin 0.8s linear infinite;
}

@keyframes up-spin {
	to { transform: rotate(360deg); }
}

.up-loading-text {
	font-size: 26rpx;
	color: var(--text-secondary);
}

/* ---- Empty ---- */
.up-empty {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 80rpx 0;
	gap: 16rpx;
}

.up-empty-icon {
	width: 200rpx;
	height: 200rpx;
	opacity: 0.5;
}

.up-empty-text {
	font-size: 26rpx;
	color: var(--text-tertiary);
}

.up-empty-action {
	margin-top: 8rpx;
	padding: 12rpx 40rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 100rpx;
	box-shadow: 0 4rpx 16rpx rgba(37, 99, 235, 0.25);

	.up-empty-action-text {
		font-size: 26rpx;
		font-weight: 700;
		color: #fff;
	}
}

/* ---- Load More ---- */
.up-load-more {
	padding: 24rpx 0;
	text-align: center;
}

.up-load-more-text {
	font-size: 26rpx;
	color: var(--primary-color);
	font-weight: 600;

	&:active { opacity: 0.7; }
}

.up-bottom-space {
	height: 120rpx;
}

/* ---- FAB ---- */
.up-fab {
	position: fixed;
	right: 32rpx;
	bottom: calc(32rpx + env(safe-area-inset-bottom));
	z-index: 100;
}

.up-fab-btn {
	width: 96rpx;
	height: 96rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	box-shadow: 0 8rpx 24rpx rgba(37, 99, 235, 0.35);
	display: flex;
	align-items: center;
	justify-content: center;
	transition: all 0.2s ease;

	&:active {
		transform: scale(0.9);
		box-shadow: 0 4rpx 12rpx rgba(37, 99, 235, 0.25);
	}
}
</style>
