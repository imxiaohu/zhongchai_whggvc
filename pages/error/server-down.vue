<template>
	<view class="error-page">
		<ServerDownPage
			:error-message="errorMessage"
			:show-details="showDetails"
			:retry-function="handleRetry"
			:show-back-button="true"
		/>
	</view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onUnload, onBackPress } from '@dcloudio/uni-app'
import { resetErrorPageNavigationState } from '@/utils/errorHandler.js'

const errorMessage = ref('服务器连接超时或暂时不可用')
const showDetails = ref(false)

onLoad((options) => {
	if (options.message) {
		errorMessage.value = decodeURIComponent(options.message)
	}

	if (options.showDetails) {
		showDetails.value = options.showDetails === 'true'
	}
})

onUnload(() => {
	resetErrorPageNavigationState()
})

onBackPress(() => {
	resetErrorPageNavigationState()
	return false
})

function handleRetry() {
	resetErrorPageNavigationState()
	uni.setStorageSync('userRetryFlag', 'true')
	uni.setStorageSync('userRetryTimestamp', Date.now().toString())

	uni.navigateBack({ delta: 1 })

	setTimeout(() => {
		uni.$emit('pageRefresh')
	}, 500)
}

function handleGoBack() {
	resetErrorPageNavigationState()
	uni.removeStorageSync('userRetryFlag')
	uni.removeStorageSync('userRetryTimestamp')

	uni.switchTab({
		url: '/pages/index/index',
		fail: () => {
			uni.reLaunch({ url: '/pages/index/index' })
		}
	})
}
</script>

<style lang="scss" scoped>
.error-page {
	width: 100vw;
	height: 100vh;
	min-height: 100vh;
	display: flex;
	flex-direction: column;
	align-items: stretch;
	justify-content: center;
	padding: 0;
	margin: 0;
	box-sizing: border-box;
	position: relative;
	overflow: hidden;
}

@media screen and (max-width: 768px) {
	.error-page {
		height: 100vh;
		height: calc(100vh - env(safe-area-inset-top) - env(safe-area-inset-bottom));
		min-height: calc(100vh - env(safe-area-inset-top) - env(safe-area-inset-bottom));
		padding: env(safe-area-inset-top) env(safe-area-inset-right) env(safe-area-inset-bottom) env(safe-area-inset-left);
	}
}

@media screen and (max-height: 600px) {
	.error-page {
		justify-content: flex-start;
		padding-top: 60px;
	}
}
</style>
