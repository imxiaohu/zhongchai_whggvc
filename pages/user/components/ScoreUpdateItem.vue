<template>
	<view class="score-update-item">
		<view class="item-info">
			<view class="item-icon" :class="iconClass">
				<l-icon :name="icon" style="font-size: 18px; color: #fff;"></l-icon>
			</view>
			<view class="item-text">
				<text class="item-title">{{ title }}</text>
			</view>
		</view>
		<switch
			:checked="enabled"
			@change="handleToggle($event.detail.value)"
			:color="color"
			:disabled="!channelEnabled"
			style="transform: scale(0.8);"
		/>
	</view>
</template>

<script>
export default {
	name: 'ScoreUpdateItem',
	props: {
		title: { type: String, required: true },
		icon: { type: String, default: 'notification' },
		iconClass: { type: String, default: 'icon-default' },
		enabled: { type: Boolean, default: false },
		channelEnabled: { type: Boolean, default: false },
		color: { type: String, default: '#6366f1' }
	},
	methods: {
		handleToggle(value) {
			this.$emit('toggle', value);
		}
	}
};
</script>

<style lang="scss" scoped>
.score-update-item {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 24rpx 0;
	border-bottom: 1rpx solid var(--border-light);

	&:last-child {
		border-bottom: none;
	}

	.item-info {
		display: flex;
		align-items: center;
		flex: 1;
		margin-right: 20rpx;
	}

	.item-icon {
		width: 72rpx;
		height: 72rpx;
		border-radius: 18rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-right: 24rpx;
		flex-shrink: 0;
		box-shadow: 0 4rpx 12rpx rgba(59, 130, 246, 0.25);

		&.icon-email {
			background: linear-gradient(135deg, #3b82f6, #60a5fa);
			box-shadow: 0 4rpx 12rpx rgba(59, 130, 246, 0.3);
		}
		&.icon-sms { background: linear-gradient(135deg, #60a5fa, #3b82f6); }
		&.icon-dingtalk { background: linear-gradient(135deg, #007fff, #0056b3); }
	}

	.item-text {
		display: flex;
		flex-direction: column;
	}

	.item-title {
		font-size: 30rpx;
		font-weight: 500;
		color: var(--text-primary);
		margin-bottom: 4rpx;
	}
}
</style>
