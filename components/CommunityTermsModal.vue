<!-- webpackChunkName: "community-terms-modal" -->
<template>

	<t-popup
		:visible="visible"
		placement="bottom"
		:overlay="true"
		:close-on-overlay-click="false"
		@update:visible="onPopupVisibleChange"
	>
		<scroll-view scroll-y>
			<view class="ctm-modal" :style="{ height: modalHeight + 'px' }">
			<!-- Header -->
			<view class="ctm-header">
				<view class="ctm-icon-wrap">
					<l-icon name="info-circle" size="40" color="#6366f1"></l-icon>
				</view>
				<text class="ctm-title">社区服务须知</text>
				<text class="ctm-subtitle">在使用社区功能前，请仔细阅读以下内容</text>
			</view>

			<!-- Content - scrollable -->
			<scroll-view class="ctm-scroll" scroll-y>
				<view class="ctm-content">
					<!-- Section 1: 第三方服务 -->
					<view class="ctm-section">
						<view class="ctm-section-header">
							<view class="ctm-section-icon">
								<l-icon name="cloud" size="18" color="#6366f1"></l-icon>
							</view>
							<text class="ctm-section-title">第三方服务说明</text>
						</view>
						<view class="ctm-section-body">
							<text class="ctm-text">本社区功能由第三方服务运营，数据存储在第三方服务器，与校园官方教务系统无直接关联。</text>
						</view>
					</view>

					<!-- Section 2: 非官方应用 -->
					<view class="ctm-section">
						<view class="ctm-section-header">
							<view class="ctm-section-icon">
								<l-icon name="lock-off" size="18" color="#f59e0b"></l-icon>
							</view>
							<text class="ctm-section-title">非官方应用声明</text>
						</view>
						<view class="ctm-section-body">
							<text class="ctm-text">本应用属于第三方开发应用，非校园官方应用。社区内的内容、信息及行为由用户自行负责，与学校官方无关。</text>
						</view>
					</view>

					<!-- Section 3: 数据授权 -->
					<view class="ctm-section">
						<view class="ctm-section-header">
							<view class="ctm-section-icon">
								<l-icon name="hard-disk-storage" size="18" color="#10b981"></l-icon>
							</view>
							<text class="ctm-section-title">数据存储授权</text>
						</view>
						<view class="ctm-section-body">
							<text class="ctm-text">点击"同意"即表示您同意将您的部分教务相关数据（用于用户身份校验、鉴权等）存储在第三方服务器。</text>
						</view>
					</view>

					<!-- Section 4: 使用规范 -->
					<view class="ctm-section">
						<view class="ctm-section-header">
							<view class="ctm-section-icon">
								<l-icon name="article" size="18" color="#ec4899"></l-icon>
							</view>
							<text class="ctm-section-title">使用规范</text>
						</view>
						<view class="ctm-section-body">
							<text class="ctm-text">请遵守相关法律法规及社区公约，文明发言。社区管理方有权对违规内容进行处理。</text>
						</view>
					</view>

					<!-- Risk Warning -->
					<view class="ctm-warning">
						<view class="ctm-warning-icon">
							<l-icon name="error-circle" size="16" color="#ef4444"></l-icon>
						</view>
						<text class="ctm-warning-text">请您在充分了解并认可上述内容后再继续使用社区功能</text>
					</view>
				</view>
			</scroll-view>

			<!-- Footer -->
			<view class="ctm-footer">
				<view class="ctm-disagree-btn" @tap="handleDisagree">
					<text class="ctm-disagree-btn-text">不同意</text>
				</view>
				<view class="ctm-agree-btn" :class="{ 'ctm-agree-btn--loading': loading }" @tap="handleAgree">
					<text v-if="!loading" class="ctm-agree-btn-text">同意并继续</text>
					<text v-else class="ctm-agree-btn-text">处理中...</text>
				</view>
			</view>
		</view>
		</scroll-view>

	</t-popup>
</template>

<script setup>
import { ref, watch } from 'vue'
import lIcon from '@/uni_modules/lime-icon/components/l-icon/l-icon.vue'
import { agreeToTerms } from '@/composables/useCommunityTerms.js'

const props = defineProps({
	visible: {
		type: Boolean,
		default: false
	}
})

const emit = defineEmits(['update:visible', 'agree', 'disagree'])

const loading = ref(false)
const modalHeight = ref(500)

function updateModalHeight() {
	try {
		const info = uni.getSystemInfoSync()
		modalHeight.value = Math.round(info.screenHeight * 0.6)
	} catch (e) {
		modalHeight.value = 500
	}
}

function onPopupVisibleChange(val) {
	emit('update:visible', val)
}

function handleDisagree() {
	emit('update:visible', false)
	emit('disagree')
	uni.showModal({
		title: '提示',
		content: '您需要同意社区服务须知才能使用社区功能，是否返回？',
		confirmText: '返回',
		cancelText: '留在当前页',
		success: (res) => {
			if (res.confirm) {
				uni.switchTab({ url: '/pages/index/index' })
			}
		}
	})
}

async function handleAgree() {
	if (loading.value) return
	loading.value = true
	try {
		const ok = await agreeToTerms()
		emit('update:visible', false)
		emit('agree')
		if (ok) {
			uni.showToast({ title: '已同意须知，正在进入社区...', icon: 'none', duration: 1500 })
		}
	} catch (error) {
		console.error('同意失败:', error)
		uni.showToast({ title: '操作失败，请重试', icon: 'none' })
	} finally {
		loading.value = false
	}
}

watch(() => props.visible, (val) => {
	if (val) {
		loading.value = false
		updateModalHeight()
	}
}, { immediate: true })
</script>

<style lang="scss" scoped>
.ctm-modal {
	width: 100%;
	background: #fff;
	border-radius: 32rpx 32rpx 0 0;
	overflow: hidden;
	display: flex;
	flex-direction: column;
}

.ctm-header {
	padding: 40rpx 32rpx 24rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	border-bottom: 1px solid rgba(148, 163, 184, 0.12);
	flex-shrink: 0;
}

.ctm-icon-wrap {
	width: 96rpx;
	height: 96rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, rgba(99, 102, 241, 0.1), rgba(99, 102, 241, 0.05));
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 20rpx;
	border: 1px solid rgba(99, 102, 241, 0.15);
}

.ctm-title {
	font-size: 36rpx;
	font-weight: 800;
	color: var(--text-primary);
	margin-bottom: 8rpx;
}

.ctm-subtitle {
	font-size: 26rpx;
	color: var(--text-tertiary);
	text-align: center;
}

.ctm-scroll {
	flex: 1;
	min-height: 0;
}

.ctm-content {
	padding: 24rpx 32rpx;
}

.ctm-section {
	margin-bottom: 28rpx;
	background: #f8fafc;
	border-radius: 16rpx;
	padding: 24rpx;
	border: 1px solid rgba(148, 163, 184, 0.08);
}

.ctm-section-header {
	display: flex;
	align-items: center;
	gap: 12rpx;
	margin-bottom: 16rpx;
}

.ctm-section-icon {
	width: 44rpx;
	height: 44rpx;
	border-radius: 12rpx;
	background: #fff;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.06);
	flex-shrink: 0;
}

.ctm-section-title {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.ctm-section-body {
	padding-left: 56rpx;
}

.ctm-text {
	font-size: 26rpx;
	color: var(--text-secondary);
	line-height: 1.8;
}

.ctm-warning {
	display: flex;
	align-items: flex-start;
	gap: 12rpx;
	padding: 20rpx 24rpx;
	background: rgba(239, 68, 68, 0.06);
	border-radius: 16rpx;
	border: 1px solid rgba(239, 68, 68, 0.12);
	margin-top: 8rpx;
}

.ctm-warning-icon {
	width: 36rpx;
	height: 36rpx;
	border-radius: 50%;
	background: rgba(239, 68, 68, 0.1);
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
	margin-top: 2rpx;
}

.ctm-warning-text {
	font-size: 25rpx;
	color: #ef4444;
	line-height: 1.7;
	flex: 1;
}

.ctm-footer {
	display: flex;
	gap: 16rpx;
	padding: 24rpx 32rpx calc(24rpx + env(safe-area-inset-bottom));
	border-top: 1px solid rgba(148, 163, 184, 0.1);
	background: #fff;
	flex-shrink: 0;
}

.ctm-disagree-btn {
	flex: 1;
	height: 88rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 44rpx;
	background: #f1f5f9;
	border: 1px solid rgba(148, 163, 184, 0.2);

	.ctm-disagree-btn-text {
		font-size: 30rpx;
		font-weight: 600;
		color: var(--text-secondary);
	}

	&:active {
		background: #e2e8f0;
	}
}

.ctm-agree-btn {
	flex: 2;
	height: 88rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 44rpx;
	background: linear-gradient(135deg, #6366f1, #818cf8);
	box-shadow: 0 8rpx 24rpx rgba(99, 102, 241, 0.3);

	.ctm-agree-btn-text {
		font-size: 30rpx;
		font-weight: 700;
		color: #fff;
	}

	&:active {
		transform: scale(0.98);
		box-shadow: 0 4rpx 12rpx rgba(99, 102, 241, 0.25);
	}

	&--loading {
		opacity: 0.7;
		pointer-events: none;
	}
}
</style>
