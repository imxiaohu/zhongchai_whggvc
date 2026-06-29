<!-- webpackChunkName: "community-index" -->
<template>
	<view class="community-page">
		<!-- 顶部蓝色渐变 Hero 区域 -->
		<view class="community-hero">
			<view class="community-hero-bg"></view>
			<view class="community-hero-overlay"></view>

			<!-- 状态栏占位 -->
			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<!-- 顶部导航栏 -->
			<view class="community-hero-nav">
				<view class="community-hero-nav-left">
					<text class="community-hero-brand">校园社区</text>
					<text class="community-hero-brand-en">CAMPUS COMMUNITY</text>
				</view>
				<view class="community-hero-nav-right">
					<view class="community-hero-icon-btn" @tap="goToSearch">
						<l-icon name="search" style="font-size: 22px; color: #fff;"></l-icon>
					</view>
					<view class="community-hero-icon-btn" @tap="goToNotifications">
						<l-icon name="notification" style="font-size: 22px; color: #fff;"></l-icon>
						<view v-if="unreadCount > 0" class="community-hero-unread-badge">{{ unreadCount > 99 ? '99+' : unreadCount }}</view>
					</view>
				</view>
			</view>

			<!-- 欢迎语 -->
			<view class="community-hero-content" style="padding-top: 8rpx;">
				<view class="community-hero-title">{{ greetingText }}</view>
				<text class="community-hero-subtitle">{{ greetingSubtitle }}</text>
			</view>
		</view>

		<!-- Tab 切换栏 -->
		<view class="page-tab-bar" :style="{ marginTop: '-0rpx' }">
			<view
				v-for="(tab, index) in tabs"
				:key="tab.key"
				class="page-tab-item"
				:class="{ 'page-tab-item--active': currentTab === index }"
				@tap="switchTab(index)"
			>
				<text class="page-tab-text">{{ tab.name }}</text>
				<text class="page-tab-text-en">{{ tab.sub }}</text>
				<view class="page-tab-indicator" v-if="currentTab === index"></view>
			</view>
		</view>

		<!-- 页面滑动切换 -->
		<swiper
			class="community-swiper"
			:current="currentTab"
			@change="onSwiperChange"
			:duration="300"
		>
			<!-- 推荐页 -->
			<swiper-item>
				<scroll-view
					class="community-scroll"
					scroll-y
					enable-back-to-top
					lower-threshold="100"
					@scrolltolower="loadMoreData"
					refresher-enabled
					:refresher-triggered="refreshing"
					@refresherrefresh="onRefresh"
				>
					<RecommendContent
						:loading="loading"
						:clubs="recommendClubs"
						:posts="recommendPosts"
						@postClick="goToPostDetail"
						@clubClick="goToClubDetail"
						@createPost="showCreateMenu"
						@createClub="showCreateMenu"
						@likePost="handleLikePost"
						@clickAuthor="goToUserProfile"
					/>
				</scroll-view>
			</swiper-item>

			<!-- 社团页 -->
			<swiper-item>
				<scroll-view
					class="community-scroll"
					scroll-y
					enable-back-to-top
					lower-threshold="100"
					@scrolltolower="loadMoreData"
					refresher-enabled
					:refresher-triggered="refreshing"
					@refresherrefresh="onRefresh"
				>
					<ClubContent
						:loading="loading"
						:clubs="allClubs"
						:myClubs="myClubs"
						@clubClick="goToClubDetail"
						@filterChange="onClubFilterChange"
						@joinClub="handleJoinClub"
					/>
				</scroll-view>
			</swiper-item>

			<!-- 广场页 -->
			<swiper-item>
				<scroll-view
					class="community-scroll"
					scroll-y
					enable-back-to-top
					lower-threshold="100"
					@scrolltolower="loadMoreData"
					refresher-enabled
					:refresher-triggered="refreshing"
					@refresherrefresh="onRefresh"
				>
					<PostList
						:loading="loading"
						:posts="allPosts"
						@postClick="goToPostDetail"
						@likePost="handleLikePost"
						@clickAuthor="goToUserProfile"
					/>
				</scroll-view>
			</swiper-item>
		</swiper>

		<!-- 悬浮发布按钮 -->
		<view class="page-fab" v-if="currentTab !== 1">
			<view class="page-fab-btn page-fab-btn--success" @tap="showCreateMenu">
				<l-icon name="add" size="32" color="#fff"></l-icon>
			</view>
		</view>

		<!-- 发布菜单弹窗 -->
		<t-popup ref="createMenuRef" v-model:visible="createMenuVisible" placement="bottom" :overlay="true" :close-on-overlay-click="true" @visible-change="onCreateMenuVisibleChange">
			<view class="community-create-menu">
				<view class="community-menu-header">
					<text class="community-menu-header-title">发布内容</text>
					<view class="community-close-btn" @tap="hideCreateMenu">
						<l-icon name="close" size="24" color="var(--text-secondary)"></l-icon>
					</view>
				</view>
				<view class="community-menu-grid">
					<view class="community-menu-item" @tap="goToCreatePost">
						<view class="community-menu-item-icon community-menu-item-icon--post">
							<l-icon name="edit" size="32" color="#fff"></l-icon>
						</view>
						<text class="community-menu-item-text">发动态</text>
					</view>
					<view class="community-menu-item" @tap="goToCreateClub">
						<view class="community-menu-item-icon community-menu-item-icon--club">
							<l-icon name="usergroup" size="32" color="#fff"></l-icon>
						</view>
						<text class="community-menu-item-text">创社团</text>
					</view>
				</view>
			</view>
		</t-popup>
		<!-- 社区服务须知弹窗 -->
		<CommunityTermsModal
			v-model:visible="termsVisible"
			@agree="onTermsAgreed"
			@disagree="onTermsDisagree"
		/>
	</view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import CustomNavBar from '@/components/CustomNavBar.vue'
import CommunityTermsModal from '@/components/CommunityTermsModal.vue'
import RecommendContent from './components/RecommendContent.vue'
import ClubContent from './components/ClubContent.vue'
import PostList from './components/PostList.vue'
import { onShow } from '@dcloudio/uni-app'
import { getRecommendData, getClubsList, getMyClubs, getPostsList, joinClub, leaveClub, likePost, unlikePost, processImages } from '../api/community.js'
import { normalizeNavHeightReadyPayload, getTabBarHeightFromSystemInfo } from '@/utils/system.js'
import { hasAgreedToTerms } from '@/composables/useCommunityTerms.js'
import unreadMessageManager from '@/utils/unreadMessageManager.js'

// 状态
const statusBarHeight = ref(20)
const navHeight = ref(0)
const tabBarHeight = ref(0)
const currentTab = ref(0)
const tabs = [
	{ name: '推荐', key: 'recommend', sub: 'FOR YOU' },
	{ name: '社团', key: 'club', sub: 'CLUBS' },
	{ name: '广场', key: 'square', sub: 'SQUARE' }
]
const loading = ref(false)
const refreshing = ref(false)
const unreadCount = ref(0)
const recommendClubs = ref([])
const recommendPosts = ref([])
const allClubs = ref([])
const myClubs = ref([])
const allPosts = ref([])
const page = ref(1)
const pageSize = ref(10)
const createMenuVisible = ref(false)
const createMenuRef = ref(null)
const termsVisible = ref(false)

// 计算属性
const greetingText = ref('发现校园精彩')
const greetingSubtitle = ref('与同学分享你的故事')

// 工具函数
function sanitizeText(value) {
	if (typeof value !== 'string') return value
	return value.replace(/\\`/g, '').replace(/`/g, '').trim()
}

function normalizeClub(club) {
	if (!club || typeof club !== 'object') return club
	const school = club.school && typeof club.school === 'object' ? {
		...club.school,
		logo: sanitizeText(club.school.logo),
		website: sanitizeText(club.school.website),
		apiBaseUrl: sanitizeText(club.school.apiBaseUrl)
	} : club.school
	return {
		...club,
		logoUrl: sanitizeText(club.logoUrl),
		tags: sanitizeText(club.tags),
		contactInfo: sanitizeText(club.contactInfo),
		school
	}
}

function normalizePost(post) {
	if (!post || typeof post !== 'object') return post
	const normalized = {
		...post,
		title: sanitizeText(post.title),
		summary: sanitizeText(post.summary),
		content: typeof post.content === 'string' ? post.content : '',
		images: processImages(post.images)
	}
	if (normalized.club && typeof normalized.club === 'object') {
		normalized.club = normalizeClub(normalized.club)
	}
	return normalized
}

// 初始化状态栏高度
function initStatusBarHeight() {
	try {
		const systemInfo = uni.getSystemInfoSync()
		statusBarHeight.value = systemInfo.statusBarHeight || 20
	} catch (e) {
		statusBarHeight.value = 20
	}
}

// 更新 TabBar 高度
function updateTabBarHeight() {
	try {
		const systemInfo = uni.getSystemInfoSync()
		tabBarHeight.value = getTabBarHeightFromSystemInfo(systemInfo, 50)
	} catch (e) {
		tabBarHeight.value = 50
	}
}

// 从缓存加载数据
function loadFromCache() {
	const cache = uni.getStorageSync('community_index_cache')
	if (cache) {
		recommendClubs.value = (cache.recommendClubs || []).map(normalizeClub)
		recommendPosts.value = (cache.recommendPosts || []).map(normalizePost)
		allClubs.value = (cache.allClubs || []).map(normalizeClub)
		myClubs.value = (cache.myClubs || []).map(normalizeClub)
		allPosts.value = (cache.allPosts || []).map(normalizePost)
	}
}

// 保存到缓存
function saveToCache() {
	uni.setStorage({
		key: 'community_index_cache',
		data: {
			recommendClubs: recommendClubs.value,
			recommendPosts: recommendPosts.value,
			allClubs: allClubs.value,
			myClubs: myClubs.value,
			allPosts: allPosts.value
		}
	})
}

async function refreshUnreadMessages() {
	try {
		await unreadMessageManager.fetchUnreadCount()
		unreadCount.value = unreadMessageManager.getUnreadCount()
		unreadMessageManager.updateTabBarBadge()
	} catch (error) {
		console.error('社区页面: 刷新未读消息数量失败:', error)
	}
}

// 初始化数据
function initData() {
	loadRecommend()
	loadClubs()
	loadPosts()
}

// 加载推荐数据
async function loadRecommend() {
	if (loading.value) return
	loading.value = true
	try {
		const res = await getRecommendData()
		const result = res && res.result ? res.result : {}
		recommendClubs.value = (result.clubs || []).map(normalizeClub)
		recommendPosts.value = (result.posts || []).map(normalizePost)
		saveToCache()
	} catch (error) {
		console.error('加载推荐数据失败:', error)
	} finally {
		loading.value = false
		refreshing.value = false
	}
}

// 加载社团数据
async function loadClubs() {
	try {
		const [allRes, myRes] = await Promise.all([
			getClubsList({ page: 1, pageSize: 50 }),
			getMyClubs()
		])
		allClubs.value = ((allRes && allRes.result && allRes.result.records) ? allRes.result.records : []).map(normalizeClub)
		myClubs.value = (myRes && myRes.result && Array.isArray(myRes.result) ? myRes.result : []).map(normalizeClub)
		saveToCache()
	} catch (error) {
		console.error('加载社团数据失败:', error)
	}
}

// 加载帖子数据
async function loadPosts() {
	try {
		const res = await getPostsList({ page: page.value, pageSize: pageSize.value })
		const records = res && res.result && Array.isArray(res.result.records) ? res.result.records : []
		if (page.value === 1) {
			allPosts.value = records.map(normalizePost)
			saveToCache()
		} else {
			allPosts.value = [...allPosts.value, ...records.map(normalizePost)]
		}
	} catch (error) {
		console.error('加载帖子数据失败:', error)
	}
}

// 切换选项卡
function switchTab(index) {
	currentTab.value = index
	uni.vibrateShort()
}

function onSwiperChange(e) {
	currentTab.value = e.detail.current
}

// 刷新数据
function onRefresh() {
	refreshing.value = true
	page.value = 1
	initData()
}

// 加载更多数据
function loadMoreData() {
	if (currentTab.value === 2 && !loading.value) {
		page.value++
		loadPosts()
	}
}

// 创建菜单
function showCreateMenu() {
	createMenuVisible.value = true
}

function hideCreateMenu() {
	createMenuVisible.value = false
}

function onCreateMenuVisibleChange(e) {
	// noop
}

// 导航
function goToPostDetail(post) {
	uni.navigateTo({ url: `/pages/community/post-detail?id=${post.id}` })
}

function goToClubDetail(club) {
	uni.navigateTo({ url: `/pages/community/club-detail?id=${club.id}` })
}

function goToCreatePost() {
	hideCreateMenu()
	uni.navigateTo({ url: '/pages/community/create-post' })
}

function goToCreateClub() {
	hideCreateMenu()
	uni.navigateTo({ url: '/pages/community/create-club' })
}

function goToSearch() {
	uni.navigateTo({ url: '/pages/community/search' })
}

function goToNotifications() {
	uni.navigateTo({ url: '/pages/user/notification-center' })
}

function goToUserProfile(author) {
	if (!author?.id) return
	uni.navigateTo({ url: `/pages/community/user-profile?id=${author.id}` })
}

// 处理点赞
async function handleLikePost(post) {
	try {
		const result = post.isLiked
			? await unlikePost(post.id)
			: await likePost(post.id)

		if (result.success) {
			post.isLiked = !post.isLiked
			post.likesCount += post.isLiked ? 1 : -1
		}
	} catch (error) {
		console.error('点赞失败:', error)
	}
}

// 社团筛选变化
function onClubFilterChange(filter) {
	console.log('Filter changed:', filter)
}

// 处理加入/退出社团
async function handleJoinClub(club) {
	const clubId = typeof club === 'object' ? club.id : club
	const isMember = typeof club === 'object' ? club.isMember : false
	try {
		if (isMember) {
			const res = await leaveClub(clubId)
			if (res && res.success) {
				uni.showToast({ title: '已退出', icon: 'none' })
				loadClubs()
			}
		} else {
			const res = await joinClub(clubId)
			if (res && res.success) {
				uni.showToast({ title: '加入成功', icon: 'success' })
				loadClubs()
			}
		}
	} catch (error) {
		console.error('社团操作失败:', error)
		uni.showToast({ title: '操作失败', icon: 'none' })
	}
}

// 社区服务须知
async function checkAndShowTerms() {
	const agreed = await hasAgreedToTerms()
	if (!agreed) {
		termsVisible.value = true
	}
}

function onTermsAgreed() {
	termsVisible.value = false
}

function onTermsDisagree() {
	termsVisible.value = false
}

// 生命周期
onMounted(() => {
	initStatusBarHeight()
	updateTabBarHeight()
	loadFromCache()
	initData()
	refreshUnreadMessages()
	checkAndShowTerms()
})

onShow(() => {
	updateTabBarHeight()
	refreshUnreadMessages()
	checkAndShowTerms()
})
</script>

<style lang="scss" scoped>
@import url('@/static/css/shared-page.css');

.community-page {
	flex: 1;
	width: 100%;
	height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
	overflow: hidden;
	min-height: 0;
	font-family: -apple-system, "SF Pro Text", "SF Pro Icons", "Helvetica Neue", Helvetica, Arial, sans-serif;
	position: relative;
}

.community-swiper {
	flex: 1;
	min-height: 0;
	height: 100%;
}

.community-swiper swiper-item {
	height: 100%;
	display: flex;
	flex-direction: column;
	min-height: 0;
}

.community-scroll {
	flex: 1;
	height: 100%;
	box-sizing: border-box;
	min-height: 0;
}

.community-create-menu {
	padding: 40rpx 32rpx;
	padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
	background-color: #fff;
	border-radius: 32rpx 32rpx 0 0;
}

.community-menu-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 40rpx;
}

.community-menu-header-title {
	font-size: 36rpx;
	font-weight: 800;
	color: var(--text-primary);
	letter-spacing: -0.5px;
}

.community-close-btn {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 50%;
	background: var(--bg-muted);
	transition: all 0.2s var(--ease-out);

	&:active {
		transform: scale(0.9);
		opacity: 0.7;
	}
}

.community-menu-grid {
	display: flex;
	justify-content: space-around;
	gap: 32rpx;
}

.community-menu-item {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 16rpx;

	&-icon {
		width: 128rpx;
		height: 128rpx;
		border-radius: 32rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.12);
		transition: transform 0.2s var(--ease-spring);

		&:active { transform: scale(0.92); }

		&--post {
			background: linear-gradient(135deg, #10b981, #059669);
			box-shadow: 0 8rpx 24rpx rgba(16, 185, 129, 0.35);
		}
		&--club {
			background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
			box-shadow: 0 8rpx 24rpx rgba(37, 99, 235, 0.35);
		}
	}

	&-text {
		font-size: 28rpx;
		color: var(--text-primary);
		font-weight: 600;
	}
}
</style>
