<template>
	<view class="redirect-container">
		<view class="loading-box">
			<view class="spinner"></view>
			<text class="loading-text">正在准备评价...</text>
		</view>
	</view>
</template>

<script>
import { navigateTo } from '@/pages/api/page.js';

export default {
	data() {
		return {
			timer: null
		}
	},
	onLoad(options) {
		// 检查登录状态
		const token = uni.getStorageSync('token');
		if (!token) {
			console.log('用户未登录，显示登录提示');
			uni.showToast({
				title: '请先登录获取评教数据',
				icon: 'none',
				duration: 3000
			});
			// 延迟返回上一页
			setTimeout(() => {
				uni.navigateBack();
			}, 1500);
			return;
		}
		
		// 获取所有参数
		const query = options || {};
		
		// 构建URL参数
		const params = Object.keys(query)
			.map(key => `${key}=${encodeURIComponent(query[key])}`)
			.join('&');
		
		// 跳转到swipe页面
		this.timer = setTimeout(() => {
			uni.redirectTo({
				url: `/pages/evaluation/swipe?${params}`
			});
		}, 300);
	},
	onUnload() {
		// 清除定时器
		if (this.timer) {
			clearTimeout(this.timer);
		}
	}
}
</script>

<style lang="scss" scoped>
.redirect-container {
	display: flex;
	justify-content: center;
	align-items: center;
	height: 100vh;
	background-color: var(--bg-secondary);
}

.loading-box {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 20rpx;
}

.spinner {
	width: 48rpx;
	height: 48rpx;
	border: 4rpx solid var(--bg-tertiary);
	border-top-color: var(--primary-color);
	border-radius: 50%;
	animation: spin 1s linear infinite;
}

.loading-text {
	font-size: 28rpx;
	color: var(--text-secondary);
	font-weight: 500;
}

@keyframes spin {
	to { transform: rotate(360deg); }
}
</style>