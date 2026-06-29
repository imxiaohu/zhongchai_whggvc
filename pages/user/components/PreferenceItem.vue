<template>
	<view class="preference-section">
		<view class="preference-header">
			<view class="preference-icon" :class="iconClass">
				<l-icon :name="icon" style="font-size: 14px; color: #fff;"></l-icon>
			</view>
			<text class="preference-title">{{ title }}</text>
		</view>
		<view class="preference-options">
			<view v-for="channel in normalizedChannels" :key="channel.key" class="preference-option">
				<text class="option-label">{{ channel.label }}</text>
				<switch
					:checked="getValue(channel.key)"
					@change="handleChange(channel.key, $event.detail.value)"
					:color="color"
					:disabled="!isChannelEnabled(channel.key)"
					style="transform: scale(0.7);"
				/>
			</view>
		</view>
	</view>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
	type: { type: String, required: true },
	title: { type: String, required: true },
	icon: { type: String, required: true },
	iconClass: { type: String, required: true },
	settings: { type: Object, required: true },
	channels: { type: Array, default: () => [] },
	color: { type: String, default: '#6366f1' }
});

const emit = defineEmits(['change']);

const normalizedChannels = computed(() => {
	return props.channels.map(c => {
		if (typeof c === 'string') {
			return { key: c, label: c === 'dingtalk' ? '钉钉' : '邮件' };
		}
		return c;
	});
});

function getValue(channelKey) {
	const key = `${props.type}${capitalize(channelKey)}`;
	return props.settings[key] || false;
}

function isChannelEnabled(channelKey) {
	if (channelKey === 'email') return props.settings.channels?.emailEnabled;
	if (channelKey === 'dingtalk') return props.settings.channels?.dingTalkEnabled;
	return false;
}

function handleChange(channelKey, value) {
	const key = `${props.type}${capitalize(channelKey)}`;
	emit('change', key, value);
}

function capitalize(str) {
	return str == null ? '' : str.charAt(0).toUpperCase() + str.slice(1);
}
</script>

<style lang="scss" scoped>
.preference-section {
	background-color: var(--bg-secondary);
	border-radius: 16rpx;
	padding: 20rpx;
	border: 1rpx solid var(--border-light);

	.preference-header {
		display: flex;
		align-items: center;
		margin-bottom: 20rpx;
	}

	.preference-icon {
		width: 44rpx;
		height: 44rpx;
		border-radius: 10rpx;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-right: 12rpx;

		&.like-icon { background: linear-gradient(135deg, #f87171, #ef4444); }
		&.bookmark-icon { background: linear-gradient(135deg, #fbbf24, #f59e0b); }
		&.comment-icon { background: linear-gradient(135deg, #60a5fa, #3b82f6); }
		&.comment-like-icon { background: linear-gradient(135deg, #34d399, #10b981); }
	}

	.preference-title {
		font-size: 26rpx;
		font-weight: 500;
		color: var(--text-primary);
	}

	.preference-options {
		display: flex;
		flex-direction: column;
		gap: 12rpx;
	}

	.preference-option {
		display: flex;
		justify-content: space-between;
		align-items: center;

		.option-label {
			font-size: 24rpx;
			color: var(--text-secondary);
		}
	}
}
</style>
