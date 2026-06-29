<template>
	<view class="log-card">
		<view class="log-card-header">
			<view class="log-type-tag" :class="log.type">{{ getLogTypeText(log.type) }}</view>
			<text class="log-time">{{ formatTime(log.createdAt) }}</text>
		</view>

		<view class="log-card-body">
			<text class="log-title">{{ log.title }}</text>
			<text class="log-content">{{ log.content }}</text>
		</view>

		<view class="log-card-footer">
			<view class="log-channel-info" v-if="log.channel">
				<l-icon name="wifi" style="font-size: 12px; margin-right: 4px;"></l-icon>
				<text>{{ log.channel }}</text>
			</view>
			<view class="log-status-tag" :class="getLogStatusClass(log.status)">
				<l-icon :name="log.status === 'success' ? 'check-circle-filled' : 'error-circle-filled'" style="font-size: 12px; margin-right: 4px;"></l-icon>
				<text>{{ getLogStatusText(log.status) }}</text>
			</view>
		</view>

		<view v-if="log.errorMsg" class="log-error-box">
			<l-icon name="info-circle" style="font-size: 14px; margin-right: 6px;"></l-icon>
			<text>{{ log.errorMsg }}</text>
		</view>
	</view>
</template>

<script>
import { getLogStatusText, getLogStatusClass, getLogTypeText, formatLogTime } from '../../../utils/scoreCheck.js';

export default {
	name: 'NotificationLogItem',
	props: {
		log: { type: Object, required: true }
	},
	methods: {
		getLogTypeText,
		getLogStatusText,
		getLogStatusClass,
		formatTime(timeStr) {
			return formatLogTime(timeStr);
		}
	}
};
</script>

<style lang="scss" scoped>
.log-card {
	background-color: var(--bg-secondary);
	border-radius: 16rpx;
	padding: 20rpx;
	margin-bottom: 20rpx;
	border: 1rpx solid var(--border-light);

	.log-card-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 16rpx;
	}

	.log-type-tag {
		padding: 4rpx 12rpx;
		border-radius: 6rpx;
		font-size: 20rpx;
		color: #fff;
		
		&.info { background-color: #3b82f6; }
		&.success { background-color: #10b981; }
		&.warning { background-color: #f59e0b; }
		&.error { background-color: #ef4444; }
	}

	.log-time {
		font-size: 22rpx;
		color: var(--text-tertiary);
	}

	.log-card-body {
		margin-bottom: 16rpx;
	}

	.log-title {
		font-size: 28rpx;
		font-weight: 500;
		color: var(--text-primary);
		display: block;
		margin-bottom: 4rpx;
	}

	.log-content {
		font-size: 24rpx;
		color: var(--text-secondary);
		line-height: 1.4;
	}

	.log-card-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.log-channel-info {
		display: flex;
		align-items: center;
		font-size: 22rpx;
		color: var(--text-tertiary);
	}

	.log-status-tag {
		display: flex;
		align-items: center;
		font-size: 22rpx;
		font-weight: 500;

		&.success { color: #10b981; }
		&.success { color: #10b981; }
		&.failed { color: #ef4444; }
	}

	.log-error-box {
		margin-top: 16rpx;
		padding: 12rpx;
		background-color: rgba(239, 68, 68, 0.05);
		border-radius: 8rpx;
		font-size: 22rpx;
		color: #ef4444;
		display: flex;
		align-items: flex-start;
	}
}
</style>
