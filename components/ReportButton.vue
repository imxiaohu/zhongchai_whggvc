<template>
	<view class="report-button" :class="themeClass">
			<button
				class="report-btn"
				@click="showReportModal"
				:disabled="loading"
			>
				<l-icon name="help-circle-filled" size="16" color="#ff6b35"></l-icon>
				<text class="rb-text">举报</text>
			</button>

		<ReportModal
			:visible="modalVisible"
			:target-type="targetType"
			:target-id="targetId"
			@close="hideReportModal"
			@success="onReportSubmitted"
		/>
	</view>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import ReportModal from './ReportModal.vue'

const props = defineProps({
	targetType: {
		type: String,
		required: true
	},
	targetId: {
		type: [Number, String],
		required: true
	}
})

const emit = defineEmits(['report-submitted'])

const modalVisible = ref(false)
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
	initTheme()
})

function initTheme() {
	try {
		const systemInfo = uni.getSystemInfoSync()
		if (systemInfo.theme) {
			isDarkMode.value = systemInfo.theme === 'dark'
		}
		listenThemeChange()
	} catch (error) {
		console.error('ReportButton: 初始化主题失败:', error)
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

function showReportModal() {
	const token = uni.getStorageSync('token') || uni.getStorageSync('accessToken')
	if (!token) {
		uni.showToast({ title: '请先登录', icon: 'none' })
		return
	}
	modalVisible.value = true
}

function hideReportModal() {
	modalVisible.value = false
}

function onReportSubmitted(data) {
	emit('report-submitted', data)
}
</script>

<style lang="scss" scoped>
.report-button {
	display: inline-block;
}

.report-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 8rpx 16rpx;
	border: none;
	border-radius: 16rpx;
	background-color: transparent;
	transition: all 0.3s ease;
	min-width: 100rpx;
	height: 56rpx;

	&:not(:disabled):active {
		transform: scale(0.95);
	}

	&:hover {
		background-color: rgba(255, 107, 53, 0.1);
	}

	&:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	}

.rb-text {
	font-size: 26rpx;
	color: #666;
	font-weight: 500;
}

/* 主题适配 */
.theme-light {
	.report-btn {
		&:hover {
			background-color: rgba(255, 107, 53, 0.1);
		}
	}

	.rb-text {
		color: #666;
	}
}

.theme-dark {
	.report-btn {
		&:hover {
			background-color: rgba(255, 107, 53, 0.2);
		}
	}

	.rb-text {
		color: #999;
	}
}

/* 响应式设计 */
@media screen and (max-width: 480px) {
	.report-btn {
		padding: 6rpx 12rpx;
		min-width: 80rpx;
		height: 48rpx;
	}

	.text {
		font-size: 24rpx;
	}
}
</style>
