<template>
	<view class="club-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="club-hero" v-if="clubDetail">
			<view class="club-hero-bg"></view>
			<view class="club-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="club-hero-nav">
				<view class="club-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="club-hero-title">{{ clubDetail.name }}</text>
				<view class="club-share-btn" @tap="shareClub">
					<l-icon name="share" style="font-size: 20px; color: #fff;"></l-icon>
				</view>
			</view>

			<view class="club-hero-info">
				<view class="club-hero-avatar-wrap">
				<ClubAvatar
					class="club-hero-avatar"
					:src="clubDetail.logoUrl || clubDetail.avatar"
					:name="clubDetail.name"
					:size="120"
				></ClubAvatar>
					<view v-if="clubDetail.isOfficial" class="club-hero-avatar-badge">
						<l-icon name="check-circle-filled" style="font-size: 12px; color: #fff;"></l-icon>
					</view>
				</view>
				<view class="club-hero-stats">
					<view class="club-hero-stat">
						<text class="club-hero-stat-num">{{ clubDetail.memberCount || 0 }}</text>
						<text class="club-hero-stat-label">成员</text>
					</view>
					<view class="club-hero-stat-divider"></view>
					<view class="club-hero-stat">
						<text class="club-hero-stat-num">{{ clubDetail.postCount || 0 }}</text>
						<text class="club-hero-stat-label">动态</text>
					</view>
				</view>
				<view class="club-hero-follow" :class="{ 'club-hero-follow--joined': clubDetail.isMember }" @tap="handleJoinToggle">
					<text class="club-hero-follow-text">{{ clubDetail.isMember ? '已加入' : '+ 加入' }}</text>
				</view>
			</view>
		</view>

		<!-- Hero 占位（加载时） -->
		<view class="club-hero club-hero--skeleton" v-else>
			<view class="club-hero-bg"></view>
			<view class="club-hero-overlay"></view>
			<view :style="{ height: statusBarHeight + 'px' }"></view>
			<view class="club-hero-nav">
				<view class="club-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="club-hero-title">社团详情</text>
				<view style="width: 64rpx;"></view>
			</view>
		</view>

		<!-- 内容区域 -->
		<view class="club-content">
			<!-- 描述与标签 -->
			<view class="club-desc-card" v-if="clubDetail && (clubDetail.description || clubDetail.tags)">
				<text class="club-desc-text" v-if="clubDetail.description">{{ clubDetail.description }}</text>
				<view class="club-tags" v-if="clubDetail.tags">
					<text v-for="tag in getTagList(clubDetail.tags)" :key="tag" class="club-tag"># {{ tag }}</text>
				</view>
			</view>

			<!-- 选项卡 -->
			<view class="club-tabs">
				<view
					v-for="(tab, index) in tabs"
					:key="tab.key"
					class="club-tab"
					:class="{ 'club-tab--active': currentTab === index }"
					@tap="switchTab(index)"
				>
					<text class="club-tab-text">{{ tab.name }}</text>
					<view class="club-tab-indicator" v-if="currentTab === index"></view>
				</view>
			</view>

			<!-- Swiper 内容 -->
			<swiper
				class="club-swiper"
				:current="currentTab"
				@change="onSwiperChange"
				:duration="300"
			>
				<!-- 动态列表 -->
				<swiper-item>
					<scroll-view
						class="club-swiper-scroll"
						scroll-y
						enable-back-to-top
						@scrolltolower="loadMorePosts"
						refresher-enabled
						:refresher-triggered="refreshingPosts"
						@refresherrefresh="onRefreshPosts"
					>
						<PostList
							:loading="loadingPosts"
							:posts="posts"
							@postClick="goToPostDetail"
							@likePost="handleLikePost"
							@clickAuthor="goToMemberProfile"
						/>
					</scroll-view>
				</swiper-item>

				<!-- 成员列表 -->
				<swiper-item>
					<scroll-view
						class="club-swiper-scroll"
						scroll-y
						enable-back-to-top
						@scrolltolower="loadMoreMembers"
					>
						<MemberList
							:loading="loadingMembers"
							:members="members"
							:isAdmin="isAdmin"
							@memberClick="goToMemberProfile"
							@updateRole="handleUpdateMemberRole"
							@removeMember="handleRemoveMember"
						/>
					</scroll-view>
				</swiper-item>
			</swiper>
		</view>

		<!-- 发布动态悬浮按钮 -->
		<view class="club-fab" v-if="clubDetail && clubDetail.isMember" @tap="goToCreatePost">
			<l-icon name="edit-1" style="font-size: 22px; color: #fff;"></l-icon>
			<text class="club-fab-text">发布动态</text>
		</view>

		<!-- 全屏加载 -->
		<view v-if="loading" class="club-loading">
			<view class="club-loading-spinner"></view>
		</view>

		<!-- 社区服务须知弹窗 -->
		<CommunityTermsModal
			v-model:visible="termsVisible"
			@agree="onTermsAgreed"
			@disagree="onTermsDisagree"
		/>
	</view>
</template>

<script setup>
import ClubAvatar from '@/components/ClubAvatar.vue'
import { ref, computed, onMounted } from 'vue'
import { getClubDetail, getPostsList, getClubMembers, joinClub, leaveClub, likePost, unlikePost, updateMemberRole } from '../api/community.js'
import PostList from './components/PostList.vue'
import MemberList from './components/MemberList.vue'
import { hasAgreedToTerms } from '@/composables/useCommunityTerms.js'
import CommunityTermsModal from '@/components/CommunityTermsModal.vue'

const statusBarHeight = ref(20)
const clubId = ref(null)
const clubDetail = ref(null)
const loading = ref(true)
const currentTab = ref(0)
const termsVisible = ref(false)
const tabs = [
	{ name: '动态', key: 'posts' },
	{ name: '成员', key: 'members' }
]
const posts = ref([])
const postsPage = ref(1)
const loadingPosts = ref(false)
const refreshingPosts = ref(false)
const members = ref([])
const membersPage = ref(1)
const loadingMembers = ref(false)

function initStatusBar() {
	try {
		const systemInfo = uni.getSystemInfoSync()
		statusBarHeight.value = systemInfo.statusBarHeight || 20
	} catch (e) {
		statusBarHeight.value = 20
	}
}

const isAdmin = computed(() => !!(clubDetail.value && clubDetail.value.isAdmin))

function goBack() {
	uni.navigateBack()
}

async function initData() {
	loading.value = true
	try {
		const res = await getClubDetail(clubId.value)
		const result = res.result || res
		clubDetail.value = { ...result.club, isAdmin: result.isAdmin, isMember: result.isMember }
		loadPosts()
		loadMembers()
	} finally {
		loading.value = false
	}
}

async function loadPosts() {
	loadingPosts.value = true
	try {
		const res = await getPostsList({ clubId: clubId.value, page: postsPage.value })
		const records = (res && res.result && Array.isArray(res.result.records)) ? res.result.records : []
		if (postsPage.value === 1) posts.value = records
		else posts.value = [...posts.value, ...records]
	} finally {
		loadingPosts.value = false
		refreshingPosts.value = false
	}
}

async function loadMembers() {
	loadingMembers.value = true
	try {
		const res = await getClubMembers(clubId.value, { page: membersPage.value })
		const records = (res && res.result && Array.isArray(res.result.records)) ? res.result.records : []
		if (membersPage.value === 1) members.value = records
		else members.value = [...members.value, ...records]
	} finally {
		loadingMembers.value = false
	}
}

function switchTab(index) { currentTab.value = index }
function onSwiperChange(e) {
	currentTab.value = e.detail.current
	postsPage.value = 1
	membersPage.value = 1
}
function onRefreshPosts() {
	refreshingPosts.value = true
	postsPage.value = 1
	loadPosts()
}
function loadMorePosts() {
	if (loadingPosts.value) return
	postsPage.value++
	loadPosts()
}
function loadMoreMembers() {
	if (loadingMembers.value) return
	membersPage.value++
	loadMembers()
}

async function handleJoinToggle() {
	if (clubDetail.value.isMember) {
		uni.showModal({
			title: '退出社团',
			content: '确定要退出该社团吗？',
			success: async (res) => {
				if (res.confirm) {
					await leaveClub(clubId.value)
					clubDetail.value.isMember = false
					clubDetail.value.memberCount--
					uni.showToast({ title: '已退出', icon: 'none' })
				}
			}
		})
	} else {
		await joinClub(clubId.value)
		clubDetail.value.isMember = true
		clubDetail.value.memberCount++
		uni.showToast({ title: '加入成功', icon: 'success' })
	}
}

async function handleUpdateMemberRole(member, newRole) {
	try {
		await updateMemberRole(clubId.value, member.id, newRole)
		member.role = newRole
		uni.showToast({ title: newRole === 'admin' ? '已设为管理员' : '已撤销管理员', icon: 'success' })
	} catch (e) {
		uni.showToast({ title: '操作失败', icon: 'none' })
	}
}

async function handleRemoveMember(member) {
	uni.showModal({
		title: '移除成员',
		content: `确定要将 ${member.user?.realname || member.user?.nickname || '该成员'} 移出社团吗？`,
		success: async (res) => {
			if (res.confirm) {
				try {
					await updateMemberRole(clubId.value, member.id, 'removed')
					const idx = members.value.findIndex(m => m.id === member.id)
					if (idx !== -1) members.value.splice(idx, 1)
					clubDetail.value.memberCount--
					uni.showToast({ title: '已移除', icon: 'success' })
				} catch (e) {
					uni.showToast({ title: '移除失败', icon: 'none' })
				}
			}
		}
	})
}

function getTagList(tags) {
	return tags ? tags.split(',').map(t => t.trim()) : []
}

function goToPostDetail(post) {
	uni.navigateTo({ url: `/pages/community/post-detail?id=${post.id}` })
}

function goToCreatePost() {
	uni.navigateTo({ url: `/pages/community/create-post?clubId=${clubId.value}` })
}

function shareClub() {
	if (!clubDetail.value) return
	// #ifdef MP-WEIXIN
	uni.showToast({ title: '点击右上角分享', icon: 'none' })
	// #endif
	// #ifndef MP-WEIXIN
	uni.setClipboardData({
		data: `一起来加入【${clubDetail.value.name}】吧！`,
		success: () => uni.showToast({ title: '链接已复制', icon: 'success' })
	})
	// #endif
}

async function handleLikePost(post) {
	post.isLiked = !post.isLiked
	post.likesCount += post.isLiked ? 1 : -1
	try {
		const res = await (post.isLiked ? likePost(post.id) : unlikePost(post.id))
		post.likesCount = res?.result?.likesCount ?? res?.result?.post?.likesCount ?? post.likesCount
	} catch (e) {
		post.isLiked = !post.isLiked
		post.likesCount += post.isLiked ? 1 : -1
	}
}

function goToMemberProfile(member) {
	const userId = member.user?.id || member.userId
	if (userId) {
		uni.navigateTo({ url: `/pages/community/user-profile?id=${userId}` })
	}
}

async function initMounted() {
	initStatusBar()
	const agreed = await hasAgreedToTerms()
	if (!agreed) {
		termsVisible.value = true
		return
	}
	const pages = getCurrentPages()
	const currentPage = pages[pages.length - 1]
	const options = currentPage?.options || {}
	clubId.value = options.id
	if (clubId.value) initData()
}

onMounted(initMounted)

function onTermsAgreed() {
	termsVisible.value = false
	const pages = getCurrentPages()
	const currentPage = pages[pages.length - 1]
	const options = currentPage?.options || {}
	clubId.value = options.id
	if (clubId.value) initData()
}

function onTermsDisagree() {
	termsVisible.value = false
	navigateBack()
}
</script>

<style lang="scss" scoped>
/* ============================================
   Club Detail - Hero Style
   ============================================ */

.club-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
	position: relative;
}

/* ---- Hero Section ---- */
.club-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 380rpx;
}

.club-hero-bg {
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

.club-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.65) 0%,
		rgba(37, 99, 235, 0.4) 50%,
		rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.club-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.club-back-btn {
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

.club-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.club-share-btn {
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

.club-hero-info {
	position: relative;
	z-index: 2;
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 0 32rpx 32rpx;
	gap: 24rpx;
}

.club-hero-avatar-wrap {
	position: relative;
}

.club-hero-avatar {
	width: 160rpx;
	height: 160rpx;
	border-radius: 40rpx;
	border: 6rpx solid rgba(255, 255, 255, 0.5);
	box-shadow: 0 8rpx 32rpx rgba(0, 0, 0, 0.2);
	object-fit: cover;
}

.club-hero-avatar-badge {
	position: absolute;
	bottom: 0;
	right: 0;
	width: 48rpx;
	height: 48rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	border: 4rpx solid #fff;
}

.club-hero-stats {
	display: flex;
	align-items: center;
	gap: 32rpx;
}

.club-hero-stat {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 4rpx;
}

.club-hero-stat-num {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
}

.club-hero-stat-label {
	font-size: 22rpx;
	color: rgba(255, 255, 255, 0.7);
}

.club-hero-stat-divider {
	width: 1px;
	height: 40rpx;
	background: rgba(255, 255, 255, 0.3);
}

.club-hero-follow {
	padding: 12rpx 48rpx;
	border-radius: 100rpx;
	background: rgba(255, 255, 255, 0.25);
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.3);
	transition: all 0.2s ease;

	.club-hero-follow-text {
		font-size: 28rpx;
		font-weight: 700;
		color: #fff;
	}

	&--joined {
		background: rgba(255, 255, 255, 0.15);
		.club-hero-follow-text { color: rgba(255, 255, 255, 0.7); }
	}

	&:active:not(.club-hero-follow--joined) {
		background: rgba(255, 255, 255, 0.35);
		transform: scale(0.95);
	}
}

/* ---- Content ---- */
.club-content {
	flex: 1;
	min-height: 0;
	display: flex;
	flex-direction: column;
}

/* ---- Desc Card ---- */
.club-desc-card {
	background: #fff;
	margin: -40rpx 24rpx 0;
	border-radius: 24rpx;
	padding: 28rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1px solid rgba(148, 163, 184, 0.12);
	position: relative;
	z-index: 2;
}

.club-desc-text {
	display: block;
	font-size: 28rpx;
	color: var(--text-secondary);
	line-height: 1.7;
	margin-bottom: 16rpx;
}

.club-tags {
	display: flex;
	flex-wrap: wrap;
	gap: 12rpx;
}

.club-tag {
	padding: 6rpx 20rpx;
	border-radius: 100rpx;
	background: rgba(59, 102, 241, 0.08);
	color: var(--primary-600);
	font-size: 24rpx;
	font-weight: 600;
}

/* ---- Tabs ---- */
.club-tabs {
	display: flex;
	background: #fff;
	margin: 16rpx 24rpx 0;
	border-radius: 24rpx;
	padding: 0 8rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	position: relative;
	z-index: 3;
}

.club-tab {
	flex: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	height: 96rpx;
	position: relative;
}

.club-tab-text {
	font-size: 28rpx;
	font-weight: 600;
	color: var(--text-secondary);
	transition: color 0.2s ease;
}

.club-tab--active .club-tab-text {
	color: var(--primary-600);
	font-weight: 700;
}

.club-tab-indicator {
	position: absolute;
	bottom: 0;
	left: 50%;
	transform: translateX(-50%);
	width: 48rpx;
	height: 6rpx;
	background: linear-gradient(90deg, var(--primary-500), var(--primary-400));
	border-radius: 3rpx;
}

/* ---- Swiper ---- */
.club-swiper {
	flex: 1;
	min-height: 0;
}

.club-swiper-scroll {
	height: 100%;
	padding-bottom: calc(140rpx + env(safe-area-inset-bottom));
}

/* ---- FAB ---- */
.club-fab {
	position: fixed;
	right: 32rpx;
	bottom: calc(40rpx + env(safe-area-inset-bottom));
	display: flex;
	align-items: center;
	gap: 12rpx;
	height: 96rpx;
	padding: 0 40rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 48rpx;
	box-shadow: 0 8rpx 32rpx rgba(37, 99, 235, 0.35);
	z-index: 100;

	&:active { transform: scale(0.95); }
}

.club-fab-text {
	font-size: 28rpx;
	font-weight: 700;
	color: #fff;
}

/* ---- Loading ---- */
.club-loading {
	position: fixed;
	top: 0; left: 0; right: 0; bottom: 0;
	background: rgba(255, 255, 255, 0.9);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 1000;
}

.club-loading-spinner {
	width: 64rpx;
	height: 64rpx;
	border: 5rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: club-spin 0.8s linear infinite;
}

@keyframes club-spin {
	to { transform: rotate(360deg); }
}
</style>
