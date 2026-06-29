<template>
	<view class="cl-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="cl-hero">
			<view class="cl-hero-bg"></view>
			<view class="cl-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="cl-hero-nav">
				<view class="cl-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="cl-hero-title">社团列表</text>
				<view class="cl-search-btn" @tap="goToSearch">
					<l-icon name="search" style="font-size: 20px; color: #fff;"></l-icon>
				</view>
			</view>

			<view class="cl-hero-content">
				<text class="cl-hero-sub">CLUB LIST</text>
			</view>
		</view>

		<!-- 内容区域 -->
		<view class="cl-content">
			<scroll-view
				class="cl-scroll"
				scroll-y
				@scrolltolower="loadMore"
				:refresher-enabled="true"
				:refresher-triggered="refreshing"
				@refresherrefresh="onRefresh"
				:refresher-threshold="45"
				:refresher-default-style="'black'"
				:refresher-background="'transparent'"
			>
				<ClubContent
					:loading="loading"
					:clubs="clubs"
					:myClubs="myClubs"
					@clubClick="goToClubDetail"
					@joinClub="handleJoinClub"
					@createClub="goToCreateClub"
					@manageClub="goToClubManagement"
					@filterChange="handleFilterChange"
				/>
			</scroll-view>
		</view>
	</view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getClubsList, getMyClubs, joinClub } from '../api/community.js'
import { showToast, navigateTo, navigateBack } from '../../pages/api/page.js'
import ClubContent from './components/ClubContent.vue'

const statusBarHeight = ref(20)
const loading = ref(false)
const refreshing = ref(false)
const clubs = ref([])
const myClubs = ref([])
const page = ref(1)
const hasMore = ref(true)
const currentFilter = ref('all')

function initStatusBar() {
	try {
		const systemInfo = uni.getSystemInfoSync()
		statusBarHeight.value = systemInfo.statusBarHeight || 20
	} catch (e) {
		statusBarHeight.value = 20
	}
}

function goBack() { navigateBack() }

async function loadClubs(refresh = false) {
	if (loading.value) return
	loading.value = true
	try {
		const currentPage = refresh ? 1 : page.value
		const params = { page: currentPage, pageSize: 10, filter: currentFilter.value }
		const [clubsResult, myClubsResult] = await Promise.all([getClubsList(params), getMyClubs()])
		if (clubsResult.success) {
			const newClubs = clubsResult.result.records || []
			if (refresh) { clubs.value = newClubs; page.value = 1 } else { clubs.value.push(...newClubs) }
			hasMore.value = newClubs.length >= 10
			if (!refresh) page.value++
		}
		if (myClubsResult.success) myClubs.value = myClubsResult.result || []
	} catch (error) {
		console.error('加载社团列表失败:', error)
		showToast({ title: '网络错误', icon: 'none' })
	} finally {
		loading.value = false
	}
}

async function onRefresh() {
	refreshing.value = true
	try { await loadClubs(true) } finally { refreshing.value = false }
}

function loadMore() {
	if (hasMore.value && !loading.value) loadClubs()
}

function handleFilterChange(filter) {
	currentFilter.value = filter.key
	loadClubs(true)
}

async function handleJoinClub(clubId) {
	try {
		const result = await joinClub(clubId)
		if (result.success) {
			showToast({ title: '加入成功', icon: 'success' })
			loadClubs(true)
		} else {
			showToast({ title: result.message || '加入失败', icon: 'none' })
		}
	} catch (error) {
		showToast({ title: '加入失败', icon: 'none' })
	}
}

function goToSearch() { navigateTo({ url: '/pages/community/search' }) }
function goToClubDetail(club) { navigateTo({ url: `/pages/community/club-detail?id=${club.id}` }) }
function goToCreateClub() { navigateTo({ url: '/pages/community/create-club' }) }
function goToClubManagement(club) { navigateTo({ url: `/pages/community/club-management?id=${club.id}` }) }

onMounted(() => {
	initStatusBar()
	loadClubs(true)
})
</script>

<style lang="scss" scoped>
.cl-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
}

.cl-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 200rpx;
}

.cl-hero-bg {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg, #1e3a8a 0%, #1e40af 25%, #2563eb 55%, #3b82f6 75%, #93c5fd 100%);
	z-index: 0;
}

.cl-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg, rgba(30, 58, 138, 0.65) 0%, rgba(37, 99, 235, 0.4) 50%, rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.cl-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.cl-back-btn {
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

.cl-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
}

.cl-search-btn {
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

.cl-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 8rpx;
}

.cl-hero-sub {
	font-size: 20rpx;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.6);
	letter-spacing: 2px;
}

.cl-content {
	flex: 1;
	min-height: 0;
}

.cl-scroll {
	height: 100%;
}
</style>
