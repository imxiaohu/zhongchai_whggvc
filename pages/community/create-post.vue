<template>
	<view class="cp-page">
		<!-- 顶部蓝色渐变 Hero -->
		<view class="cp-hero">
			<view class="cp-hero-bg"></view>
			<view class="cp-hero-overlay"></view>

			<view :style="{ height: statusBarHeight + 'px' }"></view>

			<view class="cp-hero-nav">
				<view class="cp-back-btn" @tap="goBack">
					<l-icon name="chevron-left" style="font-size: 24px; color: #fff;"></l-icon>
				</view>
				<text class="cp-hero-title">{{ isEditMode ? '编辑动态' : '发布动态' }}</text>
				<view class="cp-submit-btn" :class="{ 'cp-submit-btn--disabled': !canSubmit }" @tap="handleSubmit">
					<text class="cp-submit-btn-text">{{ submitting ? (isEditMode ? '保存中...' : '发布中...') : (isEditMode ? '保存' : '发布') }}</text>
				</view>
			</view>

			<view class="cp-hero-content">
				<text class="cp-hero-sub">CREATE POST</text>
			</view>
		</view>

		<!-- 表单区域 -->
		<scroll-view class="cp-scroll" scroll-y>
			<view class="cp-form">
				<!-- 标题 -->
				<view class="cp-card">
					<input
						class="cp-title-input"
						v-model="form.title"
						placeholder="请输入标题"
						maxlength="100"
						placeholder-class="cp-placeholder"
					/>
					<view class="cp-counter">
						<text class="cp-counter-text">{{ form.title.length }}/100</text>
					</view>
				</view>

				<!-- 正文 -->
				<view class="cp-card">
					<textarea
						class="cp-content-textarea"
						v-model="form.content"
						placeholder="分享你的校园生活..."
						maxlength="5000"
						auto-height
						placeholder-class="cp-placeholder"
					></textarea>
					<view class="cp-counter">
						<text class="cp-counter-text">{{ form.content.length }}/5000</text>
					</view>
				</view>

				<!-- 图片上传 -->
				<view class="cp-card">
					<view class="cp-section-header">
						<text class="cp-section-title">图片展示</text>
						<text class="cp-section-optional">(可选)</text>
					</view>
					<view class="cp-images-grid" :class="imagesUploadClass">
						<view v-for="(image, index) in displayImages" :key="index" class="cp-image-item">
							<image class="cp-image" :src="image" mode="aspectFill" @tap="previewImageHandle(image, form.images)"></image>
							<view class="cp-image-actions">
								<view class="cp-image-action cp-image-action--preview" @tap="previewImageHandle(image, form.images)">
									<l-icon name="browse" style="font-size: 12px; color: #fff;"></l-icon>
								</view>
								<view class="cp-image-action cp-image-action--remove" @tap="removeImage(index)">
									<l-icon name="close" style="font-size: 12px; color: #fff;"></l-icon>
								</view>
							</view>
						</view>
						<view v-if="hasMoreImages && !showAllImages" class="cp-more-btn" @tap="toggleShowAllImages">
							<l-icon name="more" style="font-size: 24px; color: var(--text-secondary);"></l-icon>
							<text class="cp-more-text">+{{ hiddenImagesCount }}</text>
							<text class="cp-more-hint">点击查看更多</text>
						</view>
						<view v-if="hasMoreImages && showAllImages" class="cp-collapse-btn" @tap="toggleShowAllImages">
							<l-icon name="chevron-up" style="font-size: 24px; color: var(--text-secondary);"></l-icon>
							<text class="cp-collapse-text">收起</text>
						</view>
						<view v-if="form.images.length < 20" class="cp-add-btn" @tap="selectImages">
							<l-icon name="photo" style="font-size: 24px; color: var(--text-tertiary);"></l-icon>
							<text class="cp-add-text">{{ form.images.length }}/20</text>
							<text class="cp-add-hint">点击添加图片</text>
						</view>
					</view>
				</view>

				<!-- 动态类型 -->
				<view class="cp-card">
					<view class="cp-section-header">
						<text class="cp-section-title">动态类型</text>
					</view>
					<view class="cp-type-selector">
						<view
							v-for="type in postTypes"
							:key="type.key"
							class="cp-type-item"
							:class="{ 'cp-type-item--active': form.type === type.key }"
							@tap="selectType(type.key)"
						>
							<text class="cp-type-text">{{ type.name }}</text>
						</view>
					</view>
				</view>

				<!-- 发布到社团 -->
				<view class="cp-card" v-if="!clubId">
					<view class="cp-section-header">
						<text class="cp-section-title">发布到社团</text>
						<text class="cp-section-optional">(可选)</text>
					</view>
					<view class="cp-club-selector" :class="{ 'cp-club-selector--official': isOfficialSelected }" @tap="showClubPicker">
						<view class="cp-club-selector-content">
							<view v-if="isOfficialSelected" class="cp-official-row">
								<l-icon name="star-filled" style="font-size: 16px; color: #fbbf24;"></l-icon>
								<text class="cp-official-text">发布到官方公告</text>
								<view class="cp-official-badge">官方</view>
							</view>
							<text v-else class="cp-club-placeholder">{{ selectedClub ? selectedClub.name : '选择发布社团' }}</text>
						</view>
						<l-icon name="chevron-right" style="font-size: 16px; color: var(--text-tertiary);"></l-icon>
					</view>
				</view>
			</view>
		</scroll-view>

		<!-- 社团选择弹窗 -->
		<t-popup v-model:visible="clubPickerVisible" placement="bottom" :overlay="true" :close-on-overlay-click="true">
			<view class="cp-picker">
				<view class="cp-picker-header">
					<text class="cp-picker-title">选择发布社团</text>
					<view class="cp-picker-close" @tap="hideClubPicker">
						<l-icon name="close" style="font-size: 20px; color: var(--text-tertiary);"></l-icon>
					</view>
				</view>
				<scroll-view class="cp-picker-scroll" scroll-y>
					<view class="cp-picker-option cp-picker-option--official" :class="{ 'cp-picker-option--selected': isOfficialSelected }" @tap="selectOfficial">
						<view class="cp-picker-official-icon">
							<l-icon name="star-filled" style="font-size: 20px; color: #fbbf24;"></l-icon>
						</view>
						<text class="cp-picker-official-text">发布到官方公告</text>
						<view class="cp-picker-official-badge">官方</view>
						<l-icon v-if="isOfficialSelected" name="check-circle-filled" style="font-size: 20px; color: var(--primary-500);"></l-icon>
					</view>
					<view
						v-for="club in myClubs"
						:key="club.id"
						class="cp-picker-option"
						:class="{ 'cp-picker-option--selected': selectedClub && selectedClub.id === club.id }"
						@tap="selectClub(club)"
					>
						<ClubAvatar
						class="cp-picker-club-logo"
						:src="club.logoUrl"
						:name="club.name"
						:size="64"
					></ClubAvatar>
						<text class="cp-picker-club-name">{{ club.name }}</text>
						<l-icon v-if="selectedClub && selectedClub.id === club.id" name="check-circle-filled" style="font-size: 20px; color: var(--primary-500);"></l-icon>
					</view>
				</scroll-view>
			</view>
		</t-popup>
		<!-- 社区服务须知弹窗 -->
		<CommunityTermsModal
			v-model:visible="termsVisible"
			@agree="onTermsAgreed"
			@disagree="onTermsDisagree"
		/>
	</view>
</template>

<script setup>
import ClubAvatar from '@/components/ClubAvatar.vue'
import CommunityTermsModal from '@/components/CommunityTermsModal.vue'
import { ref, reactive, computed, onMounted } from 'vue';
import { createPost, getPostDetail, updatePost, getClubsList, normalizePostDetailResponse, processImages } from '../api/community.js';
import { showToast, navigateBack } from '../api/page.js';
import { previewImage, isSvgImage, getImageGridClass, chooseImages } from '../../utils/imageUpload.js';
import { validatePostForm, buildPostPayload } from '../../utils/postForm.js';
import { hasAgreedToTerms } from '@/composables/useCommunityTerms.js';

const clubId = ref(null);
const editingPostId = ref(null);
const clubPickerVisible = ref(false);
const statusBarHeight = ref(20);
const form = reactive({
	title: '',
	content: '',
	images: [],
	type: 'article'
});
const submitting = ref(false);
const loading = ref(false);
const selectedClub = ref(null);
const isOfficialSelected = ref(false);
const myClubs = ref([]);
const showAllImages = ref(false);
const maxDisplayImages = ref(9);
const termsVisible = ref(false);

function initStatusBar() {
	try {
		const systemInfo = uni.getSystemInfoSync();
		statusBarHeight.value = systemInfo.statusBarHeight || 20;
	} catch (e) {
		statusBarHeight.value = 20;
	}
}

function goBack() {
	uni.navigateBack();
}

const canSubmit = computed(() => {
	return form.title.trim() && form.content.trim() && !submitting.value && !loading.value;
});

const isEditMode = computed(() => !!editingPostId.value);

const displayImages = computed(() => {
	if (showAllImages.value || form.images.length <= maxDisplayImages.value) return form.images;
	return form.images.slice(0, maxDisplayImages.value - 1);
});

const hasMoreImages = computed(() => form.images.length > maxDisplayImages.value);
const hiddenImagesCount = computed(() => Math.max(0, form.images.length - (maxDisplayImages.value - 1)));
const imagesUploadClass = computed(() => {
	const hasMore = hasMoreImages.value && !showAllImages.value;
	const canAdd = form.images.length < 20;
	return getImageGridClass(displayImages.value.length + (hasMore ? 1 : 0) + (canAdd ? 1 : 0));
});

const postTypes = [
	{ key: 'article', name: 'Article' },
	{ key: 'announcement', name: 'Announcement' },
	{ key: 'activity', name: 'Activity' }
];

async function selectImages() {
	const images = await chooseImages(20 - form.images.length);
	if (images && images.length > 0) form.images = [...form.images, ...images];
}

function removeImage(index) {
	form.images.splice(index, 1);
}

function previewImageHandle(image, images) {
	uni.previewImage({ current: image, urls: images || form.images });
}

function toggleShowAllImages() {
	showAllImages.value = !showAllImages.value;
}

function selectType(key) {
	form.type = key;
}

function showClubPicker() { clubPickerVisible.value = true; }
function hideClubPicker() { clubPickerVisible.value = false; }
function selectOfficial() {
	isOfficialSelected.value = true;
	selectedClub.value = null;
	hideClubPicker();
}
function selectClub(club) {
	selectedClub.value = club;
	isOfficialSelected.value = false;
	hideClubPicker();
}

async function handleSubmit() {
	if (!canSubmit.value) return;
	const validation = validatePostForm(form);
	if (!validation.valid) {
		showToast({ title: validation.message, icon: 'none' });
		return;
	}
	submitting.value = true;
	try {
		const payload = buildPostPayload(form, selectedClub.value, isOfficialSelected.value);
		let result;
		if (isEditMode.value) {
			result = await updatePost(editingPostId.value, payload);
		} else {
			result = await createPost(payload);
		}
		if (result.success) {
			showToast({ title: isEditMode.value ? '修改成功' : '发布成功', icon: 'success' });
			const pages = getCurrentPages();
			const prevPage = pages[pages.length - 2];
			if (prevPage) {
				if (prevPage.$vm?.refreshPost) {
					prevPage.$vm.refreshPost(editingPostId.value, { ...form, club: selectedClub.value })
				} else if (prevPage.$vm?.refresh) {
					prevPage.$vm.refresh()
				}
			}
			setTimeout(() => navigateBack(), 1500);
		} else {
			showToast({ title: result.message || (isEditMode.value ? '修改失败' : '发布失败'), icon: 'none' });
		}
	} catch (e) {
		showToast({ title: isEditMode.value ? '修改失败' : '发布失败', icon: 'none' });
	} finally {
		submitting.value = false;
	}
}

async function initMounted() {
	initStatusBar();
	const agreed = await hasAgreedToTerms();
	if (!agreed) {
		termsVisible.value = true;
		return;
	}
	const pages = getCurrentPages();
	const currentPage = pages[pages.length - 1];
	const options = currentPage?.options || {};
	clubId.value = options.clubId;

	const editPostId = options.postId;
	if (editPostId) {
		editingPostId.value = editPostId;
		loading.value = true;
		try {
			const res = await getPostDetail(editPostId);
			if (res && res.success) {
				const detail = normalizePostDetailResponse(res);
				if (detail) {
					form.title = detail.title || '';
					form.content = detail.content || '';
					form.images = processImages(detail.images) || [];
					form.type = detail.type || 'article';
					if (detail.club) {
						selectedClub.value = detail.club;
					}
				}
			}
		} catch (e) {
			showToast({ title: '加载失败', icon: 'none' });
		} finally {
			loading.value = false;
		}
	}

	getClubsList().then(res => {
		if (res.success && res.data) myClubs.value = res.data;
	});
}

onMounted(initMounted);

function onTermsAgreed() {
	termsVisible.value = false;
	getClubsList().then(res => {
		if (res.success && res.data) myClubs.value = res.data;
	});
}

function onTermsDisagree() {
	termsVisible.value = false;
	navigateBack();
}
</script>

<style lang="scss" scoped>
/* ============================================
   Create Post - Hero Style
   ============================================ */

.cp-page {
	min-height: 100vh;
	background-color: var(--bg-secondary);
	display: flex;
	flex-direction: column;
}

/* ---- Hero Section ---- */
.cp-hero {
	position: relative;
	width: 100%;
	overflow: hidden;
	min-height: 200rpx;
}

.cp-hero-bg {
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

.cp-hero-overlay {
	position: absolute;
	top: 0; left: 0; right: 0; bottom: 0;
	background: linear-gradient(180deg,
		rgba(30, 58, 138, 0.65) 0%,
		rgba(37, 99, 235, 0.4) 50%,
		rgba(147, 197, 253, 0.1) 100%);
	z-index: 1;
}

.cp-hero-nav {
	position: relative;
	z-index: 2;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	height: 88rpx;
}

.cp-back-btn {
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

.cp-hero-title {
	font-size: 36rpx;
	font-weight: 800;
	color: #fff;
	letter-spacing: 0.5px;
}

.cp-submit-btn {
	padding: 10rpx 28rpx;
	background: rgba(255, 255, 255, 0.25);
	border-radius: 100rpx;
	backdrop-filter: blur(8px);
	border: 1px solid rgba(255, 255, 255, 0.3);
	transition: all 0.2s ease;

	.cp-submit-btn-text {
		font-size: 26rpx;
		font-weight: 700;
		color: #fff;
	}

	&--disabled {
		opacity: 0.5;
	}

	&:active:not(.cp-submit-btn--disabled) {
		background: rgba(255, 255, 255, 0.35);
		transform: scale(0.95);
	}
}

.cp-hero-content {
	position: relative;
	z-index: 2;
	padding: 0 32rpx 8rpx;
}

.cp-hero-sub {
	font-size: 20rpx;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.6);
	letter-spacing: 2px;
}

/* ---- Scroll ---- */
.cp-scroll {
	flex: 1;
	min-height: 0;
}

.cp-form {
	padding: 24rpx;
	padding-bottom: 60rpx;
}

/* ---- Card ---- */
.cp-card {
	background: #fff;
	border-radius: 24rpx;
	padding: 24rpx;
	margin-bottom: 20rpx;
	box-shadow: 0 4rpx 16rpx rgba(30, 64, 175, 0.06);
	border: 1px solid rgba(148, 163, 184, 0.1);
}

/* ---- Title Input ---- */
.cp-title-input {
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

.cp-placeholder {
	color: var(--text-tertiary);
}

/* ---- Content Textarea ---- */
.cp-content-textarea {
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

/* ---- Counter ---- */
.cp-counter {
	display: flex;
	justify-content: flex-end;
	margin-top: 12rpx;
}

.cp-counter-text {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

/* ---- Section Header ---- */
.cp-section-header {
	display: flex;
	align-items: center;
	gap: 10rpx;
	margin-bottom: 16rpx;
}

.cp-section-title {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.cp-section-optional {
	font-size: 22rpx;
	color: var(--text-tertiary);
}

/* ---- Images Grid ---- */
.cp-images-grid {
	display: grid;
	gap: 12rpx;
}

.cp-image-item {
	position: relative;
	border-radius: 16rpx;
	overflow: hidden;
}

.cp-image {
	width: 100%;
	height: 100%;
	object-fit: cover;
}

.cp-image-actions {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.3);
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 16rpx;
	opacity: 0;
	transition: opacity 0.2s ease;
}

.cp-image-item:active .cp-image-actions {
	opacity: 1;
}

.cp-image-action {
	width: 56rpx;
	height: 56rpx;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;

	&--preview { background: rgba(59, 130, 246, 0.7); }
	&--remove { background: rgba(239, 68, 68, 0.7); }
}

.cp-more-btn,
.cp-collapse-btn,
.cp-add-btn {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 8rpx;
	border-radius: 16rpx;
	border: 2px dashed rgba(148, 163, 184, 0.4);
	min-height: 180rpx;
	transition: all 0.2s ease;
}

.cp-more-btn:active,
.cp-add-btn:active {
	border-color: var(--primary-400);
	background: rgba(59, 102, 241, 0.05);
}

.cp-more-text {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.cp-more-hint,
.cp-add-hint {
	font-size: 20rpx;
	color: var(--text-tertiary);
}

.cp-collapse-btn {
	border-style: solid;
	border-color: rgba(148, 163, 184, 0.3);
}

.cp-collapse-text,
.cp-add-text {
	font-size: 24rpx;
	color: var(--text-secondary);
}

/* ---- Type Selector ---- */
.cp-type-selector {
	display: flex;
	gap: 16rpx;
}

.cp-type-item {
	flex: 1;
	height: 80rpx;
	border-radius: 20rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	border: 2px solid rgba(148, 163, 184, 0.2);
	background: var(--bg-muted);
	transition: all 0.2s ease;

	&--active {
		border-color: var(--primary-500);
		background: rgba(59, 102, 241, 0.08);
	}

	&:active:not(.cp-type-item--active) {
		background: var(--bg-secondary);
	}
}

.cp-type-text {
	font-size: 26rpx;
	font-weight: 700;
	color: var(--text-secondary);

	.cp-type-item--active & {
		color: var(--primary-600);
	}
}

/* ---- Club Selector ---- */
.cp-club-selector {
	display: flex;
	align-items: center;
	justify-content: space-between;
	height: 88rpx;
	padding: 0 20rpx;
	border-radius: 16rpx;
	border: 1px solid rgba(148, 163, 184, 0.2);
	background: var(--bg-muted);

	&--official {
		border-color: rgba(251, 191, 36, 0.3);
		background: rgba(251, 191, 36, 0.05);
	}

	&:active { background: var(--bg-secondary); }
}

.cp-club-selector-content {
	flex: 1;
}

.cp-club-placeholder {
	font-size: 28rpx;
	color: var(--text-tertiary);
}

.cp-official-row {
	display: flex;
	align-items: center;
	gap: 10rpx;
}

.cp-official-text {
	font-size: 28rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.cp-official-badge {
	padding: 4rpx 12rpx;
	border-radius: 8rpx;
	background: linear-gradient(135deg, #f59e0b, #d97706);
	font-size: 18rpx;
	color: #fff;
	font-weight: 700;
}

/* ---- Picker ---- */
.cp-picker {
	background: #fff;
	border-radius: 28rpx 28rpx 0 0;
	max-height: 70vh;
	display: flex;
	flex-direction: column;
}

.cp-picker-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 32rpx;
	border-bottom: 1px solid rgba(226, 232, 240, 0.6);
}

.cp-picker-title {
	font-size: 34rpx;
	font-weight: 800;
	color: var(--text-primary);
}

.cp-picker-close {
	width: 64rpx;
	height: 64rpx;
	border-radius: 50%;
	background: var(--bg-muted);
	display: flex;
	align-items: center;
	justify-content: center;
}

.cp-picker-scroll {
	flex: 1;
	padding: 16rpx 24rpx 40rpx;
}

.cp-picker-option {
	display: flex;
	align-items: center;
	padding: 24rpx;
	border-radius: 20rpx;
	margin-bottom: 12rpx;
	background: var(--bg-muted);
	transition: background 0.15s ease;

	&--official { border: 1px solid rgba(251, 191, 36, 0.2); }
	&--selected { background: rgba(59, 102, 241, 0.08); border: 1px solid rgba(59, 102, 241, 0.3); }
	&:active { background: var(--bg-secondary); }
}

.cp-picker-official-icon {
	width: 80rpx;
	height: 80rpx;
	border-radius: 20rpx;
	background: rgba(251, 191, 36, 0.15);
	display: flex;
	align-items: center;
	justify-content: center;
	margin-right: 20rpx;
	flex-shrink: 0;
}

.cp-picker-official-text {
	flex: 1;
	font-size: 30rpx;
	font-weight: 700;
	color: var(--text-primary);
}

.cp-picker-official-badge {
	padding: 4rpx 12rpx;
	border-radius: 8rpx;
	background: linear-gradient(135deg, #f59e0b, #d97706);
	font-size: 18rpx;
	color: #fff;
	font-weight: 700;
	margin-right: 12rpx;
}

.cp-picker-club-logo {
	width: 80rpx;
	height: 80rpx;
	border-radius: 20rpx;
	object-fit: cover;
	margin-right: 20rpx;
	flex-shrink: 0;
}

.cp-picker-club-name {
	flex: 1;
	font-size: 30rpx;
	font-weight: 600;
	color: var(--text-primary);
}
</style>
