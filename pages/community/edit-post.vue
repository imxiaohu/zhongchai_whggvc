<template>
	<view class="ep-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="ep-hero">
			<view class="ep-hero-bg"></view>
			<view class="ep-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="ep-hero-nav">
				<view class="ep-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="ep-hero-title">编辑动态</text>
				<view class="ep-submit-btn" :class="{ 'ep-submit-btn--disabled': !canSubmit }" @tap="handleSubmit">
					<text class="ep-submit-btn-text">{{ submitting ? '保存中...' : '保存' }}</text>
				</view>
			</view>

			<view class="ep-hero-content">
				<text class="ep-hero-sub">EDIT POST</text>
			</view>
		</view>

		<!-- 加载状态 -->
		<view v-if="loading" class="ep-loading">
			<view class="ep-loading-spinner"></view>
			<text>加载中...</text>
		</view>

		<!-- 表单区域 -->
		<scroll-view v-else class="ep-scroll" scroll-y :style="scrollViewStyle">
			<view class="ep-form">
				<!-- 标题 -->
				<view class="ep-card">
					<input class="ep-title-input" v-model="form.title" placeholder="请输入标题" maxlength="100" placeholder-class="ep-placeholder" />
					<view class="ep-counter"><text>{{ form.title.length }}/100</text></view>
				</view>

				<!-- 内容 -->
				<view class="ep-card">
					<textarea class="ep-content-textarea" v-model="form.content" placeholder="分享你的校园生活..." maxlength="5000" auto-height placeholder-class="ep-placeholder"></textarea>
					<view class="ep-counter"><text class="ep-counter-text">{{ form.content.length }}/5000</text></view>
				</view>

				<!-- 图片上传 -->
				<view class="ep-card">
					<view class="ep-section-header">
						<text class="ep-section-title">图片展示</text>
						<text class="ep-section-optional">(可选)</text>
					</view>
					<view class="ep-images-grid" :class="getImageGridClassFn(displayImages.length + (hasMoreImages && !showAllImages ? 1 : 0) + (form.images.length < 20 ? 1 : 0))">
						<view v-for="(image, index) in displayImages" :key="index" class="ep-image-item" :class="{ 'ep-image-item--svg': isSvgImageFn(image) }">
							<image class="ep-uploaded-image" :src="image" mode="aspectFill" @tap="previewImage(image, form.images)"></image>
							<view v-if="isSvgImageFn(image)" class="ep-svg-overlay"><text>SVG</text></view>
							<view class="ep-image-actions">
								<view class="ep-image-action ep-image-action--preview" @tap="previewImage(image, form.images)">
									<l-icon name="browse" style="font-size: 12px; color: #fff;"></l-icon>
								</view>
								<view class="ep-image-action ep-image-action--remove" @tap="removeImage(index)">
									<l-icon name="close" style="font-size: 12px; color: #fff;"></l-icon>
								</view>
							</view>
						</view>
						<view v-if="form.images.length < 20" class="ep-add-btn" @tap="chooseImage">
							<l-icon name="add" style="font-size: 32px; color: var(--text-tertiary);"></l-icon>
							<text class="ep-add-text">添加图片</text>
						</view>
					</view>
				</view>
			</view>
		</scroll-view>
	</view>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { getPostDetail, updatePost } from '../api/community.js'
import { showToast, navigateBack } from '../../pages/api/page.js'
import { previewImages, isSvgImage, getImageGridClass } from '../../utils/imageUpload.js'

const statusBarHeight = ref(20)
const postId = ref(null)
const loading = ref(true)
const submitting = ref(false)
const showAllImages = ref(false)
const maxDisplayImages = ref(9)
const scrollViewHeight = ref('calc(100vh - 200rpx)')

const form = reactive({ title: '', content: '', images: [], type: 'article' })

function initStatusBar() {
	try { const info = uni.getSystemInfoSync(); statusBarHeight.value = info.statusBarHeight || 20 } catch (e) { statusBarHeight.value = 20 }
}

function goBack() { navigateBack() }

const canSubmit = computed(() => form.title.trim() && form.content.trim() && !submitting.value)

const displayImages = computed(() => {
	if (showAllImages.value || form.images.length <= maxDisplayImages.value) return form.images
	return form.images.slice(0, maxDisplayImages.value - 1)
})

const hasMoreImages = computed(() => form.images.length > maxDisplayImages.value)

const scrollViewStyle = computed(() => ({ height: `calc(100vh - ${statusBarHeight.value}px - 200rpx)` }))

async function loadPostData() {
	loading.value = true
	try {
		const res = await getPostDetail(postId.value)
		if (res.success && res.result.post) {
			const post = res.result.post
			form.title = post.title
			form.content = post.content
			form.type = post.type
			try { form.images = post.images ? JSON.parse(post.images) : [] } catch (e) { form.images = [] }
		} else {
			showToast('获取动态详情失败')
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
		const data = { ...form, images: JSON.stringify(form.images) }
		const res = await updatePost(postId.value, data)
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

function chooseImage() {
	uni.chooseImage({
		count: 20 - form.images.length,
		sizeType: ['compressed'],
		sourceType: ['album', 'camera'],
		success: (res) => { form.images = [...form.images, ...res.tempFilePaths] }
	})
}

function removeImage(index) { form.images.splice(index, 1) }
function previewImage(url, urls) { previewImages(url, urls) }
function isSvgImageFn(url) { return isSvgImage(url) }
function getImageGridClassFn(count) { return getImageGridClass(count) }

onMounted(() => {
	initStatusBar()
	const pages = getCurrentPages()
	const currentPage = pages[pages.length - 1]
	const options = currentPage?.options || {}
	postId.value = options.id
	if (!postId.value) {
		showToast('动态ID不能为空')
		navigateBack()
		return
	}
	loadPostData()
})
</script>

<style lang="scss" scoped>
.ep-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
}

.ep-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 200rpx;
}

.ep-hero-bg {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg, #1e3a8a 0%, #1e40af 25%, #2563eb 55%, #3b82f6 75%, #93c5fd 100%);
	z-index: 0;
}

.ep-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg, rgba(30, 58, 138, 0.65) 0%, rgba(37, 99, 235, 0.4) 50%, rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.ep-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.ep-back-btn {
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

.ep-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
}

.ep-submit-btn {
	padding: 10rpx 24rpx;
	background: rgba(255, 255, 255, 0.25);
	border-radius: 100rpx;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.3);

	.ep-submit-btn-text { font-size: 26rpx; font-weight: 700; color: #fff; }

	&--disabled { opacity: 0.5; }
	&:active:not(.ep-submit-btn--disabled) { background: rgba(255, 255, 255, 0.35); transform: scale(0.95); }
}

.ep-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 8rpx;
}

.ep-hero-sub {
	font-size: 20rpx;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.6);
	letter-spacing: 2px;
}

.ep-loading {
	position: fixed;
	top: 0; left: 0; right: 0; bottom: 0;
	background: rgba(255, 255, 255, 0.95);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 16rpx;
	z-index: 1000;

	.ep-loading-text { font-size: 28rpx; color: var(--text-secondary); }
}

.ep-loading-spinner {
	width: 64rpx;
	height: 64rpx;
	border: 5rpx solid rgba(99, 102, 241, 0.15);
	border-top-color: var(--primary-500);
	border-radius: 50%;
	animation: ep-spin 0.8s linear infinite;
}

@keyframes ep-spin { to { transform: rotate(360deg); } }

.ep-scroll { flex: 1; min-height: 0; }

.ep-form {
	padding: 24rpx;
	padding-bottom: 60rpx;
}

.ep-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 24rpx;
	margin-bottom: 20rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

.ep-title-input {
	width: 100%;
	height: 88rpx;
	font-size: 34rpx;
	font-weight: 700;
	color: var(--text-primary);
	padding: 0;
	border: none;
	outline: none;
	background: transparent;
}

.ep-placeholder { color: var(--text-tertiary); }

.ep-content-textarea {
	width: 100%;
	min-height: 240rpx;
	font-size: 30rpx;
	color: var(--text-primary);
	line-height: 1.8;
	padding: 0;
	border: none;
	outline: none;
	background: transparent;
	resize: none;
}

.ep-counter {
	display: flex;
	justify-content: flex-end;
	margin-top: 12rpx;

	.ep-counter-text { font-size: 22rpx; color: var(--text-tertiary); }
}

.ep-section-header {
	display: flex;
	align-items: center;
	gap: 10rpx;
	margin-bottom: 16rpx;
}

.ep-section-title { font-size: 28rpx; font-weight: 700; color: var(--text-primary); }
.ep-section-optional { font-size: 22rpx; color: var(--text-tertiary); }

.ep-images-grid {
	display: grid;
	gap: 12rpx;
}

.ep-image-item {
	position: relative;
	border-radius: 16rpx;
	overflow: hidden;
}

.ep-uploaded-image { width: 100%; height: 100%; object-fit: cover; }

.ep-svg-overlay {
	position: absolute;
	top: 0; left: 0;
	background: rgba(0, 0, 0, 0.4);
	padding: 4rpx 12rpx;
	border-radius: 0 0 12rpx 0;

	.ep-image-badge-text { font-size: 18rpx; color: #fff; font-weight: 700; }
}

.ep-image-actions {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: rgba(0, 0, 0, 0.3);
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 16rpx;
	opacity: 0;
	transition: opacity 0.2s ease;
}

.ep-image-item:active .ep-image-actions { opacity: 1; }

.ep-image-action {
	width: 56rpx;
	height: 56rpx;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;

	&--preview { background: rgba(59, 130, 246, 0.7); }
	&--remove { background: rgba(239, 68, 68, 0.7); }
}

.ep-add-btn {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 8rpx;
	border-radius: 16rpx;
	border: 2px dashed rgba(148, 163, 184, 0.4);
	min-height: 180rpx;
	transition: all 0.2s ease;

	&:active { border-color: var(--primary-400); background: rgba(59, 102, 241, 0.05); }
}

.ep-add-text {
	font-size: 24rpx;
	color: var(--text-secondary);
}
</style>
