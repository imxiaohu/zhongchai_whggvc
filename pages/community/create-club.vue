<template>
	<view class="cc-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="cc-hero">
			<view class="cc-hero-bg"></view>
			<view class="cc-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="cc-hero-nav">
				<view class="cc-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="cc-hero-title">创社团</text>
				<view class="cc-submit-btn" :class="{ 'cc-submit-btn--disabled': !canSubmit }" @tap="handleSubmit">
					<l-icon v-if="!submitting" name="check-circle-filled" style="font-size: 14px; margin-right: 8rpx; color: #fff;"></l-icon>
					<text class="cc-submit-btn-text">{{ submitting ? '创建中...' : '发布' }}</text>
				</view>
			</view>

			<view class="cc-hero-content">
				<text class="cc-hero-sub">CREATE CLUB</text>
			</view>
		</view>

		<!-- 表单区域 -->
		<scroll-view class="cc-scroll" scroll-y>
			<view class="cc-form">
				<!-- Logo 上传 -->
				<view class="cc-card cc-card--center">
					<view class="cc-logo-area" @tap="selectLogo">
						<view v-if="!form.logoUrl" class="cc-logo-placeholder">
							<l-icon name="photo" style="font-size: 48rpx; color: var(--text-tertiary); margin-bottom: 12rpx;"></l-icon>
							<text class="cc-logo-placeholder-text">选择图片</text>
							<text class="cc-logo-placeholder-hint">建议尺寸：200x200</text>
						</view>
						<view v-else class="cc-logo-preview">
							<image class="cc-logo-image" :src="form.logoUrl" mode="aspectFill"></image>
							<view class="cc-logo-overlay">
								<view class="cc-logo-overlay-btn">
									<l-icon name="photo" style="font-size: 14px; color: #fff;"></l-icon>
									<text class="cc-logo-overlay-text">更换</text>
								</view>
							</view>
						</view>
					</view>
				</view>

				<!-- 社团名称 -->
				<view class="cc-card">
					<view class="cc-field-header">
						<text class="cc-field-title">社团名称</text>
						<text class="cc-required">*</text>
					</view>
					<input
						class="cc-input"
						v-model="form.name"
						placeholder="请输入社团名称"
						maxlength="50"
						placeholder-class="cc-placeholder"
					/>
					<view class="cc-counter">
						<text class="cc-counter-text">{{ form.name.length }}/50</text>
					</view>
				</view>

				<!-- 社团描述 -->
				<view class="cc-card">
					<view class="cc-field-header">
						<text class="cc-field-title">社团描述</text>
						<text class="cc-required">*</text>
					</view>
					<textarea
						class="cc-textarea"
						v-model="form.description"
						placeholder="请输入描述内容"
						maxlength="500"
						auto-height
						placeholder-class="cc-placeholder"
					></textarea>
					<view class="cc-counter">
						<text class="cc-counter-text">{{ form.description.length }}/500</text>
					</view>
				</view>

				<!-- 社团标签 -->
				<view class="cc-card">
					<view class="cc-field-header">
						<text class="cc-field-title">社团标签</text>
						<text class="cc-optional">(可选)</text>
					</view>
					<input
						class="cc-input"
						v-model="form.tags"
						placeholder="输入标签，多个用逗号分隔"
						maxlength="100"
						placeholder-class="cc-placeholder"
					/>
					<view class="cc-tip">
						<text class="cc-tip-text">输入标签</text>
					</view>
				</view>

				<!-- 联系方式 -->
				<view class="cc-card">
					<view class="cc-field-header">
						<text class="cc-field-title">联系方式</text>
						<text class="cc-optional">(可选)</text>
					</view>
					<input
						class="cc-input"
						v-model="form.contactInfo"
						placeholder="请输入联系方式"
						maxlength="100"
						placeholder-class="cc-placeholder"
					/>
				</view>

				<!-- 提交按钮 -->
				<view class="cc-submit-section">
					<view class="cc-submit-btn-main" :class="{ 'cc-submit-btn-main--disabled': !canSubmit }" @tap="handleSubmit">
						<text class="cc-submit-btn-main-text">创建社团</text>
					</view>
				</view>
			</view>
		</scroll-view>
		<!-- 社区服务须知弹窗 -->
		<CommunityTermsModal
			v-model:visible="termsVisible"
			@agree="onTermsAgreed"
			@disagree="onTermsDisagree"
		></CommunityTermsModal>
	</view>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { createClub } from '../api/community.js'
import { showToast, navigateBack } from '../../pages/api/page.js'
import { uploadSingleImage, validateImage, previewImages } from '../../utils/imageUpload.js'
import { hasAgreedToTerms } from '@/composables/useCommunityTerms.js'
import CommunityTermsModal from '@/components/CommunityTermsModal.vue'

const statusBarHeight = ref(20)
const submitting = ref(false)
const termsVisible = ref(false)

const form = reactive({
	name: '',
	description: '',
	logoUrl: '',
	tags: '',
	contactInfo: ''
})

function initStatusBar() {
	try {
		const systemInfo = uni.getSystemInfoSync()
		statusBarHeight.value = systemInfo.statusBarHeight || 20
	} catch (e) {
		statusBarHeight.value = 20
	}
}

function goBack() {
	uni.navigateBack()
}

const canSubmit = computed(() => {
	return form.name.trim() && form.description.trim() && !submitting.value
})

async function selectLogo() {
	try {
		const filePaths = await new Promise((resolve, reject) => {
			uni.chooseImage({
				count: 1,
				sizeType: ['compressed'],
				sourceType: ['album', 'camera'],
				success: (res) => resolve(res.tempFilePaths),
				fail: reject
			})
		})
		if (filePaths && filePaths.length > 0) {
			await uploadLogo(filePaths[0])
		}
	} catch (error) {
		console.error('选择图片失败:', error)
		showToast({ title: '选择图片失败', icon: 'none' })
	}
}

async function uploadLogo(filePath) {
	try {
		await validateImage(filePath, 5 * 1024 * 1024)
		const result = await uploadSingleImage(filePath, { showLoading: true, compress: true, quality: 0.8 })
		if (result.success && result.url) {
			form.logoUrl = result.url
			showToast({ title: '上传成功', icon: 'success' })
		} else {
			throw new Error(result.error || '上传失败')
		}
	} catch (error) {
		console.error('上传logo失败:', error)
		showToast({ title: error.message || '上传失败', icon: 'none' })
	}
}

async function handleSubmit() {
	if (!canSubmit.value) return
	if (!form.name.trim()) { showToast({ title: '请输入社团名称', icon: 'none' }); return }
	if (!form.description.trim()) { showToast({ title: '请输入描述内容', icon: 'none' }); return }
	submitting.value = true
	try {
		const submitData = {
			name: form.name.trim(),
			description: form.description.trim(),
			logoUrl: form.logoUrl,
			tags: form.tags.trim(),
			contactInfo: form.contactInfo.trim()
		}
		const result = await createClub(submitData)
		if (result.success) {
			showToast({ title: '创建成功', icon: 'success' })
			uni.setStorageSync('communityNeedRefresh', true)
			setTimeout(() => navigateBack(), 1500)
		} else {
			showToast({ title: result.message || '创建失败', icon: 'none' })
		}
	} catch (error) {
		console.error('创建社团失败:', error)
		showToast({ title: '创建失败', icon: 'none' })
	} finally {
		submitting.value = false
	}
}

async function checkTerms() {
	const agreed = await hasAgreedToTerms()
	if (!agreed) {
		termsVisible.value = true
	}
}

function onTermsAgreed() {
	termsVisible.value = false
}

function onTermsDisagree() {
	termsVisible.value = false
	navigateBack()
}

initStatusBar()
checkTerms()
</script>

<style lang="scss" scoped>
/* ============================================
   Create Club - Hero Style
   ============================================ */

.cc-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
}

/* ---- Hero Section ---- */
.cc-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 200rpx;
}

.cc-hero-bg {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		#1e3a8a 0%,
		#1e40af 25%,
		#2563eb 55%,
		#3b82f6 75%,
		#93c5fd 100%);
	z-index: 0;
}

.cc-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.65) 0%,
		rgba(37, 99, 235, 0.4) 50%,
		rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.cc-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.cc-back-btn {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.18);
	border-radius: 50%;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.25);

	&:active {
		background: rgba(255, 255, 255, 0.28);
		transform: scale(0.95);
	}
}

.cc-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.cc-submit-btn {
	display: flex;
	align-items: center;
	padding: 10rpx 24rpx;
	background: rgba(255, 255, 255, 0.25);
	border-radius: 100rpx;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.3);
	transition: all 0.2s ease;

	.cc-submit-btn-text {
		font-size: 26rpx;
		font-weight: 700;
		color: #fff;
	}

	&--disabled { opacity: 0.5; }
	&:active:not(.cc-submit-btn--disabled) { background: rgba(255, 255, 255, 0.35); transform: scale(0.95); }
}

.cc-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 8rpx;
}

.cc-hero-sub {
	font-size: 20rpx;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.6);
	letter-spacing: 2px;
}

/* ---- Scroll ---- */
.cc-scroll {
	flex: 1;
	min-height: 0;
}

.cc-form {
	padding: 24rpx;
	padding-bottom: 60rpx;
}

/* ---- Card ---- */
.cc-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 24rpx;
	margin-bottom: 20rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);

	&--center {
		display: flex;
		justify-content: center;
	}
}

/* ---- Logo ---- */
.cc-logo-area {
	display: flex;
	justify-content: center;
}

.cc-logo-placeholder {
	width: 200rpx;
	height: 200rpx;
	border-radius: 50%;
	border: 2px dashed rgba(148, 163, 184, 0.4);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	background: var(--bg-muted);
	transition: all 0.2s ease;

	&:active {
		border-color: var(--primary-500);
		background: rgba(59, 102, 241, 0.05);
		transform: scale(0.98);
	}
}

.cc-logo-placeholder-text {
	font-size: 26rpx;
	color: var(--text-secondary);
	font-weight: 600;
	margin-top: 4rpx;
}

.cc-logo-placeholder-hint {
	font-size: 20rpx;
	color: var(--text-tertiary);
	margin-top: 4rpx;
}

.cc-logo-preview {
	position: relative;
	width: 200rpx;
	height: 200rpx;
	border-radius: 50%;
	overflow: hidden;
	box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.15);
	transition: all 0.2s ease;

	&:active { transform: scale(0.98); }
}

.cc-logo-image {
	width: 100%;
	height: 100%;
}

.cc-logo-overlay {
	position: absolute;
	bottom: 0;
	left: 0;
	right: 0;
	background: linear-gradient(transparent, rgba(0, 0, 0, 0.6));
	padding: 16rpx 0 12rpx;
	display: flex;
	justify-content: center;
}

.cc-logo-overlay-btn {
	display: flex;
	align-items: center;
	gap: 8rpx;
	padding: 8rpx 20rpx;
	background: rgba(255, 255, 255, 0.2);
	border-radius: 100rpx;
}

.cc-logo-overlay-text {
	font-size: 22rpx;
	color: #fff;
	font-weight: 600;
}

/* ---- Field ---- */
.cc-field-header {
	display: flex;
	align-items: center;
	gap: 8rpx;
	margin-bottom: 16rpx;
}

.cc-field-title {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.cc-required {
	font-size: 24rpx;
	color: var(--error-color);
}

.cc-optional {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.cc-input {
	width: 100%;
	height: 88rpx;
	padding: 0 20rpx;
	border-radius: 16rpx;
	background: var(--bg-muted);
	border: 1px solid rgba(148, 163, 184, 0.15);
	font-size: 30rpx;
	color: var(--text-primary);
	transition: border-color 0.2s ease;

	&:focus {
		border-color: var(--primary-500);
	}
}

.cc-placeholder {
	color: var(--text-tertiary);
}

.cc-textarea {
	width: 100%;
	min-height: 200rpx;
	padding: 20rpx;
	border-radius: 16rpx;
	background: var(--bg-muted);
	border: 1px solid rgba(148, 163, 184, 0.15);
	font-size: 30rpx;
	color: var(--text-primary);
	line-height: 1.6;
	resize: none;
	transition: border-color 0.2s ease;

	&:focus {
		border-color: var(--primary-500);
	}
}

.cc-counter {
	display: flex;
	justify-content: flex-end;
	margin-top: 10rpx;
}

.cc-counter-text {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

.cc-tip {
	margin-top: 10rpx;
}

.cc-tip-text {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

/* ---- Submit ---- */
.cc-submit-section {
	margin-top: 48rpx;
}

.cc-submit-btn-main {
	height: 100rpx;
	background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
	border-radius: 50rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 8rpx 32rpx rgba(37, 99, 235, 0.3);
	transition: all 0.2s ease;

	.cc-submit-btn-main-text {
		font-size: 32rpx;
		font-weight: 800;
		color: #fff;
	}

	&--disabled {
		background: var(--bg-muted);
		box-shadow: none;
		.cc-submit-btn-main-text { color: var(--text-tertiary); }
	}

	&:active:not(.cc-submit-btn-main--disabled) {
		transform: translateY(2rpx);
		box-shadow: 0 4rpx 16rpx rgba(37, 99, 235, 0.25);
	}
}
</style>
