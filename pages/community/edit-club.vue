<template>
	<view class="ec-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="ec-hero">
			<view class="ec-hero-bg"></view>
			<view class="ec-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="ec-hero-nav">
				<view class="ec-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="ec-hero-title">编辑社团</text>
				<view class="ec-submit-btn" :class="{ 'ec-submit-btn--disabled': !canSubmit }" @tap="handleSubmit">
					<text class="ec-submit-btn-text">{{ submitting ? '保存中...' : '保存' }}</text>
				</view>
			</view>

			<view class="ec-hero-content">
				<text class="ec-hero-sub">EDIT CLUB</text>
			</view>
		</view>

		<!-- 加载状态 -->
		<view v-if="loading" class="ec-loading">
			<view class="ec-loading-spinner"></view>
			<text class="ec-loading-text">加载中...</text>
		</view>

		<!-- 表单区域 -->
		<scroll-view v-else class="ec-scroll" scroll-y>
			<view class="ec-form">
				<!-- Logo 上传 -->
				<view class="ec-card ec-card--center">
					<view class="ec-logo-area" @tap="selectLogo">
						<view v-if="!form.logoUrl" class="ec-logo-placeholder">
							<l-icon name="photo" style="font-size: 48rpx; color: var(--text-tertiary); margin-bottom: 12rpx;"></l-icon>
							<text class="ec-logo-placeholder-text">更换Logo</text>
						</view>
						<image v-else class="ec-logo-image" :src="form.logoUrl" mode="aspectFill"></image>
					</view>
				</view>

				<!-- 社团名称 -->
				<view class="ec-card">
					<view class="ec-field-header">
						<text class="ec-field-title">社团名称</text>
						<text class="ec-required">*</text>
					</view>
					<input class="ec-input" v-model="form.name" placeholder="请输入社团名称" maxlength="50" placeholder-class="ec-placeholder" />
					<view class="ec-counter"><text class="ec-counter-text">{{ form.name.length }}/50</text></view>
				</view>

				<!-- 社团描述 -->
				<view class="ec-card">
					<view class="ec-field-header">
						<text class="ec-field-title">社团描述</text>
						<text class="ec-required">*</text>
					</view>
					<textarea class="ec-textarea" v-model="form.description" placeholder="描述一下你的社团吧..." maxlength="500" auto-height placeholder-class="ec-placeholder"></textarea>
					<view class="ec-counter"><text class="ec-counter-text">{{ form.description.length }}/500</text></view>
				</view>

				<!-- 标签 -->
				<view class="ec-card">
					<view class="ec-field-header">
						<text class="ec-field-title">标签</text>
						<text class="ec-optional">(逗号分隔)</text>
					</view>
					<input class="ec-input" v-model="form.tags" placeholder="如: 运动, 音乐, 编程" maxlength="100" placeholder-class="ec-placeholder" />
				</view>

				<!-- 联系方式 -->
				<view class="ec-card">
					<view class="ec-field-header">
						<text class="ec-field-title">联系方式</text>
					</view>
					<input class="ec-input" v-model="form.contactInfo" placeholder="QQ/微信/手机号" maxlength="100" placeholder-class="ec-placeholder" />
				</view>
			</view>
		</scroll-view>
	</view>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { getClubDetail, updateClub } from '../api/community.js'
import { showToast, navigateBack } from '../../pages/api/page.js'

const statusBarHeight = ref(20)
const clubId = ref(null)
const loading = ref(true)
const submitting = ref(false)

const form = reactive({ name: '', description: '', logoUrl: '', tags: '', contactInfo: '' })

function initStatusBar() {
	try { const info = uni.getSystemInfoSync(); statusBarHeight.value = info.statusBarHeight || 20 } catch (e) { statusBarHeight.value = 20 }
}

function goBack() { navigateBack() }

const canSubmit = computed(() => form.name.trim() && form.description.trim() && !submitting.value)

async function loadClubData() {
	loading.value = true
	try {
		const res = await getClubDetail(clubId.value)
		if (res.success && res.result) {
			const club = res.result
			form.name = club.name
			form.description = club.description
			form.logoUrl = club.logoUrl
			form.tags = club.tags
			form.contactInfo = club.contactInfo
		} else {
			showToast('获取社团详情失败')
		}
	} catch (e) {
		showToast('网络请求失败')
	} finally {
		loading.value = false
	}
}

async function handleSubmit() {
	if (!canSubmit.value) return
	submitting.value = true
	try {
		const res = await updateClub(clubId.value, form)
		if (res.success) {
			showToast('保存成功')
			setTimeout(() => navigateBack(), 1500)
		} else {
			showToast(res.message || '保存失败')
		}
	} catch (e) {
		showToast('保存失败')
	} finally {
		submitting.value = false
	}
}

function selectLogo() {
	uni.chooseImage({
		count: 1,
		sizeType: ['compressed'],
		success: (res) => { form.logoUrl = res.tempFilePaths[0] }
	})
}

onMounted(() => {
	initStatusBar()
	const pages = getCurrentPages()
	const currentPage = pages[pages.length - 1]
	const options = currentPage?.options || {}
	clubId.value = options.id
	if (!clubId.value) {
		showToast('社团ID不能为空')
		navigateBack()
		return
	}
	loadClubData()
})
</script>

<style lang="scss" scoped>
.ec-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
}

.ec-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 200rpx;
}

.ec-hero-bg {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg, #1e3a8a 0%, #1e40af 25%, #2563eb 55%, #3b82f6 75%, #93c5fd 100%);
	z-index: 0;
}

.ec-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg, rgba(30, 58, 138, 0.65) 0%, rgba(37, 99, 235, 0.4) 50%, rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.ec-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.ec-back-btn {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.18);
	border-radius: 50%;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.25);

	&:active { background: rgba(255, 255, 255, 0.28); transform: scale(0.95); }
}

.ec-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
}

.ec-submit-btn {
	padding: 10rpx 24rpx;
	background: rgba(255, 255, 255, 0.25);
	border-radius: 100rpx;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.3);

	.ec-submit-btn-text { font-size: 26rpx; font-weight: 700; color: #fff; }

	&--disabled { opacity: 0.5; }
	&:active:not(.ec-submit-btn--disabled) { background: rgba(255, 255, 255, 0.35); transform: scale(0.95); }
}

.ec-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 8rpx;
}

.ec-hero-sub {
	font-size: 20rpx;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.6);
	letter-spacing: 2px;
}

.ec-loading {
	position: fixed;
	top: 0; left: 0; right: 0; bottom: 0;
	background: rgba(255, 255, 255, 0.95);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 16rpx;
	z-index: 1000;

	.ec-loading-text { font-size: 28rpx; color: var(--text-secondary); }
}

.ec-loading-spinner {
	width: 64rpx;
	height: 64rpx;
	border: 5rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: ec-spin 0.8s linear infinite;
}

@keyframes ec-spin { to { transform: rotate(360deg); } }

.ec-scroll { flex: 1; min-height: 0; }

.ec-form {
	padding: 24rpx;
	padding-bottom: 60rpx;
}

.ec-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 24rpx;
	margin-bottom: 20rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);

	&--center { display: flex; justify-content: center; }
}

.ec-logo-area {
	width: 160rpx;
	height: 160rpx;
	border-radius: 40rpx;
	overflow: hidden;
	background: var(--bg-muted);
	border: 2px dashed rgba(148, 163, 184, 0.4);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	transition: all 0.2s ease;

	&:active { border-color: var(--primary-500); transform: scale(0.98); }
}

.ec-logo-placeholder-text { font-size: 24rpx; color: var(--text-secondary); font-weight: 600; margin-top: 4rpx; }
.ec-logo-image { width: 100%; height: 100%; }

.ec-field-header { display: flex; align-items: center; gap: 8rpx; margin-bottom: 16rpx; }
.ec-field-title { font-size: 28rpx; font-weight: 700; color: var(--text-primary); }
.ec-required { font-size: 24rpx; color: var(--error-color); }
.ec-optional { font-size: 22rpx; color: var(--text-tertiary); }

.ec-input {
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

.ec-placeholder { color: var(--text-tertiary); }

.ec-textarea {
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

.ec-counter {
	display: flex;
	justify-content: flex-end;
	margin-top: 10rpx;

	.ec-counter-text { font-size: 22rpx; color: var(--text-tertiary); }
}
</style>
