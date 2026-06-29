<template>
	<view v-if="showBanner" class="cache-banner" :class="[themeClass]">
		<view class="banner-content">
			<view class="banner-icon">
				<l-icon name="cloud" size="16" color="var(--primary-color)"></l-icon>
			</view>
			<view class="banner-text">
				<text class="banner-title">正在使用缓存数据</text>
				<text class="banner-subtitle">最后更新: {{ formattedTime }}</text>
			</view>
			<view class="banner-close" @tap="closeBanner">
				<l-icon name="close" size="14" color="var(--text-secondary)"></l-icon>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount } from 'vue'

const props = defineProps({
	visible: {
		type: Boolean,
		default: false
	},
	cacheUpdatedAt: {
		type: String,
		default: ''
	},
	themeClass: {
		type: String,
		default: 'theme-light'
	},
	autoHideDelay: {
		type: Number,
		default: 5000
	}
})

const emit = defineEmits(['close'])

const showBanner = ref(false)
let autoHideTimer = null

const formattedTime = computed(() => {
	if (!props.cacheUpdatedAt) return ''

	try {
		const date = new Date(props.cacheUpdatedAt.replace(' ', 'T'))
		const now = new Date()
		const diff = now - date

		const minutes = Math.floor(diff / (1000 * 60))
		const hours = Math.floor(diff / (1000 * 60 * 60))
		const days = Math.floor(diff / (1000 * 60 * 60 * 24))

		if (minutes < 1) {
			return '刚刚'
		} else if (minutes < 60) {
			return `${minutes}分钟前`
		} else if (hours < 24) {
			return `${hours}小时前`
		} else if (days < 7) {
			return `${days}天前`
		} else {
			return date.toLocaleDateString()
		}
	} catch (error) {
		console.error('解析缓存时间失败:', error)
		return props.cacheUpdatedAt
	}
})

watch(() => props.visible, (newVal) => {
	if (newVal) {
		showBanner.value = true
		startAutoHideTimer()
	} else {
		showBanner.value = false
		clearAutoHideTimer()
	}
}, { immediate: true })

function closeBanner() {
	showBanner.value = false
	clearAutoHideTimer()
	emit('close')
}

function startAutoHideTimer() {
	clearAutoHideTimer()
	if (props.autoHideDelay > 0) {
		autoHideTimer = setTimeout(() => {
			closeBanner()
		}, props.autoHideDelay)
	}
}

function clearAutoHideTimer() {
	if (autoHideTimer) {
		clearTimeout(autoHideTimer)
		autoHideTimer = null
	}
}

onBeforeUnmount(() => {
	clearAutoHideTimer()
})
</script>

<style lang="scss" scoped>
.cache-banner {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	z-index: 999;
	background: rgba(var(--primary-color-rgb), 0.95);
	backdrop-filter: blur(10px);
	border-bottom: 1px solid rgba(var(--primary-color-rgb), 0.2);
	animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
	from {
		transform: translateY(-100%);
		opacity: 0;
	}
	to {
		transform: translateY(0);
		opacity: 1;
	}
}

.banner-content {
	display: flex;
	align-items: center;
	padding: 16rpx 32rpx;
	gap: 12rpx;
}

.banner-icon {
	flex-shrink: 0;
	width: 32rpx;
	height: 32rpx;
	border-radius: 50%;
	background: rgba(255, 255, 255, 0.2);
	display: flex;
	align-items: center;
	justify-content: center;
}

.banner-text {
	flex: 1;
	display: flex;
	flex-direction: column;
	gap: 2rpx;
}

.banner-title {
	font-size: 26rpx;
	font-weight: 500;
	color: #fff;
	line-height: 1.2;
}

.banner-subtitle {
	font-size: 22rpx;
	color: rgba(255, 255, 255, 0.8);
	line-height: 1.2;
}

.banner-close {
	flex-shrink: 0;
	width: 32rpx;
	height: 32rpx;
	border-radius: 50%;
	background: rgba(255, 255, 255, 0.1);
	display: flex;
	align-items: center;
	justify-content: center;
	transition: all 0.3s ease;
}

.banner-close:active {
	background: rgba(255, 255, 255, 0.2);
	transform: scale(0.95);
}

/* 深色模式适配 */
.theme-dark .cache-banner {
	background: rgba(var(--bg-secondary-rgb), 0.95);
	border-bottom-color: var(--border-color);
}

.theme-dark .banner-title {
	color: var(--text-primary);
}

.theme-dark .banner-subtitle {
	color: var(--text-secondary);
}

.theme-dark .banner-icon {
	background: rgba(var(--primary-color-rgb), 0.2);
}

.theme-dark .banner-close {
	background: rgba(var(--text-secondary-rgb), 0.1);
}

.theme-dark .banner-close:active {
	background: rgba(var(--text-secondary-rgb), 0.2);
}
</style>
