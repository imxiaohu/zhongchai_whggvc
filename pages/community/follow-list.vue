<!-- webpackChunkName: "community-follow-list" -->
<template>
	<view class="fl-page">
		<!-- Hero 导航 -->
		<view class="fl-hero">
			<view class="fl-hero-bg"></view>
			<view :style="{ height: statusBarHeight + 'px' }"></view>
			<view class="fl-nav">
				<view class="fl-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="fl-title">{{ type === 'followers' ? '粉丝' : '关注' }}</text>
				<view style="width: 64rpx;"></view>
			</view>
		</view>

		<!-- 用户列表 -->
		<scroll-view
			class="fl-scroll"
			scroll-y
			@scrolltolower="loadMore"
			:refresher-enabled="true"
			:refresher-triggered="refreshing"
			@refresherrefresh="onRefresh"
		>
			<!-- 加载骨架屏 -->
			<view v-if="loading && users.length === 0" class="fl-skeleton">
				<view class="fl-skeleton-item" v-for="i in 6" :key="i">
					<view class="fl-skeleton-avatar"></view>
					<view class="fl-skeleton-info">
						<view class="fl-skeleton-name"></view>
						<view class="fl-skeleton-bio"></view>
					</view>
					<view class="fl-skeleton-btn"></view>
				</view>
			</view>

			<!-- 用户列表 -->
			<view v-else-if="users.length > 0" class="fl-list">
				<view
					class="fl-item"
					v-for="user in users"
					:key="user.id"
					@tap="goToUserProfile(user.id)"
				>
					<view class="fl-item-avatar">
						<UserAvatar :src="user.avatar" :name="user.realname || user.nickname" :size="96"></UserAvatar>
					</view>
					<view class="fl-item-info">
						<view class="fl-item-name-row">
							<text class="fl-item-name">{{ user.realname || user.nickname || '匿名用户' }}</text>
							<view v-if="user.userType === 'admin' || user.userType === 'teacher'" class="fl-official-badge">
								<text class="fl-official-badge-text">官方</text>
							</view>
						</view>
						<text class="fl-item-sub">@{{ user.nickname || '无用户名' }}</text>
					</view>
					<!-- 当前用户对列表中每个用户的关注状态 -->
					<view
						class="fl-item-action"
						v-if="!isOwnList && currentUserId && String(user.id) !== String(currentUserId)"
						:class="{ 'fl-item-action--followed': followedMap[user.id] }"
						@tap.stop="handleFollowToggle(user)"
					>
						<text class="fl-item-action-text">{{ followedMap[user.id] ? '已关注' : '关注' }}</text>
					</view>
					<!-- 自己列表显示已关注 -->
					<view v-else-if="isOwnList && type === 'following'" class="fl-item-tag">
						<text class="fl-item-tag-text">已关注</text>
					</view>
				</view>
			</view>

			<!-- 空状态 -->
			<view v-else class="fl-empty">
				<l-icon name="account-multiple" style="font-size: 80px; color: var(--primary-300, #93c5fd);"></l-icon>
				<text class="fl-empty-text">{{ type === 'followers' ? '还没有粉丝' : '还没有关注任何人' }}</text>
			</view>

			<!-- 加载更多 -->
			<view class="fl-load-more" v-if="users.length > 0 && hasMore">
				<text class="fl-load-more-text" @tap="loadMore">加载更多</text>
			</view>

			<view class="fl-bottom-space"></view>
		</scroll-view>
	</view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import UserAvatar from '@/components/UserAvatar.vue'
import { getFollowers, getFollowing, followUser, unfollowUser } from '../api/community.js'

const statusBarHeight = ref(20)
const type = ref('followers') // 'followers' | 'following'
const userId = ref(null)
const users = ref([])
const loading = ref(false)
const refreshing = ref(false)
const page = ref(1)
const pageSize = ref(20)
const hasMore = ref(false)
const currentUserId = ref(null)
const followedMap = ref({}) // 当前用户对列表中每个人的关注状态

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

const isOwnList = computed(() => {
	if (!userId.value || !currentUserId.value) return false
	return String(userId.value) === String(currentUserId.value)
})

async function fetchData(isRefresh = false) {
	if (isRefresh) {
		refreshing.value = true
		page.value = 1
	} else {
		loading.value = true
	}

	try {
		const res = type.value === 'followers'
			? await getFollowers(userId.value, { page: page.value, pageSize: pageSize.value })
			: await getFollowing(userId.value, { page: page.value, pageSize: pageSize.value })

		const records = res?.result?.records || res?.result || []
		const total = res?.result?.total || res?.total || 0
		const followedData = res?.result?.followed || {}

		if (isRefresh || page.value === 1) {
			users.value = records
		} else {
			users.value = [...users.value, ...records]
		}

		hasMore.value = users.value.length < total
		followedMap.value = followedData
	} catch (error) {
		console.error('获取列表失败:', error)
	} finally {
		loading.value = false
		refreshing.value = false
	}
}

async function onRefresh() {
	await fetchData(true)
}

async function loadMore() {
	if (!hasMore.value || loading.value) return
	page.value++
	await fetchData()
}

function goToUserProfile(uid) {
	uni.navigateTo({ url: `/pages/community/user-profile?id=${uid}` })
}

async function handleFollowToggle(user) {
	try {
		const isFollowed = followedMap.value[user.id]
		const res = isFollowed
			? await unfollowUser(user.id)
			: await followUser(user.id)
		if (res && res.success) {
			followedMap.value = { ...followedMap.value, [user.id]: !isFollowed }
		}
	} catch (error) {
		console.error('关注操作失败:', error)
		uni.showToast({ title: '操作失败', icon: 'none' })
	}
}

function goBack() {
	uni.navigateBack()
}

onMounted(() => {
	initStatusBarHeight()
	currentUserId.value = getCurrentUserId()

	const pages = getCurrentPages()
	const currentPage = pages[pages.length - 1]
	const options = currentPage?.options || {}
	userId.value = options.userId || options.id
	type.value = options.type || 'followers'

	if (userId.value) {
		fetchData()
	} else {
		uni.showToast({ title: '参数错误', icon: 'none' })
		setTimeout(() => goBack(), 1500)
	}
})
</script>

<style lang="scss" scoped>
.fl-page {
	width: 100%;
	height: 100vh;
	display: flex;
	flex-direction: column;
	background-color: var(--bg-secondary);
}

.fl-hero {
	position: relative;
	z-index: 1;
}

.fl-hero-bg {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg, #1e3a8a 0%, #2563eb 100%);
	z-index: 0;
}

.fl-nav {
	position: relative;
	z-index: 1;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 24rpx;
	height: 88rpx;
}

.fl-back-btn {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.15);
	border-radius: 50%;
	border: 1px solid rgba(255, 255, 255, 0.2);

	&:active { opacity: 0.8; transform: scale(0.95); }
}

.fl-title {
	font-size: 34rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: -0.5px;
}

.fl-scroll {
	flex: 1;
	height: 0;
}

.fl-skeleton {
	padding: 24rpx;
	display: flex;
	flex-direction: column;
	gap: 4rpx;
}

.fl-skeleton-item {
	display: flex;
	align-items: center;
	gap: 20rpx;
	padding: 20rpx;
	background: #fff;
	border-radius: 20rpx;
	margin-bottom: 4rpx;

	.fl-skeleton-avatar {
		width: 96rpx;
		height: 96rpx;
		border-radius: 50%;
		background: rgba(99, 102, 241, 0.1);
		flex-shrink: 0;
	}

	.fl-skeleton-info {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 10rpx;
	}

	.fl-skeleton-name {
		width: 180rpx;
		height: 32rpx;
		border-radius: 8rpx;
		background: rgba(99, 102, 241, 0.1);
	}

	.fl-skeleton-bio {
		width: 120rpx;
		height: 24rpx;
		border-radius: 6rpx;
		background: rgba(99, 102, 241, 0.07);
	}

	.fl-skeleton-btn {
		width: 120rpx;
		height: 56rpx;
		border-radius: 28rpx;
		background: rgba(99, 102, 241, 0.1);
	}
}

.fl-list {
	padding: 24rpx;
	display: flex;
	flex-direction: column;
	gap: 4rpx;
}

.fl-item {
	display: flex;
	align-items: center;
	gap: 20rpx;
	padding: 20rpx;
	background: #fff;
	border-radius: 20rpx;
	margin-bottom: 4rpx;

	&:active { opacity: 0.9; }
}

.fl-item-avatar {
	flex-shrink: 0;
}

.fl-item-info {
	flex: 1;
	min-width: 0;
	display: flex;
	flex-direction: column;
	gap: 4rpx;
}

.fl-item-name-row {
	display: flex;
	align-items: center;
	gap: 8rpx;
}

.fl-item-name {
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.fl-official-badge {
	padding: 2rpx 10rpx;
	border-radius: 6rpx;
	background: linear-gradient(135deg, #fbbf24, #f59e0b);
	flex-shrink: 0;

	.fl-official-badge-text {
		font-size: 16rpx;
		color: #fff;
		font-weight: 700;
	}
}

.fl-item-sub {
	font-size: 24rpx;
	color: var(--text-tertiary);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.fl-item-action {
	flex-shrink: 0;
	padding: 10rpx 28rpx;
	border-radius: 28rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	box-shadow: 0 4rpx 12rpx rgba(37, 99, 235, 0.25);
	transition: all 0.2s ease;

	.fl-item-action-text {
		font-size: 24rpx;
		font-weight: 700;
		color: #fff;
	}

	&:active { transform: scale(0.95); }

	&--followed {
		background: rgba(255, 255, 255, 0.85);
		backdrop-filter: blur(10px);
		border: 1px solid rgba(148, 163, 184, 0.25);
		box-shadow: none;

		.fl-item-action-text { color: var(--text-secondary); }
	}
}

.fl-item-tag {
	flex-shrink: 0;
	padding: 8rpx 20rpx;
	border-radius: 20rpx;
	background: rgba(99, 102, 241, 0.08);
	border: 1px solid rgba(99, 102, 241, 0.2);

	.fl-item-tag-text {
		font-size: 22rpx;
		color: var(--primary-color);
		font-weight: 600;
	}
}

.fl-empty {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 120rpx 0;
	gap: 20rpx;
}

.fl-empty-text {
	font-size: 28rpx;
	color: var(--text-tertiary);
}

.fl-load-more {
	padding: 24rpx 0;
	text-align: center;
}

.fl-load-more-text {
	font-size: 26rpx;
	color: var(--primary-color);
	font-weight: 600;
	&:active { opacity: 0.7; }
}

.fl-bottom-space {
	height: 40rpx;
}
</style>
