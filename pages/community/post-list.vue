<template>
	<view class="pl-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="pl-hero">
			<view class="pl-hero-bg"></view>
			<view class="pl-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="pl-hero-nav">
				<view class="pl-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="pl-hero-title">文章列表</text>
				<view class="pl-search-btn" @tap="goToSearch">
					<l-icon name="search" style="font-size: 20px; color: #fff;"></l-icon>
				</view>
			</view>

			<view class="pl-hero-content">
				<text class="pl-hero-sub">POST LIST</text>
			</view>
		</view>

		<!-- 内容区域 -->
		<view class="pl-content">
			<scroll-view
				class="pl-scroll"
				scroll-y
				@scrolltolower="loadMore"
				:refresher-enabled="true"
				:refresher-triggered="refreshing"
				@refresherrefresh="onRefresh"
				:refresher-threshold="45"
				:refresher-default-style="'black'"
				:refresher-background="'transparent'"
			>
				<PostContent
					:loading="loading"
					:posts="posts"
					@postClick="goToPostDetail"
					@likePost="handleLikePost"
					@createPost="goToCreatePost"
					@filterChange="handleFilterChange"
					@clickAuthor="handleClickAuthor"
				/>
			</scroll-view>
		</view>
	</view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getPostsList, likePost, unlikePost } from '../api/community.js'
import { showToast, navigateTo, navigateBack } from '../../pages/api/page.js'
import PostContent from './components/PostContent.vue'

const statusBarHeight = ref(20)
const loading = ref(false)
const refreshing = ref(false)
const posts = ref([])
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

async function loadPosts(refresh = false) {
	if (loading.value) return
	loading.value = true
	try {
		const currentPage = refresh ? 1 : page.value
		const params = { page: currentPage, pageSize: 10, filter: currentFilter.value }
		const result = await getPostsList(params)
		if (result.success) {
			const newPosts = result.result.records || []
			newPosts.forEach(post => {
				if (post.images && typeof post.images === 'string') {
					try { post.images = JSON.parse(post.images) } catch (e) { post.images = [] }
				} else if (!post.images) { post.images = [] }
			})
			if (refresh) {
				posts.value = newPosts
				page.value = 1
			} else {
				posts.value.push(...newPosts)
			}
			hasMore.value = newPosts.length >= 10
			if (!refresh) page.value++
		}
	} catch (error) {
		console.error('加载帖子列表失败:', error)
		showToast({ title: '网络错误', icon: 'none' })
	} finally {
		loading.value = false
	}
}

async function onRefresh() {
	refreshing.value = true
	try { await loadPosts(true) } finally { refreshing.value = false }
}

function loadMore() {
	if (hasMore.value && !loading.value) loadPosts()
}

function handleFilterChange(filter) {
	currentFilter.value = filter.key
	loadPosts(true)
}

async function handleLikePost(post) {
	try {
		const result = post.isLiked ? await unlikePost(post.id) : await likePost(post.id)
		if (result.success) {
			post.isLiked = !post.isLiked
			post.likesCount += post.isLiked ? 1 : -1
		}
	} catch (error) {
		console.error('点赞失败:', error)
	}
}

function goToSearch() { navigateTo({ url: '/pages/community/search' }) }
function goToPostDetail(post) { navigateTo({ url: `/pages/community/post-detail?id=${post.id}` }) }
function goToCreatePost() { navigateTo({ url: '/pages/community/create-post' }) }
function handleClickAuthor(author) {
	if (!author?.id) return
	navigateTo({ url: `/pages/community/user-profile?id=${author.id}` })
}

onMounted(() => {
	initStatusBar()
	loadPosts(true)
})
</script>

<style lang="scss" scoped>
.pl-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
}

.pl-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 200rpx;
}

.pl-hero-bg {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg, #1e3a8a 0%, #1e40af 25%, #2563eb 55%, #3b82f6 75%, #93c5fd 100%);
	z-index: 0;
}

.pl-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg, rgba(30, 58, 138, 0.65) 0%, rgba(37, 99, 235, 0.4) 50%, rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.pl-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.pl-back-btn {
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

.pl-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
}

.pl-search-btn {
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

.pl-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 8rpx;
}

.pl-hero-sub {
	font-size: 20rpx;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.6);
	letter-spacing: 2px;
}

.pl-content {
	flex: 1;
	min-height: 0;
}

.pl-scroll {
	height: 100%;
}
</style>
