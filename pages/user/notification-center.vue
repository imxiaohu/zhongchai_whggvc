<template>
	<view class="notif-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="notif-hero">
			<view class="notif-hero-bg"></view>
			<view class="notif-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="notif-hero-nav">
				<view class="notif-back-btn" @tap="handleBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="notif-hero-title">消息中心</text>
				<view class="notif-hero-actions">
					<view class="notif-hero-icon-btn" @tap="markAllAsRead">
						<l-icon name="check-double" style="font-size: 20px; color: #fff;"></l-icon>
					</view>
				</view>
			</view>

			<view class="notif-hero-content">
				<text class="notif-hero-sub">NOTIFICATION CENTER</text>
			</view>
		</view>

		<!-- 内容区域 -->
		<view class="notif-content">
			<!-- 顶部 Tab 切换 -->
			<view class="notif-tab-bar">
				<view
					class="notif-tab"
					:class="{ 'notif-tab--active': !unreadOnly }"
					@tap="setFilter(false)"
				>
					<text class="notif-tab-text">全部</text>
				</view>
				<view
					class="notif-tab"
					:class="{ 'notif-tab--active': unreadOnly }"
					@tap="setFilter(true)"
				>
					<text class="notif-tab-text">未读</text>
					<view v-if="unreadCount > 0" class="notif-tab-badge">{{ unreadCount > 99 ? '99+' : unreadCount }}</view>
				</view>
			</view>

			<!-- 通知列表 -->
			<scroll-view
				class="notif-scroll"
				scroll-y
				@scrolltolower="loadMore"
			>
				<!-- 加载状态 -->
				<view v-if="loading" class="notif-state-card">
					<view class="notif-state-spinner"></view>
					<text class="notif-state-text">正在获取消息...</text>
				</view>

				<!-- 空状态 -->
				<view v-else-if="notifications.length === 0" class="notif-empty-state">
					<view class="notif-empty-icon">
						<l-icon name="notification-circle" style="font-size: 64px; color: var(--text-tertiary);"></l-icon>
					</view>
					<text class="notif-empty-title">{{ unreadOnly ? '暂无未读消息' : '消息列表为空' }}</text>
					<text class="notif-empty-sub">收到的消息将在此显示</text>
				</view>

				<!-- 通知卡片 -->
				<view
					v-else
					v-for="(item, index) in notifications"
					:key="item.id"
					class="notif-card"
					:class="{ 'notif-card--unread': !item.isRead }"
					@tap="handleNotificationTap(item)"
				>
					<view class="notif-card-icon" :class="getIconClass(item.type)">
						<l-icon :name="getMdiIcon(item.type)" style="font-size: 20px; color: #fff;"></l-icon>
					</view>
					<view class="notif-card-body">
						<view class="notif-card-header">
							<text class="notif-card-title">{{ item.title }}</text>
							<text class="notif-card-time">{{ formatTime(item.createdAt) }}</text>
						</view>
						<text class="notif-card-content">{{ item.content }}</text>
						<view class="notif-card-footer" v-if="item.fromUser">
							<l-icon name="user-circle-filled" style="font-size: 12px; color: var(--text-tertiary);"></l-icon>
							<text class="notif-card-from">来自: {{ item.fromUser.nickname || item.fromUser.realname }}</text>
						</view>
					</view>
					<view class="notif-card-dot" v-if="!item.isRead"></view>
				</view>

				<!-- 加载更多 -->
				<view v-if="hasMore && !loading && notifications.length > 0" class="notif-load-more">
					<view class="notif-load-more-spinner"></view>
					<text class="notif-load-more-text">加载更多...</text>
				</view>
				<view v-else-if="!hasMore && notifications.length > 0" class="notif-load-more">
					<text class="notif-load-more-text">已经到底了</text>
				</view>
			</scroll-view>
		</view>

		<!-- 底部设置入口 -->
		<view class="notif-settings-bar" @tap="navigateToChannelSettings">
			<l-icon name="setting" style="font-size: 18px; color: var(--primary-600);"></l-icon>
			<text class="notif-settings-text">通知接收设置</text>
			<l-icon name="chevron-right" style="font-size: 18px; color: var(--text-tertiary);"></l-icon>
		</view>
	</view>
</template>

<script setup>
import { ref } from 'vue'
import { showToast } from '../../pages/api/page.js'
import { getNotificationsList, markNotificationAsRead, markAllNotificationsAsRead } from '../api/community.js'
import unreadMessageManager from '../../utils/unreadMessageManager.js'

const statusBarHeight = ref(20)
const notifications = ref([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const totalPages = ref(0)
const loading = ref(true)
const hasMore = ref(true)
const unreadOnly = ref(false)
const loadingMore = ref(false)
const unreadCount = ref(0)

function initStatusBarHeight() {
	try {
		const systemInfo = uni.getSystemInfoSync()
		statusBarHeight.value = systemInfo.statusBarHeight || 20
	} catch (e) {
		statusBarHeight.value = 20
	}
}

async function refreshUnreadBadge() {
	try {
		const count = await unreadMessageManager.fetchUnreadCount()
		unreadCount.value = count || 0
	} catch (e) {}
}

function setFilter(val) {
	if (unreadOnly.value === val) return
	unreadOnly.value = val
	reloadList()
}

async function reloadList() {
	page.value = 1
	total.value = 0
	totalPages.value = 0
	hasMore.value = true
	notifications.value = []
	loading.value = true
	try {
		await loadMore()
	} finally {
		loading.value = false
	}
}

async function loadMore() {
	if (loadingMore.value) return
	if (!hasMore.value) return

	loadingMore.value = true
	try {
		const res = await getNotificationsList({
			page: page.value,
			pageSize: pageSize.value,
			unreadOnly: unreadOnly.value
		})

		if (!res || !res.success) return

		const result = res.result || {}
		const list = result.notifications || []
		const pagination = result.pagination || {}

		if (page.value === 1) notifications.value = list
		else notifications.value = notifications.value.concat(list)

		total.value = pagination.total || total.value
		totalPages.value = pagination.totalPages || totalPages.value
		const currentPage = pagination.page || page.value
		hasMore.value = currentPage < (pagination.totalPages || 0)
		page.value = currentPage + 1
	} catch (error) {
		console.error('加载通知列表失败:', error)
	} finally {
		loadingMore.value = false
	}
}

async function markAllAsRead() {
	try {
		uni.showLoading({ title: '处理中...', mask: true })
		const res = await markAllNotificationsAsRead()
		if (!res || !res.success) return

		if (unreadOnly.value) {
			await reloadList()
		} else {
			notifications.value = notifications.value.map(item => ({ ...item, isRead: true }))
		}

		unreadMessageManager.reset()
		showToast({ title: '已全部标为已读', icon: 'success' })
	} catch (error) {
		console.error('全标已读失败:', error)
	} finally {
		uni.hideLoading()
	}
}

async function handleNotificationTap(item) {
	if (!item) return
	try {
		if (!item.isRead) {
			const res = await markNotificationAsRead(item.id)
			if (res && res.success) {
				item.isRead = true
				unreadMessageManager.markAsRead(1)
			}
		}
		if (item.relatedType === 'post' && item.relatedId) {
			uni.navigateTo({ url: `/pages/community/post-detail?id=${item.relatedId}` })
			return
		}
		uni.showModal({
			title: item.title || '消息',
			content: item.content || '',
			showCancel: false
		})
	} catch (error) {
		console.error('处理通知点击失败:', error)
	}
}

function navigateToChannelSettings() {
	uni.navigateTo({ url: '/pages/user/notification-settings' })
}

function formatTime(timeStr) {
	if (!timeStr) return ''
	try {
		const date = new Date(timeStr)
		const now = new Date()
		const diff = now - date
		if (diff < 60000) return '刚刚'
		if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
		if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
		return date.toLocaleDateString()
	} catch (e) {
		return String(timeStr)
	}
}

function getIconClass(type) {
	switch (type) {
		case 'like': return 'notif-icon--like'
		case 'comment': return 'notif-icon--comment'
		case 'bookmark': return 'notif-icon--system'
		case 'system': return 'notif-icon--system'
		case 'score_update': return 'notif-icon--score'
		default: return 'notif-icon--default'
	}
}

function getMdiIcon(type) {
	switch (type) {
		case 'like': return 'heart-filled'
		case 'comment': return 'chat-double'
		case 'bookmark': return 'star'
		case 'system': return 'notification'
		case 'score_update': return 'education'
		default: return 'notification'
	}
}

function handleBack() {
	uni.navigateBack()
}

initStatusBarHeight()
reloadList()
refreshUnreadBadge()
</script>

<style lang="scss" scoped>
/* ============================================
   Notification Center - Hero Style
   ============================================ */

.notif-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
	position: relative;
}

/* ---- Hero Section ---- */
.notif-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 220rpx;
}

.notif-hero-bg {
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

.notif-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.65) 0%,
		rgba(37, 99, 235, 0.4) 50%,
		rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.notif-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.notif-back-btn {
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

.notif-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.notif-hero-actions {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.notif-hero-icon-btn {
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

.notif-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 8rpx;
}

.notif-hero-sub {
	font-size: 20rpx;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.6);
	letter-spacing: 2px;
}

/* ---- Content ---- */
.notif-content {
	flex: 1;
	display: flex;
	flex-direction: column;
	min-height: 0;
}

/* ---- Tab Bar ---- */
.notif-tab-bar {
	display: flex;
	background: #fff;
	padding: 0 24rpx;
	margin: -30rpx 24rpx 0;
	border-radius: 24rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.1);
	border: 1px solid rgba(148, 163, 184, 0.12);
	position: relative;
	z-index: 2;
}

.notif-tab {
	flex: 1;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 8rpx;
	height: 96rpx;
	position: relative;
	transition: all 0.2s ease;
}

.notif-tab-text {
	font-size: 28rpx;
	font-weight: 600;
	color: var(--text-secondary);
	transition: color 0.2s ease;
}

.notif-tab--active .notif-tab-text {
	color: var(--primary-600);
	font-weight: 700;
}

.notif-tab-badge {
	min-width: 36rpx;
	height: 36rpx;
	padding: 0 10rpx;
	background: var(--error-color);
	color: #fff;
	font-size: 18rpx;
	font-weight: 700;
	border-radius: 18rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.notif-tab--active::after {
	content: '';
	position: absolute;
	bottom: 0;
	left: 50%;
	transform: translateX(-50%);
	width: 48rpx;
	height: 6rpx;
	background: linear-gradient(90deg, var(--primary-500), var(--primary-400));
	border-radius: 3rpx;
}

/* ---- Scroll ---- */
.notif-scroll {
	flex: 1;
	padding: 24rpx;
	padding-bottom: 160rpx;
	box-sizing: border-box;

}

/* ---- State Card ---- */
.notif-state-card {
	background: #fff;
	border-radius: 28rpx;
	padding: 80rpx 32rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 16rpx;
	box-shadow: 0 4rpx 20rpx rgba(30, 64, 175, 0.08);
	border: 1px solid rgba(148, 163, 184, 0.12);
}

.notif-state-spinner {
	width: 64rpx;
	height: 64rpx;
	border: 5rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: notif-spin 0.8s linear infinite;
}

@keyframes notif-spin {
	to { transform: rotate(360deg); }
}

.notif-state-text {
	font-size: 28rpx;
	color: var(--text-secondary);
}

/* ---- Empty State ---- */
.notif-empty-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 80rpx 40rpx;
	gap: 16rpx;
}

.notif-empty-icon {
	width: 160rpx;
	height: 160rpx;
	border-radius: 50%;
	background: var(--bg-muted);
	display: flex;
	align-items: center;
	justify-content: center;
}

.notif-empty-title {
	font-size: 32rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.notif-empty-sub {
	font-size: 26rpx;
	color: var(--text-tertiary);
}

/* ---- Notification Card ---- */
.notif-card {
	display: flex;
	background: #fff;
	border-radius: 24rpx;
	padding: 24rpx;
	margin-bottom: 16rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
	position: relative;
	overflow: hidden;
	transition: all 0.2s var(--ease-out);

	&:active {
		transform: scale(0.99);
		background: #f8fafc;
	}

	&--unread {
		border-left: 6rpx solid var(--primary-500);
	}
}

.notif-card-icon {
	width: 88rpx;
	height: 88rpx;
	border-radius: 24rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
	margin-right: 16rpx;

	&.notif-icon--like { background: linear-gradient(135deg, #f87171, #ef4444); }
	&.notif-icon--comment { background: linear-gradient(135deg, #60a5fa, #3b82f6); }
	&.notif-icon--system { background: linear-gradient(135deg, #f59e0b, #d97706); }
	&.notif-icon--score { background: linear-gradient(135deg, #10b981, #059669); }
	&.notif-icon--default { background: linear-gradient(135deg, #6366f1, #4f46e5); }
}

.notif-card-body {
	flex: 1;
	min-width: 0;
	overflow: hidden;
}

.notif-card-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
	margin-bottom: 8rpx;
	overflow: hidden;
}

.notif-card-title {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
	flex: 1;
	margin-right: 12rpx;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.notif-card-time {
	font-size: 20rpx;
	color: var(--text-tertiary);
	font-weight: 500;
	flex-shrink: 0;
}

.notif-card-content {
	font-size: 26rpx;
	color: var(--text-secondary);
	line-height: 1.5;
	display: -webkit-box;
	-webkit-line-clamp: 2;
	-webkit-box-orient: vertical;
	overflow: hidden;
	margin-bottom: 8rpx;
	word-break: break-all;
}

.notif-card-footer {
	display: flex;
	align-items: center;
	gap: 6rpx;
}

.notif-card-from {
	font-size: 20rpx;
	color: var(--text-tertiary);
}

.notif-card-dot {
	position: absolute;
	top: 28rpx;
	right: 28rpx;
	width: 16rpx;
	height: 16rpx;
	background: var(--error-color);
	border-radius: 50%;
	border: 3rpx solid #fff;
}

/* ---- Load More ---- */
.notif-load-more {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 12rpx;
	padding: 32rpx 0;
}

.notif-load-more-spinner {
	width: 32rpx;
	height: 32rpx;
	border: 3rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: notif-spin 0.8s linear infinite;
}

.notif-load-more-text {
	font-size: 24rpx;
	color: var(--text-tertiary);
}

/* ---- Settings Bar ---- */
.notif-settings-bar {
	position: fixed;
	bottom: calc(24rpx + env(safe-area-inset-bottom));
	left: 24rpx;
	right: 24rpx;
	height: 96rpx;
	background: #fff;
	border-radius: 48rpx;
	display: flex;
	align-items: center;
	padding: 0 32rpx;
	box-shadow: 0 8rpx 32rpx rgba(30, 64, 175, 0.15);
	border: 1px solid rgba(148, 163, 184, 0.12);
	z-index: 10;

	&:active { transform: scale(0.98); opacity: 0.9; }
}

.notif-settings-text {
	flex: 1;
	margin-left: 16rpx;
	font-size: 28rpx;
	font-weight: 600;
	color: var(--text-primary);
}
</style>
