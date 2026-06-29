<template>
	<view class="settings-item">
		<view class="item-info">
			<view v-if="icon" class="item-icon" :class="iconClass">
				<l-icon :name="icon" style="font-size: 18px; color: #fff;"></l-icon>
			</view>
			<view class="item-text">
				<text class="item-title">{{ title }}</text>
				<text v-if="description" class="item-desc">{{ description }}</text>
			</view>
		</view>
		<switch
			:checked="checked"
			@change="handleChange($event.detail.value)"
			:color="color"
			:disabled="disabled"
			style="transform: scale(0.8);"
		/>
	</view>
</template>

<script>
export default {
	name: 'ToggleItem',
	props: {
		title: { type: String, required: true },
		description: { type: String, default: '' },
		icon: { type: String, default: '' },
		iconClass: { type: String, default: 'icon-default' },
		checked: { type: Boolean, default: false },
		disabled: { type: Boolean, default: false },
		color: { type: String, default: '#6366f1' }
	},
	methods: {
		handleChange(value) {
			this.$emit('change', value);
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

		&.icon-check { background: linear-gradient(135deg, #34d399, #10b981); }
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
