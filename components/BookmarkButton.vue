<template>
	<view class="bookmark-button" :class="themeClass">
		<button
			class="bookmark-btn"
			:class="{ 'bookmarked': isBookmarked, 'loading': loading }"
			@click="toggleBookmark"
			:disabled="loading"
		>
			<view class="btn-content">
				<l-icon :name="isBookmarked ? 'bookmark-filled' : 'bookmark'" size="16" color="var(--primary-color)"></l-icon>
				<text class="bb-text">{{ isBookmarked ? '已收藏' : '收藏' }}</text>
			</view>
			<view v-if="loading" class="loading-spinner"></view>
		</button>
	</view>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { bookmarkPost, unbookmarkPost, checkBookmarkStatus as checkBookmarkStatusApi } from '@/pages/api/community.js'

const props = defineProps({
	postId: {
		type: [Number, String],
		required: true
	},
	initialBookmarked: {
		type: Boolean,
		default: false
	},
	authorId: {
		type: [Number, String],
		default: null
	},
	postTitle: {
		type: String,
		default: ''
	}
})

const emit = defineEmits(['bookmark-changed'])

const isBookmarked = ref(false)
const loading = ref(false)
const isDarkMode = ref(false)
let themeWxCallback = null
let themeH5Callback = null
let themeMediaQuery = null
let themeAppCallback = null

const themeClass = computed(() =>
	isDarkMode.value ? 'theme-dark' : 'theme-light'
)

onMounted(() => {
	isBookmarked.value = props.initialBookmarked
	initTheme()
	checkBookmarkStatus()
})

function initTheme() {
	try {
		const systemInfo = uni.getSystemInfoSync()
		if (systemInfo.theme) {
			isDarkMode.value = systemInfo.theme === 'dark'
		}
		listenThemeChange()
	} catch (error) {
		console.error('BookmarkButton: 初始化主题失败:', error)
		isDarkMode.value = false
	}
}

function listenThemeChange() {
	// #ifdef MP-WEIXIN
	themeWxCallback = (res) => {
		isDarkMode.value = res.theme === 'dark'
	}
	uni.onThemeChange(themeWxCallback)
	// #endif

	// #ifdef H5
	if (typeof window !== 'undefined' && window.matchMedia) {
		themeH5Callback = (e) => {
			isDarkMode.value = e.matches
		}
		const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
		if (mediaQuery.addEventListener) {
			mediaQuery.addEventListener('change', themeH5Callback)
		}
		themeMediaQuery = mediaQuery
		isDarkMode.value = mediaQuery.matches
	}
	// #endif

	// #ifdef APP-PLUS
	themeAppCallback = (res) => {
		isDarkMode.value = res.theme === 'dark'
	}
	uni.onThemeChange(themeAppCallback)
	// #endif
}

function cleanupThemeListeners() {
	// #ifdef MP-WEIXIN
	if (themeWxCallback) {
		uni.offThemeChange(themeWxCallback)
		themeWxCallback = null
	}
	// #endif
	// #ifdef H5
	if (themeMediaQuery && themeH5Callback) {
		themeMediaQuery.removeEventListener('change', themeH5Callback)
		themeMediaQuery = null
		themeH5Callback = null
	}
	// #endif
	// #ifdef APP-PLUS
	if (themeAppCallback) {
		uni.offThemeChange(themeAppCallback)
		themeAppCallback = null
	}
	// #endif
}

onBeforeUnmount(() => {
	cleanupThemeListeners()
})

async function checkBookmarkStatus() {
	try {
		const response = await checkBookmarkStatusApi(props.postId)
		if (response.success) {
			isBookmarked.value = response.result.isBookmarked
		}
	} catch (error) {
		console.error('检查收藏状态失败:', error)
	}
}

async function toggleBookmark() {
	if (loading.value) return

	const token = uni.getStorageSync('token') || uni.getStorageSync('accessToken')
	if (!token) {
		uni.showToast({ title: '请先登录', icon: 'none' })
		return
	}

	loading.value = true

	try {
		let response
		if (isBookmarked.value) {
			response = await unbookmarkPost(props.postId)
		} else {
			response = await bookmarkPost(props.postId)
		}

		if (response.success) {
			isBookmarked.value = !isBookmarked.value
			uni.showToast({
				title: isBookmarked.value ? '收藏成功' : '取消收藏成功',
				icon: 'success'
			})
			emit('bookmark-changed', {
				postId: props.postId,
				isBookmarked: isBookmarked.value
			})
		} else {
			throw new Error(response.message || '操作失败')
		}
	} catch (error) {
		console.error('收藏操作失败:', error)
		uni.showToast({
			title: error.message || '操作失败',
			icon: 'none'
		})
	} finally {
		loading.value = false
	}
}
</script>

<style lang="scss" scoped>
.bookmark-button {
	display: inline-block;
}

.bookmark-btn {
	position: relative;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 12rpx 24rpx;
	border: 2rpx solid #e0e0e0;
	border-radius: 20rpx;
	background-color: transparent;
	transition: all 0.3s ease;
	min-width: 120rpx;
	height: 60rpx;

	&:not(:disabled):active {
		transform: scale(0.95);
	}

	&.loading {
		opacity: 0.7;
	}

	&.bookmarked {
		border-color: #ff6b35;
		background-color: #ff6b35;
		color: white;

		.bb-text {
			color: white;
		}
	}

	&:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
}

.btn-content {
	display: flex;
	align-items: center;
	gap: 8rpx;
}

.bb-text {
	font-size: 28rpx;
	font-weight: 500;
	transition: color 0.3s ease;
}

.loading-spinner {
	position: absolute;
	top: 50%;
	left: 50%;
	transform: translate(-50%, -50%);
	width: 32rpx;
	height: 32rpx;
	border: 4rpx solid rgba(255, 255, 255, 0.3);
	border-top: 4rpx solid #ff6b35;
	border-radius: 50%;
	animation: spin 1s linear infinite;
}

@keyframes spin {
	0% { transform: translate(-50%, -50%) rotate(0deg); }
	100% { transform: translate(-50%, -50%) rotate(360deg); }
}

/* 主题适配 */
.theme-light {
	.bookmark-btn {
		border-color: #e0e0e0;
		color: #333;

		&:hover {
			border-color: #ff6b35;
			background-color: rgba(255, 107, 53, 0.1);
		}
	}

	.icon-bookmark {
		color: #666;
	}

	.bb-text {
		color: #333;
	}
}

.theme-dark {
	.bookmark-btn {
		border-color: #444;
		color: #e0e0e0;

		&:hover {
			border-color: #ff6b35;
			background-color: rgba(255, 107, 53, 0.2);
		}
	}

	.icon-bookmark {
		color: #999;
	}

	.bb-text {
		color: #e0e0e0;
	}
}

/* 响应式设计 */
@media screen and (max-width: 480px) {
	.bookmark-btn {
		padding: 10rpx 20rpx;
		min-width: 100rpx;
		height: 56rpx;
	}

	.icon {
		font-size: 28rpx;
	}

	.text {
		font-size: 26rpx;
	}
}
</style>
