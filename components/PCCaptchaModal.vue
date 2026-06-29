<template>
	<view v-if="visible" class="pc-captcha-modal">
		<view class="modal-overlay" @click="onClose"></view>
		<view class="modal-content" :class="{ 'modal-show': visible }">
			<view class="modal-header">
				<text class="modal-title">请输入验证码</text>
				<button class="close-btn" @click="onClose">
					<text class="close-icon">×</text>
				</button>
			</view>

			<view class="modal-body">
				<text class="tips-text">{{ tips }}</text>

				<view class="captcha-area">
					<view class="captcha-left">
						<input
							class="captcha-input"
							v-model="captchaValue"
							type="text"
							maxlength="4"
							placeholder="请输入4位验证码"
							placeholder-class="input-placeholder"
							:focus="inputFocused"
						/>
					</view>
					<view class="captcha-right" @click="onRefreshCaptcha">
						<image
							v-if="captchaImage"
							class="captcha-image"
							:src="captchaImage"
							mode="aspectFill"
						></image>
						<view v-else class="captcha-placeholder">
							<view class="captcha-loading"></view>
							<text class="captcha-loading-text">加载中</text>
						</view>
						<text class="refresh-hint">点击刷新</text>
					</view>
				</view>
			</view>

			<view class="modal-footer">
				<button class="cancel-btn" @click="onClose">取消</button>
				<button
					class="submit-btn"
					:class="{ 'loading': submitting }"
					@click="onSubmit"
					:disabled="submitting || captchaValue.length !== 4"
				>
					<view v-if="submitting" class="btn-spinner"></view>
					<text v-else>确认</text>
				</button>
			</view>
		</view>
	</view>
</template>

<script setup>
import { ref } from 'vue'
import { pcLoginSubmit } from '@/pages/api/discover.js'

const props = defineProps({
	visible: {
		type: Boolean,
		default: false
	},
	sessionId: {
		type: String,
		default: ''
	},
	captchaImage: {
		type: String,
		default: ''
	},
	tips: {
		type: String,
		default: '自动识别失败，请在下方输入验证码'
	}
})

const emit = defineEmits(['close', 'success', 'refresh-captcha'])

const captchaValue = ref('')
const submitting = ref(false)
const inputFocused = ref(false)

function onClose() {
	captchaValue.value = ''
	emit('close')
}

function onRefreshCaptcha() {
	captchaValue.value = ''
	emit('refresh-captcha')
}

async function onSubmit() {
	if (submitting.value || captchaValue.value.length !== 4) return
	submitting.value = true
	try {
		const res = await pcLoginSubmit(captchaValue.value)
		if (res && res.success) {
			captchaValue.value = ''
			emit('success')
		} else {
			uni.showToast({ title: res?.message || '验证码错误', icon: 'none' })
		}
	} catch (e) {
		uni.showToast({ title: '提交失败，请重试', icon: 'none' })
	} finally {
		submitting.value = false
	}
}
</script>

<style lang="scss" scoped>
.pc-captcha-modal {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	z-index: 999;
	display: flex;
	align-items: center;
	justify-content: center;
}

.modal-overlay {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.5);
}

.modal-content {
	position: relative;
	width: 600rpx;
	background: #fff;
	border-radius: 24rpx;
	overflow: hidden;
	opacity: 0;
	transform: scale(0.9);
	transition: opacity 0.2s, transform 0.2s;
}

.modal-content.modal-show {
	opacity: 1;
	transform: scale(1);
}

.modal-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 32rpx 32rpx 24rpx;
	border-bottom: 1rpx solid #f0f0f0;
}

.modal-title {
	font-size: 32rpx;
	font-weight: 600;
	color: #1f2937;
}

.close-btn {
	width: 48rpx;
	height: 48rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: none;
	border: none;
	padding: 0;
}

.close-icon {
	font-size: 40rpx;
	color: #9ca3af;
	line-height: 1;
}

.modal-body {
	padding: 32rpx;
}

.tips-text {
	display: block;
	font-size: 26rpx;
	color: #6b7280;
	margin-bottom: 24rpx;
	line-height: 1.5;
}

.captcha-area {
	display: flex;
	gap: 24rpx;
	align-items: stretch;
}

.captcha-left {
	flex: 1;
	display: flex;
	flex-direction: column;
	justify-content: center;
}

.captcha-input {
	height: 88rpx;
	padding: 0 24rpx;
	border: 2rpx solid #e5e7eb;
	border-radius: 16rpx;
	font-size: 32rpx;
	letter-spacing: 8rpx;
	text-align: center;
	background: #f9fafb;
}

.input-placeholder {
	color: #d1d5db;
	font-size: 28rpx;
	letter-spacing: 0;
}

.captcha-right {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 8rpx;
}

.captcha-image {
	width: 160rpx;
	height: 72rpx;
	border-radius: 12rpx;
	border: 2rpx solid #e5e7eb;
}

.captcha-placeholder {
	width: 160rpx;
	height: 72rpx;
	border-radius: 12rpx;
	border: 2rpx dashed #d1d5db;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 4rpx;
}

.captcha-loading {
	width: 32rpx;
	height: 32rpx;
	border: 3rpx solid #e5e7eb;
	border-top-color: #6366f1;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

@keyframes spin {
	to { transform: rotate(360deg); }
}

.captcha-loading-text {
	font-size: 20rpx;
	color: #9ca3af;
}

.refresh-hint {
	font-size: 20rpx;
	color: #9ca3af;
}

.modal-footer {
	display: flex;
	gap: 24rpx;
	padding: 24rpx 32rpx 32rpx;
}

.cancel-btn,
.submit-btn {
	flex: 1;
	height: 80rpx;
	border-radius: 16rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 30rpx;
	font-weight: 500;
	border: none;
}

.cancel-btn {
	background: #f3f4f6;
	color: #4b5563;
}

.submit-btn {
	background: #6366f1;
	color: #fff;
}

.submit-btn[disabled] {
	background: #c7c9f0;
}

.submit-btn.loading {
	background: #6366f1;
}

.btn-spinner {
	width: 32rpx;
	height: 32rpx;
	border: 3rpx solid rgba(255, 255, 255, 0.4);
	border-top-color: #fff;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

/* dark mode */
.theme-dark {
	.modal-content {
		background: #1f2937;
	}
	.modal-header {
		border-bottom-color: #374151;
	}
	.modal-title {
		color: #f3f4f6;
	}
	.close-icon {
		color: #9ca3af;
	}
	.tips-text {
		color: #9ca3af;
	}
	.captcha-input {
		border-color: #374151;
		background: #111827;
		color: #f3f4f6;
	}
	.input-placeholder {
		color: #6b7280;
	}
	.captcha-image {
		border-color: #374151;
	}
	.captcha-placeholder {
		border-color: #374151;
	}
	.captcha-loading {
		border-color: #374151;
		border-top-color: #818cf8;
	}
	.refresh-hint {
		color: #6b7280;
	}
	.cancel-btn {
		background: #374151;
		color: #d1d5db;
	}
}
</style>
