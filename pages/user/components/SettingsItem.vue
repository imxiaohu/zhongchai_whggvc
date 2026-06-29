<template>
	<view class="settings-item">
		<view class="item-info">
			<view class="item-icon" :class="iconClass">
				<l-icon :name="icon" style="font-size: 18px; color: #fff;"></l-icon>
			</view>
			<view class="item-text">
				<text class="item-title">{{ title }}</text>
				<text class="item-desc">{{ description }}</text>
			</view>
		</view>
		<switch
			:checked="enabled"
			@change="handleToggle($event.detail.value)"
			:color="color"
			:disabled="!available"
			style="transform: scale(0.8);"
		/>
	</view>
</template>

<script>
export default {
	name: 'SettingsItem',
	props: {
		title: { type: String, required: true },
		description: { type: String, default: '' },
		icon: { type: String, default: 'notification' },
		iconClass: { type: String, default: 'icon-default' },
		enabled: { type: Boolean, default: false },
		available: { type: Boolean, default: true },
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
.settings-item {
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

		&.icon-push { background: linear-gradient(135deg, #60a5fa, #3b82f6); }
		&.icon-email { background: linear-gradient(135deg, #f87171, #ef4444); }
		&.icon-news { background: linear-gradient(135deg, #fbbf24, #f59e0b); }
		&.icon-score { background: linear-gradient(135deg, #34d399, #10b981); }
		&.icon-schedule { background: linear-gradient(135deg, #818cf8, #6366f1); }
		&.icon-sms { background: linear-gradient(135deg, #60a5fa, #3b82f6); }
		&.icon-dingtalk { background: linear-gradient(135deg, #007fff, #0056b3); }
		&.icon-default { background: linear-gradient(135deg, #a1a1aa, #71717a); }
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

	.item-desc {
		font-size: 24rpx;
		color: var(--text-tertiary);
	}
}
</style>
