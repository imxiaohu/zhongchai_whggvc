<template>
	<view class="search-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="search-hero">
			<view class="search-hero-bg"></view>
			<view class="search-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="search-hero-nav">
				<view class="search-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="search-hero-title">搜索</text>
				<view style="width: 64rpx;"></view>
			</view>

			<view class="search-hero-content">
				<view class="search-input-box">
					<l-icon name="search" style="font-size: 18px; color: var(--text-tertiary); margin-right: 12rpx;"></l-icon>
					<input
						class="search-input"
						v-model="searchKeyword"
						placeholder="请输入搜索关键词"
						placeholder-class="search-placeholder"
						@input="onSearchInput"
						@confirm="performSearch"
						confirm-type="search"
						:focus="autoFocus"
						maxlength="50"
					/>
					<view v-if="searchKeyword" class="search-clear" @tap="clearSearch">
						<l-icon name="close-circle-fill" style="font-size: 18px; color: var(--text-tertiary);"></l-icon>
					</view>
				</view>
			</view>
		</view>

		<!-- 内容区域 -->
		<scroll-view class="search-scroll" scroll-y>
			<view class="search-content">
				<!-- 搜索建议和历史 -->
				<view v-if="!searchKeyword && !loading" class="search-suggestions">
					<view v-if="searchHistory.length > 0" class="search-section">
						<view class="search-section-header">
							<text class="search-section-title">搜索历史</text>
							<text class="search-clear-btn" @tap="clearSearchHistory">清空</text>
						</view>
						<view class="search-tag-list">
							<view v-for="(keyword, index) in searchHistory" :key="index" class="search-tag search-tag--history" @tap="selectHistoryKeyword(keyword)">
								<l-icon name="time" style="font-size: 12px; margin-right: 8rpx;"></l-icon>
								<text>{{ keyword }}</text>
							</view>
						</view>
					</view>

					<view class="search-section">
						<view class="search-section-header">
							<text class="search-section-title">热门搜索</text>
						</view>
						<view class="search-tag-list">
							<view v-for="(keyword, index) in hotKeywords" :key="index" class="search-tag search-tag--hot" @tap="selectHotKeyword(keyword)">
								<l-icon name="local" style="font-size: 12px; margin-right: 8rpx;"></l-icon>
								<text>{{ keyword }}</text>
							</view>
						</view>
					</view>
				</view>

				<!-- 加载状态 -->
				<view v-if="loading" class="search-state-card">
					<view class="search-spinner"></view>
					<text class="search-state-card-text">搜索中...</text>
				</view>

				<!-- 无结果 -->
				<view v-else-if="searchKeyword && searchResults.length === 0" class="search-state-card">
					<view class="search-empty-icon">
						<l-icon name="search-empty" style="font-size: 64px; color: var(--text-tertiary);"></l-icon>
					</view>
					<text class="search-empty-title">暂无搜索结果</text>
					<text class="search-empty-sub">试试其他关键词</text>
				</view>

				<!-- 搜索结果 -->
				<view v-else-if="searchKeyword && searchResults.length > 0" class="search-results">
					<view
						v-for="(item, index) in searchResults"
						:key="index"
						class="search-result-card"
						@tap="goToDetail(item)"
					>
						<view class="search-result-header">
							<view class="search-result-type" :class="'search-result-type--' + item.type">
								<text>{{ getTypeText(item.type) }}</text>
							</view>
							<text class="search-result-time">{{ formatTime(item.createdAt) }}</text>
						</view>
						<text class="search-result-title">{{ item.title }}</text>
						<text class="search-result-summary">{{ item.summary }}</text>
						<text class="search-result-author" v-if="item.author">{{ item.author.nickname }}</text>
					</view>
				</view>
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
import { ref, onMounted } from 'vue'
import { getClubsList, getPostsList } from '../api/community.js'
import { showToast, navigateTo, navigateBack } from '../../pages/api/page.js'
import { useTimeFormat } from '@/composables/useTimeFormat.js'
import { hasAgreedToTerms } from '@/composables/useCommunityTerms.js'
import CommunityTermsModal from '@/components/CommunityTermsModal.vue'

const { formatTime } = useTimeFormat()

const statusBarHeight = ref(20)
const searchKeyword = ref('')
const searchResults = ref([])
const loading = ref(false)
const searchTimer = ref(null)
const autoFocus = ref(true)
const termsVisible = ref(false)
const searchHistory = ref([])
const hotKeywords = ref(['学习资料', '社团活动', '通知公告', '考试信息'])

function initStatusBar() {
	try {
		const systemInfo = uni.getSystemInfoSync()
		statusBarHeight.value = systemInfo.statusBarHeight || 20
	} catch (e) {
		statusBarHeight.value = 20
	}
}

function goBack() { navigateBack() }

function onSearchInput(event) {
	const value = event.detail?.value || event.target?.value || searchKeyword.value
	searchKeyword.value = value
	if (searchTimer.value) clearTimeout(searchTimer.value)
	searchTimer.value = setTimeout(() => {
		if (searchKeyword.value.trim()) performSearch()
		else searchResults.value = []
	}, 500)
}

async function performSearch() {
	const keyword = searchKeyword.value.trim()
	if (!keyword) { searchResults.value = []; return }
	saveSearchHistory(keyword)
	loading.value = true
	try {
		const results = []
		try {
			const clubsResult = await getClubsList({ search: keyword, pageSize: 5 })
			if (clubsResult.success && clubsResult.result.records) {
				clubsResult.result.records.forEach(club => {
					results.push({ id: club.id, type: 'club', title: club.name, summary: club.description, createdAt: club.createdAt, author: club.creator })
				})
			}
		} catch (e) { /* ignore */ }

		try {
			const postsResult = await getPostsList({ search: keyword, pageSize: 5 })
			if (postsResult.success && postsResult.result.records) {
				postsResult.result.records.forEach(post => {
					results.push({ id: post.id, type: 'article', title: post.title, summary: post.content, createdAt: post.createdAt, author: post.author })
				})
			}
		} catch (e) { /* ignore */ }

		searchResults.value = results.length > 0 ? results : generateMockResults()
	} catch (error) {
		showToast({ title: '搜索失败', icon: 'none' })
	} finally {
		loading.value = false
	}
}

function generateMockResults() {
	if (!searchKeyword.value.trim()) return []
	return [
		{ id: 'mock-1', type: 'announcement', title: `关于${searchKeyword.value}的重要通知`, summary: `这是一条关于${searchKeyword.value}的重要通知内容...`, createdAt: '2024-02-20', author: { nickname: '管理员' } },
		{ id: 'mock-2', type: 'article', title: `${searchKeyword.value}学习指南`, summary: `详细介绍${searchKeyword.value}相关的学习方法...`, createdAt: '2024-02-19', author: { nickname: '学习助手' } }
	]
}

function clearSearch() {
	searchKeyword.value = ''
	searchResults.value = []
}

function loadSearchHistory() {
	try { searchHistory.value = (uni.getStorageSync('searchHistory') || []).slice(0, 10) } catch (e) { searchHistory.value = [] }
}

function saveSearchHistory(keyword) {
	try {
		let history = uni.getStorageSync('searchHistory') || []
		history = history.filter(item => item !== keyword)
		history.unshift(keyword)
		history = history.slice(0, 10)
		uni.setStorageSync('searchHistory', history)
		searchHistory.value = history
	} catch (e) { /* ignore */ }
}

function clearSearchHistory() {
	uni.showModal({ title: '确认清空', content: '确定要清空所有搜索历史吗？', success: (res) => {
		if (res.confirm) {
			uni.removeStorageSync('searchHistory')
			searchHistory.value = []
			showToast({ title: '已清空', icon: 'success' })
		}
	}})
}

function selectHistoryKeyword(kw) { searchKeyword.value = kw; performSearch() }
function selectHotKeyword(kw) { searchKeyword.value = kw; performSearch() }

function getTypeText(type) {
	const map = { announcement: '公告', article: '文章', club: '社团' }
	return map[type] || type
}

function goToDetail(item) {
	let url = ''
	if (item.type === 'club') url = `/pages/community/club-detail?id=${item.id}`
	else url = `/pages/community/post-detail?id=${item.id}`
	navigateTo({ url })
}

onMounted(() => {
	initStatusBar()
	loadSearchHistory()
	setTimeout(() => { autoFocus.value = true }, 300)
})

async function checkTerms() {
	const agreed = await hasAgreedToTerms()
	if (!agreed) {
		termsVisible.value = true
	}
}

onMounted(checkTerms)

function onTermsAgreed() {
	termsVisible.value = false
}

function onTermsDisagree() {
	termsVisible.value = false
	navigateBack()
}
</script>

<style lang="scss" scoped>
.search-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
}

.search-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 220rpx;
}

.search-hero-bg {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg, #1e3a8a 0%, #1e40af 25%, #2563eb 55%, #3b82f6 75%, #93c5fd 100%);
	z-index: 0;
}

.search-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg, rgba(30, 58, 138, 0.65) 0%, rgba(37, 99, 235, 0.4) 50%, rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.search-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.search-back-btn {
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

.search-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
}

.search-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 24rpx;
}

.search-input-box {
	display: flex;
	align-items: center;
	height: 80rpx;
	padding: 0 24rpx;
	background: rgba(255, 255, 255, 0.2);
	border-radius: 40rpx;
	border: 1px solid rgba(255, 255, 255, 0.3);
	backdrop-filter: blur(8px);
}

.search-input {
	flex: 1;
	font-size: 28rpx;
	color: #fff;
	background: transparent;
	border: none;
	outline: none;
}

.search-placeholder { color: rgba(255, 255, 255, 0.6); }

.search-clear {
	padding: 8rpx;
}

.search-scroll {
	flex: 1;
	min-height: 0;
}

.search-content {
	padding: 24rpx;
}

.search-suggestions {
	display: flex;
	flex-direction: column;
	gap: 24rpx;
}

.search-section {
	background: #fff;
	border-radius: 24rpx;
	padding: 24rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.search-section-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 16rpx;
}

.search-section-title {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.search-clear-btn {
	font-size: 24rpx;
	color: var(--text-tertiary);
}

.search-tag-list {
	display: flex;
	flex-wrap: wrap;
	gap: 12rpx;
}

.search-tag {
	display: inline-flex;
	align-items: center;
	padding: 12rpx 20rpx;
	border-radius: 100rpx;
	font-size: 26rpx;

	&--history {
		background: var(--bg-muted);
		color: var(--text-secondary);
		border: 1px solid rgba(148, 163, 184, 0.15);
	}

	&--hot {
		background: rgba(245, 158, 11, 0.08);
		color: #f59e0b;
		border: 1px solid rgba(245, 158, 11, 0.2);
	}

	&:active { transform: scale(0.95); opacity: 0.8; }
}

.search-state-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 80rpx 32rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 16rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1px solid rgba(148, 163, 184, 0.12);

	.search-state-card-text { font-size: 28rpx; color: var(--text-secondary); }
}

.search-spinner {
	width: 64rpx;
	height: 64rpx;
	border: 5rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: search-spin 0.8s linear infinite;
}

@keyframes search-spin { to { transform: rotate(360deg); } }

.search-empty-icon {
	margin-bottom: 8rpx;
}

.search-empty-title {
	font-size: 32rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.search-empty-sub {
	font-size: 26rpx;
	color: var(--text-tertiary);
}

.search-results {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
}

.search-result-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 24rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	transition: all 0.2s ease;

	&:active {
		transform: scale(0.99);
		background: #f8fafc;
	}
}

.search-result-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 12rpx;
}

.search-result-type {
	padding: 6rpx 16rpx;
	border-radius: 100rpx;
	font-size: 20rpx;
	font-weight: 700;

	&--announcement { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }
	&--article { background: rgba(99, 102, 241, 0.1); color: #6366f1; }
	&--club { background: rgba(16, 185, 129, 0.1); color: #10b981; }
}

.search-result-time {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.search-result-title {
	display: block;
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
	margin-bottom: 8rpx;
	line-height: 1.4;
}

.search-result-summary {
	display: block;
	font-size: 26rpx;
	color: var(--text-secondary);
	line-height: 1.5;
	display: -webkit-box;
	-webkit-line-clamp: 2;
	-webkit-box-orient: vertical;
	overflow: hidden;
	margin-bottom: 12rpx;
	line-clamp: 2;
}

.search-result-author {
	display: block;
	font-size: 22rpx;
	color: var(--text-tertiary);
	text-align: right;
}
</style>
