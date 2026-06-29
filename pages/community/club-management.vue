<template>
	<view class="cm-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="cm-hero">
			<view class="cm-hero-bg"></view>
			<view class="cm-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="cm-hero-nav">
				<view class="cm-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="cm-hero-title">社团管理</text>
				<view class="cm-save-btn" :class="{ 'cm-save-btn--disabled': !canSave }" @tap="handleSave">
					<l-icon v-if="saving" name="refresh" style="font-size: 14px; margin-right: 8rpx; color: #fff; animation: cm-spin 1s linear infinite;"></l-icon>
					<text class="cm-save-btn-text">{{ saving ? '保存中...' : '保存' }}</text>
				</view>
			</view>

			<view class="cm-hero-content">
				<text class="cm-hero-sub">CLUB MANAGEMENT</text>
			</view>
		</view>

		<!-- 表单区域 -->
		<scroll-view class="cm-scroll" scroll-y>
			<view class="cm-form">
				<!-- Logo 上传 -->
				<view class="cm-card cm-card--center">
					<view class="cm-logo-area" @tap="selectLogo">
						<view v-if="!form.logoUrl" class="cm-logo-placeholder">
							<l-icon name="photo" style="font-size: 48rpx; color: var(--text-tertiary); margin-bottom: 12rpx;"></l-icon>
							<text class="cm-logo-placeholder-text">上传图标</text>
						</view>
						<view v-else class="cm-logo-preview">
							<image class="cm-logo-image" :src="form.logoUrl" mode="aspectFill"></image>
							<view class="cm-logo-overlay">
								<view class="cm-logo-overlay-btn">
									<l-icon name="photo" style="font-size: 14px; color: #fff;"></l-icon>
									<text class="cm-logo-overlay-text">更换</text>
								</view>
							</view>
						</view>
					</view>
				</view>

				<!-- 社团名称 -->
				<view class="cm-card">
					<view class="cm-field-header">
						<text class="cm-field-title">社团名称</text>
						<text class="cm-required">*</text>
					</view>
					<input class="cm-input" v-model="form.name" placeholder="请输入社团名称" maxlength="50" placeholder-class="cm-placeholder" />
					<view class="cm-counter"><text class="cm-counter-text">{{ form.name.length }}/50</text></view>
				</view>

				<!-- 社团描述 -->
				<view class="cm-card">
					<view class="cm-field-header">
						<text class="cm-field-title">社团描述</text>
						<text class="cm-required">*</text>
					</view>
					<textarea class="cm-textarea" v-model="form.description" placeholder="请输入社团简介" maxlength="500" auto-height placeholder-class="cm-placeholder"></textarea>
					<view class="cm-counter"><text class="cm-counter-text">{{ form.description.length }}/500</text></view>
				</view>

				<!-- 社团标签 -->
				<view class="cm-card">
					<view class="cm-field-header">
						<text class="cm-field-title">社团标签</text>
						<text class="cm-optional">(可选)</text>
					</view>
					<input class="cm-input" v-model="form.tags" placeholder="输入标签" maxlength="100" placeholder-class="cm-placeholder" />
					<view class="cm-tip"><text class="cm-tip-text">输入标签，多个用逗号分隔</text></view>
				</view>

				<!-- 联系方式 -->
				<view class="cm-card">
					<view class="cm-field-header">
						<text class="cm-field-title">联系信息</text>
						<text class="cm-optional">(可选)</text>
					</view>
					<input class="cm-input" v-model="form.contactInfo" placeholder="请输入联系方式" maxlength="100" placeholder-class="cm-placeholder" />
				</view>

				<!-- 危险区域 -->
				<view class="cm-danger">
					<view class="cm-danger-header">
						<text class="cm-danger-title">危险区域</text>
					</view>
					<view class="cm-danger-warning">
						<l-icon name="info-circle" style="font-size: 16px; color: var(--warning-color); margin-right: 12rpx;"></l-icon>
						<text class="cm-danger-warning-text">删除社团后将无法恢复，请谨慎操作</text>
					</view>
					<view class="cm-danger-btn" @tap="handleDeleteClub">
						<l-icon name="delete" style="font-size: 18px; margin-right: 12rpx; color: #fff;"></l-icon>
						<text class="cm-danger-btn-text">删除社团</text>
					</view>
				</view>
			</view>
		</scroll-view>

		<!-- 社区服务须知弹窗 -->
		<CommunityTermsModal
			v-model:visible="termsVisible"
			@agree="onTermsAgreed"
			@disagree="onTermsDisagree"
		/>
	</view>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { getClubDetail, updateClub, deleteClub } from '../api/community.js'
import { showToast, showModal, navigateBack } from '../../pages/api/page.js'
import { BASE_URL } from '../../utils/request.config.js'
import { hasAgreedToTerms } from '@/composables/useCommunityTerms.js'
import CommunityTermsModal from '@/components/CommunityTermsModal.vue'

const statusBarHeight = ref(20)
const clubId = ref(null)
const saving = ref(false)
const originalForm = ref({})
const termsVisible = ref(false)

const form = reactive({ name: '', description: '', logoUrl: '', tags: '', contactInfo: '' })

function initStatusBar() {
	try { const info = uni.getSystemInfoSync(); statusBarHeight.value = info.statusBarHeight || 20 } catch (e) { statusBarHeight.value = 20 }
}

function goBack() { navigateBack() }

const canSave = computed(() => form.name.trim() && form.description.trim() && JSON.stringify(form) !== JSON.stringify(originalForm.value) && !saving.value)

async function loadClubDetail() {
	try {
		const result = await getClubDetail(clubId.value)
		if (result.success) {
			const club = result.result.club
			form.name = club.name || ''
			form.description = club.description || ''
			form.logoUrl = club.logoUrl || ''
			form.tags = club.tags || ''
			form.contactInfo = club.contactInfo || ''
			originalForm.value = { ...form }
		} else {
			showToast({ title: result.message || '加载失败', icon: 'none' })
			navigateBack()
		}
	} catch (error) {
		showToast({ title: '加载失败', icon: 'none' })
		navigateBack()
	}
}

async function selectLogo() {
	try {
		const res = await uni.chooseImage({ count: 1, sizeType: ['compressed'], sourceType: ['album', 'camera'] })
		if (res.tempFilePaths && res.tempFilePaths.length > 0) {
			uni.showLoading({ title: '上传中...' })
			const uploadResult = await uni.uploadFile({
				url: `${BASE_URL}/api/upload/image`,
				filePath: res.tempFilePaths[0],
				name: 'file',
				header: { 'Authorization': `Bearer ${uni.getStorageSync('token')}` }
			})
			uni.hideLoading()
			if (uploadResult.statusCode === 200) {
				const data = JSON.parse(uploadResult.data)
				if (data.success) form.logoUrl = data.result.url
			}
		}
	} catch (error) {
		uni.hideLoading()
		showToast({ title: '上传失败', icon: 'none' })
	}
}

async function handleSave() {
	if (!canSave.value) return
	saving.value = true
	try {
		const result = await updateClub(clubId.value, {
			name: form.name.trim(), description: form.description.trim(),
			logoUrl: form.logoUrl, tags: form.tags.trim(), contactInfo: form.contactInfo.trim()
		})
		if (result.success) {
			originalForm.value = { ...form }
			uni.setStorageSync('communityNeedRefresh', true)
			showToast({ title: '保存成功', icon: 'success' })
		} else {
			showToast({ title: result.message || '保存失败', icon: 'none' })
		}
	} catch (error) {
		showToast({ title: '保存失败', icon: 'none' })
	} finally {
		saving.value = false
	}
}

async function handleDeleteClub() {
	const confirmed = await showModal({ title: '确定', content: '内容删除后无法恢复，请谨慎操作', confirmText: '删除', cancelText: '取消' })
	if (confirmed) {
		try {
			const result = await deleteClub(clubId.value)
			if (result.success) {
				uni.setStorageSync('communityNeedRefresh', true)
				showToast({ title: '删除成功', icon: 'success' })
				setTimeout(() => navigateBack(), 1500)
			}
		} catch (error) {
			showToast({ title: '删除失败', icon: 'none' })
		}
	}
}

async function initMounted() {
	initStatusBar()
	const agreed = await hasAgreedToTerms()
	if (!agreed) {
		termsVisible.value = true
		return
	}
	const pages = getCurrentPages()
	const currentPage = pages[pages.length - 1]
	const options = currentPage?.options || {}
	if (options.id) {
		clubId.value = options.id
		loadClubDetail()
	}
}

onMounted(initMounted)

function onTermsAgreed() {
	termsVisible.value = false
	const pages = getCurrentPages()
	const currentPage = pages[pages.length - 1]
	const options = currentPage?.options || {}
	if (options.id) {
		clubId.value = options.id
		loadClubDetail()
	}
}

function onTermsDisagree() {
	termsVisible.value = false
	navigateBack()
}
</script>

<style lang="scss" scoped>
.cm-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
}

.cm-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 200rpx;
}

.cm-hero-bg {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg, #1e3a8a 0%, #1e40af 25%, #2563eb 55%, #3b82f6 75%, #93c5fd 100%);
	z-index: 0;
}

.cm-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg, rgba(30, 58, 138, 0.65) 0%, rgba(37, 99, 235, 0.4) 50%, rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.cm-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.cm-back-btn {
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

.cm-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
}

.cm-save-btn {
	display: flex;
	align-items: center;
	padding: 10rpx 24rpx;
	background: rgba(255, 255, 255, 0.25);
	border-radius: 100rpx;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.3);

	.cm-save-btn-text { font-size: 26rpx; font-weight: 700; color: #fff; }

	&--disabled { opacity: 0.5; }
	&:active:not(.cm-save-btn--disabled) { background: rgba(255, 255, 255, 0.35); transform: scale(0.95); }
}

@keyframes cm-spin { to { transform: rotate(360deg); } }

.cm-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 8rpx;
}

.cm-hero-sub {
	font-size: 20rpx;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.6);
	letter-spacing: 2px;
}

.cm-scroll { flex: 1; min-height: 0; }

.cm-form {
	padding: 24rpx;
	padding-bottom: 60rpx;
}

.cm-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 24rpx;
	margin-bottom: 20rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);

	&--center { display: flex; justify-content: center; }
}

.cm-logo-area {
	display: flex;
	justify-content: center;
}

.cm-logo-placeholder {
	width: 160rpx;
	height: 160rpx;
	border-radius: 40rpx;
	border: 2px dashed rgba(148, 163, 184, 0.4);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	background: var(--bg-muted);
	transition: all 0.2s ease;

	&:active { border-color: var(--primary-500); background: rgba(59, 102, 241, 0.05); transform: scale(0.98); }
}

.cm-logo-placeholder-text { font-size: 26rpx; color: var(--text-secondary); font-weight: 600; margin-top: 4rpx; }

.cm-logo-preview {
	position: relative;
	width: 160rpx;
	height: 160rpx;
	border-radius: 40rpx;
	overflow: hidden;
	box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.15);

	&:active { transform: scale(0.98); }
}

.cm-logo-image { width: 100%; height: 100%; }

.cm-logo-overlay {
	position: absolute;
	bottom: 0; left: 0; right: 0;
	background: linear-gradient(transparent, rgba(0, 0, 0, 0.6));
	padding: 12rpx 0 10rpx;
	display: flex;
	justify-content: center;
}

.cm-logo-overlay-btn {
	display: flex;
	align-items: center;
	gap: 8rpx;
	padding: 6rpx 16rpx;
	background: rgba(255, 255, 255, 0.2);
	border-radius: 100rpx;
}

.cm-logo-overlay-text { font-size: 20rpx; color: #fff; font-weight: 600; }

.cm-field-header { display: flex; align-items: center; gap: 8rpx; margin-bottom: 16rpx; }
.cm-field-title { font-size: 28rpx; font-weight: 700; color: var(--text-primary); }
.cm-required { font-size: 24rpx; color: var(--error-color); }
.cm-optional { font-size: 22rpx; color: var(--text-tertiary); }

.cm-input {
	width: 100%;
	height: 88rpx;
	padding: 0 20rpx;
	border-radius: 16rpx;
	background: var(--bg-muted);
	border: 1px solid rgba(148, 163, 184, 0.15);
	font-size: 30rpx;
	color: var(--text-primary);
	transition: border-color 0.2s ease;

	&:focus { border-color: var(--primary-500); }
}

.cm-placeholder { color: var(--text-tertiary); }

.cm-textarea {
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

	&:focus { border-color: var(--primary-500); }
}

.cm-counter {
	display: flex;
	justify-content: flex-end;
	margin-top: 10rpx;

	.cm-counter-text { font-size: 22rpx; color: var(--text-tertiary); }
}

.cm-tip {
	margin-top: 10rpx;

	.cm-tip-text { font-size: 22rpx; color: var(--text-tertiary); }
}

.cm-danger {
	background: rgba(239, 68, 68, 0.03);
	border-radius: 24rpx;
	padding: 24rpx;
	margin-top: 48rpx;
	border: 1px solid rgba(239, 68, 68, 0.15);
}

.cm-danger-header { margin-bottom: 16rpx; }
.cm-danger-title { font-size: 28rpx; font-weight: 700; color: var(--error-color); }

.cm-danger-warning {
	display: flex;
	align-items: center;
	padding: 16rpx;
	background: rgba(245, 158, 11, 0.08);
	border-radius: 12rpx;
	margin-bottom: 20rpx;

	.cm-danger-warning-text { font-size: 24rpx; color: var(--warning-color); }
}

.cm-danger-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	height: 88rpx;
	background: linear-gradient(135deg, #ef4444, #dc2626);
	border-radius: 20rpx;
	box-shadow: 0 8rpx 24rpx rgba(239, 68, 68, 0.3);

	.cm-danger-btn-text { font-size: 30rpx; font-weight: 700; color: #fff; }

	&:active { transform: translateY(2rpx); }
}
</style>
