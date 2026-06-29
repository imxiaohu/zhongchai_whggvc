<template>
	<view class="channel-config-expand">
		<!-- 邮箱配置 -->
		<block v-if="channelKey === 'email'">
			<view class="email-config-wrapper">
				<view class="email-config-header">
					<view class="email-icon-badge">
						<l-icon name="mail" style="font-size: 16px; color: #fff;"></l-icon>
					</view>
					<view class="email-config-info">
						<text class="email-config-label">邮箱地址</text>
						<text class="email-config-hint">用于接收成绩通知邮件</text>
					</view>
				</view>
				<view class="email-input-box">
					<t-input
						v-model="settings.channels.emailAddress"
						type="text"
						placeholder="请输入电子邮箱地址"
						prefix-icon="mail"
						:clearable="true"
						@change="handleEmailInput"
					/>
				</view>
			</view>
		</block>

		<!-- 手机号配置 -->
		<block v-if="channelKey === 'sms'">
			<view class="config-input-wrapper">
				<t-input
					v-model="settings.channels.phoneNumber"
					type="number"
					placeholder="请输入手机号码"
					prefix-icon="call"
					:clearable="true"
					@change="handlePhoneInput"
				/>
			</view>

			<!-- 短信余额小卡片 -->
			<view class="sms-balance-mini-card">
				<view class="balance-main">
					<text class="balance-label">余额</text>
					<text class="balance-value">¥{{ smsBalance.balanceYuan || '0.00' }}</text>
					<text class="balance-count">约 {{ Math.floor((smsBalance.balance || 0) / (smsBalance.smsCost || 10)) }} 条</text>
				</view>
				<view class="balance-btns">
					<view class="mini-action-btn recharge" @tap="$emit('recharge')">
						<l-icon name="wallet" style="font-size: 14px; margin-right: 4px;"></l-icon>
						<text>充值</text>
					</view>
					<view class="mini-action-btn test" @tap="$emit('test-sms')" :class="{'loading': testingSMS}">
						<l-icon name="check-circle-filled" style="font-size: 14px; margin-right: 4px;"></l-icon>
						<text>{{ testingSMS ? '发送中' : '测试' }}</text>
					</view>
				</view>
			</view>
		</block>

		<!-- 钉钉Webhook配置 -->
		<block v-if="channelKey === 'dingtalk'">
			<view class="config-input-wrapper">
				<t-input
					v-model="settings.channels.dingTalkWebhookURL"
					type="text"
					placeholder="请输入钉钉机器人 Webhook 地址"
					prefix-icon="link"
					:clearable="true"
					@change="handleWebhookInput"
				/>
			</view>
		</block>
	</view>
</template>

<script>
export default {
	name: 'ChannelConfig',
	methods: {
		handleEmailInput(e) {
			this.$emit('email-input', { detail: { value: e } });
		},
		handlePhoneInput(e) {
			this.$emit('phone-input', { detail: { value: e } });
		},
		handleWebhookInput(e) {
			this.$emit('webhook-input', { detail: { value: e } });
		}
	},
	props: {
		channelKey: { type: String, required: true },
		settings: { type: Object, required: true },
		smsBalance: { type: Object, default: () => ({}) },
		testingSMS: { type: Boolean, default: false }
	}
};
</script>

<style lang="scss" scoped>
@import '@/uni_modules/lime-style/index.scss';

.channel-config-expand {
	padding: 20rpx;
	background-color: var(--bg-secondary);
	border-radius: 16rpx;
	margin-top: -10rpx;
	margin-bottom: 20rpx;
	border: 1rpx solid var(--border-light);
}

/* ---- Email Config ---- */
.email-config-wrapper {
	background: linear-gradient(135deg, rgba(59, 130, 246, 0.06), rgba(99, 102, 241, 0.04));
	border: 1px solid rgba(59, 130, 246, 0.15);
	border-radius: 20rpx;
	padding: 24rpx;
	margin-bottom: 8rpx;
}

.email-config-header {
	display: flex;
	align-items: center;
	gap: 16rpx;
	margin-bottom: 20rpx;
}

.email-icon-badge {
	width: 64rpx;
	height: 64rpx;
	border-radius: 16rpx;
	background: linear-gradient(135deg, #3b82f6, #60a5fa);
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
	box-shadow: 0 4rpx 12rpx rgba(59, 130, 246, 0.35);
}

.email-config-info {
	display: flex;
	flex-direction: column;
	gap: 4rpx;
}

.email-config-label {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.email-config-hint {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.email-input-box {
	width: 100%;
}

.config-input-wrapper {
	padding: 0;
}

.sms-balance-mini-card {
	margin-top: 20rpx;
	background-color: var(--bg-highlight);
	border-radius: 12rpx;
	padding: 20rpx;
	display: flex;
	justify-content: space-between;
	align-items: center;

	.balance-main {
		display: flex;
		flex-direction: column;
	}

	.balance-label {
		font-size: 22rpx;
		color: var(--text-tertiary);
		margin-bottom: 4rpx;
	}

	.balance-value {
		font-size: 32rpx;
		font-weight: 600;
		color: var(--primary-color);
	}

	.balance-count {
		font-size: 20rpx;
		color: var(--text-tertiary);
	}

	.balance-btns {
		display: flex;
		gap: 12rpx;
	}

	.mini-action-btn {
		display: flex;
		align-items: center;
		padding: 10rpx 20rpx;
		border-radius: 30rpx;
		font-size: 24rpx;
		transition: all 0.2s ease;

		&.recharge {
			background-color: var(--primary-color);
			color: #fff;
		}

		&.test {
			background-color: var(--bg-card);
			color: var(--text-primary);
			border: 1rpx solid var(--border-primary);

			&.loading {
				opacity: 0.7;
			}
		}

		&:active {
			transform: scale(0.95);
		}
	}
}
</style>
