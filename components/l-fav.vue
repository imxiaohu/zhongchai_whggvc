<template>
	<view class="l-fav" :class="{ 'l-fav--visible': visible }" @tap="handleClick">
		<view class="l-fav__btn">
			<view class="l-fav__icon-wrapper">
				<l-icon name="lightning" size="20" color="#fff"></l-icon>
			</view>
			<view class="l-fav__text-group">
				<text class="l-fav__label">一键</text>
				<text class="l-fav__label">评教</text>
			</view>
		</view>
		<view v-if="showBadge && pendingCount > 0" class="l-fav__badge">{{ pendingCount > 99 ? '99+' : pendingCount }}</view>
	</view>
</template>

<script setup>
const props = defineProps({
	visible: {
		type: Boolean,
		default: true
	},
	pendingCount: {
		type: Number,
		default: 0
	},
	showBadge: {
		type: Boolean,
		default: true
	}
})

const emit = defineEmits(['click'])

function handleClick() {
	uni.vibrateShort({ type: 'medium' })
	emit('click')
}
</script>

<style lang="scss" scoped>
.l-fav {
	position: fixed;
	right: var(--spacing-lg);
	bottom: calc(var(--spacing-xl) + env(safe-area-inset-bottom));
	z-index: 999;
	opacity: 0;
	transform: scale(0.6) translateY(20px);
	pointer-events: none;
	transition: all 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);

	&--visible {
		opacity: 1;
		transform: scale(1) translateY(0);
		pointer-events: auto;
	}

	&__btn {
		display: flex;
		align-items: center;
		padding: 0 20rpx;
		height: 72rpx;
		background: linear-gradient(135deg, #6366f1, #818cf8);
		border-radius: 36rpx;
		box-shadow: 0 8rpx 32rpx rgba(99, 102, 241, 0.4), 0 2rpx 8rpx rgba(0, 0, 0, 0.1);
		gap: 8rpx;

		&:active {
			transform: scale(0.93);
			box-shadow: 0 4rpx 16rpx rgba(99, 102, 241, 0.3);
		}
	}

	&__icon-wrapper {
		width: 44rpx;
		height: 44rpx;
		border-radius: 50%;
		background: rgba(255, 255, 255, 0.25);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	&__text-group {
		display: flex;
		flex-direction: column;
		line-height: 1;
		gap: 2rpx;
	}

	&__label {
		font-size: 22rpx;
		font-weight: 700;
		color: #fff;
		white-space: nowrap;
	}

	&__badge {
		position: absolute;
		top: -8rpx;
		right: -8rpx;
		min-width: 32rpx;
		height: 32rpx;
		padding: 0 6rpx;
		background: #ef4444;
		color: #fff;
		font-size: 20rpx;
		font-weight: 700;
		border-radius: 16rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 2rpx 8rpx rgba(239, 68, 68, 0.4);
		border: 2rpx solid #fff;
	}
}
</style>
